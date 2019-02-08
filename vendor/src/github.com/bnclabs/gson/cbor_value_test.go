package gson

import "testing"
import "reflect"
import "encoding/json"
import "time"
import "regexp"
import "math/big"

func TestCborNil(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	val := config.NewValue(nil)

	if rv := val.Tocbor(cbr).Tovalue(); rv != nil {
		t.Errorf("expected %v, got %v", nil, rv)
	}
}

func TestCborBoolean(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	// test true
	val := config.NewValue(true)
	if rv := val.Tocbor(cbr).Tovalue(); rv != true {
		t.Errorf("expected %v, got %v", true, rv)
	}

	// test false
	val = config.NewValue(false)
	if rv := val.Tocbor(cbr.Reset(nil)).Tovalue(); rv != false {
		t.Errorf("expected %v, got %v", false, rv)
	}
}

func TestCborSmallNumber(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	// test uint8
	for i := uint16(0); i <= 255; i++ {
		rv := config.NewValue(uint8(i)).Tocbor(cbr.Reset(nil)).Tovalue()
		if rv.(uint64) != uint64(i) {
			t.Errorf("expected %v, got %v", uint64(i), rv)
		}
	}

	// test int8
	for i := int16(-128); i <= 127; i++ {
		rv := config.NewValue(int8(i)).Tocbor(cbr.Reset(nil)).Tovalue()
		if num1, ok := rv.(int64); ok {
			if num1 != int64(i) {
				t.Errorf("int8: expected %v, got %v", uint64(i), rv)
			}
		} else if rv.(uint64) != uint64(i) {
			t.Errorf("unt8: expected %v, got %v", uint64(i), rv)
		}
	}
}

