package coap

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
	// "runtime/debug"
)

type writeReq interface {
	sendResp(err error, timeout time.Duration)
	waitResp(timeout time.Duration) error
	data() Message
}

type writeReqBase struct {
	req      Message
	respChan chan error // channel must have size 1 for non-blocking write to channel
}

func (wreq *writeReqBase) sendResp(err error, timeout time.Duration) {
	select {
	case wreq.respChan <- err:
		return
	default:
		log.Fatal("Exactly one error can be send as resp. This is err.")
	}
}

func (wreq *writeReqBase) waitResp(timeout time.Duration) error {
	select {
	case err := <-wreq.respChan:
		return err
	case <-time.After(timeout):
		return ErrTimeout
	}
}

func (wreq *writeReqBase) data() Message {
	return wreq.req
}

type writeReqTCP struct {
	writeReqBase
}

type writeReqUDP struct {
	writeReqBase
	sessionData *SessionUDPData
	ns          *NoiseState
}

// Conn represents the connection
type Conn interface {
	// LocalAddr get local address of the connection
	LocalAddr() net.Addr
	// RemoteAddr get peer address of the connection
	RemoteAddr() net.Addr
	// Close close the connection
	Close() error

	write(w writeReq, timeout time.Duration) error
}

type connWriter interface {
	writeHandler(srv *Server) bool
	writeEndHandler(timeout time.Duration) bool
	sendFinish(timeout time.Duration)

	writeHandlerWithFunc(srv *Server, writeFunc func(srv *Server, wreq writeReq) error) bool
}

type connBase struct {
	writeChan chan writeReq
	closeChan chan bool
	finChan   chan bool
	closed    int32
	// ns        *NoiseState
}

func (conn *connBase) finishWrite() {
	if !atomic.CompareAndSwapInt32(&conn.closed, conn.closed, 1) {
		return
	}
	conn.closeChan <- true
	<-conn.finChan
}

func (conn *connBase) writeHandlerWithFunc(srv *Server, writeFunc func(srv *Server, wreq writeReq) error) bool {
	select {
	case wreq := <-conn.writeChan:
		wreq.sendResp(writeFunc(srv, wreq), srv.syncTimeout())
		return true
	case <-conn.closeChan:
		return false
	}
}

func (conn *connBase) sendFinish(timeout time.Duration) {
	select {
	case conn.finChan <- true:
	case <-time.After(timeout):
		log.Println("Client cannot recv start: Timeout")
	}
}

func (conn *connBase) writeEndHandler(timeout time.Duration) bool {
	select {
	case wreq := <-conn.writeChan:
		wreq.sendResp(ErrConnectionClosed, timeout)
		return true
	default:
		return false
	}
}

func (conn *connBase) write(w writeReq, timeout time.Duration) error {
	if atomic.LoadInt32(&conn.closed) > 0 {
		return ErrConnectionClosed
	}
	select {
	case conn.writeChan <- w:
		return w.waitResp(timeout)
	case <-time.After(timeout):
		return ErrTimeout
	}
}

type connTCP struct {
	connBase
	connection net.Conn // i/o connection if TCP was used
	num        int32
}

func (conn *connTCP) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

func (conn *connTCP) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}

func (conn *connTCP) Close() error {
	conn.finishWrite()
	return conn.connection.Close()
}

func (conn *connTCP) writeHandler(srv *Server) bool {
	return conn.writeHandlerWithFunc(srv, func(srv *Server, wreq writeReq) error {
		data := wreq.data()
		wr := srv.acquireWriter(conn.connection)
		defer srv.releaseWriter(wr)
		writeTimeout := srv.writeTimeout()
		conn.connection.SetWriteDeadline(time.Now().Add(writeTimeout))
		err := data.MarshalBinary(wr)
		if err != nil {
			return err
		}
		wr.Flush()
		return nil
	})
}

type connUDP struct {
	connBase
	connection   *net.UDPConn // i/o connection if UDP was used
	retriesQueue *RetriesQueue
	rRand        *rand.Rand
}

type retryHeaders struct {
	seqnum uint8
	nps    NoisePipeState
	msgID  uint16
}

func (conn *connUDP) LocalAddr() net.Addr {
	return conn.connection.LocalAddr()
}

func (conn *connUDP) RemoteAddr() net.Addr {
	return conn.connection.RemoteAddr()
}

func (conn *connUDP) SetReadDeadline(timeout time.Time) error {
	return conn.connection.SetReadDeadline(timeout)
}

