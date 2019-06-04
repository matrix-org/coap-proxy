package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/matrix-org/coap-proxy/common"
)

const (
	matrixClientPrefix     = "/_matrix/client"
	matrixMediaPrefix      = "/_matrix/media"
	matrixFederationPrefix = "/_matrix/federation"
)

// route is a struct that represents items in the routes.json file, which maps
// an endpoint to an ID for compression purposes.
type route struct {
	Path   string `json:"path"`
	Method string `json:"method"`
	Name   string `json:"name,allowempty"`
}

// argsAndRouteFromPath is a function that returns the routeID (encoded integer
// representing a matrix API endpoint) and any associated arguments from a given path.
// An argument being roomId in `/_matrix/client/r0/rooms/{roomId}/state` for instance.
func argsAndRouteFromPath(
	path string,
) (args []string, trailingSlash bool, routeID int, err error) {
	deconstructedPath := strings.Split(path, "/")

	common.Debugf("deconstructedPath %v", deconstructedPath)

	var r int64
	if len(deconstructedPath) == 0 {
		err = errors.New("Got empty path")
		return
	}

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
	patternMatcher := "[^/]*"
	for id, route := range routes {
		routeRgxpBase := "^" + route.Path + "$"
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

		if trailingSlash && path[len(path)-1] != '/' {
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

func isClientRoute(path string) bool {
	return (strings.HasPrefix(path, matrixClientPrefix) || strings.HasPrefix(path, matrixMediaPrefix))
}
