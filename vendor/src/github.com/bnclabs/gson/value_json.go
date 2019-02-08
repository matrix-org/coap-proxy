// transform golang native value into json encoded value.
// cnf: -

package gson

import "fmt"
import "strconv"
import "encoding/json"

func value2json(value interface{}, out []byte, config *Config) int {
	var err error
	var outsl []byte

	if value == nil {
		return copy(out, "null")
	}

	switch v := value.(type) {
	case bool:
		if v {
			return copy(out, "true")
		}
		return copy(out, "false")

	case byte:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int8:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int16:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case uint16:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int32:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case uint32:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case uint:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case int64:
		switch config.nk {
		case FloatNumber:
			out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		case SmartNumber:
			out = strconv.AppendInt(out[:0], v, 10)
		default:
			panic(fmt.Errorf("unknown number kind %v", config.nk))
		}
		return len(out)

	case uint64:
		switch config.nk {
		case FloatNumber:
			out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		case SmartNumber:
			out = strconv.AppendUint(out[:0], v, 10)
		default:
			panic(fmt.Errorf("unknown number kind %v", config.nk))
		}
		return len(out)

	case float32:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case float64:
		out = strconv.AppendFloat(out[:0], v, 'f', -1, 64)
		return len(out)

	case string:
		out, err = encodeString(str2bytes(v), out[:0])
		if err != nil {
			panic("error encoding string")
		}
		return len(out)

	case json.Number:
		return copy(out, v)

	case []interface{}:
		n := 0
		out[n] = '['
		n++
		for i, x := range v {
			n += value2json(x, out[n:], config)
			if i < len(v)-1 {
				out[n] = ','
				n++
			}
		}
		out[n] = ']'
		n++
		return n

	case map[string]interface{}:
		n := 0
		out[n] = '{'
		n++

		mkeys := config.mkeysh.getmkeys(len(v))

		count := len(v)
		for _, key := range mkeys.sortProps1(v) {
			outsl, err = encodeString(str2bytes(key), out[n:n])
			if err != nil {
				panic("error encoding key")
			}
			n += len(outsl)
			out[n] = ':'
			n++

			n += value2json(v[key], out[n:], config)

			count--
			if count > 0 {
				out[n] = ','
				n++
			}
		}
		out[n] = '}'
		n++

		config.mkeysh.putmkeys(mkeys)
		return n

	case map[string]uint64:
		n := 0
		out[n] = '{'
		n++

		mkeys := config.mkeysh.getmkeys(len(v))

		count := len(v)
		for _, key := range mkeys.sortProps2(v) {
			outsl, err = encodeString(str2bytes(key), out[n:n])
			if err != nil {
				panic("error encoding key")
			}
			n += len(outsl)
			out[n] = ':'
			n++

			n += value2json(v[key], out[n:], config)

			count--
			if count > 0 {
				out[n] = ','
				n++
			}
		}
		out[n] = '}'
		n++

		config.mkeysh.putmkeys(mkeys)
		return n

	case [][2]interface{}:
		n := 0
		out[n] = '{'
		n++

		for i, item := range v {
			n += value2json(item[0], out[n:], config)
			out[n] = ':'
			n++

			n += value2json(item[1], out[n:], config)

			if i < len(v)-1 {
				out[n] = ','
				n++
			}
		}
		out[n] = '}'
		n++
		return n
	}
	return 0
}