func (conn *connUDP) resetConnection() error {
	conn.connection.Close()
	dialer := net.Dialer{}
	if conn.RemoteAddr() == nil {
		debugf("Reset listener on %s", conn.LocalAddr())
		c, err := net.ListenUDP("udp", conn.LocalAddr().(*net.UDPAddr))
		if err != nil {
			return err
		}
		conn.connection = c
	} else {
		debugf("Reset connection from %s to %s", conn.LocalAddr().String(), conn.RemoteAddr().String())
		c, err := dialer.Dial("udp", conn.RemoteAddr().String())
		if err != nil {
			return err
		}
		conn.connection = c.(*net.UDPConn)
		debugf("+-> New local address: %s", c.LocalAddr().String())
	}
	return nil
}

func (conn *connUDP) ReadFromSessionUDP(m []byte) (int, *SessionUDPData, error) {
	n, sessionData, err := ReadFromSessionUDP(conn.connection, m)
	if _, ok := err.(net.Error); err == nil || !ok {
		return n, sessionData, err
	}

	debugf("Resetting connection because of error: %s", err.Error())
	if err = conn.resetConnection(); err != nil {
		return 0, nil, err
	}
	return ReadFromSessionUDP(conn.connection, m)
}

func (conn *connUDP) extractRetryHeaders(m []byte) (h retryHeaders, pl []byte) {
	// Check whether the message is one we hand-wrapped. If it's not, end the
	// process here.
	if m[0]>>6 != 3 {
		return retryHeaders{}, m
	}

	// Extract the headers we placed. These will take the following form:
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Ver| T |  TKL  |      Code     |          Message ID           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |   Token (if any, TKL bytes) ...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Sequence number|
	// +-+-+-+-+-+-+-+-+
	// We only really care about the type, code, message ID and sequence number.

	var t, tkl, code, seqnum uint8
	var messageID uint16

	// Type
	t = m[0] >> 4 & 0x3
	// Token length, needed to know the end of the token header so we can grab
	// the sequence number that's located right after it.
	tkl = m[0] & 0xf
	tkEnd := 4 + tkl
	// Code
	code = m[1]
	// Message ID
	messageID = uint16(m[2])<<8 | uint16(m[3])
	// Sequence number
	seqnum = m[tkEnd]

	// Copy the encrypted payload's content to the i/o slice.
	pl = m[tkEnd+1:]

	debugf("Received message using CoAP version %d, of type %d and code %d, and with message ID %d, token of length %d, and sequence number %d\n", m[0]>>6, t, code, messageID, tkl, uint8(seqnum))

	conn.retriesQueue.CancelRetrySchedule(messageID)

	// If state is READY then the payload should be decrypted using the sequence
	// number. Otherwise, we just let the noise pipeline do its job.
	state := getNoisePipelineState(t, code)

	h = retryHeaders{
		seqnum: seqnum,
		msgID:  messageID,
		nps:    state,
	}

	return
}

func getNoisePipelineState(t, code uint8) NoisePipeState {
	// Handshake | CoAP Type | CoAP Code |
	// ----------|-----------|-----------|
	// XX1       | 0 (CON)   | 250       |
	// XX2       | 2 (ACK)   | 250       |
	// XX3       | 1 (NON)   | 251       |
	// IK1       | 0 (CON)   | 252       |
	// IK2       | 2 (ACK)   | 252       |
	switch code {
	case 250:
		// XX1 or XX2
		if COAPType(t) == Confirmable {
			return XX1
		}

		if COAPType(t) == Acknowledgement {
			return XX2
		}
	case 251:
		return XX3
	case 252:
		// IK1 or IK2
		if COAPType(t) == Confirmable {
			return IK1
		}

		if COAPType(t) == Acknowledgement {
			return IK2
		}
	}

	return READY
}

func (conn *connUDP) Close() error {
	conn.finishWrite()
	return conn.connection.Close()
}

func (conn *connUDP) writeHandler(srv *Server) bool {
	return conn.writeHandlerWithFunc(srv, func(srv *Server, wreq writeReq) error {
		wreqUDP := wreq.(*writeReqUDP)

		return conn.sendMessage(wreq.data(), wreqUDP.ns, wreqUDP.sessionData, srv)
	})
}

