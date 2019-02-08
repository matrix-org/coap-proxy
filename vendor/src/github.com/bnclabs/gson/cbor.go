package gson

import "fmt"

// ContainerEncoding method to encode arrays and maps into cbor.
type ContainerEncoding byte

const (
	// LengthPrefix to encode number of items in the collection type.
	LengthPrefix ContainerEncoding = iota + 1

	// Stream to encode collection types as indefinite sequence of items.
	Stream
)

type cborConfig struct {
	ct ContainerEncoding
}

const ( // major types (3 most significant bits in the first byte)
	cborType0 byte = iota << 5 // unsigned integer
	cborType1                  // negative integer
	cborType2                  // byte string
	cborType3                  // text string
	cborType4                  // array
	cborType5                  // map
	cborType6                  // tagged data-item
	cborType7                  // floating-point, simple-types and break-stop
)

// CborMaxSmallInt maximum integer value that can be stored as associative value
// for cborType0 or cborType1.
const CborMaxSmallInt = 23

const ( // for cborType0 cborType1 (5 least significant bits in the first byte)
	// 0..23 actual value
	cborInfo24 byte = iota + 24 // followed by 1-byte data-item
	cborInfo25                  // followed by 2-byte data-item
	cborInfo26                  // followed by 4-byte data-item
	cborInfo27                  // followed by 8-byte data-item
	// 28..30 reserved
	cborIndefiniteLength = 31 // for cborType2/cborType3/cborType4/cborType5
)

// CborIndefinite code, {cborType2,Type3,Type4,Type5}/cborIndefiniteLength
type CborIndefinite byte

const ( // simple types for cborType7
	// 0..19 unassigned
	cborSimpleTypeFalse byte = iota + 20 // encodes nil type
	cborSimpleTypeTrue
	cborSimpleTypeNil
	cborSimpleUndefined
	cborSimpleTypeByte // the actual type in next byte 32..255
	cborFlt16          // IEEE 754 Half-Precision Float
	cborFlt32          // IEEE 754 Single-Precision Float
	cborFlt64          // IEEE 754 Double-Precision Float
	// 28..30 reserved
	cborItemBreak = 31 // stop-code for indefinite-length items
)

// CborUndefined simple type, cborType7/cborSimpleUndefined
type CborUndefined byte

// CborBreakStop code, cborType7/cborItemBreak
type CborBreakStop byte

const ( // pre-defined tag values
	tagDateTime        = iota // datetime as utf-8 string
	tagEpoch                  // datetime as +/- int or +/- float
	tagPosBignum              // as []bytes
	tagNegBignum              // as []bytes
	tagDecimalFraction        // decimal fraction as array of [2]num
	tagBigFloat               // as array of [2]num

	// unassigned 6..20

	tagBase64URL = iota + 15 // interpret []byte as base64 format
	tagBase64                // interpret []byte as base64 format
	tagBase16                // interpret []byte as base16 format

	tagCborEnc // embedd another CBOR message

	// unassigned 25..31

	tagURI          = iota + 22 // defined in rfc3986
	tagBase64URLEnc             // base64 encoded url as text strings
	tagBase64Enc                // base64 encoded byte-string as text strings
	tagRegexp                   // PCRE and ECMA262 regular expression
	tagMime                     // MIME defined by rfc2045

	// unassigned 38..55798

	tagCborPrefix = iota + 55783
	// unassigned 55800..
)

// CborTagEpoch codepoint-1, followed by int64 of seconds since
// 1970-01-01T00:00Z in UTC time.
type CborTagEpoch int64

// CborTagEpochMicro codepoint-1, followed by float64 of seconds/us since
// 1970-01-01T00:00Z in UTC time.
type CborTagEpochMicro float64

// CborTagFraction codepoint-4, followed by [2]int64{e,m} => m*(10**e).
type CborTagFraction [2]int64

// CborTagFloat codepoint-5, followed by [2]int64{e,m} => m*(2**e).
type CborTagFloat [2]int64

// CborTagBytes codepoint-24, bytes in cbor format.
type CborTagBytes []byte

// CborTagPrefix codepoint-5579, followed by byte-string.
type CborTagPrefix []byte

var brkstp = cborHdr(cborType7, cborItemBreak)

var hdrIndefiniteBytes = cborHdr(cborType2, cborIndefiniteLength)
var hdrIndefiniteText = cborHdr(cborType3, cborIndefiniteLength)
var hdrIndefiniteArray = cborHdr(cborType4, cborIndefiniteLength)
var hdrIndefiniteMap = cborHdr(cborType5, cborIndefiniteLength)

