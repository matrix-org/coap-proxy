// transform json encoded value into collate encoding.
// cnf: NumberKind, arrayLenPrefix, propertyLenPrefix

package gson

func collate2json(code []byte, text []byte, config *Config) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	var scratch [64]byte
	n, m := 1, 0
	switch code[0] {
	case TypeMissing:
		copy(text, MissingLiteral)
		return n + 1, m + len(MissingLiteral)

	case TypeNull:
		copy(text, "null")
		return n + 1, m + 4

	case TypeTrue:
		copy(text, "true")
		return n + 1, m + 4

	case TypeFalse:
		copy(text, "false")
		return n + 1, m + 5

	case TypeNumber:
		x := getDatum(code[n:])
		y := collated2Json(code[n:n+x-1], text, config.nk)
		return n + x, m + y

	case TypeString:
		var x int

		bufn := config.bufferh.getbuffer(len(code[n:]) * 5)
		scratch := bufn.data

		scratch, x = collate2String(code[n:], scratch[:])

		if config.strict {
			config.buf.Reset()
			if err := config.enc.Encode(bytes2str(scratch)); err != nil {
				panic(err)
			}
			s := config.buf.Bytes()
			m += copy(text[m:], s[:len(s)-1]) // -1 to strip \n
			config.bufferh.putbuffer(bufn)
			return n + x, m
		}
		remtxt, err := encodeString(scratch, text[m:m])
		if err != nil {
			panic(err)
		}
		config.bufferh.putbuffer(bufn)
		return n + x, m + len(remtxt)

	case TypeArray:
		if config.arrayLenPrefix {
			x := getDatum(code[n:])
			collated2Int(code[n:n+x-1], scratch[:])
			n += x
		}
		text[m] = '['
		m++
		for code[n] != Terminator {
			x, y := collate2json(code[n:], text[m:], config)
			n += x
			m += y
			text[m] = ','
			m++
		}
		n++ // skip terminator
		if text[m-1] == ',' {
			text[m-1] = ']'
		} else {
			text[m] = ']'
			m++
		}
		return n, m

	case TypeObj:
		if config.propertyLenPrefix {
			x := getDatum(code[n:])
			collated2Int(code[n:n+x-1], scratch[:])
			n += x
		}
		text[m] = '{'
		m++
		for code[n] != Terminator {
			x, y := collate2json(code[n:], text[m:], config)
			n, m = n+x, m+y
			text[m] = ':'
			m++
			x, y = collate2json(code[n:], text[m:], config)
			n, m = n+x, m+y
			text[m] = ','
			m++
		}
		n++ // skip terminator
		if text[m-1] == ',' {
			text[m-1] = '}'
		} else {
			text[m] = '}'
			m++
		}
		return n, m
	}
	panic("collate decode to json invalid binary")
}
