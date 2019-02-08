package gson

import "fmt"
import "strconv"

// Constants used in text representation of basic data types.
const (
	chPLUS  = 43
	chMINUS = 45
	chLT    = 60
	chGT    = 62
	chDOT   = 46
	chZERO  = 48
)

// Constants used to represent positive and negative numbers while encoding
// them.
const (
	negPrefix = byte(chMINUS)
	posPrefix = byte(chGT)
)

// Negative integers, in its text representation are 10's complement. This map
// provides the lookup table to generate the complements.
var negIntLookup = map[byte]byte{
	48: 57, // 0 - 9
	49: 56, // 1 - 8
	50: 55, // 2 - 7
	51: 54, // 3 - 6
	52: 53, // 4 - 5
	53: 52, // 5 - 4
	54: 51, // 6 - 3
	55: 50, // 7 - 2
	56: 49, // 8 - 1
	57: 48, // 9 - 0
}

// A simple lookup table to flip prefixes.
var prefixOpp = map[byte]byte{
	posPrefix: negPrefix,
	negPrefix: posPrefix,
}

// Zero is encoded as '0'
func collateInt(text, code []byte) int {
	if len(text) == 0 { // empty input
		return 0
	}
	n, ok := isZero(text, code)
	if ok {
		return n
	}

	switch text[0] {
	case chPLUS: // positive int
		n = collatePosInt(text[1:], code)
	case chMINUS: // negative int
		n = collateNegInt(text[1:], code)
	default:
		n = collatePosInt(text, code)
	}
	return n
}

// encode positive integer, local function gets called by collateInt
//  encoding 7,          >7
//  encoding 123,        >>3123
//  encoding 1234567890, >>>2101234567890
func collatePosInt(text []byte, code []byte) int {
	var scratch [32]byte
	code[0] = posPrefix
	n, ln := 1, int64(len(text))
	if ln > 1 {
		n += collatePosInt(strconv.AppendInt(scratch[:0], ln, 10), code[n:])
	}
	copy(code[n:], text)
	return n + int(ln)
}

// encode negative integer, local function gets called by collateInt
//  encoding -1,         -8
//  encoding -2,         -7
//  encoding -9,         -0
//  encoding -10,        --789
//  encoding -11,        --788
//  encoding -1234567891 ---7898765432108
//  encoding -1234567890 ---7898765432109
//  encoding -1234567889 ---7898765432110
func collateNegInt(text []byte, code []byte) int {
	var scratch [32]byte
	code[0] = negPrefix
	n, ln := 1, int64(len(text))
	if ln > 1 {
		n += collateNegInt(strconv.AppendInt(scratch[:0], ln, 10), code[n:])
	}
	for i, x := range text {
		code[n+i] = negIntLookup[x]
	}
	return n + int(ln)
}

// handle different forms of zero, like 0 +0 -0.
func isZero(text, code []byte) (int, bool) {
	switch len(text) {
	case 1:
		if text[0] == chZERO {
			code[0] = chZERO
			return 1, true
		}

	case 2:
		if (text[0] == chPLUS || text[0] == chMINUS) && text[1] == chZERO {
			code[0] = chZERO
			return 1, true
		}
	}
	return 0, false
}

func collated2Int(code, text []byte) (int, int) {
	if len(code) == 0 { // empty input
		return 0, 0

	} else if code[0] == chZERO {
		text[0] = code[0]
		return 1, 1

	} else if code[0] == posPrefix {
		text[0] = chPLUS
		s, e := doCollated2Int(code[1:])
		copy(text[1:], code[1+s:1+e])
		return 1 + e, 1 + (e - s)
	}
	text[0] = chMINUS
	s, e := doCollated2Int(code[1:])
	for i, x := range code[1+s : 1+e] {
		text[1+i] = negIntLookup[x]
	}
	return 1 + e, 1 + (e - s)
}

func doCollated2Int(code []byte) (int, int) {
	var scratch [32]byte
	switch code[0] {
	case posPrefix:
		s, e := doCollated2Int(code[1:])
		l, err := strconv.Atoi(bytes2str(code[1+s : 1+e]))
		if err != nil {
			panic(err)
		}
		return 1 + e, 1 + e + l

	case negPrefix:
		s, e := doCollated2Int(code[1:])
		for i, x := range code[1+s : 1+e] {
			scratch[i] = negIntLookup[x]
		}
		l, err := strconv.Atoi(bytes2str(scratch[:e-s]))
		if err != nil {
			panic(err)
		}
		return 1 + e, 1 + e + l
	}
	return 0, 1 // <-- exit case.
}

