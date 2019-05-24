// Copyright 2019 New Vector Ltd
//
// This file is part of coap-proxy.
//
// coap-proxy is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// coap-proxy is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with coap-proxy.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/matrix-org/coap-proxy/common"
	"github.com/matrix-org/coap-proxy/types"

	"github.com/matrix-org/go-coap"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	olog "github.com/opentracing/opentracing-go/log"

	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

// route is a struct that represents items in the routes.json file, which maps
// an endpoint to an ID for compression purposes.
type route struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Name   string `json:"name,allowempty"`
}

// openConn is a struct that represents an open CoAP connection to another
// coap-proxy instance. We keep a map of these for timeout tracking purposes.
type openConn struct {
	*coap.ClientConn
	lastMsg    time.Time
	killswitch chan bool
	dead       bool
}

// Patterns for identifying arguments in Matrix API endpoint query paths
const (
	patternRoomID        = "{roomId}"
	patternRoomAlias     = "{roomAlias}"
	patternRoomIDOrAlias = "{roomIdOrAlias}"
	patternEventID       = "{eventId}"
	patternUserID        = "{userId}"
	patternEventType     = "{eventType}"
)

var (
	// CLI flags
	onlyCoAP     = flag.Bool("only-coap", false, "Only proxy CoAP requests to HTTP and not the other way around")
	onlyHTTP     = flag.Bool("only-http", false, "Only proxy HTTP requests to CoAP and not the other way around")
	noEncryption = flag.Bool("disable-encryption", false, "Disable noise encryption")
	debugLog     = flag.Bool("debug-log", false, "Output debug logs")
	mapsDir      = flag.String("maps-dir", "maps", "Directory in which the JSON maps live")
	coapTarget   = flag.String("coap-target", "", "Force the host+port of the CoAP server to talk to")
	httpTarget   = flag.String("http-target", "http://127.0.0.1:8008", "Force the host+port of the HTTP server to talk to")
	coapPort     = flag.String("coap-port", "5683", "The CoAP port to listen on")
	httpPort     = flag.String("http-port", "8888", "The HTTP port to listen on")

	// Env flags
	jaegerHost, useJaeger = os.LookupEnv("SYNAPSE_JAEGER_HOST")

	fedAuthPrefix = "X-Matrix origin="
	fedAuthSuffix = ",key=\"\",sig=\"\""

	routePatternRgxp = regexp.MustCompile("{[^/]+}")
	fedAuthRgxp      = regexp.MustCompile(fedAuthPrefix + "origin=([^,]+)")

	// Slices to keep parsed json dictionary data in
	routes      = make([]route, 0)
	eventTypes  = make([]string, 0)
	commonKeys  = make([]string, 0)
	errorCodes  = make([]string, 0)
	queryParams = make([]string, 0)
	eduTypes    = make([]string, 0)

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

	// CBOR encoder/decoder
	cbor = new(types.CBOR)

	// JSON encoder/decoder
	json = new(types.JSON)

	// Implementation of go-coap compressor struct
	compressor *types.Compressor

	s1 = rand.NewSource(time.Now().UnixNano())
	r1 = rand.New(s1)

	// Map of open connections with the host address as the key. Allows us to keep
	// track of the last time a message was sent for timeout purposes.
	conns map[string]*openConn

	// Client for outbound HTTP requests to homeservers
	httpClient = &http.Client{}
)

func randSlice(n int) []byte {
	token := make([]byte, n)
	r1.Read(token)
	return token
}

// handleErr is a function that takes an error and an opentracing span and
// performs the necessary error handling functions such and printing relevant
// information and adding the error to the span.
func handleErr(err error, serverSpan opentracing.Span) {
	ext.Error.Set(serverSpan, true)
	serverSpan.LogFields(olog.Error(err))
	log.Println("ERROR:", err)
}

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
		w.Write(pl)
	} else {
		w.WriteHeader(int(statusCoAPToHTTP(statusCode)))
	}

	common.Debugf("CoAP server responded with code %s", statusCode.String())
	common.Debug("HTTP: Sending response")
}

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