// Cbor encapsulates configuration and a cbor buffer. Use config
// object's NewCbor() method to Create new instance of Cbor.
// Map element in cbor encoding should have its keys sorted.
type Cbor struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return the byte-slice holding the CBOR data.
func (cbr *Cbor) Bytes() []byte {
	return cbr.data[:cbr.n]
}

// Reset overwrite buffer with data, or if data is nil,
// reset buffer to zero-length.
func (cbr *Cbor) Reset(data []byte) *Cbor {
	if data == nil {
		cbr.n = 0
		return cbr
	}
	cbr.data, cbr.n = data, len(data)
	return cbr
}

// Tovalue convert to golang native value.
func (cbr *Cbor) Tovalue() interface{} {
	if cbr.n == 0 {
		panic("cannot convert empty cbor bytes to value")
	}
	value, _ /*rb*/ := cbor2value(cbr.data[:cbr.n], cbr.config)
	return value
}

// Tojson convert to json encoded value.
func (cbr *Cbor) Tojson(jsn *Json) *Json {
	if cbr.n == 0 {
		panic("cannot convert empty cbor bytes to json")
	}
	in, out := cbr.data[:cbr.n], jsn.data[jsn.n:cap(jsn.data)]
	_ /*rb*/, m /*wb*/ := cbor2json(in, out, cbr.config)
	jsn.n += m
	return jsn
}

// Tocollate convert to binary-collation.
func (cbr *Cbor) Tocollate(clt *Collate) *Collate {
	if cbr.n == 0 {
		panic("cannot convert empty cbor bytes to binary-collation")
	}
	in, out := cbr.data[:cbr.n], clt.data[clt.n:cap(clt.data)]
	_ /*rb*/, m /*wb*/ := cbor2collate(in, out, cbr.config)
	clt.n += m
	return clt
}

// EncodeSmallint to encode tiny integers between -23..+23 into cbor buffer.
func (cbr *Cbor) EncodeSmallint(item int8) *Cbor {
	cbr.data = cbr.data[:cbr.n+1]
	if item < 0 {
		cbr.data[cbr.n] = cborHdr(cborType1, byte(-(item + 1))) // -23 to -1
	} else {
		cbr.data[cbr.n] = cborHdr(cborType0, byte(item)) // 0 to 23
	}
	cbr.n++
	return cbr
}

// EncodeSimpletype to encode simple type into cbor buffer.
// Code points 0..19 and 32..255 are un-assigned.
func (cbr *Cbor) EncodeSimpletype(typcode byte) *Cbor {
	cbr.n += simpletypeToCbor(typcode, cbr.data[cbr.n:cap(cbr.data)])
	cbr.data = cbr.data[:cbr.n]
	return cbr
}

// EncodeMapslice to encode key,value pairs into cbor buffer. Whether
// to encode them as indefinite-sequence of pairs, or as length prefixed
// pairs is decided by config.ContainerEncoding.
func (cbr *Cbor) EncodeMapslice(items [][2]interface{}) *Cbor {
	cbr.n += mapl2cbor(items, cbr.data[cbr.n:cap(cbr.data)], cbr.config)
	cbr.data = cbr.data[:cbr.n]
	return cbr
}

// EncodeBytechunks to encode several chunks of bytes as an
// indefinite-sequence of byte-blocks.
func (cbr *Cbor) EncodeBytechunks(chunks [][]byte) *Cbor {
	cbr.n += bytesStart(cbr.data[cbr.n:cap(cbr.data)])
	for _, chunk := range chunks {
		cbr.n += valbytes2cbor(chunk, cbr.data[cbr.n:cap(cbr.data)])
	}
	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[cbr.n] = cborType7 | cborItemBreak
	cbr.n++
	return cbr
}

// EncodeTextchunks to encode several chunks of text as an
// indefinite-sequence of byte-blocks.
func (cbr *Cbor) EncodeTextchunks(chunks []string) *Cbor {
	cbr.n += textStart(cbr.data[cbr.n:cap(cbr.data)])
	for _, chunk := range chunks {
		cbr.n += valtext2cbor(chunk, cbr.data[cbr.n:cap(cbr.data)])
	}
	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[cbr.n] = cborType7 | cborItemBreak
	cbr.n++
	return cbr
}

// Get field or nested field specified by json-pointer.
func (cbr *Cbor) Get(jptr *Jsonpointer, item *Cbor) *Cbor {
	config, segments := cbr.config, jptr.Segments()
	doc, itemb := cbr.Bytes(), item.data[:cap(item.data)]
	item.n = cborGet(doc, segments, itemb, config)
	return item
}

