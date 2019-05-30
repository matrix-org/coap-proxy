package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/matrix-org/coap-proxy/common"
	"github.com/matrix-org/coap-proxy/types"

	"github.com/matrix-org/go-coap"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"
)

var (
	// In-memory store for crypto keys which implements the go-coap.KeyStore
	// interface.
	keyStore = types.NewKeyStore()
	// Instance of go-coap.RetriesQueue we'll give to our server and clients so
	// they can handle retries. This needs to be dont that way and at the
	// application layer since the server needs to match responses to requests
	// the client sends.
	// 40 seconds as an initial delay should be enough to prevent sync responses
	// from being sent twice (since Riot's timeout for syncs is 30s).
	retriesQueue = coap.NewRetriesQueue(40*time.Second, 1)
)

// ServeCOAP is a function that listens for CoAP requests and responds accordingly.
// It:
//   * Takes in a CoAP request
//   * Decompresses and CBOR decodes the payload if there is one
//   * Decompresses the request path and query parameters
//   * Creates an HTTP request with carried over and decompressed headers, path, body etc.
//   * Sends the HTTP request to an attached Homeserver, retrieves the response
//   * Compresses the response
//   * Returns it over CoAP to the requester
func ServeCOAP(w coap.ResponseWriter, req *coap.Request) {
	ctx := context.Background()

	m := req.Msg

	if !m.IsConfirmable() {
		log.Printf("Got unconfirmable message")
		return
	}

	pl := m.Payload()

	var serverSpan opentracing.Span
	var body interface{}
	if len(pl) > 0 {
		// Decompress and decode the payload body if it exists
		var err error
		if pl, err = compressor.DecompressPayload(pl); err != nil {
			handleErr(err, serverSpan)
			return
		}
		body = cbor.Decode(pl)

		var carrier interface{}
		if bodyMap, ok := body.(map[interface{}]interface{}); ok {
			carrier = bodyMap["XJG"]
			delete(bodyMap, "XJG")
		}

		if carrierBytes, ok := carrier.([]byte); ok {
			common.Debugf("XFG header received len %d", len(carrierBytes))
			wireContext, err := opentracing.GlobalTracer().Extract(
				opentracing.Binary,
				bytes.NewReader(carrierBytes),
			)

			if err != nil {
				common.Debugf("Failed to extract wire context %v", err)
			}

			serverSpan = opentracing.StartSpan(
				"coap-server",
				ext.RPCServerOption(wireContext),
			)

			defer serverSpan.Finish()

			ctx = opentracing.ContextWithSpan(ctx, serverSpan)
		}
	}

	if serverSpan == nil {
		serverSpan, ctx = opentracing.StartSpanFromContext(ctx, "coap-server")
		defer serverSpan.Finish()

		ext.SpanKindRPCServer.Set(serverSpan)
	}

	path := m.PathString()

	common.Debugf("CoAP - %X: Got request on path %s", req.Msg.Token(), path)

	var query string
	s := strings.Split(path, "?")
	if len(s) > 1 {
		query = s[1]
		path = s[0]
	} else {
		query = ""
	}

	// Get routeID and path arguments from request path
	args, trailingSlash, routeID, err := argsAndRouteFromPath(path)

	// Get decompressed path and query parameters from the request
	var method, routeName string
	if err == nil {
		r := routes[routeID]

		common.Debugf(
			"CoAP - %X: Got request on route #%d (%s %s)\n",
			req.Msg.Token(), routeID, strings.ToUpper(r.Method), r.Path,
		)

		path, err = genExpandedPath(path, args, query, trailingSlash, routeID)
		if err != nil {
			handleErr(err, serverSpan)
			return
		}

		common.Debugf("CoAP - %X: Got request on URL %s", req.Msg.Token(), path)

		method = r.Method
		routeName = r.Name
	} else {
		c := req.Msg.Code()
		if c >= coap.GET && c <= coap.DELETE {
			method = c.String()
			if path[:1] != "/" {
				path = "/" + path
			}

			path, err = genExpandedPath(path, args, query, trailingSlash, -1)
			if err != nil {
				handleErr(err, serverSpan)
				return
			}
		} else {
			err = errors.New("Wrong method code: " + c.String())
			handleErr(err, serverSpan)
			return
		}

		common.Debugf(
			"CoAP - %X: Got request on unknown route %s %s",
			req.Msg.Token(), method, path,
		)
	}

	common.Debugf("CoAP - %X: Sending HTTP request", req.Msg.Token())
	common.Debug("routeName", routeName)

	// Encode the CBOR-decoded body into JSON
	if len(pl) > 0 {
		if routeName == "send_transaction" {
			body = compressor.DecompressTransaction(body)
		}
		pl = json.Encode(body)
	}

	// XXX: HACK: Use the coap library's MaxAge option to store origin (due to the
	// lib ignoring unknown options). Do not rely on this to remain functional across
	// lib upgrades.
	originOpt := m.Option(coap.MaxAge)
	common.Debugf("opt: %v", originOpt)

	// TODO: Handle fs-specific stuff
	var origin string
	if s, ok := originOpt.(string); ok {
		origin = s
	}

	ext.PeerHostname.Set(serverSpan, origin)

	// Send an HTTP request to a homeserver and receive a response
	pl, statusCode, err := sendHTTPRequest(
		ctx,
		method,
		path,
		pl,
		origin,
	)
	if err != nil {
		handleErr(err, serverSpan)
		return
	}

	common.Debugf("CoAP - %X: Got status %d", req.Msg.Token(), statusCode)
	common.Debugf("CoAP - %X: Sending response", req.Msg.Token())

	// Convert the receive HTTP status code to a CoAP one and add to response
	w.SetCode(statusHTTPToCoAP(statusCode))

	// Re-encode the JSON body into CBOR and write out
	if len(pl) > 0 {
		pl = cbor.Encode(json.Decode(pl))

		if pl, err = compressor.CompressPayload(pl); err != nil {
			handleErr(err, serverSpan)
			return
		}

		w.SetContentFormat(coap.AppOctets)

		if _, err = w.Write(pl); err != nil {
			handleErr(err, serverSpan)
			return
		}
	}
}

