package main

import (
	"time"

	"github.com/matrix-org/coap-proxy/common"

	"github.com/matrix-org/go-coap"
)

// Map of open connections with the host address as the key. Allows us to keep
// track of the last time a message was sent for timeout purposes.
var conns map[string]*openConn

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

// resetConn is a function that given a CoAP target (address and port), closes
// any existing connections to it and opens a new one.
func resetConn(target string) (*openConn, error) {
	if c, exists := conns[target]; exists {
		common.Debugf("Closing UDP connection to %s", target)
		_ = c.Close()
	}

	common.Debugf("Creating new UDP connection to %s", target)
	return newOpenConn(target)
}
