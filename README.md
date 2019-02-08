# CoAP <-> HTTP proxy

## Build

You'll need [gb](https://github.com/constabulary/gb) to build this project. You
can install it with:

```
go get github.com/constabulary/gb/...
```

Then run `gb build`

You can also use the `build-release.sh` script (which also requires `gb` to be
installed) which provides a lighter binary and strips part of the hard-coded
paths from it so it doesn't leak the project's full path in stack traces.

## Run

Once built, the proxy's binary is located in `bin/coap-proxy`. You can give it
the following flags:

* `--only-coap`: Only proxy CoAP requests to HTTP and not the other way around.
* `--only-http`: Only proxy HTTP requests to CoAP and not the other way around.
* `--debug-log`: Print debugging logs. Coupling it with setting the environment
  variable `PROXY_DUMP_PAYLOADS` to any value will also dump requests and
  responses payloads along with their transformations during the process.
* `--maps-dir DIR`: Tell the proxy to look for map files in `DIR`. Defaults to
  `./maps`.
* `--coap-target`: Tell the proxy where to send CoAP requests.

If no flag is provided, the proxy <!-- will use CBOR for every CoAP request, and  -->will
listen for both HTTP and CoAP requests on port:

* `8888` for HTTP
* `5683` for CoAP

## How it works

The proxy works as follows:

```
---------------            ----------------            ----------------            ---------------
| HTTP client | <--------> | HTTP to CoAP | <--------> | CoAP to HTTP | <--------> | HTTP server |
|             |    HTTP    |     proxy    |   CoAP +   |     proxy    |    HTTP    |             |
---------------            ----------------    CBOR    ----------------            ---------------

                           |__________________________________________|
                                                 |
                                             coap-proxy
```

## Run the proxy for meshsim

* Build the proxy
* Run it by telling it to talk to the HS's proxy:

```bash
./bin/coap-proxy --coap-target localhost:20001 # Ports are 20000 + hsid
```

* Make clients talk to http://localhost:8888
* => profit

## Connections

Whenever it's possible, the proxy will try to reuse existing connections to
specific destinations when sending CoAP requests.

The only reasons that can lead it to (re)create a connection to a given
destination are:

* no connection to that destination exist.
* writing to the connection returned raised an error (in which case we
  recreate the connection and retry sending the message).
* the time delta between now and when the latest message written to the
  connection was sent is higher than the 180s sync timeout.

The third reason triggering a re-creation of the connection (therefore another
handshake) can lead to discussions, here are the thought behind it:

1. the server (and client) track sessions in networksession, which keys them
   based on a given {srcport, dstport} tuple.
2. the client, having connect()ed its socket, will not change src port.
3. the server always has the same dest port (5683).
4. the traffic may be NATed however, which means that after a ~180s gap in
   traffic, the srcport that the server sees may change, making the server
   see it as a new session (unless we do something to tie sessions to CoAP
   IDs rather than srcport/dstport tuples)
5. therefore the behaviour here probably *is* right (for now) and we
   should handshake after 180s of any pause of traffic.