// collateSD encodes small-decimal, values that are greater than
// -1.0 and less than +1.0.
//
// small decimals is greater than -1.0 and less than 1.0
//
// Input `text` is also in textual representation, that is,
// strconv.ParseFloat(text, 64) is the actual integer that is encoded.
//
//  encoding -0.9995    -0004>
//  encoding -0.999     -000>
//  encoding -0.0123    -9876>
//  encoding -0.00123   -99876>
//  encoding -0.0001233 -9998766>
//  encoding -0.000123  -999876>
//  encoding +0.000123  >000123-
//  encoding +0.0001233 >0001233-
//  encoding +0.00123   >00123-
//  encoding +0.0123    >0123-
//  encoding +0.999     >999-
//  encoding +0.9995    >9995-
//
// Caveats:
//  -0.0, 0.0 and +0.0 must be filtered out as integer ZERO `0`.
func collateSD(text, code []byte) int {
	if len(text) == 0 { // empty input
		return 0
	}

	prefix, n := signPrefix(text)
	code[0] = prefix
	m := 1

	// remove decimal point and all zeros before that.
	for _, x := range text[n:] {
		n++
		if x == chDOT {
			break
		}
	}
	if prefix == negPrefix { // do inversion if negative number
		for _, x := range text[n:] {
			code[m] = negIntLookup[x]
			m++
		}
	} else { // if positive number just copy the text
		copy(code[m:], text[n:])
		m += len(text[n:])
	}
	code[m] = prefixOpp[prefix]
	return m + 1
}

// collated2SD complements collateSD
func collated2SD(code, text []byte) (int, int) {
	if len(code) == 0 {
		return 0, 0
	}

	prefix, sign := code[0], prefixSign(code)
	text[0], text[1], text[2] = sign, chZERO, chDOT

	// if negative number invert the digits.
	codesz := len(code)
	if prefix == negPrefix {
		if codesz > 1 {
			for i, x := range code[1 : codesz-1] {
				text[3+i] = negIntLookup[x]
			}
		}

	} else if codesz > 1 {
		copy(text[3:], code[1:codesz-1])
	}
	return codesz, 3 + codesz - 2
}

// a floating point number f takes a mantissa m ∈ [1/10 , 1) and an integer
// exponent e such that f = (10^e) * ±m.
//
//  encoding −0.1 × 10^11    - --7888+
//  encoding −0.1 × 10^10    - --7898+
//  encoding -1.4            - -885+
//  encoding -1.3            - -886+
//  encoding -1              - -88+
//  encoding -0.123          - 0876+
//  encoding -0.0123         - +1876+
//  encoding -0.001233       - +28766+
//  encoding -0.00123        - +2876+
//  encoding 0               0
//  encoding +0.00123        + -7123-
//  encoding +0.001233       + -71233-
//  encoding +0.0123         + -8123-
//  encoding +0.123          + 0123-
//  encoding +1              + +11-
//  encoding +1.3            + +113-
//  encoding +1.4            + +114-
//  encoding +0.1 × 10^10    + ++2101-
//  encoding +0.1 × 10^11    + ++2111-
func collateFloat(text, code []byte) int {
	if len(text) == 0 { // empty input
		return 0
	}
	val, e := strconv.ParseFloat(bytes2str(text), 64)
	if e == nil && val == 0 {
		code[0] = chZERO
		return 1
	}

	prefix, n := signPrefix(text)
	code[0] = prefix
	m := 1

	var exp, mant []byte // gather exponent and mantissa
	for i, x := range text[n:] {
		if x == 'e' {
			mant = text[n : n+i]
			exp = text[n+i+1:]
			break
		}
	}

	var mantissa [64]byte // fix mantessa
	mantissa[0], mantissa[1], mantissa[2] = prefix, '0', '.'
	p := 0
	for _, x := range mant {
		if x == '.' {
			continue
		}
		mantissa[3+p] = x
		p++
	}
	mant = mantissa[:3+p]

	var exponent [64]byte // fix exponent
	expi, err := strconv.Atoi(bytes2str(exp))
	if err != nil {
		panic(fmt.Errorf("%v", err))
	}
	expi++
	if prefix == negPrefix {
		expi = -expi
	}
	exp = strconv.AppendInt(exponent[:0], int64(expi), 10)

	m += collateInt(exp, code[m:])
	q := collateSD(mant, code[m:])
	copy(code[m:], code[m+1:m+q])
	return m + q - 1
}

var flipmap = map[byte]byte{chPLUS: chMINUS, chMINUS: chPLUS}

