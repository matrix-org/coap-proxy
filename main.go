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
	"errors"
	"flag"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sync"

	"github.com/matrix-org/coap-proxy/common"
	"github.com/matrix-org/coap-proxy/types"

	"github.com/matrix-org/go-coap"
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
	coapBindHost = flag.String("coap-bind-host", "0.0.0.0", "The COAP host to listen on")
	httpPort     = flag.String("http-port", "8888", "The HTTP port to listen on")

	fedAuthPrefix = "X-Matrix origin="
	fedAuthSuffix = ",key=\"\",sig=\"\""

	routePatternRgxp = regexp.MustCompile("{[^/]+}")
	fedAuthRgxp      = regexp.MustCompile(fedAuthPrefix + "([^,]+)")

	// Slices to keep parsed json dictionary data in
	routes      = make([]route, 0)
	eventTypes  = make([]string, 0)
	errorCodes  = make([]string, 0)
	queryParams = make([]string, 0)

	// CBOR encoder/decoder
	cbor = new(types.CBOR)

	// JSON encoder/decoder
	json = new(types.JSON)

	// Implementation of go-coap compressor struct
	compressor *types.Compressor
)

func init() {
	log.Printf("Starting up...")

	flag.Parse()

	if *debugLog {
		common.EnableDebugLogging()
	}

	conns = make(map[string]*sync.Pool)

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

	log.Println("Finished loading compression maps")
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
			coapAddr := *coapBindHost + ":" + *coapPort
			log.Printf("Setting up CoAP to HTTP proxy on %s", coapAddr)
			log.Println(listenAndServe(coapAddr, "udp", coapRecoverWrap(coap.HandlerFunc(ServeCOAP)), compressor))
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
			log.Println(http.ListenAndServe(httpAddr, httpRecoverWrap(h)))
			log.Println("HTTP to CoAP proxy exited")
		}()
	}

	wg.Wait()

	// Close all open CoAP connections on program termination
	// for _, c := range conns {
	// 	if err := c.Close(); err != nil {
	// 		panic(err)
	// 	}
	// }
}

func httpRecoverWrap(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Printf("Recovered from panic: %v", err)
				log.Println("Stacktrace:\n" + string(debug.Stack()))
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func coapRecoverWrap(h coap.Handler) coap.Handler {
	return coap.HandlerFunc(func(w coap.ResponseWriter, r *coap.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				log.Printf("Recovered from panic: %v", err)
				log.Println("Stacktrace:\n" + string(debug.Stack()))
			}
		}()
		h.ServeCOAP(w, r)
	})
}