func (conn *connUDP) sendMessage(data Message, ns *NoiseState, sessionData *SessionUDPData, srv *Server) error {
	writeTimeout := srv.writeTimeout()
	buf := &bytes.Buffer{}
	err := data.MarshalBinary(buf)
	if err != nil {
		return err
	}

	// TODO:
	//
	// before compressing, we have to move the coap headers to the cleartext payload
	// we move:
	//
	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Ver| T |  TKL  |      Code     |          Message ID           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |   Token (if any, TKL bytes) ...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

	// We leave the options and subsequent CBOR payloads however encrypted.

	// We calculate an 8-bit sequence number at this point from the noise
	// handshake state's nonce, and append it to the headers - we need
	// this to do reordering on receipt to hand the
	// packets to noise in the write order and without dups or gaps.

	// We package up our XX and IK noise handshakes with the same headers
	// as if they were CoAP.  No token is needed.  We'll need to pick a
	// custom CoAP code.  Suggestion:

	// Handshake | CoAP Type | CoAP Code |
	// ----------|-----------|-----------|
	// XX1       | 0 (CON)   | 250       |
	// XX2       | 2 (ACK)   | 250       |
	// XX3       | 1 (NON)   | 251       |
	// IK1       | 0 (CON)   | 252       |
	// IK2       | 2 (ACK)   | 252       |

	// We could be naughty and set Ver=011b rather than 001b to indicate
	// that encryption is turned on, in order to negotiate it more elegantly

	//  On sending the response to a

	var compressed []byte
	if srv.Compressor != nil {
		compressed, err = srv.Compressor.CompressPayload(buf.Bytes())
		if err != nil {
			return err
		}
		//log.Printf("Compressed packet: %d -> %d bytes", len(buf.Bytes()), len(compressed))
	} else {
		compressed = buf.Bytes()
	}

	conn.connection.SetWriteDeadline(time.Now().Add(writeTimeout))

	var msg []byte
	if srv.Encryption {
		buf = new(bytes.Buffer)
		ns := ns

		nps := ns.PipeState

		// If we're doing a handshake we'll want to queue up the message
		// until the handshake is over.
		if nps != READY {
			debugf("Queuing %X", compressed)
			conn.retriesQueue.PushHS(HSQueueMsg{
				b:           compressed,
				m:           data,
				conn:        conn,
				sessionData: sessionData,
			})
		} else {
			var msg *HSQueueMsg
			for {
				if msg = srv.RetriesQueue.PopHS(sessionData.raddr.IP.String()); msg == nil {
					break
				}
				// Run the sending in a goroutine since it's not
				// supposed to have side-effects and we don't want it to
				// block the request we're currently processing.
				go msg.conn.sendMessage(msg.m, ns, msg.sessionData, srv)
			}
		}

		// log.Printf("encrypting %d bytes with %+v as %v", len(compressed), ns, compressed)
		msg, err = ns.EncryptMessage(compressed, conn, sessionData)
		if err != nil {
			log.Printf("failed to encrypt message: %v", err)
			return err
		}

		var mID uint16
		if mID, err = conn.SetCoapHeaders(buf, data, nps, 0); err != nil {
			return err
		}

		if _, err = buf.Write(msg); err != nil {
			return err
		}

		if data.Type() == Confirmable {
			// TODO: Figure out a sensible value for timeToRetry.
			go conn.retriesQueue.ScheduleRetry(mID, buf.Bytes(), sessionData, conn)

			// Increment the sequence number for the next message.
			conn.retriesQueue.seqnum++
		}

		return conn.writeToSession(buf.Bytes(), sessionData)
	}

	msg = compressed
	_, err = WriteToSessionUDP(conn.connection, msg, sessionData)
	return err

	// TODO:
	// Rather than having noise send directly or handle retries itself, noise needs to pass
	// back the payload and we then retry (re)sending it here, as a bunch of bits.
	//
	// We need to track the msgid+token pair of the confirmable messages being sent, so we know when to
	// keep retrying.  (As when we receive the ID of the response, we should stop retrying.)
	//
	// We may need a mechanism to unwedge wedged noisepipes	(e.g. actively rehandshake if the retry
	// schedule expires or if we have a gap of > 128 in the queue)
}

func (conn *connUDP) writeToSession(b []byte, sessionData *SessionUDPData) error {
	_, err := WriteToSessionUDP(conn.connection, b, sessionData)
	if _, ok := err.(net.Error); err == nil || !ok {
		return err
	}

	debugf("Resetting connection because of error: %s", err.Error())
	if err = conn.resetConnection(); err != nil {
		return err
	}

	_, err = WriteToSessionUDP(conn.connection, b, sessionData)

	return err
}