// sendCoAPRequest is a function that sends a CoAP request to another instance
// of the CoAP proxy.
func sendCoAPRequest(
	ctx context.Context, method, host, path string, routeName string, body interface{},
	origin *string,
) (payload []byte, statusCode coap.COAPCode, err error) {
	var c *openConn
	var exists bool

	// Setup OpenTracing
	var clientSpan opentracing.Span
	clientSpan, ctx = opentracing.StartSpanFromContext(ctx, "coap-client")
	defer clientSpan.Finish()

	ext.SpanKindRPCClient.Set(clientSpan)

	clientSpan.SetTag("coap.path", path)
	clientSpan.SetTag("coap.method", method)

	// Send to request's host unless the target has been forced
	var target string
	if len(*coapTarget) > 0 {
		target = *coapTarget
	} else {
		target = host + ":" + *coapPort
	}

	common.Debugf("Proxying request to %s", target)

	// if c, err = dialTimeout("udp", target, 300*time.Second); err != nil {
	// 	ext.Error.Set(clientSpan, true)
	// 	clientSpan.LogFields(olog.Error(err))
	// 	return
	// }
	//
	// defer c.Close()

	// If there is an existing connection, use it, otherwise provision a new one
	if c, exists = conns[target]; !exists || (c != nil && c.dead) {
		if c, err = resetConn(target); err != nil {
			return
		}
		println(c)
		// } else if time.Now().Add(-180 * time.Second).After(c.lastMsg) {
		// 	// Reset an existing connection if the latest message sent is older than
		// 	// go-coap's syncTimeout.
		// 	if c, err = resetConn(target); err != nil {
		// 		return
		// 	}
	}

	// Record the destination in the trace
	hostAddr := strings.Split(target, ":")[0]
	ext.PeerHostname.Set(clientSpan, hostAddr)
	ext.PeerAddress.Set(clientSpan, target)

	if useJaeger {
		carrier := &bytes.Buffer{}
		_ = opentracing.GlobalTracer().Inject(
			clientSpan.Context(),
			opentracing.Binary,
			carrier,
		)

		// Add a tracing header to the request
		if bodyMap, ok := body.(map[interface{}]interface{}); ok {
			carrierBytes := carrier.Bytes()

			common.Debugf("XFG header len: %d", len(carrierBytes))

			clientSpan.LogFields(olog.Int("jaeger-bytes", len(carrierBytes)))
			bodyMap["XJG"] = carrierBytes
		}

	}

	// Map for translating HTTP method codes to CoAP
	methodCodes := map[string]coap.COAPCode{
		"GET":    coap.GET,
		"POST":   coap.POST,
		"PUT":    coap.PUT,
		"DELETE": coap.DELETE,
	}

	var bodyBytes []byte
	if body != nil {
		// Compress transaction if this a federation transaction request
		if routeName == "send_transaction" {
			body = compressor.CompressTransaction(body)
			common.DumpPayload("Encoded transaction", body)
		}

		// Encode body as CBOR
		bodyBytes = cbor.Encode(body)

		common.DumpPayload("Encoded body", bodyBytes)

		// Compress body
		if bodyBytes, err = compressor.CompressPayload(bodyBytes); err != nil {
			ext.Error.Set(clientSpan, true)
			clientSpan.LogFields(olog.Error(err))
			return
		}
	}

	common.Debugf("Sending %d bytes in compressed payload", len(bodyBytes))
	clientSpan.LogFields(olog.Int("payload-bytes", len(bodyBytes)))

	// Create a new CoAP request
	req := c.NewMessage(coap.MessageParams{
		Type:      coap.Confirmable,
		Code:      methodCodes[strings.ToUpper(method)],
		MessageID: uint16(r1.Intn(100000)),
		Token:     randSlice(2),
		Payload:   bodyBytes,
	})

	if origin != nil {
		// XXX: HACK: Use the coap library's MaxAge option to store origin (due to the
		// lib ignoring unknown options). Do not rely on this to remain functional across
		// lib upgrades.
		req.SetOption(coap.MaxAge, *origin)
	}

	// This option should be set last, to aid compression of the packet
	req.SetOption(coap.ContentFormat, coap.AppOctets)
	req.SetPathString(path)

	log.Printf("HTTP: Sending CoAP request with token %X", req.Token())

	// Send the CoAP request and receive a response
	common.Debugf("opts %v", req.AllOptions())
	res, err := c.Exchange(req)

	// Check for errors
	if err != nil {
		log.Printf("Closing CoAP connection because of error: %v", err)

		if c, err = resetConn(target); err != nil {
			return
		}

		if res, err = c.Exchange(req); err != nil {
			ext.Error.Set(clientSpan, true)
			clientSpan.LogFields(olog.Error(err))
			log.Printf("HTTP failed to exchange coap: %v", err)
			return
		}
	}

	// Receive and decompress the response payload
	rawPayload := res.Payload()
	clientSpan.LogFields(olog.Int("response-payload-bytes", len(rawPayload)))

	common.Debugf("HTTP: Got response to CoAP request %X with %d bytes in response payload", res.Token(), len(rawPayload))

	pl, err := compressor.DecompressPayload(rawPayload)
	// common.Debugf("Got %d bytes in response payload (%d decompressed)", len(rawPayload), len(pl))

	// Keep track of the last successfully received message for connection timeout purposes
	c.lastMsg = time.Now()

	return pl, res.Code(), err
}
