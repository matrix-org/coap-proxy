package gson

import "encoding/json"
import "math/big"
import "bytes"

import "golang.org/x/text/collate"

// Collation order for supported types. Applications desiring different
// ordering between types can initialize these byte values before
// instantiating a config object.
var (
	Terminator  byte = 0
	TypeMissing byte = 49
	TypeNull    byte = 50
	TypeFalse   byte = 60
	TypeTrue    byte = 70
	TypeNumber  byte = 80
	TypeString  byte = 90
	TypeLength  byte = 100
	TypeArray   byte = 110
	TypeObj     byte = 120
	TypeBinary  byte = 130
)

// Missing denotes a special type for an item that evaluates to _nothing_.
type Missing string

// MissingLiteral is undocumented, for now.
const MissingLiteral = Missing("~[]{}falsenilNA~")

type collateConfig struct {
	doMissing         bool // handle missing values (for N1QL)
	arrayLenPrefix    bool // first sort arrays based on its length
	propertyLenPrefix bool // first sort properties based on length
	enc               *json.Encoder
	buf               *bytes.Buffer
	zf                *big.Float
	tcltbuffer        *collate.Buffer
	textcollator      *collate.Collator
}

// Collate abstraction for value encoded into binary-collation.
type Collate struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return the byte-slice holding the collated data.
func (clt *Collate) Bytes() []byte {
	return clt.data[:clt.n]
}

// Reset overwrite buffer with data, or if data is nil,
// reset buffer to zero-length.
func (clt *Collate) Reset(data []byte) *Collate {
	if data == nil {
		clt.n = 0
		return clt
	}
	clt.data, clt.n = data, len(data)
	return clt
}

// Tovalue convert to golang native value.
func (clt *Collate) Tovalue() interface{} {
	if clt.n == 0 {
		panic("cannot convert empty binary-collate to value")
	}
	value, _ /*rb*/ := collate2gson(clt.data[:clt.n], clt.config)
	return value
}

// Tojson convert to json encoded text.
func (clt *Collate) Tojson(jsn *Json) *Json {
	if clt.n == 0 {
		panic("cannot convert empty binary-collate to json")
	}
	in, out := clt.data[:clt.n], jsn.data[jsn.n:cap(jsn.data)]
	_ /*rb*/, m /*wb*/ := collate2json(in, out, clt.config)
	jsn.n += m
	return jsn
}

// Tocbor convert to cbor encoded value.
func (clt *Collate) Tocbor(cbr *Cbor) *Cbor {
	if clt.n == 0 {
		panic("cannot convert empty binary-collate to cbor")
	}
	in, out := clt.data[:clt.n], cbr.data[cbr.n:cap(cbr.data)]
	_ /*rb*/, m /*wb*/ := collate2cbor(in, out, clt.config)
	cbr.n += m
	return cbr
}

// Equal checks wether n is MissingLiteral
func (m Missing) Equal(n string) bool {
	s := string(m)
	if len(n) == len(s) && n[0] == '~' && n[1] == '[' {
		return s == n
	}
	return false
}
