// Package coap provides a CoAP client and server.
package coap

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Interval for stop worker if no load
const idleWorkerTimeout = 10 * time.Second

// Maximum number of workers
const maxWorkersCount = 10

const coapTimeout time.Duration = 3600 * time.Second

const syncTimeout time.Duration = 180 * time.Second

const maxMessageSize = 1152

const (
	defaultReadBufferSize  = 4096
	defaultWriteBufferSize = 4096
)

//DefaultPort default unsecure port for COAP server
const DefaultPort = 5683

//DefaultSecurePort default secure port for COAP server
const DefaultSecurePort = 5684

//const tcpIdleTimeout time.Duration = 8 * time.Second

// Handler is implemented by any value that implements ServeCOAP.
type Handler interface {
	ServeCOAP(w ResponseWriter, r *Request)
}

type Compressor interface {
	CompressPayload(j []byte) ([]byte, error)
	DecompressPayload(j []byte) ([]byte, error)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as COAP handlers.  If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler object that calls f.
type HandlerFunc func(ResponseWriter, *Request)

// ServeCOAP calls f(w, r).
func (f HandlerFunc) ServeCOAP(w ResponseWriter, r *Request) {
	f(w, r)
}

// HandleFailed returns a HandlerFunc that returns NotFound for every request it gets.
func HandleFailed(w ResponseWriter, req *Request) {
	msg := req.Client.NewMessage(MessageParams{
		Type:      Acknowledgement,
		Code:      NotFound,
		MessageID: req.Msg.MessageID(),
		Token:     req.Msg.Token(),
	})
	w.WriteMsg(msg)
}

func failedHandler() Handler { return HandlerFunc(HandleFailed) }

// ListenAndServe Starts a server on address and network specified Invoke handler
// for incoming queries.
func ListenAndServe(addr string, network string, handler Handler) error {
	server := &Server{Addr: addr, Net: network, Handler: handler}
	return server.ListenAndServe()
}

// ListenAndServeTLS acts like http.ListenAndServeTLS, more information in
// http://golang.org/pkg/net/http/#ListenAndServeTLS
func ListenAndServeTLS(addr, certFile, keyFile string, handler Handler) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	server := &Server{
		Addr:      addr,
		Net:       "tcp-tls",
		TLSConfig: &config,
		Handler:   handler,
	}

	return server.ListenAndServe()
}

// ActivateAndServe activates a server with a listener from systemd,
// l and p should not both be non-nil.
// If both l and p are not nil only p will be used.
// Invoke handler for incoming queries.
func ActivateAndServe(l net.Listener, p net.Conn, handler Handler) error {
	server := &Server{Listener: l, Conn: p, Handler: handler}
	return server.ActivateAndServe()
}

// A Server defines parameters for running an COAP server.
type Server struct {
	// Address to listen on, ":COAP" if empty.
	Addr string
	// if "tcp" or "tcp-tls" (COAP over TLS) it will invoke a TCP listener, otherwise an UDP one
	Net string
	// TCP Listener to use, this is to aid in systemd's socket activation.
	Listener net.Listener
	// TLS connection configuration
	TLSConfig *tls.Config
	// UDP/TCP "Listener/Connection" to use, this is to aid in systemd's socket activation.
	Conn net.Conn
	// Handler to invoke, COAP.DefaultServeMux if nil.
	Handler Handler
	// Max message size that could be received from peer. Min 16bytes. If not set
	// it defaults to 1152 B.
	MaxMessageSize uint32
	// The net.Conn.SetReadTimeout value for new connections, defaults to 1hour.
	ReadTimeout time.Duration
	// The net.Conn.SetWriteTimeout value for new connections, defaults to 1hour.
	WriteTimeout time.Duration
	// If NotifyStartedFunc is set it is called once the server has started listening.
	NotifyStartedFunc func()
	// The maximum of time for synchronization go-routines, defaults to 30 seconds, if it occurs, then it call log.Fatal
	SyncTimeout time.Duration
	// If newSessionUDPFunc is set it is called when session UDP want to be created
	newSessionUDPFunc func(connection Conn, srv *Server, sessionUDPData *SessionUDPData, initiator bool) (networkSession, error)
	// If newSessionUDPFunc is set it is called when session TCP want to be created
	newSessionTCPFunc func(connection Conn, srv *Server) (networkSession, error)
	// If NotifyNewSession is set it is called when new TCP/UDP session was created.
	NotifySessionNewFunc func(w *ClientCommander)
	// If NotifyNewSession is set it is called when TCP/UDP session was ended.
	NotifySessionEndFunc func(w *ClientCommander, err error)
	// The interfaces that will be used for udp-mcast (default uses the system assigned for multicast)
	UDPMcastInterfaces []net.Interface
	// Use blockWise transfer for transfer payload (default for UDP it's enabled, for TCP it's disable)
	BlockWiseTransfer *bool
	// Set maximal block size of payload that will be send in fragment
	BlockWiseTransferSzx *BlockWiseSzx

	// Used to compress packets
	Compressor Compressor

	// is encryption on or off?
	Encryption bool

	// what keystore to use for tracking remote server keys
	KeyStore KeyStore

	// psk if any
	Psk []byte

	TCPReadBufferSize  int
	TCPWriteBufferSize int

	readerPool sync.Pool
	writerPool sync.Pool
	// UDP packet or TCP connection queue
	queue chan *Request
	// Workers count
	workersCount int32
	// Shutdown handling
	lock    sync.RWMutex
	started bool

	sessionUDPMapLock sync.Mutex
	sessionUDPMap     map[string]networkSession

	// Queue used to schedule, operate and cancel retries of sent messages
	RetriesQueue *RetriesQueue
}

