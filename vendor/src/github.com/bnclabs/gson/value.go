package gson

import "bytes"
import "sort"
import "fmt"

// Value abstractions for golang-native value.
type Value struct {
	config *Config
	data   interface{}
}

// Data returns the golang data.
func (val *Value) Data() interface{} {
	return val.data
}

// Tojson encode golang native value to json text.
func (val *Value) Tojson(jsn *Json) *Json {
	out := jsn.data[jsn.n:cap(jsn.data)]
	jsn.n += value2json(val.data, out, val.config)
	return jsn
}

// Tocbor encode golang native into cbor binary.
func (val *Value) Tocbor(cbr *Cbor) *Cbor {
	out := cbr.data[cbr.n:cap(cbr.data)]
	cbr.n += value2cbor(val.data, out, val.config)
	return cbr
}

// Tocollate encode golang native into binary-collation.
func (val *Value) Tocollate(clt *Collate) *Collate {
	out := clt.data[clt.n:cap(clt.data)]
	clt.n += gson2collate(val.data, out, val.config)
	return clt
}

// ListPointers all possible pointers in value.
func (val *Value) ListPointers(ptrs []string) []string {
	bufn := val.config.bufferh.getbuffer(val.config.ptrlen)
	ptrs = allpaths(val.data, ptrs, bufn.data[:0])
	ptrs = append(ptrs, "")
	val.config.bufferh.putbuffer(bufn)
	return ptrs
}

// Get field or nested field specified by json pointer.
func (val *Value) Get(jptr *Jsonpointer) (item interface{}) {
	return valGet(jptr.Segments(), val.data)
}

// Set field or nested field specified by json pointer. While
// `newval` is guaranteed to contain the `item`, `val` _may_ not be.
// Suggested usage,
//      val := config.NewValue([]interface{}{"hello"})
//      newval, _ = val.Set("/-", "world")
func (val *Value) Set(jptr *Jsonpointer, item interface{}) (newval, oldval interface{}) {
	return valSet(jptr.Segments(), val.data, item)
}

// Delete field or nested field specified by json pointer. While
// `newval` is guaranteed to be updated, `val` _may_ not be.
// Suggested usage,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Delete("/1")
func (val *Value) Delete(jptr *Jsonpointer) (newval, deleted interface{}) {
	return valDel(jptr.Segments(), val.data)
}

// Append item to end of an array pointed by json-pointer.
// returns `newval`, is guaranteed to be updated,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Append("", "welcome")
func (val *Value) Append(jptr *Jsonpointer, item interface{}) interface{} {
	return valAppend(jptr.Segments(), val.data, item)
}

// Prepend an item to the beginning of an array.
// returns `newval`, is guaranteed to be updated,
//      val := NewValue([]interface{}{"hello", "world"})
//      newval, _ = val.Append("", "welcome")
func (val *Value) Prepend(jptr *Jsonpointer, item interface{}) interface{} {
	return valPrepend(jptr.Segments(), val.data, item)
}

// Compare to value object.
func (val *Value) Compare(other *Value) int {
	return valuecompare(val.data, other.data)
}

