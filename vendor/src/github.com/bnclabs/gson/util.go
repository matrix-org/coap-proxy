package gson

import "fmt"
import "reflect"
import "unsafe"
import "unicode"
import "strconv"
import "encoding/json"

func fixbuffer(buffer []byte, size int64) []byte {
	if buffer == nil || int64(cap(buffer)) < size {
		buffer = make([]byte, size)
	}
	return buffer[:size]
}

func bytes2str(bytes []byte) string {
	if bytes == nil {
		return ""
	}
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	st := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}
	return *(*string)(unsafe.Pointer(st))
}

func str2bytes(str string) []byte {
	if str == "" {
		return nil
	}
	st := (*reflect.StringHeader)(unsafe.Pointer(&str))
	sl := &reflect.SliceHeader{Data: st.Data, Len: st.Len, Cap: st.Len}
	return *(*[]byte)(unsafe.Pointer(sl))
}

// CborMap2golangMap used by validation tools.
// Transforms [][2]interface{} to map[string]interface{} that is required for
// converting golang to cbor and vice-versa.
func CborMap2golangMap(value interface{}) interface{} {
	switch items := value.(type) {
	case []interface{}:
		for i, item := range items {
			items[i] = CborMap2golangMap(item)
		}
		return items
	case [][2]interface{}:
		m := make(map[string]interface{})
		for _, item := range items {
			m[item[0].(string)] = CborMap2golangMap(item[1])
		}
		return m
	}
	return value
}

// GolangMap2cborMap used by validation tools.
// Transforms map[string]interface{} to [][2]interface{}
// that is required for converting golang to cbor and vice-versa.
func GolangMap2cborMap(value interface{}) interface{} {
	switch items := value.(type) {
	case []interface{}:
		for i, item := range items {
			items[i] = GolangMap2cborMap(item)
		}
		return items
	case map[string]interface{}:
		sl := make([][2]interface{}, 0, len(items))
		for k, v := range items {
			sl = append(sl, [2]interface{}{k, GolangMap2cborMap(v)})
		}
		return sl
	}
	return value
}

// Fixtojson used by validation tools.
func Fixtojson(config *Config, val interface{}) interface{} {
	var err error

	if val == nil {
		return nil
	}

	if s, ok := val.(json.Number); ok {
		val, err = strconv.ParseFloat(string(s), 64)
		if err != nil {
			panic(err)
		}
	}

	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v
	case int8:
		return float64(v)
	case uint8:
		return float64(v)
	case int16:
		return float64(v)
	case uint16:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	case int64:
		if config.nk == FloatNumber {
			return float64(v)
		} else if config.nk == SmartNumber {
			return v
		}
	case uint64:
		if config.nk == FloatNumber {
			return float64(v)
		} else if config.nk == SmartNumber {
			return v
		}
	case float32:
		return float64(v)
	case float64:
		return v
	case map[string]interface{}:
		for key, x := range v {
			v[key] = Fixtojson(config, x)
		}
		return v
	case [][2]interface{}:
		m := make(map[string]interface{})
		for _, item := range v {
			m[item[0].(string)] = Fixtojson(config, item[1])
		}
		return m
	case []interface{}:
		for i, x := range v {
			v[i] = Fixtojson(config, x)
		}
		return v
	}
	panic(fmt.Errorf("unreachable code, unexpected %T", val))
}