func init() {
	flag.Parse()

	if *debugLog {
		common.EnableDebugLogging()
	}

	conns = make(map[string]*openConn)

	var err error

	// Parse maps for later compression purposes. These allow for compression
	// something like a known Matrix API endpoint route down into a single integer.
	if err = json.ParseFile(
		filepath.Join(*mapsDir, "routes.json"),
		&routes,
	); err != nil {
		panic(err)
	}

	if err = json.ParseFile(
		filepath.Join(*mapsDir, "query_params.json"),
		&queryParams,
	); err != nil {
		panic(err)
	}

	if err = json.ParseFile(
		filepath.Join(*mapsDir, "event_types.json"),
		&eventTypes,
	); err != nil {
		panic(err)
	}

	if compressor, err = types.NewCompressor(*mapsDir, []string{
		"event_types.json",
		"common_keys.json",
		"error_codes.json",
		"edu_types.json",
	}, cbor); err != nil {
		panic(err)
	}
}

func main() {
	closer := setupJaegerTracing()
	if closer != nil {
		defer closer.Close()
	}

	// Create a wait group to keep main routine alive while HTTP and CoAP servers run in separate routines
	wg := sync.WaitGroup{}
	var h *handler

	// Start CoAP listener
	// Listens for CoAP requests and sends out HTTP
	if !*onlyHTTP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coapAddr := "0.0.0.0:" + *coapPort
			log.Printf("Setting up CoAP to HTTP proxy on %s", coapAddr)
			log.Println(listenAndServe(coapAddr, "udp", coap.HandlerFunc(ServeCOAP), compressor))
			log.Println("CoAP to HTTP proxy exited")
		}()
	}

	// Start HTTP listener
	// Listens for HTTP requests and sends out CoAP
	if !*onlyCoAP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpAddr := "0.0.0.0:" + *httpPort
			log.Printf("Setting up HTTP to CoAP proxy on %s", httpAddr)
			log.Println(http.ListenAndServe(httpAddr, h))
			log.Println("HTTP to CoAP proxy exited")
		}()
	}

	wg.Wait()

	// Close all open CoAP connections on program termination
	for _, c := range conns {
		if err := c.Close(); err != nil {
			panic(err)
		}
	}
}

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

