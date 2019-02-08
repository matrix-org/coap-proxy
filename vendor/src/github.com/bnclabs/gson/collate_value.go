// transform collated value into golang native value.
// cnf: NumberKind, arrayLenPrefix, propertyLenPrefix

package gson

import "strconv"

func collate2gson(code []byte, config *Config) (interface{}, int) {
	if len(code) == 0 {
		return nil, 0
	}

	var scratch [64]byte
	n := 1
	switch code[0] {
	case TypeMissing:
		return MissingLiteral, 2

	case TypeNull:
		return nil, 2

	case TypeTrue:
		return true, 2

	case TypeFalse:
		return false, 2

	case TypeNumber:
		m := getDatum(code[n:])
		ui, i, f, what := collated2Number(code[n:n+m-1], config.nk)
		switch what {
		case 1:
			return ui, n + m
		case 2:
			return i, n + m
		case 3:
			return f, n + m
		}

	case TypeString:
		s := make([]byte, encodedStringSize(code[n:]))
		s, x := collate2String(code[n:], s)
		return bytes2str(s), n + x

	case TypeBinary:
		m := getDatum(code[n:])
		bs := make([]byte, m-1)
		copy(bs, code[n:n+m-1])
		return bs, n + m

	case TypeArray:
		var arr []interface{}
		if config.arrayLenPrefix {
			if code[n] != TypeLength {
				panic("collate decode expected array length prefix")
			}
			n++
			m := getDatum(code[n:])
			_, y := collated2Int(code[n:n+m], scratch[:])
			ln, err := strconv.Atoi(bytes2str(scratch[:y]))
			if err != nil {
				panic(err)
			}
			arr = make([]interface{}, 0, ln)
			n += m
		} else {
			arr = make([]interface{}, 0, 8)
		}
		for code[n] != Terminator {
			item, y := collate2gson(code[n:], config)
			arr = append(arr, item)
			n += y
		}
		return arr, n + 1 // +1 to skip terminator

	case TypeObj:
		obj := make(map[string]interface{})
		if config.propertyLenPrefix {
			if code[n] != TypeLength {
				panic("collate decode expected object length prefix")
			}
			n++
			m := getDatum(code[n:])
			_, y := collated2Int(code[n:n+m], scratch[:])
			_, err := strconv.Atoi(bytes2str(scratch[:y])) // just skip
			if err != nil {
				panic(err)
			}
			n += m
		}
		for code[n] != Terminator {
			key, m := collate2gson(code[n:], config)
			n += m
			value, m := collate2gson(code[n:], config)
			obj[key.(string)] = value
			n += m
		}
		return obj, n + 1 // +1 to skip terminator
	}
	panic("collate decode invalid binary")
}

// get the collated datum based on Terminator and return the length
// of the datum.
func getDatum(code []byte) int {
	var i int
	var b byte
	for i, b = range code {
		if b == Terminator {
			return i + 1
		}
	}
	panic("collate decode terminator not found")
}
