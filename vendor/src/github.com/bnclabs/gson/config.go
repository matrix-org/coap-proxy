package gson

import "bytes"
import "fmt"
import "math/big"
import "encoding/json"

import "golang.org/x/text/collate"

// NumberKind how to treat numbers.
type NumberKind byte

const (
	// FloatNumber to treat number as float64.
	FloatNumber NumberKind = iota + 1

	// SmartNumber to treat number as either integer or fall back to float64.
	SmartNumber
)

// MaxKeys maximum number of keys allowed in a property object. Affects
// memory pool. Changing this value will affect all new configuration
// instances.
var MaxKeys = 1024

// MaxStringLen maximum length of string value inside json document.
// Affects memory pool. Changing this value will affect all new
// configuration instances.
var MaxStringLen = 1024 * 1024

// MaxCollateLen maximum length of collated value. Affects memory pool.
// Changing this value will affect all new configuration instances.
var MaxCollateLen = 1024 * 1024

type memConfig struct {
	strlen  int // maximum length of string value inside JSON document
	numkeys int // maximum number of keys that a property object can have
	itemlen int // maximum length of collated value.
	ptrlen  int // maximum length of json-pointer can take
}

// Config is the root object to access all transformations and APIs
// exported by this package. Before calling any of the config-methods,
// make sure to initialize them with desired settings.
//
// NOTE: Config objects are immutable.
type Config struct {
	nk      NumberKind
	bufferh *bufferhead
	mkeysh  *mkeyshead
	kvh     *kvhead

	cborConfig
	jsonConfig
	collateConfig
	jptrConfig
	memConfig
}

// NewDefaultConfig return a new configuration with default settings:
//		+FloatNumber        +Stream
//		+UnicodeSpace       -strict
//		+doMissing          -arrayLenPrefix     +propertyLenPrefix
//		MaxJsonpointerLen   MaxKeys
//		MaxStringLen
//		MaxCollateLen
// Several methods are available to change configuration parameters.
func NewDefaultConfig() *Config {
	config := &Config{
		nk:      FloatNumber,
		bufferh: &bufferhead{},
		mkeysh:  &mkeyshead{},
		kvh:     &kvhead{},

		cborConfig: cborConfig{
			ct: Stream,
		},
		jsonConfig: jsonConfig{
			ws:     UnicodeSpace,
			strict: false,
		},
		collateConfig: collateConfig{
			doMissing:         true,
			arrayLenPrefix:    false,
			propertyLenPrefix: true,
		},
		memConfig: memConfig{
			strlen:  MaxStringLen,
			numkeys: MaxKeys,
			itemlen: MaxCollateLen,
			ptrlen:  MaxJsonpointerLen,
		},
	}
	config = config.SetJptrlen(MaxJsonpointerLen)
	return config.init()
}

func (config *Config) init() *Config {
	// collateConfig
	config.buf = bytes.NewBuffer(make([]byte, 0, 1024)) // start with 1K
	config.enc = json.NewEncoder(config.buf)
	config.zf = big.NewFloat(0)
	config.zf.SetPrec(64)
	config.tcltbuffer = &collate.Buffer{}
	return config
}

// SetNumberKind configure to interpret number values.
func (config Config) SetNumberKind(nk NumberKind) *Config {
	config.nk = nk
	return &config
}

// SetContainerEncoding configure to encode / decode cbor
// arrays and maps.
func (config Config) SetContainerEncoding(ct ContainerEncoding) *Config {
	config.ct = ct
	return &config
}

// SetSpaceKind setting to interpret whitespaces in json text.
func (config Config) SetSpaceKind(ws SpaceKind) *Config {
	config.ws = ws
	return &config
}

// SetStrict setting to enforce strict transforms to and from JSON.
// If set to true,
//   a. IntNumber configuration float numbers in JSON text still are parsed.
//   b. Use golang stdlib encoding/json for transforming strings to JSON.
func (config Config) SetStrict(what bool) *Config {
	config.strict = what
	return &config
}

// SortbyArrayLen setting to sort array of smaller-size before larger ones.
func (config Config) SortbyArrayLen(what bool) *Config {
	config.arrayLenPrefix = what
	return &config
}

// SortbyPropertyLen setting to sort properties of smaller size before
// larger ones.
func (config Config) SortbyPropertyLen(what bool) *Config {
	config.propertyLenPrefix = what
	return &config
}

