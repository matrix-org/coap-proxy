// transform cbor encoded value into json encoding.
// cnf: -

package gson

import "math"
import "strconv"
import "encoding/binary"

var nullBin = []byte("null")
var trueBin = []byte("true")
var falseBin = []byte("false")

func cbor2json(in, out []byte, config *Config) (int, int) {
	n, m := cbor2jsonM[in[0]](in, out, config)
	return n, m
}

func cbor2jsonnull(buf, out []byte, config *Config) (int, int) {
	copy(out, nullBin)
	return 1, 4
}

func cbor2jsonfalse(buf, out []byte, config *Config) (int, int) {
	copy(out, falseBin)
	return 1, 5
}

func cbor2jsontrue(buf, out []byte, config *Config) (int, int) {
	copy(out, trueBin)
	return 1, 4
}

func cbor2jsonfloat32(buf, out []byte, config *Config) (int, int) {
	item, n := uint64(binary.BigEndian.Uint32(buf[1:])), 5
	f := math.Float32frombits(uint32(item))
	out = strconv.AppendFloat(out[:0], float64(f), 'f', -1, 32)
	return n, len(out)
}

func cbor2jsonfloat64(buf, out []byte, config *Config) (int, int) {
	item, n := uint64(binary.BigEndian.Uint64(buf[1:])), 9
	f := math.Float64frombits(item)
	out = strconv.AppendFloat(out[:0], f, 'f', -1, 64)
	return n, len(out)
}