func collateFloat64(value float64, code []byte) int {
	var num [64]byte
	bs := strconv.AppendFloat(num[:0], value, 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateUint64(value uint64, code []byte, config *Config) int {
	var fltx [64]byte
	if value > 9007199254740991 {
		config.zf.SetUint64(value)
		bs := config.zf.Append(fltx[:0], 'e', -1)
		return collateFloat(bs, code)
	}
	bs := strconv.AppendFloat(fltx[:0], float64(value), 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateInt64(value int64, code []byte, config *Config) int {
	var fltx [64]byte
	if value > 9007199254740991 || value < -9007199254740992 {
		config.zf.SetInt64(value)
		bs := config.zf.Append(fltx[:0], 'e', -1)
		return collateFloat(bs, code)
	}
	bs := strconv.AppendFloat(fltx[:0], float64(value), 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateJsonNumber(value string, code []byte, config *Config) int {
	var fltx [64]byte

	if len(value) > len("9007199254740991") { // may need higher precision
		zf, _, err := config.zf.Parse(value, 0)
		if err != nil {
			panic(fmt.Errorf("collateJsonNumber(%q): %v", value, err))
		}
		bs := zf.Append(fltx[:0], 'e', -1)
		return collateFloat(bs, code)
	}

	num53, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Errorf("strconv.Atoi(%q): %v", value, err))
	}
	bs := strconv.AppendFloat(fltx[:0], float64(num53), 'e', -1, 64)
	return collateFloat(bs, code)
}

func collateString(str string, code []byte, config *Config) (n int) {
	if config.doMissing && MissingLiteral.Equal(str) {
		code[0], code[1] = TypeMissing, Terminator
		return 2
	}
	strcode := str2bytes(str)
	if config.textcollator != nil {
		config.tcltbuffer.Reset()
		strcode = config.textcollator.Key(config.tcltbuffer, strcode)
		strcode = strcode[:len(strcode)-1] // return text is null terminated
	}

	code[n] = TypeString
	n++
	n += suffixEncodeString(strcode, code[n:])
	code[n] = Terminator
	n++
	return n
}

func collate2String(code []byte, str []byte) ([]byte, int) {
	x, y := suffixDecodeString(code, str)
	return str[:y], x
}

func collated2Number(code []byte, nk NumberKind) (uint64, int64, float64, int) {
	var mantissa, scratch [64]byte
	_, y := collated2Float(code, scratch[:])
	if nk == SmartNumber {
		dotat, _, exp, mant := parseFloat(scratch[:y], mantissa[:0])
		if exp >= 15 {
			x := len(mant) - dotat
			for i := 0; i < (exp - x); i++ {
				mant = append(mant, '0')
			}
			if mant[0] == '+' {
				mant = mant[1:]
			}
			ui, err := strconv.ParseUint(bytes2str(mant), 10, 64)
			if err == nil {
				return ui, 0, 0, 1
			}
			i, err := strconv.ParseInt(bytes2str(mant), 10, 64)
			if err == nil {
				return 0, i, 0, 2
			}
			panic(fmt.Errorf("unexpected number %v", string(scratch[:y])))
		}
	}
	f, err := strconv.ParseFloat(bytes2str(scratch[:y]), 64)
	if err != nil {
		panic(err)
	}
	return 0, 0, f, 3

}

func parseFloat(text []byte, m []byte) (int, int, int, []byte) {
	var err error
	var exp int

	dotat, expat := -1, -1
	for i, ch := range text {
		if ch == '.' {
			dotat = i
		} else if ch == 'e' || ch == 'E' {
			expat = i
		} else if expat > -1 && ch == '+' {
			expat = i
		} else if expat == -1 {
			m = append(m, ch)
		}
	}
	if expat > -1 {
		exp, err = strconv.Atoi(bytes2str(text[expat+1:]))
		if err != nil {
			panic(err)
		}
	}
	return dotat, expat, exp, m
}

func isnegative(s json.Number) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		} else if r == '-' {
			return true
		}
		break
	}
	return false
}

func collated2Json(code []byte, text []byte, nk NumberKind) int {
	var num [64]byte

	ui, i, f, what := collated2Number(code, nk)
	switch what {
	case 1:
		nm := strconv.AppendUint(num[:0], ui, 10)
		return copy(text, nm)
	case 2:
		nm := strconv.AppendInt(num[:0], i, 10)
		return copy(text, nm)
	case 3:
		nm := strconv.AppendFloat(num[:0], f, 'e', -1, 64)
		return copy(text, nm)
	}
	panic("unreachable code")
}

// NOTE: using built in sort incurs mem-allocation.
func sortStrings(strs []string) []string {
	for ln := len(strs) - 1; ; ln-- {
		changed := false
		for i := 0; i < ln; i++ {
			if strs[i] > strs[i+1] {
				strs[i], strs[i+1] = strs[i+1], strs[i]
				changed = true
			}
		}
		if changed == false {
			break
		}
	}
	return strs
}
