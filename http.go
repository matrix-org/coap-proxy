package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/matrix-org/coap-proxy/common"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"
)

// Client for outbound HTTP requests to homeservers
var httpClient = &http.Client{}

// handler is a struct that acts as an http.Handler, where its ServeHTTP method is used to handle HTTP requests
type handler struct{}

// ServeHTTP is a function implemented on handler which handles HTTP requests.
// It:
//   * Takes in an HTTP request
//   * Compresses the path, query parameters and body if possible
//   * Creates a CoAP request with carried over and compressed headers, path, body etc.
//   * Sends the CoAP request to another proxy, retrieves the response
//   * Decompresses the response back into normal HTTP
//   * Returns it to the original sender
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		common.Debug("Got preflight request")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")
		return
	}

	common.Debugf("HTTP: Got request on path %s", r.URL.Path)

	// Set up an OpenTracing span to track this request's lifecycle
	var serverSpan opentracing.Span
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))

	serverSpan = opentracing.StartSpan(
		"http_server",
		ext.RPCServerOption(wireContext))

	defer serverSpan.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), serverSpan)

	ext.HTTPMethod.Set(serverSpan, r.Method)
	ext.HTTPUrl.Set(serverSpan, r.URL.Path)

	// Convert the path and HTTP method into an identifier represented by a single
	// integer (for compression purposes)
	routeID, foundRoute := identifyRoute(r.URL.Path, r.Method)

	var path, method, routeName string
	if foundRoute {
		common.Debugf(
			"HTTP: Got request on route #%d (%s %s)\n",
			routeID, strings.ToUpper(routes[routeID].Method),
			routes[routeID].Path,
		)

		// Generate a compressed path using the found routeID
		path = genCompressedPath(r.URL, routeID)
		method = strings.ToUpper(routes[routeID].Method)
		routeName = routes[routeID].Name
	} else {
		common.Debugf(
			"HTTP: Got request on unknown route %s %s\n",
			r.Method,
			r.URL.Path,
		)

		// Generate a compressed path
		path = genCompressedPath(r.URL, -1)
		method = r.Method
	}

	serverSpan.SetTag("route.name", routeName)
	common.Debug("routeName", routeName)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleErr(err, serverSpan)
		return
	}

	// Unmarshal request body JSON
	var decodedBody interface{}
	if len(body) > 0 {
		decodedBody = json.Decode(body)
	}

	// Add authentication header to query parameters of CoAP request
	var origin *string
	if authHeader := r.Header.Get("Authorization"); len(authHeader) > 0 {
		var k, v string
		// Assume that the CoAP target will only be forced for the CS API
		// TODO: More flexibility
		if len(*coapTarget) > 0 {
			k = "access_token"
			v = strings.Replace(authHeader, "Bearer ", "", 1)

			var sep string
			if strings.Contains(path, "?") {
				sep = "&"
			} else {
				sep = "?"
			}

			path = path + sep + k + "=" + v
		} else {
			submatch := fedAuthRgxp.FindAllStringSubmatch(authHeader, 1)[0][0]
			origin = &submatch
		}
	}

	if len(path) == 0 {
		path = "/"
	}

	common.Debugf("Final path: %s", path)

	// Send the CoAP request to another instance of the CoAP proxy and receive a response
	pl, statusCode, err := sendCoAPRequest(ctx, method, r.Host, path, routeName, decodedBody, origin)
	if err != nil {
		handleErr(err, serverSpan)
		return
	}

	ext.HTTPStatusCode.Set(serverSpan, statusCoAPToHTTP(statusCode))

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET,PUT,DELETE,OPTIONS")

	// CoAP requests use CBOR as their encoding scheme. Decode CBOR and encode back
	// into JSON (if this response has a body)
	if len(pl) > 0 {
		pl = json.Encode(cbor.Decode(pl))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(int(statusCoAPToHTTP(statusCode)))
		if _, err = w.Write(pl); err != nil {
			log.Printf("Failed to write HTTP response: %s", err.Error())
		}
	} else {
		w.WriteHeader(int(statusCoAPToHTTP(statusCode)))
	}

	common.Debugf("CoAP server responded with code %s", statusCode.String())
	common.Debug("HTTP: Sending response")
}

// sendHTTPRequest is a function that sends an HTTP request to a homeserver
// either from a client or another homeserver in the case of federation.
func sendHTTPRequest(
	ctx context.Context, method string, path string, payload []byte, origin string,
) (resBody []byte, statusCode int, err error) {
	// OpenTracing setup
	span, ctx := opentracing.StartSpanFromContext(ctx, "http_request")
	defer span.Finish()

	ext.SpanKindRPCClient.Set(span)

	// Create the request
	url := fmt.Sprintf("%s%s", *httpTarget, path)
	hReq, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewReader(payload))
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(olog.Error(err))
		return
	}

	// Set headers
	hReq.Header.Add("Content-Type", "application/json")
	if len(origin) > 0 {
		hReq.Header.Add("Authorization", fedAuthPrefix+origin+fedAuthSuffix)
	}

	// Record request details in OpenTracing
	ext.HTTPMethod.Set(span, method)
	ext.HTTPUrl.Set(span, path)
	_ = opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(hReq.Header),
	)

	// Perform the request and receive the response
	hRes, err := httpClient.Do(hReq)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(olog.Error(err))
		return
	}

	// Receive the response body
	resBody, err = ioutil.ReadAll(hRes.Body)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogFields(olog.Error(err))
		return
	}

	// Record response status code in OpenTracing
	ext.HTTPStatusCode.Set(span, uint16(hRes.StatusCode))

	statusCode = hRes.StatusCode

	return
}
