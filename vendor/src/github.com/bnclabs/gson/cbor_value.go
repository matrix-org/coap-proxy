// transform cbor encoded data into golang native data.
// cnf: -

package gson

import "math"
import "math/big"
import "regexp"
import "time"
import "encoding/binary"
import "fmt"

func cbor2value(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2valueM[buf[0]](buf, config)
	if _, ok := item.(CborIndefinite); ok {
		switch cborMajor(buf[0]) {
		case cborType4:
			arr := make([]interface{}, 0, 2)
			for buf[n] != brkstp {
				item, n1 := cbor2value(buf[n:], config)
				arr = append(arr, item)
				n += n1
			}
			return arr, n + 1

		case cborType5:
			mapv := make(map[string]interface{})
			for buf[n] != brkstp {
				key, n1 := cbor2value(buf[n:], config)
				value, n2 := cbor2value(buf[n+n1:], config)
				mapv[key.(string)] = value
				n = n + n1 + n2
			}
			return mapv, n + 1
		}
	}
	return item, n
}

func cbor2tag(buf []byte, config *Config) (interface{}, int) {
	byt := (buf[0] & 0x1f) | cborType0 // fix as positive num
	item, n := cbor2valueM[byt](buf, config)
	switch item.(uint64) {
	case tagDateTime:
		item, m := cbor2dtval(buf[n:], config)
		return item, n + m

	case tagEpoch:
		item, m := cbor2epochval(buf[n:], config)
		return item, n + m

	case tagPosBignum:
		item, m := cbor2bignumval(buf[n:], config)
		return item, n + m

	case tagNegBignum:
		item, m := cbor2bignumval(buf[n:], config)
		return big.NewInt(0).Mul(item.(*big.Int), big.NewInt(-1)), n + m

	case tagDecimalFraction:
		item, m := cbor2decimalval(buf[n:], config)
		return item, n + m

	case tagBigFloat:
		item, m := cbor2bigfloatval(buf[n:], config)
		return item, n + m

	case tagCborEnc:
		item, m := cbor2cborval(buf[n:], config)
		return item, n + m

	case tagRegexp:
		item, m := cbor2regexpval(buf[n:], config)
		return item, n + m

	case tagCborPrefix:
		item, m := cbor2cborprefixval(buf[n:], config)
		return item, n + m
	}
	// skip tags
	item, m := cbor2value(buf[n:], config)
	return item, n + m
}

func cbor2valnull(buf []byte, config *Config) (interface{}, int) {
	return nil, 1
}

func cbor2valfalse(buf []byte, config *Config) (interface{}, int) {
	return false, 1
}

func cbor2valtrue(buf []byte, config *Config) (interface{}, int) {
	return true, 1
}

func cbor2stbyte(buf []byte, config *Config) (interface{}, int) {
	return buf[1], 2
}

func cbor2valfloat16(buf []byte, config *Config) (interface{}, int) {
	panic("cbor2valfloat16 not supported")
}

func cbor2valfloat32(buf []byte, config *Config) (interface{}, int) {
	item, n := binary.BigEndian.Uint32(buf[1:]), 5
	return math.Float32frombits(item), n
}

func cbor2valfloat64(buf []byte, config *Config) (interface{}, int) {
	item, n := binary.BigEndian.Uint64(buf[1:]), 9
	return math.Float64frombits(item), n
}

func cbor2valt0smallint(buf []byte, config *Config) (interface{}, int) {
	return uint64(cborInfo(buf[0])), 1
}

func cbor2valt1smallint(buf []byte, config *Config) (interface{}, int) {
	return -int64(cborInfo(buf[0]) + 1), 1
}

func cbor2valt0info24(buf []byte, config *Config) (interface{}, int) {
	return uint64(buf[1]), 2
}

func cbor2valt1info24(buf []byte, config *Config) (interface{}, int) {
	return -int64(buf[1] + 1), 2
}

func cbor2valt0info25(buf []byte, config *Config) (interface{}, int) {
	return uint64(binary.BigEndian.Uint16(buf[1:])), 3
}

func cbor2valt1info25(buf []byte, config *Config) (interface{}, int) {
	return -int64(binary.BigEndian.Uint16(buf[1:]) + 1), 3
}