func valuecompare(v1, v2 interface{}) int {
	t1, t2 := valueType(v1), valueType(v2)
	if t1 < t2 {
		return -1

	} else if t1 > t2 {
		return 1

	} else {
		switch t1 {
		case TypeNull:
			return 0

		case TypeFalse:
			if t2 == TypeTrue {
				return -1
			}
			return 0

		case TypeTrue:
			if t2 == TypeFalse {
				return 1
			}
			return 0

		case TypeNumber:
			cmp := cmpnumber(v1, v2)
			return cmp

		case TypeString:
			b1 := str2bytes(v1.(string))
			b2 := str2bytes(v2.(string))
			return bytes.Compare(b1, b2)

		case TypeArray:
			a1, a2 := v1.([]interface{}), v2.([]interface{})
			for i, item := range a1 {
				if i > len(a2)-1 {
					break
				} else if cmp := valuecompare(item, a2[i]); cmp != 0 {
					return cmp
				}
			}
			if len(a1) < len(a2) {
				return -1
			} else if len(a1) > len(a2) {
				return 1
			}
			return 0

		case TypeObj:
			o1 := v1.(map[string]interface{})
			o2 := v2.(map[string]interface{})
			switch {
			case len(o1) < len(o2):
				return -1
			case len(o1) > len(o2):
				return 1
			}

			keys1 := make([]string, 0)
			for key := range o1 {
				keys1 = append(keys1, key)
			}
			sort.Strings(keys1)
			keys2 := make([]string, 0)
			for key := range o2 {
				keys2 = append(keys2, key)
			}
			sort.Strings(keys2)

			for i, key1 := range keys1 {
				key2 := keys2[i]
				val1, val2 := o1[key1], o2[key2]
				switch {
				case key1 < key2:
					return -1
				case key1 > key2:
					return 1
				case key1 == key2:
					cmp := valuecompare(val1, val2)
					if cmp != 0 {
						return cmp
					}
				}
			}
			return 0
		}
	}
	panic(fmt.Errorf("unknown value type: %T", v1))
}

func valueType(data interface{}) byte {
	if data == nil {
		return TypeNull
	}
	switch v := data.(type) {
	case bool:
		if v == false {
			return TypeFalse
		}
		return TypeTrue
	case float64:
		return TypeNumber
	case int64, uint64, int, uint:
		return TypeNumber
	case string:
		return TypeString
	case []interface{}:
		return TypeArray
	case map[string]interface{}:
		return TypeObj
	}
	panic(fmt.Errorf("unknown value type: %T", data))
}

func cmpnumber(n1, n2 interface{}) int {
	switch v1 := n1.(type) {
	case float64:
		return float64t(v1).cmp(n2)
	case int:
		return int64t(v1).cmp(n2)
	case int64:
		return int64t(v1).cmp(n2)
	case uint:
		return uint64t(v1).cmp(n2)
	case uint64:
		return uint64t(v1).cmp(n2)
	}
	panic(fmt.Errorf("unknown value type: %T", n1))
}

type float64t float64

func (f float64t) cmp(other interface{}) int {
	var rc int
	switch {
	case float64(f) < float64(int64(f)):
		rc = -1
	case float64(f) > float64(int64(f)):
		rc = 1
	}

	switch val := other.(type) {
	case float64:
		if float64(f) < val {
			return -1
		} else if float64(f) == val {
			return 0
		}
		return 1

	case int64:
		if v := int64(f); v < val {
			return -1
		} else if v == val {
			return rc
		}
		return 1

	case uint64:
		if v := uint64(f); v < val {
			return -1
		} else if v == val {
			return rc
		}
		return 1
	}
	panic(fmt.Errorf("cannot handle type %T", other))
}

type int64t int64

func (f int64t) cmp(other interface{}) int {
	switch val := other.(type) {
	case float64:
		return -float64t(val).cmp(int64(f))
	case int:
		if int64(f) < int64(val) {
			return -1
		} else if int64(f) == int64(val) {
			return 0
		}
		return 1

	case int64:
		if int64(f) < val {
			return -1
		} else if int64(f) == val {
			return 0
		}
		return 1

	case uint64:
		if val >= 9223372036854775808 {
			return -1
		} else if v := int64(val); f < int64t(v) {
			return -1
		} else if f == int64t(v) {
			return 0
		}
		return 1
	}
	panic(fmt.Errorf("cannot handle type %T", other))
}

type uint64t int64

func (f uint64t) cmp(other interface{}) int {
	switch val := other.(type) {
	case float64:
		return -float64t(val).cmp(uint64(f))

	case int64:
		if val < 0 {
			return 1
		} else if v := int64(val); int64(f) < v {
			return -1
		} else if int64(f) == v {
			return 0
		}
		return 1

	case uint64:
		if uint64(f) < val {
			return -1
		} else if uint64(f) == val {
			return 0
		}
		return 1
	}
	panic(fmt.Errorf("cannot handle type %T", other))
}