func cbor2jsont0smallint(buf, out []byte, config *Config) (int, int) {
	val, n := uint64(cborInfo(buf[0])), 1
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont1smallint(buf, out []byte, config *Config) (int, int) {
	val, n := -int64(cborInfo(buf[0])+1), 1
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont0info24(buf, out []byte, config *Config) (int, int) {
	val, n := uint64(buf[1]), 2
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont1info24(buf, out []byte, config *Config) (int, int) {
	val, n := -int64(buf[1]+1), 2
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont0info25(buf, out []byte, config *Config) (int, int) {
	val, n := uint64(binary.BigEndian.Uint16(buf[1:])), 3
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont1info25(buf, out []byte, config *Config) (int, int) {
	val, n := -int64(binary.BigEndian.Uint16(buf[1:])+1), 3
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont0info26(buf, out []byte, config *Config) (int, int) {
	val, n := uint64(binary.BigEndian.Uint32(buf[1:])), 5
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont1info26(buf, out []byte, config *Config) (int, int) {
	val, n := -int64(binary.BigEndian.Uint32(buf[1:])+1), 5
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont0info27(buf, out []byte, config *Config) (int, int) {
	val, n := uint64(binary.BigEndian.Uint64(buf[1:])), 9
	out = strconv.AppendUint(out[:0], val, 10)
	return n, len(out)
}

func cbor2jsont1info27(buf, out []byte, config *Config) (int, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbo->json number exceeds the limit of int64")
	}
	val, n := int64(-x)-1, 9
	out = strconv.AppendInt(out[:0], val, 10)
	return n, len(out)
}

// this is to support strings that are encoded via golang,
// but used by cbor->json decoder.
func cbor2jsont3(buf, out []byte, config *Config) (int, int) {
	ln, n := cborItemLength(buf)

	if config.strict {
		config.buf.Reset()
		if err := config.enc.Encode(bytes2str(buf[n : n+ln])); err != nil {
			panic(err)
		}
		s := config.buf.Bytes()
		return n + ln, copy(out, s[:len(s)-1]) // -1 to strip \n
	}

	out1, err := encodeString(buf[n:n+ln], out[:0])
	if err != nil {
		panic(err)
	}
	return n + ln, len(out1)
}

// this to support arrays thar are encoded via golang,
// but used by cbor->json decoder
func cbor2jsont4(buf, out []byte, config *Config) (int, int) {
	ln, n := cborItemLength(buf)
	out[0] = '['
	if ln == 0 {
		out[1] = ']'
		return n, 2
	}
	m := 1
	for ; ln > 0; ln-- {
		x, y := cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = ']'
	return n, m
}

func cbor2jsont4indefinite(buf, out []byte, config *Config) (int, int) {
	out[0] = '['
	if buf[1] == brkstp {
		out[1] = ']'
		return 2, 2
	}
	n, m := 1, 1
	for buf[n] != brkstp {
		x, y := cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = ']'
	return n + 1, m
}

// this to support maps thar are encoded via golang,
// but used by cbor->json decoder
func cbor2jsont5(buf, out []byte, config *Config) (int, int) {
	ln, n := cborItemLength(buf)
	out[0] = '{'
	if ln == 0 {
		out[1] = '}'
		return n, 2
	}
	m := 1
	for ; ln > 0; ln-- {
		x, y := cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ':', m+1

		x, y = cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = '}'
	return n, m
}

func cbor2jsont5indefinite(buf, out []byte, config *Config) (int, int) {
	out[0] = '{'
	if buf[1] == brkstp {
		out[1] = '}'
		return 2, 2
	}
	n, m := 1, 1
	for buf[n] != brkstp {
		x, y := cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ':', m+1

		x, y = cbor2jsonM[buf[n]](buf[n:], out[m:], config)
		m, n = m+y, n+x
		out[m], m = ',', m+1
	}
	out[m-1] = '}'
	return n + 1, m
}

func tag2json(buf, out []byte, config *Config) (int, int) {
	byt := (buf[0] & 0x1f) | cborType0 // fix as positive num
	_ /*item*/, n := cbor2valueM[byt](buf, config)
	return n, 0 // skip this tag
}

var cbor2jsonM = make(map[byte]func([]byte, []byte, *Config) (int, int))

func init() {
	makePanic := func(msg string) func([]byte, []byte, *Config) (int, int) {
		return func(_, _ []byte, _ *Config) (int, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2jsonM[cborHdr(cborType0, i)] = cbor2jsont0smallint
	}
	// 1st-byte 24..27
	cbor2jsonM[cborHdr(cborType0, cborInfo24)] = cbor2jsont0info24
	cbor2jsonM[cborHdr(cborType0, cborInfo25)] = cbor2jsont0info25
	cbor2jsonM[cborHdr(cborType0, cborInfo26)] = cbor2jsont0info26
	cbor2jsonM[cborHdr(cborType0, cborInfo27)] = cbor2jsont0info27
	// 1st-byte 28..31
	msg := "cbor->json decode type0 reserved info"
	cbor2jsonM[cborHdr(cborType0, 28)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType0, 29)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType0, 30)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType0, cborIndefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2jsonM[cborHdr(cborType1, i)] = cbor2jsont1smallint
	}
	// 1st-byte 24..27
	cbor2jsonM[cborHdr(cborType1, cborInfo24)] = cbor2jsont1info24
	cbor2jsonM[cborHdr(cborType1, cborInfo25)] = cbor2jsont1info25
	cbor2jsonM[cborHdr(cborType1, cborInfo26)] = cbor2jsont1info26
	cbor2jsonM[cborHdr(cborType1, cborInfo27)] = cbor2jsont1info27
	// 1st-byte 28..31
	msg = "cbor->json type1 decode reserved info"
	cbor2jsonM[cborHdr(cborType1, 28)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType1, 29)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType1, 30)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType1, cborIndefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	msg = "cbor->json byte string not supported"
	for i := 0; i < 28; i++ {
		cbor2jsonM[cborHdr(cborType2, byte(i))] = makePanic(msg)
	}
	// 1st-byte 28..31
	cbor2jsonM[cborHdr(cborType2, 28)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType2, 29)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType2, 30)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType2, cborIndefiniteLength)] = makePanic(msg)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2jsonM[cborHdr(cborType3, byte(i))] = cbor2jsont3
	}
	// 1st-byte 28..31
	cbor2jsonM[cborHdr(cborType3, 28)] = cbor2jsont3
	cbor2jsonM[cborHdr(cborType3, 29)] = cbor2jsont3
	cbor2jsonM[cborHdr(cborType3, 30)] = cbor2jsont3
	msg = "cbor->json indefinite string not supported"
	cbor2jsonM[cborHdr(cborType3, cborIndefiniteLength)] = makePanic(msg)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2jsonM[cborHdr(cborType4, byte(i))] = cbor2jsont4
	}
	// 1st-byte 28..31
	cbor2jsonM[cborHdr(cborType4, 28)] = cbor2jsont4
	cbor2jsonM[cborHdr(cborType4, 29)] = cbor2jsont4
	cbor2jsonM[cborHdr(cborType4, 30)] = cbor2jsont4
	cbor2jsonM[cborHdr(cborType4, cborIndefiniteLength)] = cbor2jsont4indefinite

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2jsonM[cborHdr(cborType5, byte(i))] = cbor2jsont5
	}
	// 1st-byte 28..31
	cbor2jsonM[cborHdr(cborType5, 28)] = cbor2jsont5
	cbor2jsonM[cborHdr(cborType5, 29)] = cbor2jsont5
	cbor2jsonM[cborHdr(cborType5, 30)] = cbor2jsont5
	cbor2jsonM[cborHdr(cborType5, cborIndefiniteLength)] = cbor2jsont5indefinite

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2jsonM[cborHdr(cborType6, i)] = tag2json
	}
	// 1st-byte 24..27
	cbor2jsonM[cborHdr(cborType6, cborInfo24)] = tag2json
	cbor2jsonM[cborHdr(cborType6, cborInfo25)] = tag2json
	cbor2jsonM[cborHdr(cborType6, cborInfo26)] = tag2json
	cbor2jsonM[cborHdr(cborType6, cborInfo27)] = tag2json
	// 1st-byte 28..31
	msg = "cbor->json type6 decode reserved info"
	cbor2jsonM[cborHdr(cborType6, 28)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType6, 29)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType6, 30)] = makePanic(msg)
	msg = "cbor->json indefinite type6 not supported"
	cbor2jsonM[cborHdr(cborType6, cborIndefiniteLength)] = makePanic(msg)

	//-- type7                  (simple values / floats / break-stop)
	msg = "cbor->json simple-type < 20 not supported"
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cbor2jsonM[cborHdr(cborType7, i)] = makePanic(msg)
	}
	// 1st-byte 20..23
	cbor2jsonM[cborHdr(cborType7, cborSimpleTypeFalse)] = cbor2jsonfalse
	cbor2jsonM[cborHdr(cborType7, cborSimpleTypeTrue)] = cbor2jsontrue
	cbor2jsonM[cborHdr(cborType7, cborSimpleTypeNil)] = cbor2jsonnull
	msg = "cbor->json simple-type-undefined not supported"
	cbor2jsonM[cborHdr(cborType7, cborSimpleUndefined)] = makePanic(msg)

	msg = "cbor->json simple-type > 31 not supported"
	cbor2jsonM[cborHdr(cborType7, cborSimpleTypeByte)] = makePanic(msg)
	msg = "cbor->json float16 not supported"
	cbor2jsonM[cborHdr(cborType7, cborFlt16)] = makePanic(msg)
	cbor2jsonM[cborHdr(cborType7, cborFlt32)] = cbor2jsonfloat32
	cbor2jsonM[cborHdr(cborType7, cborFlt64)] = cbor2jsonfloat64
	// 1st-byte 28..31
	msg = "cbor->json simple-type 28 not supported"
	cbor2jsonM[cborHdr(cborType7, 28)] = makePanic(msg)
	msg = "cbor->json simple-type 29 not supported"
	cbor2jsonM[cborHdr(cborType7, 29)] = makePanic(msg)
	msg = "cbor->json simple-type 30 not supported"
	cbor2jsonM[cborHdr(cborType7, 30)] = makePanic(msg)
	msg = "cbor->json simple-type break-code not supported"
	cbor2jsonM[cborHdr(cborType7, cborItemBreak)] = makePanic(msg)
}
