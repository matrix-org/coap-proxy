package gson

// transform json encoded value into cbor encoded value.
// cnf: SpaceKind, NumberKind, ContainerEncoding, strict

import "strconv"

var nullStr = "null"
var trueStr = "true"
var falseStr = "false"

func json2cbor(txt string, out []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)

	if len(txt) < 1 {
		panic("cbor scanner empty json text")
	}

	if numCheck[txt[0]] == 1 {
		return jsonNumToCbor(txt, out, config)
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == nullStr {
			n := cborNull(out)
			return txt[4:], n
		}
		panic("cbor scanner expected null")

	case 't':
		if len(txt) >= 4 && txt[:4] == trueStr {
			n := cborTrue(out)
			return txt[4:], n
		}
		panic("cbor scanner expected true")

	case 'f':
		if len(txt) >= 5 && txt[:5] == falseStr {
			n := cborFalse(out)
			return txt[5:], n
		}
		panic("cbor scanner expected false")

	case '"':
		n := 0
		txt, x := scanString(txt, out[n+16:]) // 16 reserved for cbor hdr
		n += valtext2cbor(bytes2str(out[n+16:n+16+x]), out[n:])
		return txt, n

	case '[':
		n, m, nn, nnn := 0, 0, 0, 0
		switch config.ct {
		case LengthPrefix:
			nn, nnn = n+32, n+32
		case Stream:
			nnn += arrayStart(out[nnn:])
		}

		var ln int
		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("cbor scanner expected ']'")
		} else if txt[0] != ']' {
			for {
				txt, m = json2cbor(txt, out[nnn:], config)
				nnn += m
				ln++
				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("cbor scanner expected ']'")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == ']' {
					break
				} else {
					panic("cbor scanner expected ']'")
				}
			}
		}
		switch config.ct {
		case LengthPrefix:
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType4 // fix type from type0->type4
			n += x
			n += copy(out[n:], out[nn:nnn])
		case Stream:
			nnn += breakStop(out[nnn:])
			n = nnn
		}
		return txt[1:], n

	case '{':
		n, m, nn, nnn := 0, 0, 0, 0
		switch config.ct {
		case LengthPrefix:
			nn, nnn = n+32, n+32
		case Stream:
			nnn += mapStart(out[nnn:])
		}

		var ln int
		txt = skipWS(txt[1:], config.ws)
		if txt[0] == '}' {
			// pass
		} else if txt[0] != '"' {
			panic("cbor scanner expected property key")
		} else {
			for {
				// 16 reserved for cbor hdr
				txt, m = scanString(txt, out[nnn+16:])
				nnn += valtext2cbor(bytes2str(out[nnn+16:nnn+16+m]), out[nnn:])

				if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
					panic("cbor scanner expected property colon")
				}
				txt, m = json2cbor(skipWS(txt[1:], config.ws), out[nnn:], config)
				nnn += m
				ln++

				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("cbor scanner expected '}'")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == '}' {
					break
				} else {
					panic("cbor scanner expected '}'")
				}
			}
		}
		switch config.ct {
		case LengthPrefix:
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType5 // fix type from type0->type5
			n += x
			n += copy(out[n:], out[nn:nnn])
		case Stream:
			nnn += breakStop(out[nnn:])
			n = nnn
		}
		return txt[1:], n

	default:
		panic("cbor scanner expected token")
	}
}

func jsonNumToCbor(txt string, out []byte, config *Config) (string, int) {
	s, e, l, flt := 0, 1, len(txt), false
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
			flt = flt || fltCheck[txt[e]] == 1 // detected as float
		}
	}
	if config.nk != FloatNumber && !flt {
		if i, err := strconv.ParseInt(txt[s:e], 10, 64); err == nil {
			n := valint642cbor(i, out)
			return txt[e:], n
		} else if ui, err := strconv.ParseUint(txt[s:e], 10, 64); err == nil {
			n := valuint642cbor(ui, out)
			return txt[e:], n
		}
	}
	num, err := strconv.ParseFloat(txt[s:e], 64)
	if err != nil { // once parsing logic is bullet proof remove this
		panic(err)
	}
	n := valfloat642cbor(num, out)
	return txt[e:], n
}