func (c *openConn) Close() error {
	c.killswitch <- true
	return c.ClientConn.Close()
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

// resetConn is a function that given a CoAP target (address and port), closes
// any existing connections to it and opens a new one.
func resetConn(target string) (*openConn, error) {
	if c, exists := conns[target]; exists {
		common.Debugf("Closing UDP connection to %s", target)
		c.Close()
	}

	common.Debugf("Creating new UDP connection to %s", target)
	return newOpenConn(target)
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

// argsAndRouteFromPath is a function that returns the routeID (encoded integer
// representing a matrix API endpoint) and any associated arguments from a given path.
// An argument being roomId in `/_matrix/client/r0/rooms/{roomId}/state` for instance.
func argsAndRouteFromPath(
	path string,
) (args []string, trailingSlash bool, routeID int, err error) {
	deconstructedPath := strings.Split(path, "/")

	var r int64
	if deconstructedPath[0] == "" {
		r, err = strconv.ParseInt(deconstructedPath[1], 32, 64)
		routeID = int(r)
		args = deconstructedPath[2:]
	} else {
		r, err = strconv.ParseInt(deconstructedPath[0], 32, 64)
		routeID = int(r)
		args = deconstructedPath[1:]
	}

	if deconstructedPath[len(deconstructedPath)-1] == "" {
		trailingSlash = true
	}

	return
}

// identifyRoute receives a route path and converts it to an identifier known by
// both proxies connected to each homeserver.
// Essentially something like /_matrix/federation/v1/send/{txnId} becomes `1`
// The proxy then sends `1` over the wire to the other proxy, and as long as
// they have the same mapping between paths and IDs, then the proxy on the other
// end knows what the correct path is.
func identifyRoute(path, method string) (routeID int, found bool) {
	patternMatcher := "[^/]+"
	for id, route := range routes {
		routeRgxpBase := route.Path
		matches := routePatternRgxp.FindAllString(routeRgxpBase, -1)
		for _, match := range matches {
			routeRgxpBase = strings.Replace(routeRgxpBase, match, patternMatcher, -1)
		}
		if regexp.MustCompile(routeRgxpBase).MatchString(path) {
			if strings.EqualFold(method, route.Method) {
				routeID = id
				found = true
				break
			}
		}
	}

	if found {
		common.Debugf("Identified route #%d", routeID)
	} else {
		common.Debugf("No route matching %s %s", strings.ToUpper(method), path)
	}

	return
}

// genExpandedPath decodes an encoded path retrieved from a CoAP request. It
// does so using a map from compressed to expanded path and query parameter
// values. This map must be the same and/or compatible on both proxies for this
// to function.
func genExpandedPath(
	srcPath string, args []string, query string, trailingSlash bool, routeID int,
) (path string, err error) {
	q, err := url.ParseQuery(query)
	if err != nil {
		return
	}

	if len(q.Encode()) > 0 {
		var buf url.Values
		buf = make(url.Values)

		for key, values := range q {
			if i, err := strconv.Atoi(key); err == nil {
				buf[queryParams[i]] = values
			} else {
				buf[key] = values
			}
		}

		q = buf
	}

	if routeID >= 0 {
		path = routes[routeID].Path

		if len(args) > 0 {
			matches := routePatternRgxp.FindAllString(path, -1)
			var arg string
			for i := 0; i < len(args) && i < len(matches); i++ {
				arg, err = getArgFromReq(matches[i], args[i])
				if err != nil {
					return
				}

				path = strings.Replace(path, matches[i], arg, -1)
			}
		}

		if trailingSlash {
			path = path + "/"
		}
	} else {
		path = srcPath
	}

	if len(q.Encode()) > 0 {
		path = path + "?" + q.Encode()
	}

	return
}

// genCompressedPath gets given a request path, attempts to compress the query
// parameters using a map, and afterwards stitches together the potentially
// compressed path and query parameters into one, which it then returns.
func genCompressedPath(uri *url.URL, routeID int) string {
	common.Debugf("Compressing %s", uri.String())

	if len(uri.RawQuery) > 1 {
		var buf url.Values
		buf = make(url.Values)

		for key, values := range uri.Query() {
			common.Debugf("Compression: Processing query param %s", key)

			index, found := queryParamsIndex(key)
			if found {
				buf[strconv.Itoa(index)] = values
			} else {
				buf[key] = values
			}
		}

		common.Debugf("Ended up with query %s", buf.Encode())

		uri.RawQuery = buf.Encode()
	}

	var path string

	if routeID >= 0 {
		deconstructedPath := strings.Split(uri.Path, "/")
		deconstructedRoute := strings.Split(routes[routeID].Path, "/")

		args := make([]string, 0)
		for i := 0; i < len(deconstructedRoute) && i < len(deconstructedPath); i++ {
			if routePatternRgxp.MatchString(deconstructedRoute[i]) {
				arg := compressReqArg(deconstructedRoute[i], deconstructedPath[i])
				args = append(args, arg)
			}
		}

		if len(args) > 0 {
			path = fmt.Sprintf("/%s/%s", strconv.FormatInt(int64(routeID), 32), strings.Join(args, "/"))
		} else {
			path = fmt.Sprintf("/%s", strconv.FormatInt(int64(routeID), 32))
		}
	} else {
		path = uri.Path
	}

	s := strings.Split(uri.String(), "?")
	if s[0][len(s[0])-1:] == "/" {
		path = path + "/"
	}

	if len(uri.RawQuery) > 0 {
		path = path + "?" + uri.Query().Encode()
	}

	common.Debugf("Ended up with compressed path %s", path)

	return path
}

// getArgFromReq is a function that retreives an argument from a request given a
// pattern type.
func getArgFromReq(match, arg string) (string, error) {
	switch match {
	case patternEventType:
		typeID, err := strconv.Atoi(arg)
		if err == nil {
			arg = eventTypes[typeID]
		}
	case patternRoomID, patternEventID, patternRoomAlias, patternUserID, patternRoomIDOrAlias:
		arg = getSigil(match) + arg
	default:
	}

	return url.PathEscape(arg), nil
}

// getSigil is a function that returns a sigil for a given pattern type
func getSigil(pattern string) string {
	switch pattern {
	case patternRoomID:
		return "!"
	case patternRoomAlias:
		return "#"
	case patternEventID:
		return "$"
	case patternUserID:
		return "@"
	default:
		return ""
	}
}

// removeSigil is a function that removes a sigil from a string given its pattern
func removeSigil(pattern, arg string) string {
	if pattern == patternEventID || pattern == patternRoomID || pattern == patternUserID || pattern == patternRoomAlias {
		if arg[0] == '%' {
			return arg[3:]
		}

		return arg[1:]
	}

	return arg
}

// compressReqArg is a function that compresses a request argument using its
// corresponding pattern type
func compressReqArg(pattern, arg string) string {
	oldVal := arg

	switch pattern {
	case patternEventType:
		index, found := eventTypeIndex(arg)
		if found {
			arg = strconv.Itoa(index)
		}
	case patternRoomID, patternEventID, patternRoomAlias, patternUserID, patternRoomIDOrAlias:
		arg = removeSigil(pattern, arg)
	default:
	}

	common.Debugf("Compressing special arg %s (value: %s) into %s", pattern, oldVal, arg)

	return arg
}

// eventTypeIndex is a function that encodes an event type to an integer
// integer using the eventTypes map.
// Found is false if encoding was not possible, otherwise true.
func eventTypeIndex(t string) (index int, found bool) {
	for i, eventType := range eventTypes {
		if strings.EqualFold(t, eventType) {
			index = i
			found = true
			break
		}
	}

	return
}

// queryParamsIndex is a function that encodes a query parameter key as an
// integer using the queryParams map.
// Found is false if encoding was not possible, otherwise true.
func queryParamsIndex(key string) (index int, found bool) {
	for i, qp := range queryParams {
		if key == qp {
			index = i
			found = true
			break
		}
	}

	return
}

// matrixErrorIndex is a function that encodes a known matrix error (e.g.
// M_UNKNOWN) as an integer using the errorCodes map.
// Found is false if encoding was not possible, otherwise true.
func matrixErrorIndex(errCode string) (index int, found bool) {
	for i, code := range errorCodes {
		if strings.EqualFold(errCode, code) {
			index = i
			found = true
			break
		}
	}

	return
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
	opentracing.GlobalTracer().Inject(
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
		opentracing.GlobalTracer().Inject(
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

// setupJaegerTracing is a function that sets up OpenTracing with a
// configuration specific to the CoAP proxy.
func setupJaegerTracing() io.Closer {
	jaegerHost := jaegerHost
	serverName := os.Getenv("SYNAPSE_SERVER_NAME")

	serviceName := fmt.Sprintf("proxy-%s", serverName)

	var cfg jaegercfg.Configuration

	if useJaeger {
		cfg = jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:           true,
				LocalAgentHostPort: fmt.Sprintf("%s:6831", jaegerHost),
			},
			ServiceName: serviceName,
		}
	} else {
		cfg = jaegercfg.Configuration{}
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return nil
	}

	return closer
}