func (srv *Server) workerChannelHandler(inUse bool, timeout *time.Timer) bool {
	select {
	case w, ok := <-srv.queue:
		if !ok {
			return false
		}
		inUse = true
		srv.serve(w)
	case <-timeout.C:
		if !inUse {
			return false
		}
		inUse = false
		timeout.Reset(idleWorkerTimeout)
	}
	return true
}

func (srv *Server) worker(w *Request) {
	srv.serve(w)

	for {
		count := atomic.LoadInt32(&srv.workersCount)
		if count > maxWorkersCount {
			return
		}
		if atomic.CompareAndSwapInt32(&srv.workersCount, count, count+1) {
			break
		}
	}

	defer atomic.AddInt32(&srv.workersCount, -1)

	inUse := false
	timeout := time.NewTimer(idleWorkerTimeout)
	defer timeout.Stop()
	for srv.workerChannelHandler(inUse, timeout) {
	}
}

func (srv *Server) spawnWorker(w *Request) {
	select {
	case srv.queue <- w:
	default:
		go srv.worker(w)
	}
}

// ListenAndServe starts a coapserver on the configured address in *Server.
func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		switch {
		case strings.Contains(srv.Net, "-tls"):
			addr = ":" + strconv.Itoa(DefaultSecurePort)
		default:
			addr = ":" + strconv.Itoa(DefaultPort)
		}
	}

	switch srv.Net {
	case "tcp", "tcp4", "tcp6":
		a, err := net.ResolveTCPAddr(srv.Net, addr)
		if err != nil {
			return err
		}
		l, err := net.ListenTCP(srv.Net, a)
		if err != nil {
			return err
		}
		srv.Listener = l
	case "tcp-tls", "tcp4-tls", "tcp6-tls":
		network := strings.TrimSuffix(srv.Net, "-tls")

		l, err := tls.Listen(network, addr, srv.TLSConfig)
		if err != nil {
			return err
		}
		srv.Listener = l
	case "udp", "udp4", "udp6":
		a, err := net.ResolveUDPAddr(srv.Net, addr)
		if err != nil {
			return err
		}
		l, err := net.ListenUDP(srv.Net, a)
		if err != nil {
			return err
		}
		if err := setUDPSocketOptions(l); err != nil {
			return err
		}
		srv.Conn = l
	case "udp-mcast", "udp4-mcast", "udp6-mcast":
		network := strings.TrimSuffix(srv.Net, "-mcast")

		a, err := net.ResolveUDPAddr(network, addr)
		if err != nil {
			return err
		}
		l, err := net.ListenUDP(network, a)
		if err != nil {
			return err
		}
		if err := setUDPSocketOptions(l); err != nil {
			return err
		}
		if len(srv.UDPMcastInterfaces) > 0 {
			for _, ifi := range srv.UDPMcastInterfaces {
				if err := joinGroup(l, &ifi, &net.UDPAddr{IP: a.IP, Zone: a.Zone}); err != nil {
					return err
				}
			}
		} else {
			if err := joinGroup(l, nil, &net.UDPAddr{IP: a.IP, Zone: a.Zone}); err != nil {
				return err
			}
		}
		srv.Conn = l
	default:
		return ErrInvalidNetParameter
	}
	return srv.ActivateAndServe()
}