func collated2Float(code, text []byte) (int, int) {
	if len(code) == 0 { // empty input
		return 0, 0
	} else if len(code) == 1 && code[0] == chZERO {
		text[0] = chZERO
		return 1, 1
	}

	var exponent [64]byte
	var exp []byte
	n := 1
	switch code[n] {
	case negPrefix, posPrefix:
		x, y := collated2Int(code[n:], exponent[:])
		n += x
		exp = exponent[:y]

	default:
		exponent[0] = chPLUS
		x, y := collated2Int(code[n:], exponent[1:])
		n += x
		exp = exponent[:1+y]
	}

	if code[0] == negPrefix {
		exponent[0] = flipmap[exponent[0]]
	}

	var mantissa [64]byte
	mantissa[0] = code[0]
	copy(mantissa[1:], code[n:])
	x, y := collated2SD(mantissa[:1+len(code[n:])], text)
	if y > 3 {
		var newexp [64]byte

		// adjust 0.xyz as x.yz
		text[1] = text[3]
		copy(text[3:], text[4:y])
		y--
		text[y] = 'e'
		// adjust exponent by 1
		i, err := strconv.Atoi(bytes2str(exp))
		if err != nil {
			panic(err)
		}
		nexp := strconv.AppendInt(newexp[:0], int64(i-1), 10)
		j := copy(text[y+1:], nexp)
		return n + x - 1, y + 1 + j
	}
	text[y] = 'e'
	copy(text[y+1:], exp)
	return n + x - 1, y + 1 + len(exp)
}

// collateLD encodes large-decimal, values that are greater than or equal to
// +1.0 and less than or equal to -1.0.
//
// Input `text` is also in textual representation, that is,
// strconv.ParseFloat(text, 64) is the actual integer that is encoded.
//
//  encoding -100.5         --68994>
//  encoding -10.5          --7>
//  encoding -3.145         -3854>
//  encoding -3.14          -385>
//  encoding -1.01          -198>
//  encoding -1             -1>
//  encoding -0.0001233     -09998766>
//  encoding -0.000123      -0999876>
//  encoding +0.000123      >0000123-
//  encoding +0.0001233     >00001233-
//  encoding +1             >1-
//  encoding +1.01          >101-
//  encoding +3.14          >314-
//  encoding +3.145         >3145-
//  encoding +10.5          >>2105-
//  encoding +100.5         >>31005-
func collateLD(text, code []byte) int {
	if len(text) == 0 { // empty input
		return 0
	}

	prefix, _ := signPrefix(text)
	m := 0

	var integer, decimal []byte // split integer and decimal parts
	integer = text[:]           // optimistically assume that no decimal part
	for i, x := range text {
		if x == chDOT {
			integer = text[:i]
			decimal = text[i+1:]
		}
	}

	// encode integer and decimal part.
	if len(integer) == 1 && integer[0] == chZERO {
		code[m], code[m+1] = prefix, chZERO
		m += 2

	} else if len(integer) == 2 {
		ch := integer[0]
		if ch == chPLUS && integer[1] == chZERO {
			code[m], code[m+1] = prefix, chZERO
			m += 2
		} else if ch == chMINUS && integer[1] == chZERO {
			code[m], code[m+1] = prefix, '9'
			m += 2
		} else {
			m += collateInt(integer, code[m:])
		}

	} else {
		m += collateInt(integer, code[m:])
	}

	var dec, sd [64]byte
	var n int
	if ln := len(decimal); ln > 0 {
		dec[0], dec[1], dec[2] = text[0], chZERO, chDOT
		copy(dec[3:], decimal)
		n = collateSD(dec[:3+ln], sd[:])
	} else {
		dec[0], dec[1], dec[2], dec[3] = text[0], chZERO, chDOT, chZERO
		n = collateSD(dec[:4], sd[:])
	}

	// Adjust the decimal part
	sd[n-1] = prefixOpp[prefix]
	if n == 3 {
		if prefix == negPrefix && sd[1] == '9' {
			copy(code[m:], sd[2:n])
			return m + 1
		} else if prefix == posPrefix && sd[1] == chZERO {
			copy(code[m:], sd[2:n])
			return m + 1
		}
	}
	copy(code[m:], sd[1:n])
	return m + n - 1
}

// collated2LD complements collateLD
func collated2LD(code, text []byte) (int, int) {
	if len(code) == 0 { // empty input
		return 0, 0
	}

	prefix, sign := code[0], prefixSign(code)
	text[0] = sign
	n, m := 1, 1
	s, e := doCollated2Int(code[n:])
	copy(text[m:], code[n+s:n+e])

	if sign == chMINUS { // negative number invert the digits
		for i, x := range text[m : m+e-s] {
			text[m+i] = negIntLookup[x]
		}
	}
	n, m = n+e, m+(e-s)

	var sdcode [64]byte
	sdcode[0] = prefix
	ln := len(code[n:])
	copy(sdcode[1:], code[n:])
	x, y := collated2SD(sdcode[:ln+1], text[m:])
	copy(text[m:], text[m+2:m+y])
	return n + x - 1, m + y - 2
}

func signPrefix(text []byte) (byte, int) {
	switch text[0] {
	case chPLUS:
		return posPrefix, 1
	case chMINUS:
		return negPrefix, 1
	default:
		return posPrefix, 0
	}
}

func prefixSign(code []byte) byte {
	var sign byte
	switch code[0] {
	case posPrefix:
		sign = chPLUS
	case negPrefix:
		sign = chMINUS
	}
	return sign
}
