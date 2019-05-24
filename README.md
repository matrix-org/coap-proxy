# CoAP <-> HTTP proxy

## Introduction

coap-proxy is a **proof of concept experiment** for converting normal Matrix HTTPS+JSON
traffic into an ultra-low-bandwidth CoAP+CBOR+Flate+Noise+UDP network transport.
The resulting transport typically uses 35x-70x less bandwidth than HTTPS+JSON, and
attempts to fit typical Matrix transactions into a single roundtrip on a 100bps network link.

coap-proxy depends heavily on our experimental fork of [go-ocf/go-coap](https://github.com/matrix-org/go-coap),
which implements Noise-based encryption, compression hooks and retry semantics.

More details on the transport can be found in our
"Breaking the 100 bits per second barrier with Matrix" FOSDEM 2019 talk:
https://matrix.org/blog/2019/03/12/breaking-the-100bps-barrier-with-matrix-meshsim-coap-proxy/

The typical way to run coap-proxy is via docker using the [meshsim](https://github.com/matrix-org/meshsim)
network simulator, which fires up docker containers in a simulated bad network environment
containing both synapse and coap-proxy configured such that coap-proxy intercepts both client-server and
server-server traffic.

## Limitations

As a proof-of-concept, coap-proxy currently has some major limitations:

 * Encryption using [Noise](https://noiseprotocol.org) is highly experimental:
   * The transport layer security here is entirely custom and hand-wrapped
     (to optimise for bandwidth at any cost), and **should not yet be trusted
     as secure or production ready**, nor has it been audited or reviewed or fully tested.
   * We use custom CoAP message codes (250, 251 and 252) to describe Noise
     handshake payloads - see
     https://github.com/matrix-org/go-coap/blob/b6887beb140cb1cd287e2cc7c4307ebe2263db90/conn.go#L360-L370.
     Instead we should probably be using CoAP's OSCORE rather than creating our own thing;
     see https://tools.ietf.org/id/draft-mattsson-lwig-security-protocol-comparison-00.html
   * Security is trust-on-first-use (although theoretically could support pre-shared static certificates)
   * IK handshakes are currently not being used due to them making go-coap's handshake state machine unreliable,
     which needs to be investigated in the future
 * coap-proxy currently assumes it runs under a trusted private network (i.e. not the internet).  This means:
   * Signatures and checks are removed from the Matrix S2S API to save bandwidth (given the network is assumed trustworthy)
   * Minimal bandwidth depends on picking predictable compact hostnames which compress easily
     (e.g. synapse1, synapse2, synapse3...)
   * No work has yet been done on congestion control; the CoAP stack uses simple exponential backoff to retry connections.
 * We currently compress data using pre-shared static deflate compression maps.
   All nodes have to share precisely the same map files.
   * Ideally we should support streaming compression and dynamic maps.
 * CoAP doesn't support querystrings longer than 255 bytes; the proxy needs to spot these and work
   around the limit (but doesn't yet).  In practice these are quite rare however.
 * Multiple overlapping blockwise CoAP requests to the same endpoint may get entangled, and may
   require application-layer mitigation.  This is a design flaw in CoAP (see RFC7959 §2.4).
 * Encryption re-handshakes after network interruptions do not yet work.
   * After a sync timeout, the session cache is not updated with the source port of the new UDP flow,
     and so will try to send responses to the old source port.
 * No IPv6 support.
 * access_tokens currently suck up bandwidth on every request. We should swap them out for PSKs or similar.

Development of coap-proxy is dependent on commercial interest - please contact
`support at vector.im` if you're interested in a production grade coap-proxy!

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

**Note:** We've observed some build issues when building coap-proxy against Go
1.12. If the build fails, re-building against Go 1.11 should fix it:

```bash
wget https://dl.google.com/go/go1.11.linux-amd64.tar.gz # replace with the archive for your platform if necessary (see https://golang.org/dl/#go1.11.5)
tar -xvf go1.11.linux-amd64.tar.gz
rm -r pkg
GOROOT=$PWD/go gb build
```

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

Whenever it's possible, the proxy will try to reuse existing connections
(connect()ed UDP sockets) to specific destinations when sending CoAP requests.

The only reasons that can lead it to (re)create a connection to a given
destination are:

* no connection to that destination exist.
* writing to the connection returned raised an error (in which case we
  recreate the connection and retry sending the message).
* the time delta between now and when the latest message written to the
  connection was sent is higher than the 180s sync timeout.

The reason why we currently re-handshake after the 180s sync timeout
is:

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

## License

Copyright 2019 New Vector Ltd

This file is part of coap-proxy.

coap-proxy is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

coap-proxy is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with coap-proxy.  If not, see <https://www.gnu.org/licenses/>.
