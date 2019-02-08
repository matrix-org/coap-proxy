package gson

import "testing"
import "reflect"
import "encoding/json"

func TestSkipWS(t *testing.T) {
	ref := "hello  "
	if got := skipWS("  hello  ", AnsiSpace); got != ref {
		t.Errorf("expected %v got %v", ref, got)
	}
}

func TestJsonEmptyToCbor(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	config.NewJson(make([]byte, 0, 1024)).Tocbor(cbr)
}

func TestCbor2Json(t *testing.T) {
	testcases := []string{
		// null
		"null",
		// boolean
		"true",
		"false",
		// integers
		"10",
		"0.1",
		"-0.1",
		"10.1",
		"-10.1",
		"-10E-1",
		"-10e+1",
		"10E-1",
		"10e+1",
		// string
		`"true"`,
		`"tru\"e"`,
		`"tru\\e"`,
		`"tru\be"`,
		`"tru\fe"`,
		`"tru\ne"`,
		`"tru\re"`,
		`"tru\te"`,
		`"tru\u0123e"`,
		`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`,
		// array
		`[]`,
		` [null,true,false,10,"tru\"e"]`,
		// object
		`{}`,
		`{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`,
	}

	dotest := func(config *Config) {
		jsn := config.NewJson(make([]byte, 0, 1024))
		cbr := config.NewCbor(make([]byte, 0, 1024))
		jsnback := config.NewJson(make([]byte, 0, 1024))

		var ref1, ref2 interface{}

		for _, tcase := range testcases {
			t.Logf("testcase - %v", tcase)
			json.Unmarshal([]byte(tcase), &ref1)

			jsn.Reset([]byte(tcase))
			jsn.Tocbor(cbr.Reset(nil))

			t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

			cbr.Tojson(jsnback.Reset(nil))
			if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
				t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
			}

			if !reflect.DeepEqual(ref1, ref2) {
				t.Errorf("mismatch %v, got %v", ref1, ref2)
			}
		}
	}
	dotest(NewDefaultConfig().SetStrict(false))
	dotest(NewDefaultConfig().SetStrict(true))
}

func TestCbor2JsonLengthPrefix(t *testing.T) {
	testcases := []string{
		`[null,true,false,10,"tru\"e"]`,
		`[]`,
		`{}`,
		`{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e","f":[1,2]}`,
	}

	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	config = config.SetContainerEncoding(LengthPrefix)

	jsn := config.NewJson(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsnback := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		jsn.Reset([]byte(tcase))

		jsn.Tocbor(cbr.Reset(nil))

		jsnback.Reset(nil)
		cbr.Tojson(jsnback)

		err := compareJSONs(t, tcase, string(jsnback.Bytes()))
		if err != nil {
			t.Errorf("%v", err)
		}
	}
}

func TestCbor2JsonNum(t *testing.T) {
	// test FloatNumber
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(UnicodeSpace)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	jsnback := config.NewJson(make([]byte, 0, 1024))

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// test SmartNumber (integer)
	config = NewDefaultConfig()
	config = config.SetNumberKind(SmartNumber).SetSpaceKind(UnicodeSpace)
	cbr = config.NewCbor(make([]byte, 0, 1024))
	jsn = config.NewJson(make([]byte, 0, 1024))
	jsnback = config.NewJson(make([]byte, 0, 1024))

	jsn.Reset([]byte("10"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10" {
		t.Errorf("expected %q, got %q", "10", s)
	}

	// test SmartNumber (float)
	config = NewDefaultConfig()
	config = config.SetNumberKind(SmartNumber).SetSpaceKind(UnicodeSpace)
	jsn = config.NewJson(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	jsnback = config.NewJson(make([]byte, 0, 1024))

	jsn.Reset([]byte("10.2"))
	jsn.Tocbor(cbr)
	cbr.Tojson(jsnback)
	if s := string(jsnback.Bytes()); s != "10.2" {
		t.Errorf("expected %q, got %q", "10.2", s)
	}
}

func TestCbor2JsonNumber(t *testing.T) {
	// for number as integer.
	var ref1, ref2 interface{}
	testcases := []string{
		"255", "256", "-255", "-256", "65535", "65536", "-65535", "-65536",
		"4294967295", "4294967296", "-4294967295", "-4294967296",
		"9223372036854775807", "-9223372036854775807", "-9223372036854775808",
	}
	config := NewDefaultConfig()
	config = config.SetNumberKind(SmartNumber).SetSpaceKind(UnicodeSpace)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	jsnback := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		t.Logf("testcase - %v", tcase)
		json.Unmarshal([]byte(tcase), &ref1)

		jsn.Reset([]byte(tcase))
		jsn.Tocbor(cbr.Reset(nil))
		t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

		cbr.Tojson(jsnback.Reset(nil))
		if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
			t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
		}
		if !reflect.DeepEqual(ref1, ref2) {
			t.Errorf("mismatch %v, got %v", ref1, ref2)
		}
	}

	// test float-number
	tcase := "10.2"
	json.Unmarshal([]byte(tcase), &ref1)

	config = NewDefaultConfig()
	jsn = config.NewJson(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))

	jsn.Reset([]byte(tcase))
	jsn.Tocbor(cbr)

	t.Logf("%v %v", len(cbr.Bytes()), cbr.Bytes())

	cbr.Tojson(jsnback.Reset(nil))
	if err := json.Unmarshal(jsnback.Bytes(), &ref2); err != nil {
		t.Errorf("json.Unmarshal() failed for cbor %v: %v", tcase, err)
	}
	if !reflect.DeepEqual(ref1, ref2) {
		t.Errorf("mismatch %v, got %v", ref1, ref2)
	}

	// negative small integers
	buf, out := make([]byte, 64), make([]byte, 64)
	n := valint82cbor(-1, buf)
	_, m := cbor2jsont1smallint(buf[:n], out, config)
	if v := string(out[:m]); v != "-1" {
		t.Errorf("expected -1, got %v", v)
	}
}