func TestCborNum(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	tests := [][2]interface{}{
		{'a', uint64(97)},
		{byte(0), uint64(0)},
		{byte(23), uint64(23)},
		{byte(24), uint64(24)},
		{byte(255), uint64(255)},
		{uint8(0), uint64(0)},
		{uint8(23), uint64(23)},
		{uint8(24), uint64(24)},
		{uint8(255), uint64(255)},
		{int8(-128), int64(-128)},
		{int8(-24), int64(-24)},
		{int8(-24), int64(-24)},
		{int8(-1), int64(-1)},
		{int8(-0), uint64(-0)},
		{int8(0), uint64(0)},
		{int8(23), uint64(23)},
		{int8(24), uint64(24)},
		{int8(127), uint64(127)},
		{uint16(0), uint64(0)},
		{uint16(23), uint64(23)},
		{uint16(24), uint64(24)},
		{uint16(255), uint64(255)},
		{uint16(65535), uint64(65535)},
		{int16(-32768), int64(-32768)},
		{int16(-256), int64(-256)},
		{int16(-255), int64(-255)},
		{int16(-129), int64(-129)},
		{int16(-128), int64(-128)},
		{int16(-127), int64(-127)},
		{int16(-24), int64(-24)},
		{int16(-23), int64(-23)},
		{int16(-1), int64(-1)},
		{int16(-0), uint64(0)},
		{int16(0), uint64(0)},
		{int16(23), uint64(23)},
		{int16(24), uint64(24)},
		{int16(127), uint64(127)},
		{int16(255), uint64(255)},
		{int16(32767), uint64(32767)},
		{uint32(0), uint64(0)},
		{uint32(23), uint64(23)},
		{uint32(24), uint64(24)},
		{uint32(255), uint64(255)},
		{uint32(65535), uint64(65535)},
		{uint32(4294967295), uint64(4294967295)},
		{int32(-2147483648), int64(-2147483648)},
		{int32(-32769), int64(-32769)},
		{int32(-32768), int64(-32768)},
		{int32(-32767), int64(-32767)},
		{int32(-256), int64(-256)},
		{int32(-255), int64(-255)},
		{int32(-129), int64(-129)},
		{int32(-128), int64(-128)},
		{int32(-127), int64(-127)},
		{int32(-24), int64(-24)},
		{int32(-23), int64(-23)},
		{int32(-1), int64(-1)},
		{int32(-0), uint64(-0)},
		{int32(0), uint64(0)},
		{int32(23), uint64(23)},
		{int32(24), uint64(24)},
		{int32(127), uint64(127)},
		{int32(32767), uint64(32767)},
		{int32(65535), uint64(65535)},
		{int32(2147483647), uint64(2147483647)},
		{int(-2147483648), int64(-2147483648)},
		{uint(2147483647), uint64(2147483647)},
		{uint64(0), uint64(0)},
		{uint64(23), uint64(23)},
		{uint64(24), uint64(24)},
		{uint64(255), uint64(255)},
		{uint64(65535), uint64(65535)},
		{uint64(4294967295), uint64(4294967295)},
		{
			uint64(18446744073709551615), uint64(18446744073709551615)},
		{
			int64(-9223372036854775808), int64(-9223372036854775808)},
		{int64(-2147483649), int64(-2147483649)},
		{int64(-2147483648), int64(-2147483648)},
		{int64(-2147483647), int64(-2147483647)},
		{int64(-32769), int64(-32769)},
		{int64(-32768), int64(-32768)},
		{int64(-32767), int64(-32767)},
		{int64(-256), int64(-256)},
		{int64(-255), int64(-255)},
		{int64(-129), int64(-129)},
		{int64(-128), int64(-128)},
		{int64(-127), int64(-127)},
		{int64(-24), int64(-24)},
		{int64(-23), int64(-23)},
		{int64(-1), int64(-1)},
		{int64(-0), uint64(-0)},
		{int64(0), uint64(0)},
		{int64(23), uint64(23)},
		{int64(24), uint64(24)},
		{int64(127), uint64(127)},
		{int64(32767), uint64(32767)},
		{int64(2147483647), uint64(2147483647)},
		{int64(4294967295), uint64(4294967295)},
		{int64(9223372036854775807), uint64(9223372036854775807)},
	}
	for _, test := range tests {
		t.Logf("executing test case %v", test[0])

		rv := config.NewValue(test[0]).Tocbor(cbr.Reset(nil)).Tovalue()
		if !reflect.DeepEqual(rv, test[1]) {
			t.Errorf("expected {%v,%T}, got {%v,%T}", test[1], test[1], rv, rv)
		}
	}

	// test case for number exceeding int64
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic decoding int64 > 9223372036854775807")
			}
		}()
		config.NewValue(uint64(9223372036854775808)).Tocbor(cbr.Reset(nil))
		cbr.data[0] = (cbr.data[0] & 0x1f) | cborType1 // chg to neg. integer
		cbr.Tovalue()
	}()
}

func TestCborFloat16(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding float16")
		}
	}()
	config.NewCbor([]byte{0xf9, 0, 0, 0, 0}).Tovalue()
}

