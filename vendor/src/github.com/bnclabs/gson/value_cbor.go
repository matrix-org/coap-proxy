// transform golang native value into cbor encoded value.
// cnf: ContainerEncoding

package gson

import "math"
import "regexp"
import "strconv"
import "time"
import "fmt"
import "math/big"
import "encoding/binary"
import "encoding/json"

func value2cbor(item interface{}, out []byte, config *Config) int {
	n := 0
	switch v := item.(type) {
	case nil:
		n += cborNull(out)
	case bool:
		if v {
			n += cborTrue(out)
		} else {
			n += cborFalse(out)
		}
	case int8:
		n += valint82cbor(v, out)
	case uint8:
		n += valuint82cbor(v, out)
	case int16:
		n += valint162cbor(v, out)
	case uint16:
		n += valuint162cbor(v, out)
	case int32:
		n += valint322cbor(v, out)
	case uint32:
		n += valuint322cbor(v, out)
	case int:
		n += valint642cbor(int64(v), out)
	case int64:
		n += valint642cbor(v, out)
	case uint:
		n += valuint642cbor(uint64(v), out)
	case uint64:
		n += valuint642cbor(v, out)
	case float32:
		n += valfloat322cbor(v, out)
	case float64:
		n += valfloat642cbor(v, out)
	case []byte:
		n += valbytes2cbor(v, out)
	case string:
		n += valtext2cbor(v, out)

	case json.Number:
		if isnegative(v) {
			vi, err := strconv.ParseInt(string(v), 10, 64)
			if err != nil {
				panic(err)
			}
			n += valint642cbor(vi, out)
		} else {
			vu, err := strconv.ParseUint(string(v), 10, 64)
			if err != nil {
				panic(err)
			}
			n += valuint642cbor(vu, out)
		}

	case []interface{}:
		n += valarray2cbor(v, out, config)

	case [][2]interface{}:
		n += valmap2cbor(v, out, config)

	case map[string]interface{}:
		mkeys := config.mkeysh.getmkeys(len(v))

		if config.ct == LengthPrefix {
			n += valuint642cbor(uint64(len(v)), out[n:])
			out[n-1] = (out[n-1] & 0x1f) | cborType5 // fix the type
			for _, key := range mkeys.sortProps1(v) {
				value := v[key]
				n += valtext2cbor(key, out[n:])
				n += value2cbor(value, out[n:], config)
			}
		} else if config.ct == Stream {
			n += mapStart(out[n:])
			for _, key := range mkeys.sortProps1(v) {
				value := v[key]
				n += valtext2cbor(key, out[n:])
				n += value2cbor(value, out[n:], config)
			}
			n += breakStop(out[n:])
		}
		config.mkeysh.putmkeys(mkeys)

	// simple types
	case CborUndefined:
		n += valundefined2cbor(out)
	// tagged encoding
	case time.Time: // tag-0
		n += valtime2cbor(v, out, config)
	case CborTagEpoch: // tag-1
		n += valtime2cbor(v, out, config)
	case CborTagEpochMicro: // tag-1
		n += valtime2cbor(v, out, config)
	case *big.Int: // tag-2 (positive) or tag-3 (negative)
		n += valbignum2cbor(v, out, config)
	case CborTagFraction: // tag-4
		n += valdecimal2cbor(v, out, config)
	case CborTagFloat: // tag-5
		n += valbigfloat2cbor(v, out, config)
	case CborTagBytes: // tag-24
		n += valcbor2cbor(v, out)
	case *regexp.Regexp: // tag-35
		n += valregexp2cbor(v, out)
	case CborTagPrefix: // tag-55799
		n += valcborprefix2cbor(v, out)
	default:
		panic(fmt.Errorf("cbor encode unknownType %T", v))
	}
	return n
}

func tag2cbor(tag uint64, buf []byte) int {
	n := valuint642cbor(tag, buf)
	buf[0] = (buf[0] & 0x1f) | cborType6 // fix the type as tag.
	return n
}