func TestScanBadToken(t *testing.T) {
	out := make([]byte, 64)
	panicfn := func(in string, config *Config) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		json2cbor(in, out, config)
	}
	testcases := []string{
		"    ",
		"nil",
		"treu",
		"fale",
		"[  ",
		"[10  ",
		"[10,  ",
		"[10true",
		"{10",
		`{"10"true`,
		`{"10":true  `,
		`{"10":true10`,
		`(`,
		`"`,
	}
	config := NewDefaultConfig()
	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		panicfn(tcase, config)
	}
}

func TestCbor2JsonFloat32(t *testing.T) {
	var ref1, ref2 interface{}

	config := NewDefaultConfig()

	buf, out := make([]byte, 64), make([]byte, 64)
	n := valfloat322cbor(float32(10.2), buf)
	if err := json.Unmarshal([]byte("10.2"), &ref1); err != nil {
		t.Errorf("json.Unmarshal() failed for %v: %v", buf[:n], err)
	}

	_, m := cbor2jsonfloat32(buf, out, config)
	t.Logf("json - %v", string(out[:m]))
	if err := json.Unmarshal(out[:m], &ref2); err != nil {
		t.Errorf("json.Unmarshal() failed for cbor %v: %v", buf[:n], err)
	}
	if !reflect.DeepEqual(ref1, ref2) {
		t.Errorf("mismatch %v, got %v", ref1, ref2)
	}
}

func TestCbor2JsonString(t *testing.T) {
	config := NewDefaultConfig()
	buf, out := make([]byte, 64), make([]byte, 64)

	var s string
	ref := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	json.Unmarshal([]byte(ref), &s)
	n := valtext2cbor(s, buf)

	_, m := cbor2json(buf[:n], out, config)
	if err := compareJSONs(t, ref, string(out[:m])); err != nil {
		t.Errorf("%v", err)
	}
}

func TestCbor2JsonBytes(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	buf := make([]byte, 16)
	n := valbytes2cbor([]byte{0xf5}, buf)
	config := NewDefaultConfig()
	config.NewCbor(buf[:n]).Tojson(config.NewJson(make([]byte, 0, 16)))
}

//---- benchmarks

func BenchmarkCbor2JsonNull(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonInt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456567"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonFlt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("1234.12312"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonBool(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonStr(b *testing.B) {
	config := NewDefaultConfig().SetStrict(false)
	jsn := config.NewJson([]byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonStrS(b *testing.B) {
	config := NewDefaultConfig().SetStrict(true)
	jsn := config.NewJson([]byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonArr(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonMap(b *testing.B) {
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocbor(cbr)

	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkCbor2JsonTyp(b *testing.B) {
	in := testdataFile("testdata/typical.json")
	config := NewDefaultConfig()
	jsn := config.NewJson(in)
	cbr := config.NewCbor(make([]byte, 0, 10*1024))

	jsn.Tocbor(cbr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}
