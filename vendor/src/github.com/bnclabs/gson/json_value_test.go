package gson

import "encoding/json"
import "reflect"
import "testing"

func TestJsonEmpty2Value(t *testing.T) {
	config := NewDefaultConfig()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic")
		}
	}()
	config.NewJson([]byte("")).Tovalue()
}

func TestScanNull(t *testing.T) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	if jsnrem, val := jsn.Reset([]byte("null")).Tovalue(); jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if val != nil {
		t.Errorf("`null` should be parsed to nil")
	}

	// test bad input
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Reset([]byte("nil")).Tovalue()
	}()
}

func TestScanBool(t *testing.T) {
	testcases := []string{"true", "false"}

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		var refval interface{}

		jsn.Reset([]byte(tcase))
		json.Unmarshal([]byte(tcase), &refval)
		if jsnrem, val := jsn.Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(bool); !ok || v != refval.(bool) {
			t.Errorf("%q should be parsed to %v", tcase, refval)
		}
	}

	// test bad input
	fn := func(input string) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Reset([]byte(input)).Tovalue()
	}
	fn("trrr")
	fn("flse")
}

func TestScanIntegers(t *testing.T) {
	testcases := []string{"10", "-10"}

	var ref interface{}

	for _, tcase := range testcases {
		json.Unmarshal([]byte(tcase), &ref)

		config := NewDefaultConfig()
		config = config.SetNumberKind(SmartNumber).SetSpaceKind(AnsiSpace)
		jsn := config.NewJson(make([]byte, 0, 1024))
		if jsnrem, val := jsn.Reset([]byte(tcase)).Tovalue(); jsnrem != nil {
			t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
		} else if v, ok := val.(int64); !ok || v != int64(ref.(float64)) {
			t.Errorf("%q int should be parsed to %T %v", tcase, val, ref)
		}
	}

	testcases = []string{
		"0.1", "-0.1", "10.1", "-10.1", "-10E-1", "-10e+1", "10E-1", "10e+1",
	}
	for _, tcase := range testcases {
		json.Unmarshal([]byte(tcase), &ref)

		config := NewDefaultConfig()
		config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
		jsn := config.NewJson(make([]byte, 0, 1024))
		_, val := jsn.Reset([]byte(tcase)).Tovalue()
		if v, ok := val.(float64); !ok || v != ref.(float64) {
			t.Errorf("%q int should be parsed to %v", tcase, ref)
		}
	}
}

func TestScanMalformed(t *testing.T) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(SmartNumber).SetSpaceKind(AnsiSpace)
	config = config.SetStrict(true)

	for _, tcase := range scaninvalid {
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("expected panic")
				}
			}()
			t.Logf("%v", tcase)
			json2value(tcase, config)
		}()
	}
}

func TestScanValues(t *testing.T) {
	testcases := append(scanvalid, []string{
		string(mapValue),
		string(allValueIndent),
		string(allValueCompact),
		string(pallValueIndent),
		string(pallValueCompact),
	}...)

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024*1024))

	for _, tcase := range testcases {
		var ref interface{}

		t.Logf("%v", tcase)
		json.Unmarshal([]byte(tcase), &ref)

		jsn.Reset([]byte(tcase))
		_, val := jsn.Tovalue()
		if reflect.DeepEqual(val, ref) == false {
			t.Errorf("%q should be parsed as: %v, got %v", tcase, ref, val)
		}
	}
}

func TestCodeJSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping code.json.gz")
	}

	var ref interface{}
	data := testdataFile("testdata/code.json.gz")
	json.Unmarshal(data, &ref)

	config := NewDefaultConfig()
	jsn := config.NewJson(data)
	if jsnrem, val := jsn.Tovalue(); jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if reflect.DeepEqual(val, ref) == false {
		t.Errorf("codeJSON parsing failed with reference: %v", ref)
	}
}

func BenchmarkJson2ValNil(b *testing.B) {
	config := NewDefaultConfig()
	in := "null"
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalNil(b *testing.B) {
	var val interface{}

	in := []byte("null")
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValBool(b *testing.B) {
	config := NewDefaultConfig()
	in := "true"
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalBool(b *testing.B) {
	var val interface{}

	in := []byte("true")
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValNum(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	in := "10.2"
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalNum(b *testing.B) {
	var val interface{}

	in := []byte("100000.23")
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValString(b *testing.B) {
	config := NewDefaultConfig()
	in := `"汉语 / 漢語; Hàn\b \tyǔ "`
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalStr(b *testing.B) {
	var val interface{}

	in := []byte(`"汉语 / 漢語; Hàn\b \tyǔ "`)
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &val)
	}
}

func BenchmarkJson2ValArr5(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalArr5(b *testing.B) {
	var a []interface{}

	in := []byte(` [null,true,false,10,"tru\"e"]`)

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &a)
	}
}

func BenchmarkJson2ValMap5(b *testing.B) {
	in := `{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in))

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalMap5(b *testing.B) {
	var m map[string]interface{}

	in := []byte(`{"a": null, "b" : true,"c":false, "d\"":-10E-1, "e":"tru\"e" }`)

	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(in, &m)
	}
}

func BenchmarkJson2ValTyp(b *testing.B) {
	in := string(testdataFile("testdata/typical.json"))

	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(in))
	b.SetBytes(int64(len(in)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalTyp(b *testing.B) {
	var m map[string]interface{}

	data := testdataFile("testdata/typical.json")
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &m)
	}
}

func BenchmarkJson2ValCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(data)
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		jsn.Tovalue()
	}
}

func BenchmarkUnmarshalCgz(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	var m map[string]interface{}

	data := testdataFile("testdata/code.json.gz")
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		json.Unmarshal(data, &m)
	}
}
