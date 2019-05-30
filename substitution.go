package main

import "strings"

// Patterns for identifying arguments in Matrix API endpoint query paths
const (
	patternRoomID        = "{roomId}"
	patternRoomAlias     = "{roomAlias}"
	patternRoomIDOrAlias = "{roomIdOrAlias}"
	patternEventID       = "{eventId}"
	patternUserID        = "{userId}"
	patternEventType     = "{eventType}"
)

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