func TestCborFloat32(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	val := config.NewValue(float32(10.11))

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborFloat64(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	val := config.NewValue(float64(10.11))

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborTagBytes(t *testing.T) {
	ref := make([]uint8, 100)
	for i := 0; i < len(ref); i++ {
		ref[i] = uint8(i)
	}

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 200))
	val := config.NewValue(ref)

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborText(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 200))
	val := config.NewValue("hello world")

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborArray(t *testing.T) {
	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	val := config.NewValue([]interface{}{10.2, "hello world"})

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}

	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(Stream)
	cbr = config.NewCbor(make([]byte, 0, 1024))
	val = config.NewValue([]interface{}{10.2, "hello world"})

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(val.data, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborMapSlice(t *testing.T) {
	ref := [][2]interface{}{
		{"10.2", "hello world"},
		{"hello world", 10.2},
	}
	refm := CborMap2golangMap(ref)

	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	val := config.NewValue(ref)

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(refm, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}

	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(Stream)
	cbr = config.NewCbor(make([]byte, 0, 1024))
	val = config.NewValue(ref)

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(refm, rv) {
		t.Errorf("expected %v, got %v", val.data, rv)
	}
}

func TestCborMap(t *testing.T) {
	ref := map[string]interface{}{
		"10.2":        "hello world",
		"hello world": float64(10.2),
	}
	refm := CborMap2golangMap(ref)

	// encoding use LengthPrefix
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(LengthPrefix)
	val := config.NewValue(ref)
	cbr := config.NewCbor(make([]byte, 0, 1024))

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(refm, rv) {
		t.Errorf("expected %v, got %v", refm, rv)
	}

	// encoding use Stream
	config = NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	config = config.SetContainerEncoding(Stream)
	val = config.NewValue(ref)
	cbr = config.NewCbor(make([]byte, 0, 1024))

	if rv := val.Tocbor(cbr).Tovalue(); !reflect.DeepEqual(refm, rv) {
		t.Errorf("expected %v, got %v", refm, rv)
	}
}

func TestCborMaster(t *testing.T) {
	var outval, ref interface{}

	testcases := append(scanvalid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	jsn := config.NewJson(nil)
	cbr := config.NewCbor(make([]byte, 0, 1024*1024))
	jsn1 := config.NewJson(make([]byte, 0, 1024*1024))
	cbr1 := config.NewCbor(make([]byte, 0, 1024*1024))
	jsn2 := config.NewJson(make([]byte, 0, 1024*1024))

	for _, tcase := range testcases {
		t.Logf("%v", tcase)

		json.Unmarshal([]byte(tcase), &ref)
		jsn.Reset([]byte(tcase))

		// test json->cbor->json->value
		jsn.Tocbor(cbr.Reset(nil))
		cbr.Tojson(jsn1.Reset(nil))
		if err := json.Unmarshal(jsn1.Bytes(), &outval); err != nil {
			t.Fatalf("error parsing %v: %v", string(jsn1.Bytes()), err)
		} else if !reflect.DeepEqual(outval, ref) {
			t.Fatalf("expected '%v', got '%v'", ref, outval)
		}

		// cbr->value->cbr->json->value
		val := config.NewValue(cbr.Tovalue())
		val.Tocbor(cbr1.Reset(nil))
		cbr1.Tojson(jsn2.Reset(nil))
		if err := json.Unmarshal(jsn2.Bytes(), &outval); err != nil {
			t.Fatalf("error parsing %q: %v", string(jsn2.Bytes()), err)
		} else if !reflect.DeepEqual(outval, ref) {
			fmsg := "expected {%T,%v}, got {%T,%v}"
			t.Fatalf(fmsg, val.data, val.data, outval, outval)
		}
	}
}

func TestCborSmartnum(t *testing.T) {
	var outval, ref interface{}

	data := testdataFile("testdata/smartnum")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data)
	cbr := config.NewCbor(make([]byte, 0, 1024*1024))
	jsnback := config.NewJson(make([]byte, 0, 1024*1024))

	// test json->cbor->json->value
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if err := json.Unmarshal(jsnback.Bytes(), &outval); err != nil {
		t.Logf("%v", string(jsnback.Bytes()))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else if !reflect.DeepEqual(ref, outval) {
		t.Errorf("expected %v", ref)
		t.Errorf("got-json %v", string(jsnback.Bytes()))
		t.Fatalf("got %v", outval)
	}

	// test cbor->value->cbor->json->value
	val := config.NewValue(cbr.Tovalue())
	jsn = config.NewJson(make([]byte, 0, 1024*1024))
	cbr = config.NewCbor(make([]byte, 0, 1024*1024))

	val.Tocbor(cbr)
	cbr.Tojson(jsn)
	if err := json.Unmarshal(jsn.Bytes(), &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else if err := json.Unmarshal(data, &val.data); err != nil {
		t.Fatalf("error parsing code.json: %v", err)
	} else if !reflect.DeepEqual(outval, val.data) {
		t.Errorf("expected %v", val.data)
		t.Fatalf("got %v", outval)
	}
}

func TestCborMalformed(t *testing.T) {
	for _, tcase := range scaninvalid {
		t.Logf("%v", tcase)

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("expected panic")
				}
			}()

			config := NewDefaultConfig()
			config = config.SetNumberKind(SmartNumber).SetSpaceKind(AnsiSpace)
			jsn := config.NewJson(make([]byte, 0, 1024))
			jsn.Tocbor(config.NewCbor(make([]byte, 0, 1024)))
		}()
	}
}

