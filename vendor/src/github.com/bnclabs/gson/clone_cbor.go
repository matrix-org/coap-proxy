package gson

func cborclone(in, out []byte, config *Config) int {
	return cborcloneM[in[0]](in, out, config)
}

func cborclonet0smallint(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborclonet0info24(in, out []byte, config *Config) int {
	return copy(out, in[:2])
}

func cborclonet0info25(in, out []byte, config *Config) int {
	return copy(out, in[:3])
}

func cborclonet0info26(in, out []byte, config *Config) int {
	return copy(out, in[:5])
}

func cborclonet0info27(in, out []byte, config *Config) int {
	return copy(out, in[:9])
}

func cborclonet1smallint(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborclonet1info24(in, out []byte, config *Config) int {
	return copy(out, in[:2])
}

func cborclonet1info25(in, out []byte, config *Config) int {
	return copy(out, in[:3])
}

func cborclonet1info26(in, out []byte, config *Config) int {
	return copy(out, in[:5])
}

func cborclonet1info27(in, out []byte, config *Config) int {
	return copy(out, in[:9])
}

func cborclonet2(in, out []byte, config *Config) int {
	ln, n := cborItemLength(in)
	return copy(out, in[:n+ln])
}

func cborclonet2indefinite(in, out []byte, config *Config) int {
	out[0] = in[0]
	n := 1
	for in[n] != brkstp {
		n += cborclone(in[n:], out[n:], config)
	}
	out[n] = in[n]
	return n + 1
}

func cborclonet3(in, out []byte, config *Config) int {
	ln, n := cborItemLength(in)
	return copy(out, in[:n+ln])
}

func cborclonet3indefinite(in, out []byte, config *Config) int {
	out[0] = in[0]
	n := 1
	for in[n] != brkstp {
		n += cborclone(in[n:], out[n:], config)
	}
	out[n] = in[n]
	return n + 1
}

func cborclonet4(in, out []byte, config *Config) int {
	ln, n := cborItemLength(in)
	copy(out, in[:n])
	for i := 0; i < ln; i++ {
		n += cborclone(in[n:], out[n:], config)
	}
	return n
}

func cborclonet4indefinite(in, out []byte, config *Config) int {
	out[0] = in[0]
	n := 1
	for in[n] != brkstp {
		n += cborclone(in[n:], out[n:], config)
	}
	out[n] = in[n]
	return n + 1
}

func cborclonet5(in, out []byte, config *Config) int {
	ln, n := cborItemLength(in)
	copy(out, in[:n])
	for i := 0; i < ln; i++ {
		n += cborclone(in[n:], out[n:], config)
		n += cborclone(in[n:], out[n:], config)
	}
	return n
}

func cborclonet5indefinite(in, out []byte, config *Config) int {
	out[0] = in[0]
	n := 1
	for in[n] != brkstp {
		n += cborclone(in[n:], out[n:], config)
		n += cborclone(in[n:], out[n:], config)
	}
	out[n] = in[n]
	return n + 1
}

func cborcloneFalse(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborcloneTrue(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborcloneNull(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborcloneUndefined(in, out []byte, config *Config) int {
	out[0] = in[0]
	return 1
}

func cborcloneSimpleType(in, out []byte, config *Config) int {
	out[0], out[1] = in[0], in[1]
	return 2
}

func cborcloneFloat16(in, out []byte, config *Config) int {
	return copy(out, in[:3])
}

func cborcloneFloat32(in, out []byte, config *Config) int {
	return copy(out, in[:5])
}

func cborcloneFloat64(in, out []byte, config *Config) int {
	return copy(out, in[:9])
}

func cborcloneTag(in, out []byte, config *Config) int {
	byt := (in[0] & 0x1f) | cborType0 // fix as positive num
	in[0] = byt
	n := cborclone(in, out, config)
	in[0], out[0] = byt, byt
	n += cborclone(in[n:], out[n:], config)
	return n
}

var cborcloneM = make(map[byte]func([]byte, []byte, *Config) int)

func init() {
	makePanic := func(msg string) func([]byte, []byte, *Config) int {
		return func(_, _ []byte, _ *Config) int { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cborcloneM[cborHdr(cborType0, i)] = cborclonet0smallint
	}
	// 1st-byte 24..27
	cborcloneM[cborHdr(cborType0, cborInfo24)] = cborclonet0info24
	cborcloneM[cborHdr(cborType0, cborInfo25)] = cborclonet0info25
	cborcloneM[cborHdr(cborType0, cborInfo26)] = cborclonet0info26
	cborcloneM[cborHdr(cborType0, cborInfo27)] = cborclonet0info27
	// 1st-byte 28..31
	msg := "cbor decode value type0 reserved info"
	cborcloneM[cborHdr(cborType0, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType0, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType0, 30)] = makePanic(msg)
	msg = "cbor decode value type0 indefnite"
	cborcloneM[cborHdr(cborType0, cborIndefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cborcloneM[cborHdr(cborType1, i)] = cborclonet1smallint
	}
	// 1st-byte 24..27
	cborcloneM[cborHdr(cborType1, cborInfo24)] = cborclonet1info24
	cborcloneM[cborHdr(cborType1, cborInfo25)] = cborclonet1info25
	cborcloneM[cborHdr(cborType1, cborInfo26)] = cborclonet1info26
	cborcloneM[cborHdr(cborType1, cborInfo27)] = cborclonet1info27
	// 1st-byte 28..31
	msg = "cbor decode value type1 reserved info"
	cborcloneM[cborHdr(cborType1, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType1, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType1, 30)] = makePanic(msg)
	msg = "cbor decode value type1 indefnite"
	cborcloneM[cborHdr(cborType1, cborIndefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborcloneM[cborHdr(cborType2, byte(i))] = cborclonet2
	}
	// 1st-byte 28..31
	msg = "cbor decode value type2 reserved info"
	cborcloneM[cborHdr(cborType2, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType2, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType2, 30)] = makePanic(msg)
	msg = "cbor decode value type2 indefnite"
	cborcloneM[cborHdr(cborType2, cborIndefiniteLength)] = cborclonet2indefinite

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborcloneM[cborHdr(cborType3, byte(i))] = cborclonet3
	}
	// 1st-byte 28..31
	msg = "cbor decode value type3 reserved info"
	cborcloneM[cborHdr(cborType3, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType3, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType3, 30)] = makePanic(msg)
	cborcloneM[cborHdr(cborType3, cborIndefiniteLength)] = cborclonet3indefinite

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborcloneM[cborHdr(cborType4, byte(i))] = cborclonet4
	}
	// 1st-byte 28..31
	msg = "cbor decode value type4 reserved info"
	cborcloneM[cborHdr(cborType4, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType4, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType4, 30)] = makePanic(msg)
	cborcloneM[cborHdr(cborType4, cborIndefiniteLength)] = cborclonet4indefinite

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cborcloneM[cborHdr(cborType5, byte(i))] = cborclonet5
	}
	// 1st-byte 28..31
	msg = "cbor decode value type5 reserved info"
	cborcloneM[cborHdr(cborType5, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType5, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType5, 30)] = makePanic(msg)
	cborcloneM[cborHdr(cborType5, cborIndefiniteLength)] = cborclonet5indefinite

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cborcloneM[cborHdr(cborType6, i)] = cborcloneTag
	}
	// 1st-byte 24..27
	cborcloneM[cborHdr(cborType6, cborInfo24)] = cborcloneTag
	cborcloneM[cborHdr(cborType6, cborInfo25)] = cborcloneTag
	cborcloneM[cborHdr(cborType6, cborInfo26)] = cborcloneTag
	cborcloneM[cborHdr(cborType6, cborInfo27)] = cborcloneTag
	// 1st-byte 28..31
	msg = "cbor decode value type6 reserved info"
	cborcloneM[cborHdr(cborType6, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType6, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType6, 30)] = makePanic(msg)
	msg = "cbor decode value type6 indefnite"
	cborcloneM[cborHdr(cborType6, cborIndefiniteLength)] = makePanic(msg)

	//-- type7                  (simple types / floats / break-stop)
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cborcloneM[cborHdr(cborType7, i)] =
			func(i byte) func([]byte, []byte, *Config) int {
				return func(in, out []byte, _ *Config) int {
					out[0] = in[0]
					return 1
				}
			}(i)
	}
	// 1st-byte 20..23
	cborcloneM[cborHdr(cborType7, cborSimpleTypeFalse)] = cborcloneFalse
	cborcloneM[cborHdr(cborType7, cborSimpleTypeTrue)] = cborcloneTrue
	cborcloneM[cborHdr(cborType7, cborSimpleTypeNil)] = cborcloneNull
	cborcloneM[cborHdr(cborType7, cborSimpleUndefined)] = cborcloneUndefined

	cborcloneM[cborHdr(cborType7, cborSimpleTypeByte)] = cborcloneSimpleType
	cborcloneM[cborHdr(cborType7, cborFlt16)] = cborcloneFloat16
	cborcloneM[cborHdr(cborType7, cborFlt32)] = cborcloneFloat32
	cborcloneM[cborHdr(cborType7, cborFlt64)] = cborcloneFloat64
	// 1st-byte 28..31
	msg = "cbor decode value type7 simple type"
	cborcloneM[cborHdr(cborType7, 28)] = makePanic(msg)
	cborcloneM[cborHdr(cborType7, 29)] = makePanic(msg)
	cborcloneM[cborHdr(cborType7, 30)] = makePanic(msg)
	cborcloneM[cborHdr(cborType7, cborItemBreak)] = makePanic(msg)
}
