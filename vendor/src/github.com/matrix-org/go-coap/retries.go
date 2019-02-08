package coap

import (
	"strings"
	"sync"
	"time"
)

type RetriesQueue struct {
	q map[uint16]queueEl
	// Maps ip addresses (or hostnames) to message IDs so we can know which
	// destination should be sending us the response for a given message.
	mIDs   map[string][]uint16
	seqnum uint8 // Next sequence number, differs from the one in
	mut    *sync.Mutex
	// We store messages we put aside during handshakes here.
	// TODO: This will eventually need to be moved to a dedicated structure.
	hsQueue    map[string][]HSQueueMsg
	initDelay  time.Duration
	multiplier time.Duration
}

type queueEl struct {
	ch chan bool
	// We need to store sequence numbers in the queue in order to detect and
	// measure holes in the message queue, e.g. we'd need to re-handshake if we
	// observe a gap higher than 128 msgs (max(seqnum)/2).
	seqnum uint8
}

type HSQueueMsg struct {
	b           []byte  // We might do compression so we can't just call MarshalBinary
	m           Message // Kept for metadata access
	conn        *connUDP
	sessionData *SessionUDPData
}

func NewRetriesQueue(initDelay time.Duration, multiplier int) *RetriesQueue {
	rq := new(RetriesQueue)
	rq.q = make(map[uint16]queueEl)
	rq.mIDs = make(map[string][]uint16)
	rq.mut = new(sync.Mutex)
	rq.hsQueue = make(map[string][]HSQueueMsg)
	rq.initDelay = initDelay
	rq.multiplier = time.Duration(multiplier)
	return rq
}

func (rq *RetriesQueue) ScheduleRetry(mID uint16, b []byte, session *SessionUDPData, conn *connUDP) {
	debugf("Scheduling retries for message %d", mID)

	rq.mut.Lock()

	ch := make(chan bool)
	if _, ok := rq.q[mID]; !ok {
		rq.q[mID] = queueEl{
			seqnum: rq.seqnum,
			ch:     ch,
		}
	}

	destination := strings.Split(session.RemoteAddr().String(), ":")[0]

	// Store which destination this message is for.
	if _, ok := rq.mIDs[destination]; !ok {
		rq.mIDs[destination] = make([]uint16, 1)
		rq.mIDs[destination][0] = mID
	} else {
		rq.mIDs[destination] = append(rq.mIDs[destination], mID)
	}

	rq.mut.Unlock()

	rq.retryIfNoResp(mID, ch, rq.initDelay, b, session, conn)
}

func (rq *RetriesQueue) retryIfNoResp(mID uint16, ch chan bool, delay time.Duration, b []byte, session *SessionUDPData, conn *connUDP) {
	select {
	case <-ch:
		// Cancel retry
		debugf("Received response for message %d, not retrying", mID)
		return
	case <-time.After(rq.initDelay):
		debugf("No response for message %d, retrying", mID)
		if err := conn.writeToSession(b, session); err != nil {
			debugf("Retried failed: %s", err.Error())
		}
		// Wait a bit more then retry
		rq.retryIfNoResp(mID, ch, delay*rq.multiplier, b, session, conn)
	}
}

func (rq *RetriesQueue) PopMID(dest string) *uint16 {
	if mIDs, ok := rq.mIDs[dest]; !ok || len(mIDs) == 0 {
		return nil
	}

	mID := rq.mIDs[dest][0]

	if len(rq.mIDs[dest]) > 1 {
		rq.mIDs[dest] = rq.mIDs[dest][1:]
	} else {
		rq.mIDs[dest] = make([]uint16, 0)
	}

	return &mID
}

func (rq *RetriesQueue) CancelRetrySchedule(mID uint16) {
	rq.mut.Lock()

	if _, ok := rq.q[mID]; ok {
		debugf("Cancelling retry schedule for message %d", mID)

		rq.q[mID].ch <- true
		delete(rq.q, mID)
	}

	rq.mut.Unlock()
}

func (rq *RetriesQueue) PushHS(msg HSQueueMsg) {
	if _, ok := rq.hsQueue[msg.sessionData.raddr.IP.String()]; !ok {
		rq.hsQueue[msg.sessionData.raddr.IP.String()] = make([]HSQueueMsg, 0)
	}
	rq.hsQueue[msg.sessionData.raddr.IP.String()] = append(rq.hsQueue[msg.sessionData.raddr.IP.String()], msg)
	return
}

func (rq *RetriesQueue) PopHS(host string) *HSQueueMsg {
	if q, ok := rq.hsQueue[host]; !ok || len(q) == 0 {
		return nil
	}

	bm := rq.hsQueue[host][0]

	if len(rq.hsQueue[host]) > 1 {
		rq.hsQueue[host] = rq.hsQueue[host][1:]
	} else {
		rq.hsQueue[host] = make([]HSQueueMsg, 0)
	}

	return &bm
}
