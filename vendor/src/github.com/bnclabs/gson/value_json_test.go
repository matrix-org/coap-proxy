package gson

import "reflect"
import "testing"
import "encoding/json"

func TestNil2Json(t *testing.T) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewValue(nil).Tojson(jsn)
	if s := string(jsn.Bytes()); s != "null" {
		t.Errorf("expected %q, got %q", "null", s)
	}
}

func TestBool2Json(t *testing.T) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewValue(true).Tojson(jsn)
	if s := string(jsn.Bytes()); s != "true" {
		t.Errorf("expected %q, got %q", "true", s)
	}

	config.NewValue(false).Tojson(jsn.Reset(nil))
	if s := string(jsn.Bytes()); s != "false" {
		t.Errorf("expected %q, got %q", "false", s)
	}
}

func TestNumber2Json(t *testing.T) {
	testcases := []interface{}{
		10.0, -10.0, 0.1, -0.1, 10.1, -10.1, -10E-1, -10e+1, 10E-1, 10e+1,
		byte(10), int8(-10), int16(1024), uint16(1024), int32(-1048576),
		uint32(1048576), int(-1048576), uint(1048576),
		int64(-1099511627776), uint64(1099511627776),
	}
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		var ref, out interface{}

		config.NewValue(tcase).Tojson(jsn.Reset(nil))
		json.Unmarshal(jsn.Bytes(), &out)
		data, _ := json.Marshal(tcase)
		json.Unmarshal(data, &ref)
		if !reflect.DeepEqual(out, ref) {
			t.Errorf("expected %v, got %v", ref, out)
		}
	}

	// convert float32
	var ref, out float32
	config.NewValue(float32(109.951162)).Tojson(jsn.Reset(nil))
	json.Unmarshal(jsn.Bytes(), &out)
	data, _ := json.Marshal(float32(109.951162))
	json.Unmarshal(data, &ref)
	if !reflect.DeepEqual(out, ref) {
		t.Errorf("expected %v, got %v", ref, out)
	}
}

func TestValNumber2Json(t *testing.T) {
	testcases := [][2]interface{}{
		{json.Number("9223372036854775808"), uint64(9223372036854775808)},
		{json.Number("-9223372036854775808"), int64(-9223372036854775808)},
	}

	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	jsn := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		nums := tcase[0].(json.Number)
		config.NewValue(nums).Tojson(jsn.Reset(nil))
		if s := string(jsn.Bytes()); s != string(nums) {
			t.Errorf("expected %q, got %q", nums, s)
		}
		_, value := jsn.Tovalue()
		if reflect.DeepEqual(value, tcase[1]) == false {
			t.Errorf("expected %v, got %v", tcase[1], value)
		}
	}
}

func TestMaptoJson(t *testing.T) {
	txt := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))

	_, value := config.NewJson([]byte(txt)).Tovalue()
	config.NewValue(GolangMap2cborMap(value)).Tojson(jsn)
	if _, val := jsn.Tovalue(); !reflect.DeepEqual(value, val) {
		t.Errorf("expected %v, got %v", value, val)
	}
}

func TestMapUint64toJson(t *testing.T) {
	txt := `{"key1":10, "key2":9223372036854775808, "key3":-9223372036854775808}`
	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	jsn := config.NewJson(make([]byte, 0, 1024))

	_, value := config.NewJson([]byte(txt)).Tovalue()
	config.NewValue(value).Tojson(jsn)
	if _, val := jsn.Tovalue(); !reflect.DeepEqual(value, val) {
		t.Errorf("expected %v, got %v", value, val)
	}
}

func TestValues2Json(t *testing.T) {
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
		var value interface{}

		t.Logf("%v", tcase)
		json.Unmarshal([]byte(tcase), &value)
		config.NewValue(value).Tojson(jsn.Reset(nil))

		_, outval := jsn.Tovalue()
		if reflect.DeepEqual(outval, value) == false {
			t.Errorf("expected %v, got %v", value, outval)
		}
	}
}

func TestCodeVal2Json(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping code.json.gz")
	}

	var value interface{}

	data := testdataFile("testdata/code.json.gz")

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, len(data)*2))
	json.Unmarshal(data, &value)
	jsnrem, outval := config.NewValue(value).Tojson(jsn).Tovalue()

	if jsnrem != nil {
		t.Errorf("remaining text after parsing should be empty, %q", jsnrem)
	} else if reflect.DeepEqual(outval, value) == false {
		t.Errorf("codeJSON expected %v, got %v", value, outval)
	}
}

func BenchmarkVal2JsonNil(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(nil)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalNil(b *testing.B) {
	var data []byte

	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(nil)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonBool(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(true)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalBool(b *testing.B) {
	var data []byte

	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(true)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonNum(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(10.2)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalNum(b *testing.B) {
	var data []byte

	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(10.2)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonString(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(`"汉语 / 漢語; Hàn\b \tyǔ "`)

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalString(b *testing.B) {
	var data []byte

	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(`"汉语 / 漢語; Hàn\b \tyǔ "`)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonArr5(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue([]interface{}{nil, true, false, 10, "tru\"e"})

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalArr5(b *testing.B) {
	var data []byte

	val := []interface{}{nil, true, false, 10, "tru\"e"}
	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(val)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonMap5(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(map[string]interface{}{
		"a": nil, "b": true, "c": false, "d\"": -10E-1, "e": "tru\"e",
	})

	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkVal2JsonMapUint64(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(SmartNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 0, 1024))
	val := config.NewValue(map[string]interface{}{
		"key1": 10,
		"key2": int64(9223372036854775807),
		"key3": int64(-9223372036854775808),
	})
	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalMap5(b *testing.B) {
	var data []byte

	val := map[string]interface{}{
		"a": nil, "b": true, "c": false, "d\"": -10E-1, "e": "tru\"e",
	}
	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(val)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonTyp(b *testing.B) {
	var ref interface{}

	data := testdataFile("testdata/typical.json")
	if err := json.Unmarshal(data, &ref); err != nil {
		b.Fatal(err)
	}

	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 0, len(data)*2))
	val := config.NewValue(ref)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalTyp(b *testing.B) {
	var value interface{}

	data := testdataFile("testdata/typical.json")
	if err := json.Unmarshal(data, &value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(value)
	}
	b.SetBytes(int64(len(data)))
}

func BenchmarkVal2JsonCgz(b *testing.B) {
	var ref interface{}

	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	if err := json.Unmarshal(data, &ref); err != nil {
		b.Fatal(err)
	}
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	jsn := config.NewJson(make([]byte, 0, len(data)*2))
	val := config.NewValue(ref)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tojson(jsn.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkMarshalCgz(b *testing.B) {
	var value interface{}

	if testing.Short() {
		b.Skip("skipping code.json.gz")
	}

	data := testdataFile("testdata/code.json.gz")
	if err := json.Unmarshal(data, &value); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ = json.Marshal(value)
	}
	b.SetBytes(int64(len(data)))
}
