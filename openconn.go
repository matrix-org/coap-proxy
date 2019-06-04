package main

import (
	"sync"
	"time"

	"github.com/matrix-org/coap-proxy/common"

	"github.com/matrix-org/go-coap"
)

// Map of open connections with the host address as the key. Allows us to keep
// track of the last time a message was sent for timeout purposes.
var conns map[string]*sync.Pool
var connsLock sync.Mutex

// openConn is a struct that represents an open CoAP connection to another
// coap-proxy instance. We keep a map of these for timeout tracking purposes.
type openConn struct {
	*coap.ClientConn
	lastMsg    time.Time
	killswitch chan bool
	dead       bool
}

func newOpenConn(target string) (c *openConn, err error) {
	c = new(openConn)
	if c.ClientConn, err = dialTimeout("udp", target, 300*time.Second); err != nil {
		return
	}
	c.killswitch = make(chan bool)

	//go c.heartbeat()

	return
}

func (c *openConn) Close() error {
	c.killswitch <- true
	return c.ClientConn.Close()
}

func (c *openConn) heartbeat() {
	for {
		// Wait before sending the first heatbeat so that the handshake and the
		// first exchange can happen.
		select {
		case <-c.killswitch:
			common.Debugf("Got killswitch signal for connection to %s", c.ClientConn.RemoteAddr().String())
			return
		case <-time.After(30 * time.Second):
			common.Debugf("Sending heartbeat to %s", c.ClientConn.RemoteAddr().String())
		}

		if err := c.ClientConn.Ping(10 * time.Second); err != nil {
			common.Debugf("Connection to %s is dead", c.ClientConn.RemoteAddr().String())
			c.dead = true
			return
		}

		common.Debugf("Connection to %s is alive", c.ClientConn.RemoteAddr().String())
	}
}

// getConn is a function that given a CoAP target (address and port), gets
// an existing idle connection or creates a new one.
func getConn(target string) (*openConn, error) {
	connsLock.Lock()
	defer connsLock.Unlock()

	pool := conns[target]
	if pool == nil {
		pool = &sync.Pool{}
		conns[target] = pool
	}

	conn := pool.Get()
	if conn != nil {
		common.Debugf("Reusing UDP connection to %s", target)
		return conn.(*openConn), nil
	}

	common.Debugf("Creating new UDP connection to %s", target)

	c, err := newOpenConn(target)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func putConn(target string, conn *openConn) {
	connsLock.Lock()
	defer connsLock.Unlock()

	common.Debugf("Adding UDP connection to %s to pool", target)

	pool := conns[target]
	if pool == nil {
		pool = &sync.Pool{}
		conns[target] = pool
	}

	pool.Put(conn)
}