// Set field or nested field specified by json-pointer.
func (cbr *Cbor) Set(jptr *Jsonpointer, item, newdoc, old *Cbor) *Cbor {
	config, segments := cbr.config, jptr.Segments()
	if len(segments) == 0 {
		panic("cbor.Set(): empty pointer")
	}
	olditemb := old.data[:cap(old.data)]
	newdocb := newdoc.data[:cap(newdoc.data)]
	doc, itemb := cbr.Bytes(), item.Bytes()
	newdoc.n, old.n = cborSet(doc, segments, itemb, newdocb, olditemb, config)
	return newdoc
}

// Prepend item to the beginning of an array field specified by json-pointer.
func (cbr *Cbor) Prepend(jptr *Jsonpointer, item, newdoc *Cbor) *Cbor {
	config, segments := cbr.config, jptr.Segments()
	newdocb := newdoc.data[:cap(newdoc.data)]
	doc, itemb := cbr.Bytes(), item.Bytes()
	newdoc.n = cborPrepend(doc, segments, itemb, newdocb, config)
	return newdoc
}

// Append item at the end of an array field specified by json-pointer.
func (cbr *Cbor) Append(jptr *Jsonpointer, item, newdoc *Cbor) *Cbor {
	config, segments := cbr.config, jptr.Segments()
	newdocb := newdoc.data[:cap(newdoc.data)]
	doc, itemb := cbr.Bytes(), item.Bytes()
	newdoc.n = cborAppend(doc, segments, itemb, newdocb, config)
	return newdoc
}

// Delete field or nested field specified by json-pointer.
func (cbr *Cbor) Delete(jptr *Jsonpointer, newdoc, deleted *Cbor) *Cbor {
	config, segments := cbr.config, jptr.Segments()
	if len(segments) == 0 {
		panic("cbor.Set(): empty pointer")
	}
	newdocb := newdoc.data[:cap(newdoc.data)]
	doc, deletedb := cbr.Bytes(), deleted.data[:cap(deleted.data)]
	newdoc.n, deleted.n = cborDel(doc, segments, newdocb, deletedb, config)
	return newdoc
}

//---- help functions.

func cborMajor(b byte) byte {
	return b & 0xe0
}

func cborInfo(b byte) byte {
	return b & 0x1f
}

func cborHdr(major, info byte) byte {
	return (major & 0xe0) | (info & 0x1f)
}

func cborItem(doc []byte) (start, end int) {
	var ln, n int

	major, info := cborMajor(doc[0]), cborInfo(doc[0])
	switch major {
	case cborType0, cborType1:
		if info < cborInfo24 {
			return start, 1
		}
		switch info {
		case cborInfo24:
			return start, 2
		case cborInfo25:
			return start, 3
		case cborInfo26:
			return start, 5
		case cborInfo27:
			return start, 9
		default:
			panic(fmt.Sprintf("invalid major, info {%v,%v}", major, info))
		}

	case cborType2, cborType3:
		if info == cborIndefiniteLength {
			for end = 1; doc[end] != brkstp; end += n {
				_, n = cborItem(doc[end:])
			}
			return start, end + 1

		}
		ln, n = cborItemLength(doc)
		return start, start + n + ln

	case cborType4:
		if info == cborIndefiniteLength {
			for end = 1; doc[end] != brkstp; end += n {
				_, n = cborItem(doc[end:])
			}
			return start, end + 1
		}
		ln, end = cborItemLength(doc)
		for i := 0; i < ln; i++ {
			_, n = cborItem(doc[end:])
			end += n
		}
		return start, end

	case cborType5:
		if info == cborIndefiniteLength {
			for end = 1; doc[end] != brkstp; {
				_, n = cborItem(doc[end:])
				end += n
				_, n = cborItem(doc[end:])
				end += n
			}
			return start, end + 1
		}
		ln, end = cborItemLength(doc)
		for i := 0; i < ln; i++ {
			_, n = cborItem(doc[end:])
			end += n
			_, n = cborItem(doc[end:])
			end += n
		}
		return start, end

	case cborType6:
		_, end = cborItemLength(doc)
		_, n = cborItem(doc[end:])
		end += n
		return start, end

	case cborType7:
		if info < 23 {
			return start, 1
		}
		switch info {
		case cborSimpleTypeByte:
			return start, 2
		case cborFlt16:
			return start, 3
		case cborFlt32:
			return start, 5
		case cborFlt64:
			return start, 9
		default:
			panic(fmt.Sprintf("invalid major, info {%v,%v}", major, info))
		}
	}
	panic("unreachable code")
}