func TestCborCodeJSON(t *testing.T) {
	var ref, outval interface{}

	data := testdataFile("testdata/code.json.gz")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data)
	cbr := config.NewCbor(make([]byte, 0, 10*1024*1024))
	jsnback := config.NewJson(make([]byte, 0, 10*1024*1024))

	t.Logf("%v %v", len(data), len(jsnback.Bytes()))

	// test json->cbor->json->value
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if err := json.Unmarshal(jsnback.Bytes(), &outval); err != nil {
		t.Logf("%v", string(jsnback.Bytes()))
		t.Fatalf("error parsing code.json.gz: %v", err)
	} else {
		if !reflect.DeepEqual(ref, outval) {
			t.Errorf("expected %v", ref)
			t.Fatalf("got %v", outval)
		}
	}

	// cbor->value->cbor->json->value
	jsn = config.NewJson(make([]byte, 0, 10*1024*1024))
	cbrback := config.NewCbor(make([]byte, 0, 10*1024*1024))
	val := config.NewValue(cbr.Tovalue())
	val.Tocbor(cbrback)
	cbrback.Tojson(jsn)
	if err := json.Unmarshal(jsn.Bytes(), &outval); err != nil {
		t.Fatalf("error parsing %v", err)
	} else {
		if !reflect.DeepEqual(outval, ref) {
			t.Errorf("expected %v", val.data)
			t.Fatalf("got %v", outval)
		}
	}
}

func TestCborTypical(t *testing.T) {
	var ref, out interface{}

	data := testdataFile("testdata/typical.json")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data)
	cbr := config.NewCbor(make([]byte, 0, 1024*1024))
	jsnback := config.NewJson(make([]byte, 0, 1024*1024))

	// test json->cbor->json->value
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if err := json.Unmarshal(jsnback.Bytes(), &out); err != nil {
		t.Errorf("error parsing typical.json: %v", err)
	} else if !reflect.DeepEqual(ref, out) {
		t.Errorf("expected %v", ref)
		t.Errorf("got      %v", out)
	}
}

func TestCborUndefined(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	if n := valundefined2cbor(buf); n != 1 {
		t.Errorf("fail value2cbor undefined: %v want 1", n)
	} else if item, m := cbor2value(buf, config); m != 1 {
		t.Errorf("fail cbor2value on undefined len: %v want 1", m)
	} else if item.(CborUndefined) != CborUndefined(cborSimpleUndefined) {
		t.Errorf("fail cbor2value on undefined: %T %v", item, item)
	}
}

func TestCborReserved(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic while decoding reserved")
		}
	}()
	cbor2value([]byte{cborHdr(cborType0, 28)}, config)
}

//---- test cases for tag function

func TestDateTime(t *testing.T) {
	// test time.Time
	ref, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	if err != nil {
		t.Errorf("time.Parse() failed: %v", err)
	}

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	val := config.NewValue(ref)
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !ref.Equal(item.(time.Time)) {
		t.Errorf("expected %v got %v", ref, item.(time.Time))
	}

	// malformed.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.data[5] = 'a'
		cbr.Tovalue()
	}()
}

func TestTagEpoch(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	// positive and negative epoch
	for _, v := range [2]int64{1000000, -100000} {
		val := config.NewValue(CborTagEpoch(v))
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
			t.Errorf("expected %v got %v", val.data, item)
		}
	}

	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.data[1] = 0x5a // instead of 0x3a
		cbr.Tovalue()
	}()

	// malformed epoch
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf := make([]byte, 16)
		n := tag2cbor(tagEpoch, buf)
		valbytes2cbor([]byte{1, 2}, buf[n:])
		config.NewCbor(buf).Tovalue()
	}()
}

func TestTagEpochMicro(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	// positive and negative epoch in uS.
	for _, v := range [2]float64{1000000.123456, -100000.123456} {
		val := config.NewValue(CborTagEpochMicro(v))
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
			t.Errorf("expected %v got %v", val.data, item)
		}
	}
}

