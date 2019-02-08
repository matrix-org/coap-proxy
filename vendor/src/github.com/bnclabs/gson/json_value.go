// transform json encoded data into golang native value.
// cnf: SpaceKind, NumberKind, strict

package gson

import "strconv"

// primary interface to scan JSON text and return,
// a. text remaining to be parsed.
// b. as go-native value.
// calling this function will scan for exactly one JSON value
func json2value(txt string, config *Config) (string, interface{}) {
	txt = skipWS(txt, config.ws)

	if len(txt) < 1 {
		panic("gson scanner jsonEmpty")
	}

	if digitCheck[txt[0]] == 1 {
		return jsonnum2value(txt, config)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == "null" {
			return txt[4:], nil
		}
		panic("gson scanner expectedNil")

	case 't':
		if len(txt) >= 4 && txt[:4] == "true" {
			return txt[4:], true
		}
		panic("gson scanner expectedTrue")

	case 'f':
		if len(txt) >= 5 && txt[:5] == "false" {
			return txt[5:], false
		}
		panic("gson scanner expectedFalse")

	case '"':
		bufn := config.bufferh.getbuffer(len(txt) * 5)
		scratch := bufn.data
		remtxt, n := scanString(txt, scratch)
		value := string(scratch[:n]) // this will copy the content.
		config.bufferh.putbuffer(bufn)
		return remtxt, value

	case '[':
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("gson scanner expectedCloseArray")
		} else if txt[0] == ']' {
			return txt[1:], []interface{}{}
		}
		arr := make([]interface{}, 0, 4)
		for {
			var tok interface{}
			txt, tok = json2value(txt, config)
			arr = append(arr, tok)
			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("gson scanner expectedCloseArray")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == ']' {
				break
			} else {
				panic("gson scanner expectedCloseArray")
			}
		}
		return txt[1:], arr

	case '{':
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("gson scanner expectedCloseobject")
		} else if txt[0] == '}' {
			return txt[1:], map[string]interface{}{}
		} else if txt[0] != '"' {
			panic("gson scanner expectedKey")
		}

		var tok interface{}
		var n int

		m := make(map[string]interface{})
		bufn := config.bufferh.getbuffer(len(txt) * 5)
		scratch := bufn.data
		for {
			txt, n = scanString(txt, scratch) // empty string is also valid key
			key := string(scratch[:n])

			if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
				panic("gson scanner expectedColon")
			}
			txt, tok = json2value(skipWS(txt[1:], config.ws), config)
			m[key] = tok
			if txt = skipWS(txt, config.ws); len(txt) == 0 {
				panic("gson scanner expectedCloseobject")
			} else if txt[0] == ',' {
				txt = skipWS(txt[1:], config.ws)
			} else if txt[0] == '}' {
				break
			} else {
				panic("gson scanner expectedCloseobject")
			}
		}
		config.bufferh.putbuffer(bufn)
		return txt[1:], m
	}
	panic("gson scanner expectedToken")
}

func jsonnum2value(txt string, config *Config) (string, interface{}) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}

	switch config.nk {
	case FloatNumber:
		f, err := strconv.ParseFloat(txt[s:e], 64)
		if err != nil {
			panic(err)
		}
		return txt[e:], f

	case SmartNumber:
		if i, err := strconv.ParseInt(txt[s:e], 10, 64); err == nil {
			return txt[e:], i
		} else if ui, err := strconv.ParseUint(txt[s:e], 10, 64); err == nil {
			return txt[e:], ui
		}
		f, err := strconv.ParseFloat(txt[s:e], 64)
		if err != nil {
			panic(err)
		}
		return txt[e:], f
	}
	panic("unreachable code")
}