func (srv *Server) initServeUDP(conn *net.UDPConn) error {
	srv.lock.Lock()
	srv.started = true
	srv.lock.Unlock()
	return srv.serveUDP(conn)
}

func (srv *Server) initServeTCP(conn net.Conn) error {
	srv.lock.Lock()
	srv.started = true
	srv.lock.Unlock()
	if srv.NotifyStartedFunc != nil {
		srv.NotifyStartedFunc()
	}
	return srv.serveTCPconnection(conn)
}

// ActivateAndServe starts a coapserver with the PacketConn or Listener
// configured in *Server. Its main use is to start a server from systemd.
func (srv *Server) ActivateAndServe() error {
	srv.lock.Lock()
	if srv.started {
		srv.lock.Unlock()
		return ErrServerAlreadyStarted
	}
	srv.lock.Unlock()

	pConn := srv.Conn
	l := srv.Listener

	if srv.MaxMessageSize == 0 {
		srv.MaxMessageSize = maxMessageSize
	}
	if srv.MaxMessageSize < uint32(szxToBytes[BlockWiseSzx16]) {
		return ErrInvalidMaxMesssageSizeParameter
	}

	srv.sessionUDPMap = make(map[string]networkSession)

	srv.queue = make(chan *Request)
	defer close(srv.queue)

	if srv.newSessionTCPFunc == nil {
		srv.newSessionTCPFunc = func(connection Conn, srv *Server) (networkSession, error) {
			session, err := newSessionTCP(connection, srv)
			if err != nil {
				return nil, err
			}
			if session.blockWiseEnabled() {
				return &blockWiseSession{networkSession: session}, nil
			}
			return session, nil
		}
	}

	if srv.newSessionUDPFunc == nil {
		srv.newSessionUDPFunc = func(connection Conn, srv *Server, sessionUDPData *SessionUDPData, initiator bool) (networkSession, error) {
			session, err := newSessionUDP(connection, srv, sessionUDPData, initiator)
			if err != nil {
				return nil, err
			}
			if session.blockWiseEnabled() {
				return &blockWiseSession{networkSession: session}, nil
			}
			return session, nil
		}
	}

	if srv.NotifySessionNewFunc == nil {
		srv.NotifySessionNewFunc = func(w *ClientCommander) {}
	}

	if srv.NotifySessionEndFunc == nil {
		srv.NotifySessionEndFunc = func(w *ClientCommander, err error) {}
	}

	if pConn != nil {
		switch pConn.(type) {
		case *net.TCPConn, *tls.Conn:
			return srv.initServeTCP(pConn)
		case *net.UDPConn:
			return srv.initServeUDP(pConn.(*net.UDPConn))
		}
		return ErrInvalidServerConnParameter
	}
	if l != nil {
		srv.lock.Lock()
		srv.started = true
		srv.lock.Unlock()
		return srv.serveTCP(l)
	}

	return ErrInvalidServerListenerParameter
}

// Shutdown shuts down a server. After a call to Shutdown, ListenAndServe and
// ActivateAndServe will return.
func (srv *Server) Shutdown() error {
	srv.lock.Lock()
	if !srv.started {
		srv.lock.Unlock()
		return ErrServerNotStarted
	}
	srv.started = false
	srv.lock.Unlock()

	if srv.Conn != nil {
		srv.Conn.Close()
	}
	if srv.Listener != nil {
		srv.Listener.Close()
	}
	return nil
}

// readTimeout is a helper func to use system timeout if server did not intend to change it.
func (srv *Server) readTimeout() time.Duration {
	if srv.ReadTimeout != 0 {
		return srv.ReadTimeout
	}
	return coapTimeout
}

// readTimeout is a helper func to use system timeout if server did not intend to change it.
func (srv *Server) writeTimeout() time.Duration {
	if srv.WriteTimeout != 0 {
		return srv.WriteTimeout
	}
	return coapTimeout
}

