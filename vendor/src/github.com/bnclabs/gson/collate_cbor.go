// transform binary-collated data into cbor encoding.
// cnf: NumberKind, ContainerEncoding, arrayLenPrefix, propertyLenPrefix

package gson

//---- collate to cbor

func collate2cbor(code, out []byte, config *Config) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}
	var scratch [64]byte
	m, n := 1, 0
	switch code[0] {
	case TypeMissing:
		n += valtext2cbor(string(MissingLiteral), out[n:])
		return m + 1, n

	case TypeNull:
		n += cborNull(out[n:])
		return m + 1, n

	case TypeTrue:
		n += cborTrue(out[n:])
		return m + 1, n

	case TypeFalse:
		n += cborFalse(out[n:])
		return m + 1, n

	case TypeNumber:
		x := getDatum(code[m:])
		// -1 is to skip terminator
		ui, i, f, what := collated2Number(code[m:m+x-1], config.nk)
		switch what {
		case 1:
			n += valuint642cbor(ui, out[n:])
		case 2:
			n += valint642cbor(i, out[n:])
		case 3:
			n += valfloat642cbor(f, out[n:])
		}
		return m + x, n

	case TypeString:
		var x int

		bufn := config.bufferh.getbuffer(len(code[m:]) * 5)
		scratch := bufn.data
		scratch, x = collate2String(code[m:], scratch[:])
		n += valtext2cbor(bytes2str(scratch), out[n:])
		config.bufferh.putbuffer(bufn)
		return m + x, n

	case TypeBinary:
		x := getDatum(code[m:])
		n += valbytes2cbor(code[m:m+x-1], out[n:])
		return m + x, n

	case TypeArray:
		if config.arrayLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected array length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			collated2Int(code[m:m+x-1], scratch[:]) // skip length
			m += x
		}
		nn, nnn := n, n
		if config.ct == LengthPrefix {
			nn, nnn = n+32, n+32
		} else if config.ct == Stream {
			nnn += arrayStart(out[nnn:])
		}
		ln := 0
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[nnn:], config)
			m, nnn = m+x, nnn+y
			ln++
		}
		if config.ct == LengthPrefix {
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType4 // fix type from type0->type4
			n += x
			n += copy(out[n:], out[nn:nnn])
		} else if config.ct == Stream {
			nnn += breakStop(out[nnn:])
			n = nnn
		}
		return m + 1, n

	case TypeObj:
		if config.propertyLenPrefix {
			if code[m] != TypeLength {
				panic("collate decode expected property length prefix")
			}
			x := getDatum(code[m:])
			// -1 skip terminator
			collated2Int(code[m:m+x-1], scratch[:]) // skip length
			m += x
		}
		nn, nnn := n, n
		if config.ct == LengthPrefix {
			nn, nnn = n+32, n+32
		} else if config.ct == Stream {
			nnn += mapStart(out[nnn:])
		}

		ln := 0
		for code[m] != Terminator {
			x, y := collate2cbor(code[m:], out[nnn:], config)
			m, nnn = m+x, nnn+y
			x, y = collate2cbor(code[m:], out[nnn:], config)
			m, nnn = m+x, nnn+y
			ln++
		}
		if config.ct == LengthPrefix {
			x := valuint642cbor(uint64(ln), out[n:])
			out[n] = (out[n] & 0x1f) | cborType5 // fix type from type0->type5
			n += x
			n += copy(out[n:], out[nn:nnn])
		} else if config.ct == Stream {
			nnn += breakStop(out[nnn:])
			n = nnn
		}
		return m + 1, n
	}
	panic("collate decode to cbor invalid binary")
}