func (conn *connUDP) SetCoapHeaders(buf io.Writer, m Message, nps NoisePipeState, msgID uint16) (mID uint16, err error) {

	debugf("SetCoapHeaders with nps=%+v, msgID=%v", nps, msgID)

	var v, t, c uint8

	// Token is only needed if we're not handshaking.
	var token = make([]byte, 0)

	// We need to define a message ID here if we're handshaking.
	var handshakeMsgID = uint16(conn.rRand.Uint32())

	// Ver = 011b instead of 001b indicates that encryption is turned on.
	v = 3

	// Handshake | CoAP Type | CoAP Code |
	// ----------|-----------|-----------|
	// XX1       | 0 (CON)   | 250       |
	// XX2       | 2 (ACK)   | 250       |
	// XX3       | 1 (NON)   | 251       |
	// IK1       | 0 (CON)   | 252       |
	// IK2       | 2 (ACK)   | 252       |
	switch nps {
	case XX1:
		t = 0
		c = 250
		mID = handshakeMsgID
	case XX2:
		t = 2
		c = 250
		mID = handshakeMsgID
	case XX3:
		t = 1
		c = 251
		mID = handshakeMsgID
	case IK1:
		t = 0
		c = 252
		mID = handshakeMsgID
	case IK2:
		t = 2
		c = 252
		mID = handshakeMsgID
	default:
		if m == nil {
			err = errors.New("Message can't be nil outside of handshake")
			return
		}

		t = uint8(m.Type())
		c = uint8(m.Code())
		token = m.Token()
		mID = m.MessageID()
	}

	if msgID != 0 {
		mID = msgID
	}

	// MessageID is 16bit, so we split it into two bytes (since we can only put
	// bytes into our writer)
	tmpbuf := []byte{0, 0}
	binary.BigEndian.PutUint16(tmpbuf, mID)

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |Ver| T |  TKL  |      Code     |          Message ID           |
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	if _, err = buf.Write([]byte{
		(v << 6) | (t << 4) | uint8(len(token)),
		byte(c),
		tmpbuf[0], tmpbuf[1],
	}); err != nil {
		return
	}

	//  0                   1                   2                   3
	//  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// |   Token (if any, TKL bytes) ...
	// +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
	// Token can be 0 to 8 bytes long
	if len(token) > MaxTokenSize {
		err = ErrInvalidTokenLen
		return
	}
	if _, err = buf.Write(token); err != nil {
		return
	}

	//  0
	//  0 1 2 3 4 5 6 7
	// +-+-+-+-+-+-+-+-+
	// |    Seq num    |
	// +-+-+-+-+-+-+-+-+
	if _, err = buf.Write([]byte{byte(conn.retriesQueue.seqnum)}); err != nil {
		return
	}

	debugf("Sending message using CoAP version %d, of type %d and code %d, and with message ID %d, token of length %d, and sequence number %d\n", v, t, c, mID, len(token), conn.retriesQueue.seqnum)

	return
}

func newConnectionTCP(c net.Conn, srv *Server) Conn {
	connection := &connTCP{connBase: connBase{writeChan: make(chan writeReq, 10000), closeChan: make(chan bool), finChan: make(chan bool), closed: 0}, connection: c}
	go writeToConnection(connection, srv)
	return connection
}

func newConnectionUDP(c *net.UDPConn, srv *Server) Conn {
	rSource := rand.NewSource(time.Now().UnixNano())
	rRand := rand.New(rSource)

	connection := &connUDP{
		connBase:     connBase{writeChan: make(chan writeReq, 10000), closeChan: make(chan bool), finChan: make(chan bool), closed: 0},
		connection:   c,
		retriesQueue: srv.RetriesQueue,
		rRand:        rRand,
	}

	debugf("Creating new connection to %v with address %p and queue %p", c.RemoteAddr(), connection, connection.retriesQueue)

	//log.Printf("newConnectionUDP called with conn=%p", connection)
	//debug.PrintStack()

	go writeToConnection(connection, srv)
	return connection
}

func writeToConnection(conn connWriter, srv *Server) {
	for conn.writeHandler(srv) {
	}
	for conn.writeEndHandler(srv.syncTimeout()) {
	}
	conn.sendFinish(srv.syncTimeout())
}
