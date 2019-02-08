package gson

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
	// AnsiSpace will skip white space characters defined by ANSI spec.
	AnsiSpace SpaceKind = iota + 1

	// UnicodeSpace will skip white space characters defined by Unicode spec.
	UnicodeSpace
)

type jsonConfig struct {
	// if `strict` use encoding/json for string conversion, else use custom
	// encoder that is memory optimized.
	strict bool
	ws     SpaceKind
}

// Json abstraction for value encoded as json text.
type Json struct {
	config *Config
	data   []byte
	n      int
}

// Bytes return reference to byte-slice of valid json-buffer.
func (jsn *Json) Bytes() []byte {
	return jsn.data[:jsn.n]
}

// Reset overwrite buffer with data, or if data is nil,
// reset buffer to zero-length.
func (jsn *Json) Reset(data []byte) *Json {
	if data == nil {
		jsn.n = 0
		return jsn
	}
	jsn.data, jsn.n = data, len(data)
	return jsn
}

// Tovalue parse json text to golang native value. Return remaining text.
func (jsn *Json) Tovalue() (*Json, interface{}) {
	remaining, value := json2value(bytes2str(jsn.data[:jsn.n]), jsn.config)
	if remaining != "" {
		return jsn.config.NewJson(str2bytes(remaining)), value
	}
	return nil, value
}

// Tovalues parse json text to one or more go native values.
func (jsn *Json) Tovalues() []interface{} {
	var values []interface{}
	var tok interface{}
	txt := bytes2str(jsn.data[:jsn.n])
	for len(txt) > 0 {
		txt, tok = json2value(txt, jsn.config)
		values = append(values, tok)
	}
	return values
}

// Tocbor convert json encoded value into cbor encoded binary string.
func (jsn *Json) Tocbor(cbr *Cbor) *Cbor {
	in, out := bytes2str(jsn.data[:jsn.n]), cbr.data[cbr.n:cap(cbr.data)]
	_ /*remning*/, m := json2cbor(in, out, jsn.config)
	cbr.n += m
	return cbr
}

// Tocollate convert json encoded value into binary-collation.
func (jsn *Json) Tocollate(clt *Collate) *Collate {
	in, out := bytes2str(jsn.data[:jsn.n]), clt.data[clt.n:cap(clt.data)]
	_ /*remn*/, m := json2collate(in, out, jsn.config)
	clt.n += m
	return clt
}

func (ws SpaceKind) String() string {
	switch ws {
	case AnsiSpace:
		return "AnsiSpace"
	case UnicodeSpace:
		return "UnicodeSpace"
	default:
		panic("new space-kind")
	}
}