func TestBigNum(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	// positive and negative bignums
	for _, v := range [2]int64{1000, -1000} {
		z := big.NewInt(0).Mul(big.NewInt(9223372036854775807), big.NewInt(v))
		val := config.NewValue(z)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); z.Cmp(item.(*big.Int)) != 0 {
			t.Errorf("expected %v got %v", z, item.(*big.Int))
		}
	}
}

func TestDecimalFraction(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	refs := []CborTagFraction{
		CborTagFraction([2]int64{int64(-10), int64(-23)}),
		CborTagFraction([2]int64{int64(-10), int64(23)}),
		CborTagFraction([2]int64{int64(10), int64(-23)}),
		CborTagFraction([2]int64{int64(10), int64(23)}),
	}

	for _, ref := range refs {
		val := config.NewValue(ref)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %v got %v", ref, item)
		}
	}
}

func TestBigFloat(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	refs := []CborTagFloat{
		CborTagFloat([2]int64{int64(-10), int64(-23)}),
		CborTagFloat([2]int64{int64(-10), int64(23)}),
		CborTagFloat([2]int64{int64(10), int64(-23)}),
		CborTagFloat([2]int64{int64(10), int64(23)}),
	}
	for _, ref := range refs {
		val := config.NewValue(ref)
		val.Tocbor(cbr.Reset(nil))
		if item := cbr.Tovalue(); !reflect.DeepEqual(ref, item) {
			t.Errorf("expected %T got %T", ref, item)
		}
	}
}

func TestCbor(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	val := config.NewValue(CborTagBytes([]byte("hello world")))
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
		t.Errorf("exptected %v got %v", val.data, item)
	}
}

func TestRegexp(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	ref, _ := regexp.Compile(`a([0-9]t*)+`)
	val := config.NewValue(ref)
	val.Tocbor(cbr)
	item := cbr.Tovalue()
	if ref.String() != (item.(*regexp.Regexp)).String() {
		t.Errorf("expected %v got %v", ref, item)
	}
	// malformed reg-ex
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		buf := make([]byte, 1024)
		n := tag2cbor(tagRegexp, buf)
		valtext2cbor(`a([0-9]t*+`, buf[n:])
		config.NewCbor(buf).Tovalue()
	}()
}

func TestCborTagPrefix(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)

	val := config.NewValue(CborTagPrefix([]byte("hello world")))
	val.Tocbor(cbr)
	if item := cbr.Tovalue(); !reflect.DeepEqual(val.data, item) {
		t.Errorf("exptected %v got %v", val.data, item)
	}
}

func TestCborBreakStop(t *testing.T) {
	buf := make([]byte, 10)
	config := NewDefaultConfig()
	if n := breakStop(buf); n != 1 {
		t.Errorf("fail code text-start len: %v wanted 1", n)
	} else if val, m := cbor2value(buf, config); m != n {
		t.Errorf("fail code text-start: %v wanted %v", m, n)
	} else if !reflect.DeepEqual(val, CborBreakStop(0xff)) {
		t.Errorf("fail code text-start: %x wanted 0xff", buf[0])
	}
}

//---- benchmarks

func BenchmarkCbor2ValNull(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(nil).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValTrue(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(true).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValFalse(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(false).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValUint8(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(uint8(255)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValInt8(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(int8(-128)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValUint16(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(uint16(65535)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValInt16(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(int16(-32768)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValUint32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(uint32(4294967295)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValInt32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(int32(-2147483648)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValUint64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(uint64(18446744073709551615)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValInt64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(int64(-2147483648)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValFlt32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(float32(10.2)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValFlt64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(float64(10.2)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValBytes(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue([]byte("hello world")).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValText(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue("hello world").Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValArr0(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(make([]interface{}, 0)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValArr5(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	value := []interface{}{5, 5.0, "hello world", true, nil}
	config.NewValue(value).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValMap0(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	config.NewValue(make([][2]interface{}, 0)).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValMap5(b *testing.B) {
	value := [][2]interface{}{
		{"key0", 5}, {"key1", 5.0},
		{"key2", "hello world"},
		{"key3", true}, {"key4", nil},
	}

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	config.NewValue(value).Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}

func BenchmarkCbor2ValTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"))
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	jsn.Tocbor(cbr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Tovalue()
	}
}
