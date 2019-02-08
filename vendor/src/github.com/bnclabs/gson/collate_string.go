package gson

func suffixEncodeString(s []byte, code []byte) int {
	n := 0
	for _, x := range s {
		code[n] = x
		n++
		if x == Terminator {
			code[n] = 1
			n++
		}
	}
	code[n] = Terminator
	return n + 1
}

func suffixDecodeString(code []byte, text []byte) (int, int) {
	for i, j := 0, 0; i < len(code); i++ {
		x := code[i]
		if x == Terminator {
			i++
			switch x = code[i]; x {
			case 1:
				text[j] = 0
				j++
			case Terminator:
				return i + 1, j
			default:
				panic("collate decode invalid escape sequence")
			}
			continue
		}
		text[j] = x
		j++
	}
	panic("collate decode invalid string")
}

func encodedStringSize(code []byte) int {
	for i := 0; i < len(code); {
		if code[i] == Terminator {
			switch code[i+1] {
			case 1:
				i++
			case Terminator:
				return i
			default:
				panic("collate decode invalid escape sequence")
			}
		}
		i++
	}
	panic("collate decode invalid string")
}