func cbor2valt0info26(buf []byte, config *Config) (interface{}, int) {
	return uint64(binary.BigEndian.Uint32(buf[1:])), 5
}

func cbor2valt1info26(buf []byte, config *Config) (interface{}, int) {
	return -int64(binary.BigEndian.Uint32(buf[1:]) + 1), 5
}

func cbor2valt0info27(buf []byte, config *Config) (interface{}, int) {
	return uint64(binary.BigEndian.Uint64(buf[1:])), 9
}

func cbor2valt1info27(buf []byte, config *Config) (interface{}, int) {
	x := uint64(binary.BigEndian.Uint64(buf[1:]))
	if x > 9223372036854775807 {
		panic("cbor decoding integer exceeds int64")
	}
	return int64(-x) - 1, 9
}

func cborItemLength(buf []byte) (int, int) {
	if y := cborInfo(buf[0]); y < cborInfo24 {
		return int(y), 1
	} else if y == cborInfo24 {
		return int(buf[1]), 2
	} else if y == cborInfo25 {
		return int(binary.BigEndian.Uint16(buf[1:])), 3
	} else if y == cborInfo26 {
		return int(binary.BigEndian.Uint32(buf[1:])), 5
	}
	return int(binary.BigEndian.Uint64(buf[1:])), 9 // info27
}

func cbor2valt2(buf []byte, config *Config) (interface{}, int) {
	ln, n := cborItemLength(buf)
	dst := make([]byte, ln)
	copy(dst, buf[n:n+ln])
	return dst, n + ln
}

func cbor2valt2indefinite(buf []byte, config *Config) (interface{}, int) {
	value, n := make([][]byte, 0), 1
	for buf[n] != brkstp {
		val, m := cbor2value(buf[n:], config)
		n += m
		value = append(value, val.([]byte))
	}
	return value, n
}

func cbor2valt3(buf []byte, config *Config) (interface{}, int) {
	ln, n := cborItemLength(buf)
	dst := make([]byte, ln)
	copy(dst, buf[n:n+ln])
	return bytes2str(dst), n + ln
}

func cbor2valt3indefinite(buf []byte, config *Config) (interface{}, int) {
	value, n := make([]string, 0), 1
	for buf[n] != brkstp {
		val, m := cbor2value(buf[n:], config)
		n += m
		value = append(value, val.(string))
	}
	return value, n
}

func cbor2valt4(buf []byte, config *Config) (interface{}, int) {
	ln, n := cborItemLength(buf)
	arr := make([]interface{}, ln)
	for i := 0; i < ln; i++ {
		item, n1 := cbor2value(buf[n:], config)
		arr[i], n = item, n+n1
	}
	return arr, n
}

func cbor2valt4indefinite(buf []byte, config *Config) (interface{}, int) {
	return CborIndefinite(buf[0]), 1
}

func cbor2valt5(buf []byte, config *Config) (interface{}, int) {
	ln, n := cborItemLength(buf)
	mapv := make(map[string]interface{})
	for i := 0; i < ln; i++ {
		key, n1 := cbor2value(buf[n:], config)
		value, n2 := cbor2value(buf[n+n1:], config)
		mapv[key.(string)] = value
		n = n + n1 + n2
	}
	return mapv, n
}

func cbor2valt5indefinite(buf []byte, config *Config) (interface{}, int) {
	return CborIndefinite(buf[0]), 1
}

func cbor2valbreakcode(buf []byte, config *Config) (interface{}, int) {
	return CborBreakStop(buf[0]), 1
}

func cbor2valundefined(buf []byte, config *Config) (interface{}, int) {
	return CborUndefined(cborSimpleUndefined), 1
}

//---- decode tags

func cbor2dtval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	item, err := time.Parse(time.RFC3339, item.(string))
	if err != nil {
		panic("cbor2dtval(): malformed time.RFC3339")
	}
	return item, n
}

func cbor2epochval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	switch v := item.(type) {
	case int64:
		return CborTagEpoch(v), n
	case uint64:
		return CborTagEpoch(v), n
	case float64:
		return CborTagEpochMicro(v), n
	}
	fmsg := "cbor2bignumval(): neither int64 nor float64: %T"
	panic(fmt.Errorf(fmsg, item))
}