func cborNull(buf []byte) int {
	buf[0] = cborHdr(cborType7, cborSimpleTypeNil)
	return 1
}

func cborTrue(buf []byte) int {
	buf[0] = cborHdr(cborType7, cborSimpleTypeTrue)
	return 1
}

func cborFalse(buf []byte) int {
	buf[0] = cborHdr(cborType7, cborSimpleTypeFalse)
	return 1
}

func valuint82cbor(item byte, buf []byte) int {
	if item <= CborMaxSmallInt {
		buf[0] = cborHdr(cborType0, item) // 0..23
		return 1
	}
	buf[0] = cborHdr(cborType0, cborInfo24)
	buf[1] = item // 24..255
	return 2
}

func valint82cbor(item int8, buf []byte) int {
	if item > CborMaxSmallInt {
		buf[0] = cborHdr(cborType0, cborInfo24)
		buf[1] = byte(item) // 24..127
		return 2
	} else if item < -CborMaxSmallInt {
		buf[0] = cborHdr(cborType1, cborInfo24)
		buf[1] = byte(-(item + 1)) // -128..-24
		return 2
	} else if item < 0 {
		buf[0] = cborHdr(cborType1, byte(-(item + 1))) // -23..-1
		return 1
	}
	buf[0] = cborHdr(cborType0, byte(item)) // 0..23
	return 1
}

func valuint162cbor(item uint16, buf []byte) int {
	if item < 256 {
		return valuint82cbor(byte(item), buf)
	}
	buf[0] = cborHdr(cborType0, cborInfo25)
	binary.BigEndian.PutUint16(buf[1:], item) // 256..65535
	return 3
}

func valint162cbor(item int16, buf []byte) int {
	if item > 127 {
		if item < 256 {
			buf[0] = cborHdr(cborType0, cborInfo24)
			buf[1] = byte(item) // 128..255
			return 2
		}
		buf[0] = cborHdr(cborType0, cborInfo25)
		binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 256..32767
		return 3

	} else if item < -128 {
		if item > -256 {
			buf[0] = cborHdr(cborType1, cborInfo24)
			buf[1] = byte(-(item + 1)) // -255..-129
			return 2
		}
		buf[0] = cborHdr(cborType1, cborInfo25) // -32768..-256
		binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
		return 3
	}
	return valint82cbor(int8(item), buf)
}

func valuint322cbor(item uint32, buf []byte) int {
	if item < 65536 {
		return valuint162cbor(uint16(item), buf) // 0..65535
	}
	buf[0] = cborHdr(cborType0, cborInfo26)
	binary.BigEndian.PutUint32(buf[1:], item) // 65536 to 4294967295
	return 5
}

func valint322cbor(item int32, buf []byte) int {
	if item > 32767 {
		if item < 65536 {
			buf[0] = cborHdr(cborType0, cborInfo25)
			binary.BigEndian.PutUint16(buf[1:], uint16(item)) // 32768..65535
			return 3
		}
		buf[0] = cborHdr(cborType0, cborInfo26) // 65536 to 2147483647
		binary.BigEndian.PutUint32(buf[1:], uint32(item))
		return 5

	} else if item < -32768 {
		if item > -65536 {
			buf[0] = cborHdr(cborType1, cborInfo25) // -65535..-32769
			binary.BigEndian.PutUint16(buf[1:], uint16(-(item + 1)))
			return 3
		}
		buf[0] = cborHdr(cborType1, cborInfo26) // -2147483648..-65536
		binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
		return 5
	}
	return valint162cbor(int16(item), buf)
}

func valuint642cbor(item uint64, buf []byte) int {
	if item < 4294967296 {
		return valuint322cbor(uint32(item), buf) // 0..4294967295
	}
	// 4294967296 .. 18446744073709551615
	buf[0] = cborHdr(cborType0, cborInfo27)
	binary.BigEndian.PutUint64(buf[1:], item)
	return 9
}