func (srv *Server) syncTimeout() time.Duration {
	if srv.SyncTimeout != 0 {
		return srv.SyncTimeout
	}
	return syncTimeout
}

func (srv *Server) serveTCPconnection(conn net.Conn) error {
	conn.SetReadDeadline(time.Now().Add(srv.readTimeout()))
	session, err := srv.newSessionTCPFunc(newConnectionTCP(conn, srv), srv)
	if err != nil {
		return err
	}
	srv.NotifySessionNewFunc(&ClientCommander{session})
	br := srv.acquireReader(conn)
	defer srv.releaseReader(br)
	for {
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()
			return session.Close()
		}
		srv.lock.RUnlock()
		mti, err := readTcpMsgInfo(br)
		if err != nil {
			return session.closeWithError(err)
		}
		o, p, err := readTcpMsgBody(mti, br)
		if err != nil {
			return session.closeWithError(err)
		}
		msg := new(TcpMessage)
		//msg := TcpMessage{MessageBase{}}

		msg.fill(mti, o, p)

		// We will block poller wait loop when
		// all pool workers are busy.
		srv.spawnWorker(&Request{Client: &ClientCommander{session}, Msg: msg})
	}
}

// serveTCP starts a TCP listener for the server.
func (srv *Server) serveTCP(l net.Listener) error {
	defer l.Close()

	if srv.NotifyStartedFunc != nil {
		srv.NotifyStartedFunc()
	}

	doneDescChan := make(chan bool)
	numRunningDesc := 0

	for {
	LOOP_CLOSE_CHANNEL:
		for {
			select {
			case <-doneDescChan:
				numRunningDesc--
			default:
				break LOOP_CLOSE_CHANNEL
			}
		}

		rw, err := l.Accept()
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()
			if rw != nil {
				rw.Close()
			}
			for numRunningDesc > 0 {
				<-doneDescChan
				numRunningDesc--
			}
			return nil
		}
		srv.lock.RUnlock()
		if err != nil {
			if neterr, ok := err.(net.Error); ok && neterr.Temporary() {
				continue
			}
			return err
		}

		numRunningDesc++

		go func() {
			srv.serveTCPconnection(rw)
			doneDescChan <- true
		}()
	}
}

func (srv *Server) closeSessions(err error) {
	srv.sessionUDPMapLock.Lock()
	tmp := srv.sessionUDPMap
	srv.sessionUDPMap = make(map[string]networkSession)
	srv.sessionUDPMapLock.Unlock()
	for _, v := range tmp {
		srv.NotifySessionEndFunc(&ClientCommander{v}, err)
	}
}

