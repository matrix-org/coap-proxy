package main

import (
	"net/http"
	"time"

	"github.com/matrix-org/coap-proxy/common"

	"github.com/matrix-org/go-coap"
)

// listenAndServe is a function that wraps around a CoAP server with a
// specialised configuration and asks it to listen on the given address and
// port.
func listenAndServe(addr string, network string, handler coap.Handler, comp coap.Compressor) error {
	blockWiseTransfer := true
	blockWiseTransferSzx := coap.BlockWiseSzx1024
	server := &coap.Server{
		Addr:                 addr,
		Net:                  network,
		Handler:              handler,
		BlockWiseTransfer:    &blockWiseTransfer,
		BlockWiseTransferSzx: &blockWiseTransferSzx,
		MaxMessageSize:       ^uint32(0),
		Encryption:           !(*noEncryption),
		KeyStore:             keyStore,
		Compressor:           comp,
		RetriesQueue:         retriesQueue,
	}
	return server.ListenAndServe()
}

// dialTimeout is a function that dials (connects to) a CoAP server as a CoAP
// client and times out on a given timeout.Duration.
func dialTimeout(network, address string, timeout time.Duration) (*coap.ClientConn, error) {
	blockWiseTransfer := true
	blockWiseTransferSzx := coap.BlockWiseSzx1024
	client := coap.Client{
		Net:                  network,
		DialTimeout:          timeout,
		BlockWiseTransfer:    &blockWiseTransfer,
		BlockWiseTransferSzx: &blockWiseTransferSzx,
		MaxMessageSize:       ^uint32(0),
		Encryption:           !(*noEncryption),
		KeyStore:             keyStore,
		Compressor:           compressor,
		RetriesQueue:         retriesQueue,
	}
	return client.Dial(address)
}

// statusCoAPToHTTP is a function that converts a CoAP status code to its
// equivalent HTTP status code.
func statusCoAPToHTTP(coapCode coap.COAPCode) uint16 {
	switch coapCode {
	case coap.Content:
		return http.StatusOK
	case coap.Changed:
		return http.StatusFound
	case coap.BadRequest:
		return http.StatusBadRequest
	case coap.Unauthorized:
		return http.StatusUnauthorized
	case coap.BadOption:
		return http.StatusConflict
	case coap.Forbidden:
		return http.StatusForbidden
	case coap.NotFound:
		return http.StatusNotFound
	case coap.MethodNotAllowed:
		return http.StatusMethodNotAllowed
	case coap.RequestEntityTooLarge:
		return http.StatusTooManyRequests
	case coap.InternalServerError:
		return http.StatusInternalServerError
	case coap.BadGateway:
		return http.StatusBadGateway
	case coap.ServiceUnavailable:
		return http.StatusServiceUnavailable
	case coap.GatewayTimeout:
		return http.StatusGatewayTimeout
	default:
		common.Debugf("Unsupported CoAP code %s", coapCode.String())
		return http.StatusInternalServerError
	}
}

// statusHTTPToCoAP is a function that converts an HTTP status code to its
// equivalent CoAP status code.
func statusHTTPToCoAP(httpCode int) coap.COAPCode {
	switch httpCode {
	case http.StatusOK:
		return coap.Content
	case http.StatusFound:
		return coap.Changed
	case http.StatusBadRequest:
		return coap.BadRequest
	case http.StatusUnauthorized:
		return coap.Unauthorized
	case http.StatusForbidden:
		return coap.Forbidden
	case http.StatusNotFound:
		return coap.NotFound
	case http.StatusTooManyRequests:
		return coap.RequestEntityTooLarge
	case http.StatusConflict:
		return coap.BadOption
	case http.StatusInternalServerError:
		return coap.InternalServerError
	default:
		common.Debugf("Unsupported HTTP code %d", httpCode)
		return coap.InternalServerError
	}
}