func valint642cbor(item int64, buf []byte) int {
	if item > 2147483647 {
		if item < 4294967296 {
			buf[0] = cborHdr(cborType0, cborInfo26) // 2147483647..4294967296
			binary.BigEndian.PutUint32(buf[1:], uint32(item))
			return 5
		}
		// 4294967296..9223372036854775807
		buf[0] = cborHdr(cborType0, cborInfo27)
		binary.BigEndian.PutUint64(buf[1:], uint64(item))
		return 9

	} else if item < -2147483648 {
		if item > -4294967296 {
			// -4294967295..-2147483649
			buf[0] = cborHdr(cborType1, cborInfo26)
			binary.BigEndian.PutUint32(buf[1:], uint32(-(item + 1)))
			return 5
		}
		// -9223372036854775808..-4294967296
		buf[0] = cborHdr(cborType1, cborInfo27)
		binary.BigEndian.PutUint64(buf[1:], uint64(-(item + 1)))
		return 9
	}
	return valint322cbor(int32(item), buf)
}

func valfloat322cbor(item float32, buf []byte) int {
	buf[0] = cborHdr(cborType7, cborFlt32)
	binary.BigEndian.PutUint32(buf[1:], math.Float32bits(item))
	return 5
}

func valfloat642cbor(item float64, buf []byte) int {
	buf[0] = cborHdr(cborType7, cborFlt64)
	binary.BigEndian.PutUint64(buf[1:], math.Float64bits(item))
	return 9
}

func valbytes2cbor(item []byte, buf []byte) int {
	n := valuint642cbor(uint64(len(item)), buf)
	buf[0] = (buf[0] & 0x1f) | cborType2 // fix the type from type0->type2
	copy(buf[n:], item)
	return n + len(item)
}

func bytesStart(buf []byte) int {
	// indefinite chunks of byte string
	buf[0] = cborHdr(cborType2, byte(cborIndefiniteLength))
	return 1
}

func valtext2cbor(item string, buf []byte) int {
	n := valbytes2cbor(str2bytes(item), buf)
	buf[0] = (buf[0] & 0x1f) | cborType3 // fix the type from type2->type3
	return n
}

func textStart(buf []byte) int {
	// indefinite chunks of text
	buf[0] = cborHdr(cborType3, byte(cborIndefiniteLength))
	return 1
}

func valarray2cbor(items []interface{}, buf []byte, config *Config) int {
	if config.ct == LengthPrefix {
		n := valuint642cbor(uint64(len(items)), buf)
		buf[0] = (buf[0] & 0x1f) | cborType4 // fix the type from type0->type4
		n += arrayitems2cbor(items, buf[n:], config)
		return n
	}
	// Stream encoding
	n := arrayStart(buf)
	n += arrayitems2cbor(items, buf[n:], config)
	n += breakStop(buf[n:])
	return n
}

func arrayitems2cbor(items []interface{}, buf []byte, config *Config) int {
	n := 0
	for _, item := range items {
		n += value2cbor(item, buf[n:], config)
	}
	return n
}

func arrayStart(buf []byte) int {
	// indefinite length array
	buf[0] = cborHdr(cborType4, byte(cborIndefiniteLength))
	return 1
}

func valmap2cbor(items [][2]interface{}, buf []byte, config *Config) int {
	if config.ct == LengthPrefix {
		n := valuint642cbor(uint64(len(items)), buf)
		buf[0] = (buf[0] & 0x1f) | cborType5 // fix the type from type0->type5
		n += mapl2cbor(items, buf[n:], config)
		return n
	}
	// Stream encoding
	n := mapStart(buf)
	n += mapl2cbor(items, buf[n:], config)
	n += breakStop(buf[n:])
	return n
}