// serveUDP starts a UDP listener for the server.
func (srv *Server) serveUDP(conn *net.UDPConn) error {
	defer conn.Close()

	if srv.NotifyStartedFunc != nil {
		srv.NotifyStartedFunc()
	}

	rtimeout := srv.readTimeout()
	// deadline is not used here

	connUDP := newConnectionUDP(conn, srv).(*connUDP)

	for {
		m := make([]byte, ^uint16(0))
		srv.lock.RLock()
		if !srv.started {
			srv.lock.RUnlock()
			srv.closeSessions(nil)
			return nil
		}
		srv.lock.RUnlock()

		err := connUDP.SetReadDeadline(time.Now().Add(rtimeout))
		if err != nil {
			srv.closeSessions(err)
			return err
		}
		m = m[:cap(m)]
		n, s, err := connUDP.ReadFromSessionUDP(m)
		if err != nil {
			debugf("error in ReadFromSessionUDP: %v", err)
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				continue
			}
			srv.closeSessions(err)
			return err
		}
		m = m[:n]

		debugf("Received %d bytes into %X", n, m)
		var h retryHeaders
		h, m = connUDP.extractRetryHeaders(m)

		srv.sessionUDPMapLock.Lock()
		debugf("Looking at session with key %s", s.Key())
		session := srv.sessionUDPMap[s.Key()]
		if session == nil {
			session, err = srv.newSessionUDPFunc(connUDP, srv, s, false)
			if err != nil {
				return err
			}
			srv.NotifySessionNewFunc(&ClientCommander{session})
			srv.sessionUDPMap[s.Key()] = session
			srv.sessionUDPMapLock.Unlock()
		} else {
			srv.sessionUDPMapLock.Unlock()
		}

		// TODO:
		//    Check the CoAP headers (whether the payload is encrypted or not).
		//    If it's a response to a confirmable message, then we mark this msgid+token pair
		// 	  as having been received, so we stop retrying to send on the send path.
		//
		//    If it's encrypted, we then look at the seqnum, and reorder the packets
		//	  in order to ensure we feed them to noise in the right order without dups or gaps
		//
		//    We have to expand out the seqnum from being 8-bits to a larger value for reordering
		//	  purposes.   (We assume that seqnum will never jump more than 2^7 bytes at a time).
		//	  We then have to pass the expanded seqnum to noise via SetNonce() so that noise
		//	  knows how to decrypt the packet.  Therefore, we don't actually need a reordering
		//	  buffer \o/

		if srv.Encryption {

			ns := session.GetNoiseState()

			// If we receive a XX1 msg, it means the remote server initiated a
			// re-handshake.
			if h.nps == XX1 && ns.PipeState != XX1 {
				ns.Initiator = false
				ns.SetupXX()
			}

			var queuedMsg *HSQueueMsg
			// Zero value of []byte is nil, which is what we want if we don't
			// want to append a payload to a handshake message.
			var piggybacked []byte
			// We only want a piggybacked payload if we're about to send a XX3
			// message (which has the prerequisites that our current state is XX2
			// and we're the initiator).
			if ns.PipeState == XX2 && ns.Initiator {
				if queuedMsg = connUDP.retriesQueue.PopHS(s.raddr.IP.String()); queuedMsg != nil {
					piggybacked = queuedMsg.b
				}
				debugf("Popped %p", queuedMsg)
			}

			// XXX: ideally we'd decrypt in connUDP.ReadFromSessionUDP, but we don't
			// know which session we're part of at that point. So instead we decrypt here.
			// We might want to call through to Conn for this?
			// log.Printf("decrypting %d bytes with %+v as %v", n, ns, m)

			var toSend []byte
			var decrypted bool

			// At this stage, ns.PipeState will be equal to the value we expect
			// from the incoming message.
			expectedPipeState := ns.PipeState

			m, toSend, decrypted, err = ns.DecryptMessage(m, piggybacked, h.seqnum, connUDP, s)
			// TODO: Move this to noise.go
			if err == ErrIncoherentHandshakeMsg || expectedPipeState != h.nps {
				// If we get an incoherent handshake message, initiate a
				// re-handshake.
				// Here, an incoherent handshake message is either a
				// non-handshake message that arrives before the handshake is
				// complete (i.e. before we have values for Cs0 and Cs1), or a
				// handshake message that arrives out of sequence (i.e. XX2 when
				// we're waiting for XX3).
				// It can also be e.g. a XX3 message received by the initiator,
				// in which case we consider the remote server as confused and
				// restart the handshake from scratch.
				debugf("Got incoherent handshake message, re-handshaking")
				ns.Initiator = true
				ns.SetupXX()
				// Encrypt the XX1 message to send
				toSend, err = ns.EncryptMessage(nil, connUDP, s)
				if err != nil {
					log.Printf("ERROR: Failed to restart handshake: %v", err)
					continue
				}
				// Drop all retries to this destination on the floor.
				dest := strings.Split(s.RemoteAddr().String(), ":")[0]
				debugf("Cancelling all pending retries to %s", dest)
				var mIDToCancel *uint16
				for {
					if mIDToCancel = srv.RetriesQueue.PopMID(dest); mIDToCancel == nil {
						break
					}
					srv.RetriesQueue.CancelRetrySchedule(*mIDToCancel)
				}
			} else if err != nil {
				log.Printf("ERROR: Failed to decrypt message: %v", err)
				continue
			}

			debugf("Having decrypted message, toSend is: %X", toSend)
			if toSend != nil {
				buf := &bytes.Buffer{}
				var mID uint16

				// If queuedMsg isn't nil, then we're both in XX3 and the
				// initiator. In this case we need to use the message of the
				// original payload instead of a new "fake" one, because it's
				// the one the response we use, which will allow us to correctly
				// remove the message from the retry schedule once the response
				// arrives.
				if queuedMsg == nil {
					mID = h.msgID
					debugf("Got messageID from incoming msg: %d", mID)
				} else {
					mID = queuedMsg.m.MessageID()
					debugf("Got messageID from queued msg: %d", mID)
				}

				// If we ever have toSend set, surely it means that we are either:
				// sending a XX2 response having received an XX1 (in which case our NPS is XX3 and must be decremented)
				// or sending a XX3 response having received an XX2 (in which case our NPS is READY and must be decremented).
				if _, err = connUDP.SetCoapHeaders(buf, nil, ns.PipeState-1, mID); err != nil {
					log.Printf("ERROR: Failed to SetCoapHeaders: %v", err)
					continue
				}

				if _, err = buf.Write(toSend); err != nil {
					// If we fail to send the handshake response, we can drop
					// the message on the floor because the remote server will
					// likely retry sending the initial message.
					continue
				}
				// bm != nil is only true when we're sending XX3
				if queuedMsg != nil {
					go srv.RetriesQueue.ScheduleRetry(mID, buf.Bytes(), s, connUDP)
				}
				if err = connUDP.writeToSession(buf.Bytes(), s); err != nil {
					continue
				}

				if ns.PipeState == READY {
					var msg *HSQueueMsg
					for {
						if msg = srv.RetriesQueue.PopHS(s.raddr.IP.String()); msg == nil {
							break
						}
						// Run the sending in a goroutine since it's not
						// supposed to have side-effects and we don't want it to
						// block the request we're currently processing.
						go msg.conn.sendMessage(msg.m, ns, msg.sessionData, srv)
					}
				}
			}
			// log.Printf("decrypted %d bytes with %+v as %v", len(m), ns, m)
			// log.Printf("decrypted %d->%d bytes with %+v", n, len(m), ns)

			if !decrypted || len(m) == 0 {
				// it'll be a dummy zero-length msg generated by a noise handshake
				continue
			}
		}

		var decompressed []byte
		if srv.Compressor != nil {
			decompressed, err = srv.Compressor.DecompressPayload(m)
			if err != nil {
				log.Printf("ERROR: Failed to decompress payload: %v", err)
				continue
			}
		} else {
			decompressed = m
		}

		msg, err := ParseDgramMessage(decompressed)
		if err != nil {
			log.Printf("ERROR: failed to ParseDgramMessage: %v", err)
			continue
		}

		// log.Printf("Decompressed msg %s: %d -> %d bytes", msg.Token(), len(m), len(decompressed))

		srv.spawnWorker(&Request{Msg: msg, Client: &ClientCommander{session}})
	}
}