func cbor2bignumval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	num := big.NewInt(0).SetBytes(item.([]byte))
	return num, n
}

func cbor2decimalval(buf []byte, config *Config) (interface{}, int) {
	x, n := cbor2value(buf, config)
	arr := x.([]interface{})
	if len(arr) != 2 {
		panic("malformed tagDecimalFraction")
	}
	e, m := arr[0], arr[1]

	if a, ok := e.(uint64); ok {
		if b, ok := m.(uint64); ok {
			return CborTagFraction([2]int64{int64(a), int64(b)}), n
		}
		return CborTagFraction([2]int64{int64(a), m.(int64)}), n

	} else if b, ok := m.(uint64); ok {
		return CborTagFraction([2]int64{e.(int64), int64(b)}), n
	}
	return CborTagFraction([2]int64{e.(int64), m.(int64)}), n
}

func cbor2bigfloatval(buf []byte, config *Config) (interface{}, int) {
	x, n := cbor2value(buf, config)
	arr := x.([]interface{})
	if len(arr) != 2 {
		panic("malformed tagBigFloat")
	}
	e, m := arr[0], arr[1]

	if a, ok := e.(uint64); ok {
		if b, ok := m.(uint64); ok {
			return CborTagFloat([2]int64{int64(a), int64(b)}), n
		}
		return CborTagFloat([2]int64{int64(a), m.(int64)}), n

	} else if b, ok := m.(uint64); ok {
		return CborTagFloat([2]int64{e.(int64), int64(b)}), n
	}
	return CborTagFloat([2]int64{e.(int64), m.(int64)}), n
}

func cbor2cborval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	return CborTagBytes(item.([]uint8)), n
}

func cbor2regexpval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	s := item.(string)
	re, err := regexp.Compile(s)
	if err != nil {
		panic(fmt.Errorf("compiling regexp %q: %v", s, err))
	}
	return re, n
}

func cbor2cborprefixval(buf []byte, config *Config) (interface{}, int) {
	item, n := cbor2value(buf, config)
	return CborTagPrefix(item.([]byte)), n
}

var cbor2valueM = make(map[byte]func([]byte, *Config) (interface{}, int))

