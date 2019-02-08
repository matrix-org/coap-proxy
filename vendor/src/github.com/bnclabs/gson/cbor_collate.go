// transform cbor encoded data into binary-collation.
// cnf: NumberKind, doMissing, arrayLenPrefix, propertyLenPrefix

package gson

import "math"
import "fmt"
import "strconv"
import "encoding/binary"

func cbor2collate(in, out []byte, config *Config) (int, int) {
	return cbor2collateM[in[0]](in, out, config)
}

func collateCborNull(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeNull, Terminator
	return 1, 2
}

func collateCborTrue(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeTrue, Terminator
	return 1, 2
}

func collateCborFalse(buf, out []byte, config *Config) (int, int) {
	out[0], out[1] = TypeFalse, Terminator
	return 1, 2
}

func collateCborFloat32(buf, out []byte, config *Config) (int, int) {
	item := uint64(binary.BigEndian.Uint32(buf[1:]))
	f, n := math.Float32frombits(uint32(item)), 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(float64(f), out[n:])
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborFloat64(buf, out []byte, config *Config) (int, int) {
	item := uint64(binary.BigEndian.Uint64(buf[1:]))
	f, n := math.Float64frombits(item), 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(float64(f), out[n:])
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT0SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(float64(cborInfo(buf[0])), out[n:])
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborT1SmallInt(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(-float64(cborInfo(buf[0])+1), out[n:])
	out[n] = Terminator
	n++
	return 1, n
}

func collateCborT0Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(float64(buf[1]), out[n:])
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborT1Info24(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	n += collateFloat64(-float64(buf[1]+1), out[n:])
	out[n] = Terminator
	n++
	return 2, n
}

func collateCborT0Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	f := float64(binary.BigEndian.Uint16(buf[1:]))
	n += collateFloat64(f, out[n:])
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborT1Info25(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	f := -float64(binary.BigEndian.Uint16(buf[1:]) + 1)
	n += collateFloat64(f, out[n:])
	out[n] = Terminator
	n++
	return 3, n
}

func collateCborT0Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	f := float64(binary.BigEndian.Uint32(buf[1:]))
	n += collateFloat64(f, out[n:])
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborT1Info26(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	f := -float64(binary.BigEndian.Uint32(buf[1:]) + 1)
	n += collateFloat64(f, out[n:])
	out[n] = Terminator
	n++
	return 5, n
}

func collateCborT0Info27(buf, out []byte, config *Config) (int, int) {
	n := 0
	out[n] = TypeNumber
	n++
	i := binary.BigEndian.Uint64(buf[1:])
	switch config.nk {
	case FloatNumber:
		n += collateFloat64(float64(i), out[n:])
	case SmartNumber:
		n += collateUint64(i, out[n:], config)
	default:
		panic(fmt.Errorf("unknown number kind, %v", config.nk))
	}
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT1Info27(buf, out []byte, config *Config) (int, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbo->collate number exceeds the limit of int64")
	}
	val, n := (int64(-x) - 1), 0
	out[n] = TypeNumber
	n++
	switch config.nk {
	case FloatNumber:
		n += collateFloat64(float64(val), out[n:])
	case SmartNumber:
		n += collateInt64(val, out[n:], config)
	default:
		panic(fmt.Errorf("unknown number kind, %v", config.nk))
	}
	out[n] = Terminator
	n++
	return 9, n
}

func collateCborT2(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeBinary
	n++
	copy(out[n:], buf[m:m+ln])
	n += ln
	out[n] = Terminator
	n++
	return m + ln, n
}

func collateCborT3(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := collateString(bytes2str(buf[m:m+ln]), out, config)
	return m + ln, n
}

func collateCborLength(length int, out []byte, config *Config) int {
	var text [64]byte
	txt := strconv.AppendInt(text[:0], int64(length), 10)

	n := 0
	out[n] = TypeLength
	n++
	n += collateInt(txt, out[n:])
	out[n] = Terminator
	n++
	return n
}

func collateCborT4(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeArray
	n++
	if config.arrayLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}
	for ; ln > 0; ln-- {
		x, y := cbor2collate(buf[m:], out[n:], config)
		m, n = m+x, n+y
	}
	out[n] = Terminator
	n++
	return m, n
}

func collateCborT4Indef(buf, out []byte, config *Config) (m int, n int) {
	ln := 0
	out[n] = TypeArray
	n++
	nn, nnn := n, n
	if config.arrayLenPrefix {
		nn, nnn = n+32, n+32 // length encoding can go upto max of 32 bytes
	}

	defer func() {
		if config.arrayLenPrefix {
			n += collateCborLength(ln, out[n:], config)
		}
		copy(out[n:], out[nn:nnn])
		n += (nnn - nn)
		out[n] = Terminator
		n++
		return
	}()

	m = 1
	if buf[1] == brkstp {
		m = 2
		return
	}
	for buf[m] != brkstp {
		x, y := cbor2collate(buf[m:], out[nnn:], config)
		m, nnn = m+x, nnn+y
		ln++
	}
	m++
	return
}

func collateCborT5(buf, out []byte, config *Config) (int, int) {
	ln, m := cborItemLength(buf)
	n := 0
	out[n] = TypeObj
	n++
	if config.propertyLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}

	bufn, p := config.bufferh.getbuffer(len(buf)*5), 0
	altcode := bufn.data

	kv := config.kvh.getkv(ln)

	for i := 0; i < ln; i++ {
		x, y := cbor2collate(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = cbor2collate(buf[m:], altcode[p:], config) // value
		kv.refs = append(kv.refs, kvref{bytes2str(key), altcode[p : p+y]})
		m, p = m+x, p+y
	}
	(kv.refs[:ln]).sort()
	for i := 0; i < ln; i++ {
		ref := kv.refs[i]
		copy(out[n:], ref.key)
		n += len(ref.key)
		copy(out[n:], ref.code)
		n += len(ref.code)
	}

	out[n] = Terminator
	n++

	config.bufferh.putbuffer(bufn)
	config.kvh.putkv(kv)
	return m, n
}

func collateCborT5Indef(buf, out []byte, config *Config) (m int, n int) {
	ln := 0
	out[n] = TypeObj
	n++

	bufn, p := config.bufferh.getbuffer(len(buf)*5), 0
	altcode := bufn.data

	kv := config.kvh.getkv(config.numkeys)

	m = 1
	for buf[m] != brkstp {
		x, y := cbor2collate(buf[m:], altcode[p:], config) // key
		key := altcode[p : p+y]
		m, p = m+x, p+y
		x, y = cbor2collate(buf[m:], altcode[p:], config) // value
		kv.refs = append(kv.refs, kvref{bytes2str(key), altcode[p : p+y]})
		m, p = m+x, p+y
		ln++
	}
	m++

	(kv.refs[:ln]).sort()

	if config.propertyLenPrefix {
		n += collateCborLength(ln, out[n:], config)
	}
	for i := 0; i < ln; i++ {
		ref := kv.refs[i]
		copy(out[n:], ref.key)
		n += len(ref.key)
		copy(out[n:], ref.code)
		n += len(ref.code)
	}
	out[n] = Terminator
	n++

	config.bufferh.putbuffer(bufn)
	config.kvh.putkv(kv)
	return
}

func collateCborTag(buf, out []byte, config *Config) (int, int) {
	_ /*item*/, m := cborItemLength(buf)
	return m, 0 // skip this tag
}

var cbor2collateM = make(map[byte]func([]byte, []byte, *Config) (int, int))

func init() {
	makePanic := func(msg string) func([]byte, []byte, *Config) (int, int) {
		return func(_, _ []byte, _ *Config) (int, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType0, i)] = collateCborT0SmallInt
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType0, cborInfo24)] = collateCborT0Info24
	cbor2collateM[cborHdr(cborType0, cborInfo25)] = collateCborT0Info25
	cbor2collateM[cborHdr(cborType0, cborInfo26)] = collateCborT0Info26
	cbor2collateM[cborHdr(cborType0, cborInfo27)] = collateCborT0Info27
	// 1st-byte 28..31
	msg := "cbor->collate decode type0 reserved info"
	cbor2collateM[cborHdr(cborType0, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, 30)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType0, cborIndefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType1, i)] = collateCborT1SmallInt
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType1, cborInfo24)] = collateCborT1Info24
	cbor2collateM[cborHdr(cborType1, cborInfo25)] = collateCborT1Info25
	cbor2collateM[cborHdr(cborType1, cborInfo26)] = collateCborT1Info26
	cbor2collateM[cborHdr(cborType1, cborInfo27)] = collateCborT1Info27
	// 1st-byte 28..31
	msg = "cbor->collate cborType1 decode reserved info"
	cbor2collateM[cborHdr(cborType1, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, 30)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType1, cborIndefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType2, byte(i))] = collateCborT2
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType2, 28)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, 29)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, 30)] = collateCborT2
	cbor2collateM[cborHdr(cborType2, cborIndefiniteLength)] = makePanic(msg)

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType3, byte(i))] = collateCborT3
	}

	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType3, 28)] = collateCborT3
	cbor2collateM[cborHdr(cborType3, 29)] = collateCborT3
	cbor2collateM[cborHdr(cborType3, 30)] = collateCborT3
	msg = "cbor->collate indefinite string not supported"
	cbor2collateM[cborHdr(cborType3, cborIndefiniteLength)] = makePanic(msg)

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType4, byte(i))] = collateCborT4
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType4, 28)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, 29)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, 30)] = collateCborT4
	cbor2collateM[cborHdr(cborType4, cborIndefiniteLength)] = collateCborT4Indef

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2collateM[cborHdr(cborType5, byte(i))] = collateCborT5
	}
	// 1st-byte 28..31
	cbor2collateM[cborHdr(cborType5, 28)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, 29)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, 30)] = collateCborT5
	cbor2collateM[cborHdr(cborType5, cborIndefiniteLength)] = collateCborT5Indef

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2collateM[cborHdr(cborType6, i)] = collateCborTag
	}
	// 1st-byte 24..27
	cbor2collateM[cborHdr(cborType6, cborInfo24)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo25)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo26)] = collateCborTag
	cbor2collateM[cborHdr(cborType6, cborInfo27)] = collateCborTag
	// 1st-byte 28..31
	msg = "cbor->collate type6 decode reserved info"
	cbor2collateM[cborHdr(cborType6, 28)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType6, 29)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType6, 30)] = makePanic(msg)
	msg = "cbor->collate indefinite type6 not supported"
	cbor2collateM[cborHdr(cborType6, cborIndefiniteLength)] = makePanic(msg)

	//-- type7                  (simple values / floats / break-stop)
	msg = "cbor->collate simple-type < 20 not supported"
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cbor2collateM[cborHdr(cborType7, i)] = makePanic(msg)
	}
	// 1st-byte 20..23
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeFalse)] = collateCborFalse
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeTrue)] = collateCborTrue
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeNil)] = collateCborNull
	msg = "cbor->collate simple-type-undefined not supported"
	cbor2collateM[cborHdr(cborType7, cborSimpleUndefined)] = makePanic(msg)

	msg = "cbor->collate simple-type > 31 not supported"
	cbor2collateM[cborHdr(cborType7, cborSimpleTypeByte)] = makePanic(msg)
	msg = "cbor->collate float16 not supported"
	cbor2collateM[cborHdr(cborType7, cborFlt16)] = makePanic(msg)
	cbor2collateM[cborHdr(cborType7, cborFlt32)] = collateCborFloat32
	cbor2collateM[cborHdr(cborType7, cborFlt64)] = collateCborFloat64
	// 1st-byte 28..31
	msg = "cbor->collate simple-type 28 not supported"
	cbor2collateM[cborHdr(cborType7, 28)] = makePanic(msg)
	msg = "cbor->collate simple-type 29 not supported"
	cbor2collateM[cborHdr(cborType7, 29)] = makePanic(msg)
	msg = "cbor->collate simple-type 30 not supported"
	cbor2collateM[cborHdr(cborType7, 30)] = makePanic(msg)
	msg = "cbor->collate simple-type break-code not supported"
	cbor2collateM[cborHdr(cborType7, cborItemBreak)] = makePanic(msg)
}