func (srv *Server) serve(r *Request) {
	w := responseWriterFromRequest(r)

	handlePairMsg(w, r, func(w ResponseWriter, r *Request) {
		handleSignalMsg(w, r, func(w ResponseWriter, r *Request) {
			handleBySessionTokenHandler(w, r, func(w ResponseWriter, r *Request) {
				handleBlockWiseMsg(w, r, srv.serveCOAP)
			})
		})
	})
}

func (srv *Server) serveCOAP(w ResponseWriter, r *Request) {
	handler := srv.Handler
	if handler == nil || reflect.ValueOf(handler).IsNil() {
		handler = DefaultServeMux
	}
	handler.ServeCOAP(w, r) // Writes back to the client
}

func (srv *Server) acquireReader(tcp net.Conn) *bufio.Reader {
	v := srv.readerPool.Get()
	if v == nil {
		n := srv.TCPReadBufferSize
		if n <= 0 {
			n = defaultReadBufferSize
		}
		return bufio.NewReaderSize(tcp, n)
	}
	r := v.(*bufio.Reader)
	r.Reset(tcp)
	return r
}

func (srv *Server) releaseReader(r *bufio.Reader) {
	srv.readerPool.Put(r)
}

func (srv *Server) acquireWriter(tcp net.Conn) *bufio.Writer {
	v := srv.writerPool.Get()
	if v == nil {
		n := srv.TCPWriteBufferSize
		if n <= 0 {
			n = defaultWriteBufferSize
		}
		return bufio.NewWriterSize(tcp, n)
	}
	wr := v.(*bufio.Writer)
	wr.Reset(tcp)
	return wr
}

func (srv *Server) releaseWriter(wr *bufio.Writer) {
	srv.writerPool.Put(wr)
}
