package gson

import "unicode"
import "unicode/utf8"
import "unicode/utf16"
import "strconv"

var spaceCode = [256]byte{
	'\t': 1,
	'\n': 1,
	'\v': 1,
	'\f': 1,
	'\r': 1,
	' ':  1,
}

func skipWS(txt string, ws SpaceKind) string {
	switch ws {
	case UnicodeSpace:
		for i, ch := range txt {
			if unicode.IsSpace(ch) {
				continue
			}
			return txt[i:]
		}
		return ""

	case AnsiSpace:
		i := 0
		for i < len(txt) && spaceCode[txt[i]] == 1 {
			i++
		}
		txt = txt[i:]
	}
	return txt
}

var escapeCode = [256]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'\'': '\'',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}

func scanString(txt string, out []byte) (string, int) {
	if len(txt) < 2 {
		panic("scanner expectedString")
	}

	e := 1
	for txt[e] != '"' {
		c := txt[e]
		if c == '\\' || c == '"' || c < ' ' {
			break
		}
		if c < utf8.RuneSelf {
			e++
			continue
		}
		r, size := utf8.DecodeRuneInString(txt[e:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		e += size
		if e == len(txt) {
			panic("scanner expectedString")
		}
	}

	if txt[e] == '"' { // done we have nothing to unquote
		return txt[e+1:], copy(out, txt[1:e])
	}

	oute := copy(out, txt[1:e]) // copy so far

loop:
	for e < len(txt) {
		switch c := txt[e]; {
		case c == '"':
			out[oute] = c
			e++
			break loop

		case c == '\\':
			if txt[e+1] == 'u' {
				r := getu4(txt[e:])
				if r < 0 { // invalid
					panic("scanner expectedString")
				}
				e += 6
				if utf16.IsSurrogate(r) {
					nextr := getu4(txt[e:])
					dec := utf16.DecodeRune(r, nextr)
					if dec != unicode.ReplacementChar { // A valid pair consume
						oute += utf8.EncodeRune(out[oute:], dec)
						e += 6
						break loop
					}
					// Invalid surrogate; fall back to replacement rune.
					r = unicode.ReplacementChar
				}
				oute += utf8.EncodeRune(out[oute:], r)

			} else { // escaped with " \ / ' b f n r t
				out[oute] = escapeCode[txt[e+1]]
				e += 2
				oute++
			}

		case c < ' ': // control character is invalid
			panic("scanner expectedString")

		case c < utf8.RuneSelf: // ASCII
			out[oute] = c
			oute++
			e++

		default: // coerce to well-formed UTF-8
			r, size := utf8.DecodeRuneInString(txt[e:])
			e += size
			oute += utf8.EncodeRune(out[oute:], r)
		}
	}

	if out[oute] == '"' {
		return txt[e:], oute
	}
	panic("scanner expectedString")
}

// getu4 decodes \uXXXX from the beginning of s, returning the hex value,
// or it returns -1.
func getu4(s string) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}
	r, err := strconv.ParseUint(s[2:6], 16, 64)
	if err != nil {
		return -1
	}
	return rune(r)
}

var intCheck = [256]byte{}
var digitCheck = [256]byte{}
var numCheck = [256]byte{}
var fltCheck = [256]byte{}

func init() {
	for i := 48; i <= 57; i++ {
		intCheck[i] = 1
		numCheck[i] = 1
	}
	intCheck['-'] = 1
	intCheck['+'] = 1
	intCheck['.'] = 1
	intCheck['e'] = 1
	intCheck['E'] = 1

	numCheck['-'] = 1
	numCheck['+'] = 1
	numCheck['.'] = 1

	fltCheck['.'] = 1
	fltCheck['e'] = 1
	fltCheck['E'] = 1

	for i := 48; i <= 57; i++ {
		digitCheck[i] = 1
	}
	digitCheck['-'] = 1
	digitCheck['+'] = 1
	digitCheck['.'] = 1
}
