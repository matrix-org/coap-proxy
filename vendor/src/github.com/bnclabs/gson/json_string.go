// Package json implements encoding and decoding of JSON objects as defined in
// RFC 4627. The mapping between JSON objects and Go values is described
// in the documentation for the Marshal and Unmarshal functions.
//
// See "JSON and Go" for an introduction to this package:
// http://golang.org/doc/articles/json_and_go.html

package gson

import "unicode/utf8"

var hex = "0123456789abcdef"

// Copied from golang src/pkg/encoding/json/encode.go
// Modified to use []byte as input and use DecodeRune() instead of
// DecodeRuneInString.
func encodeString(s, text []byte) ([]byte, error) {
	text = append(text, '"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue
			}
			if start < i {
				text = append(text, s[start:i]...)
			}
			switch b {
			case '\\', '"':
				text = append(text, '\\')
				text = append(text, b)
			case '\n':
				text = append(text, '\\')
				text = append(text, 'n')
			case '\r':
				text = append(text, '\\')
				text = append(text, 'r')
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as <, > and &. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into JSON and served to some browsers.

				text = append(text, `\u00`...)
				text = append(text, hex[b>>4])
				text = append(text, hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRune(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				text = append(text, s[start:i]...)
			}
			text = append(text, `\ufffd`...)
			i += size
			start = i
			continue
		}
		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				text = append(text, s[start:i]...)
			}
			text = append(text, `\u202`...)
			text = append(text, hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		text = append(text, s[start:]...)
	}
	text = append(text, '"')
	return text, nil
}