// UseMissing setting to use TypeMissing collation.
func (config Config) UseMissing(what bool) *Config {
	config.doMissing = what
	return &config
}

// SetMaxkeys configure to set the maximum number of keys
// allowed in property item.
func (config Config) SetMaxkeys(n int) *Config {
	config.numkeys = n
	return config.init()
}

// SetTextCollator for string type. If collator is not set, strings will
// be treated as byte-array and compared as such.
func (config Config) SetTextCollator(collator *collate.Collator) *Config {
	config.textcollator = collator
	return &config
}

// ResetPools configure a new set of pools with specified size, instead
// of using the default size: MaxStringLen, MaxKeys, MaxCollateLen, and,
// MaxJsonpointerLen.
//	 strlen  - maximum length of string value inside JSON document
//	 numkeys - maximum number of keys that a property object can have
//	 itemlen - maximum length of collated value.
//	 ptrlen  - maximum possible length of json-pointer.
func (config Config) ResetPools(strlen, numkeys, itemlen, ptrlen int) *Config {
	config.memConfig = memConfig{
		strlen: strlen, numkeys: numkeys, itemlen: itemlen, ptrlen: ptrlen,
	}
	return config.init()
}

// NewCbor factory to create a new Cbor instance. If buffer is nil,
// a new buffer of 128 byte capacity will be allocated. Cbor object can
// be re-used after a Reset() call.
func (config *Config) NewCbor(buffer []byte) *Cbor {
	if buffer == nil {
		buffer = make([]byte, 0, 128)
	}
	if len(buffer) == 0 && cap(buffer) < 128 {
		panic("cbor buffer should atleast be 128 bytes")
	}
	return &Cbor{config: config, data: buffer, n: len(buffer)}
}

// NewJson factory to create a new Json instance. If buffer is nil,
// a new buffer of 128 byte capacity will be allocated. Json object can
// be re-used after a Reset() call.
func (config *Config) NewJson(buffer []byte) *Json {
	if buffer == nil {
		buffer = make([]byte, 0, 128)
	}
	if len(buffer) == 0 && cap(buffer) < 128 {
		panic("json buffer should atleast be 128 bytes")
	}
	return &Json{config: config, data: buffer, n: len(buffer)}
}

// NewCollate factor to create a new Collate instance. If buffer is nil,
// a new buffer of 128 byte capacity will be allocated. Collate object can
// be re-used after a Reset() call.
func (config *Config) NewCollate(buffer []byte) *Collate {
	if buffer == nil {
		buffer = make([]byte, 0, 128)
	}
	if len(buffer) == 0 && cap(buffer) < 128 {
		panic("collate buffer should atleast be 128 bytes")
	}
	return &Collate{config: config, data: buffer, n: len(buffer)}
}

// NewValue factory to create a new Value instance. Value instances are
// immutable, and can be used and re-used any number of times.
func (config *Config) NewValue(value interface{}) *Value {
	return &Value{config: config, data: value}
}

// NewJsonpointer create a instance of Jsonpointer.
func (config *Config) NewJsonpointer(path string) *Jsonpointer {
	if len(path) > config.jptrMaxlen {
		panic("jsonpointer path exceeds configured length")
	}
	jptr := &Jsonpointer{
		config:   config,
		path:     make([]byte, config.jptrMaxlen+16),
		segments: make([][]byte, config.jptrMaxseg),
	}
	for i := 0; i < config.jptrMaxseg; i++ {
		jptr.segments[i] = make([]byte, 0, 16)
	}
	n := copy(jptr.path, path)
	jptr.path = jptr.path[:n]
	return jptr
}

func (config *Config) String() string {
	return fmt.Sprintf(
		"nk:%v, ws:%v, ct:%v, arrayLenPrefix:%v, "+
			"propertyLenPrefix:%v, doMissing:%v",
		config.nk, config.ws, config.ct,
		config.arrayLenPrefix, config.propertyLenPrefix,
		config.doMissing)
}

func (nk NumberKind) String() string {
	switch nk {
	case SmartNumber:
		return "SmartNumber"
	case FloatNumber:
		return "FloatNumber"
	default:
		panic("new number-kind")
	}
}

func (ct ContainerEncoding) String() string {
	switch ct {
	case LengthPrefix:
		return "LengthPrefix"
	case Stream:
		return "Stream"
	default:
		panic("new space-kind")
	}
}