func mapl2cbor(items [][2]interface{}, buf []byte, config *Config) int {
	n := 0
	for _, item := range items {
		n += value2cbor(item[0], buf[n:], config)
		n += value2cbor(item[1], buf[n:], config)
	}
	return n
}

func mapStart(buf []byte) int {
	// indefinite length map
	buf[0] = cborHdr(cborType5, byte(cborIndefiniteLength))
	return 1
}

func breakStop(buf []byte) int {
	// break stop for indefinite array or map
	buf[0] = cborHdr(cborType7, byte(cborItemBreak))
	return 1
}

func valundefined2cbor(buf []byte) int {
	buf[0] = cborHdr(cborType7, cborSimpleUndefined)
	return 1
}

func simpletypeToCbor(typcode byte, buf []byte) int {
	if typcode < 32 {
		buf[0] = cborHdr(cborType7, typcode)
		return 1
	}
	buf[0] = cborHdr(cborType7, cborSimpleTypeByte)
	buf[1] = typcode
	return 2
}

//---- encode tags

func valtime2cbor(dt interface{}, buf []byte, config *Config) int {
	n := 0
	switch v := dt.(type) {
	case time.Time: // rfc3339, as refined by section 3.3 rfc4287
		item := v.Format(time.RFC3339)
		n += tag2cbor(tagDateTime, buf)
		n += value2cbor(item, buf[n:], config)
	case CborTagEpoch:
		n += tag2cbor(tagEpoch, buf)
		n += valint642cbor(int64(v), buf[n:])
	case CborTagEpochMicro:
		n += tag2cbor(tagEpoch, buf)
		n += valfloat642cbor(float64(v), buf[n:])
	}
	return n
}

func valbignum2cbor(num *big.Int, buf []byte, config *Config) int {
	n := 0
	bytes := num.Bytes()
	if num.Sign() < 0 {
		n += tag2cbor(tagNegBignum, buf)
	} else {
		n += tag2cbor(tagPosBignum, buf)
	}
	n += valbytes2cbor(bytes, buf[n:])
	return n
}

func valdecimal2cbor(v CborTagFraction, buf []byte, config *Config) int {
	var item interface{}

	n := tag2cbor(tagDecimalFraction, buf)
	if config.ct == LengthPrefix {
		m := valuint642cbor(uint64(len(v)), buf[n:])
		buf[n] = (buf[n] & 0x1f) | cborType4 // fix the type from type0->type4
		n += m
		for _, item = range v {
			n += value2cbor(item, buf[n:], config)
		}
		return n
	}
	n += arrayStart(buf[n:])
	for _, item = range v {
		n += value2cbor(item, buf[n:], config)
	}
	n += breakStop(buf[n:])
	return n
}

func valbigfloat2cbor(v CborTagFloat, buf []byte, config *Config) int {
	var item interface{}

	n := tag2cbor(tagBigFloat, buf)
	if config.ct == LengthPrefix {
		m := valuint642cbor(uint64(len(v)), buf[n:])
		buf[n] = (buf[n] & 0x1f) | cborType4 // fix the type from type0->type4
		n += m
		for _, item = range v {
			n += value2cbor(item, buf[n:], config)
		}
		return n
	}
	n += arrayStart(buf[n:])
	for _, item = range v {
		n += value2cbor(item, buf[n:], config)
	}
	n += breakStop(buf[n:])
	return n
}

func valcbor2cbor(item, buf []byte) int {
	n := tag2cbor(tagCborEnc, buf)
	n += valbytes2cbor(item, buf[n:])
	return n
}

func valregexp2cbor(item *regexp.Regexp, buf []byte) int {
	n := tag2cbor(tagRegexp, buf)
	n += valtext2cbor(item.String(), buf[n:])
	return n
}

func valcborprefix2cbor(item, buf []byte) int {
	n := tag2cbor(tagCborPrefix, buf)
	n += valbytes2cbor(item, buf[n:])
	return n
}