func init() {
	makePanic := func(msg string) func([]byte, *Config) (interface{}, int) {
		return func(_ []byte, _ *Config) (interface{}, int) { panic(msg) }
	}
	//-- type0                  (unsigned integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2valueM[cborHdr(cborType0, i)] = cbor2valt0smallint
	}
	// 1st-byte 24..27
	cbor2valueM[cborHdr(cborType0, cborInfo24)] = cbor2valt0info24
	cbor2valueM[cborHdr(cborType0, cborInfo25)] = cbor2valt0info25
	cbor2valueM[cborHdr(cborType0, cborInfo26)] = cbor2valt0info26
	cbor2valueM[cborHdr(cborType0, cborInfo27)] = cbor2valt0info27
	// 1st-byte 28..31
	msg := "cbor decode value type0 reserved info"
	cbor2valueM[cborHdr(cborType0, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType0, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType0, 30)] = makePanic(msg)
	msg = "cbor decode value type0 indefnite"
	cbor2valueM[cborHdr(cborType0, cborIndefiniteLength)] = makePanic(msg)

	//-- type1                  (signed integer)
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2valueM[cborHdr(cborType1, i)] = cbor2valt1smallint
	}
	// 1st-byte 24..27
	cbor2valueM[cborHdr(cborType1, cborInfo24)] = cbor2valt1info24
	cbor2valueM[cborHdr(cborType1, cborInfo25)] = cbor2valt1info25
	cbor2valueM[cborHdr(cborType1, cborInfo26)] = cbor2valt1info26
	cbor2valueM[cborHdr(cborType1, cborInfo27)] = cbor2valt1info27
	// 1st-byte 28..31
	msg = "cbor decode value type1 reserved info"
	cbor2valueM[cborHdr(cborType1, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType1, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType1, 30)] = makePanic(msg)
	msg = "cbor decode value type1 indefnite"
	cbor2valueM[cborHdr(cborType1, cborIndefiniteLength)] = makePanic(msg)

	//-- type2                  (byte string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2valueM[cborHdr(cborType2, byte(i))] = cbor2valt2
	}
	// 1st-byte 28..31
	msg = "cbor decode value type2 reserved info"
	cbor2valueM[cborHdr(cborType2, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType2, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType2, 30)] = makePanic(msg)
	msg = "cbor decode value type2 indefnite"
	cbor2valueM[cborHdr(cborType2, cborIndefiniteLength)] = cbor2valt2indefinite

	//-- type3                  (string)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2valueM[cborHdr(cborType3, byte(i))] = cbor2valt3
	}
	// 1st-byte 28..31
	msg = "cbor decode value type3 reserved info"
	cbor2valueM[cborHdr(cborType3, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType3, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType3, 30)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType3, cborIndefiniteLength)] = cbor2valt3indefinite

	//-- type4                  (array)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2valueM[cborHdr(cborType4, byte(i))] = cbor2valt4
	}
	// 1st-byte 28..31
	msg = "cbor decode value type4 reserved info"
	cbor2valueM[cborHdr(cborType4, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType4, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType4, 30)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType4, cborIndefiniteLength)] = cbor2valt4indefinite

	//-- type5                  (map)
	// 1st-byte 0..27
	for i := 0; i < 28; i++ {
		cbor2valueM[cborHdr(cborType5, byte(i))] = cbor2valt5
	}
	// 1st-byte 28..31
	msg = "cbor decode value type5 reserved info"
	cbor2valueM[cborHdr(cborType5, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType5, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType5, 30)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType5, cborIndefiniteLength)] = cbor2valt5indefinite

	//-- type6
	// 1st-byte 0..23
	for i := byte(0); i < cborInfo24; i++ {
		cbor2valueM[cborHdr(cborType6, i)] = cbor2tag
	}
	// 1st-byte 24..27
	cbor2valueM[cborHdr(cborType6, cborInfo24)] = cbor2tag
	cbor2valueM[cborHdr(cborType6, cborInfo25)] = cbor2tag
	cbor2valueM[cborHdr(cborType6, cborInfo26)] = cbor2tag
	cbor2valueM[cborHdr(cborType6, cborInfo27)] = cbor2tag
	// 1st-byte 28..31
	msg = "cbor decode value type6 reserved info"
	cbor2valueM[cborHdr(cborType6, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType6, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType6, 30)] = makePanic(msg)
	msg = "cbor decode value type6 indefnite"
	cbor2valueM[cborHdr(cborType6, cborIndefiniteLength)] = makePanic(msg)

	//-- type7                  (simple types / floats / break-stop)
	// 1st-byte 0..19
	for i := byte(0); i < 20; i++ {
		cbor2valueM[cborHdr(cborType7, i)] =
			func(i byte) func([]byte, *Config) (interface{}, int) {
				return func(buf []byte, _ *Config) (interface{}, int) {
					return i, 1
				}
			}(i)
	}
	// 1st-byte 20..23
	cbor2valueM[cborHdr(cborType7, cborSimpleTypeFalse)] = cbor2valfalse
	cbor2valueM[cborHdr(cborType7, cborSimpleTypeTrue)] = cbor2valtrue
	cbor2valueM[cborHdr(cborType7, cborSimpleTypeNil)] = cbor2valnull
	cbor2valueM[cborHdr(cborType7, cborSimpleUndefined)] = cbor2valundefined

	cbor2valueM[cborHdr(cborType7, cborSimpleTypeByte)] = cbor2stbyte
	cbor2valueM[cborHdr(cborType7, cborFlt16)] = cbor2valfloat16
	cbor2valueM[cborHdr(cborType7, cborFlt32)] = cbor2valfloat32
	cbor2valueM[cborHdr(cborType7, cborFlt64)] = cbor2valfloat64
	// 1st-byte 28..31
	msg = "cbor decode value type7 simple type"
	cbor2valueM[cborHdr(cborType7, 28)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType7, 29)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType7, 30)] = makePanic(msg)
	cbor2valueM[cborHdr(cborType7, cborItemBreak)] = cbor2valbreakcode
}
