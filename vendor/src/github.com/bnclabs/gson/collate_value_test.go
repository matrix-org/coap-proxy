package gson

import "testing"
import "fmt"
import "reflect"

func TestVal2CollateNil(t *testing.T) {
	ref := `2\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(interface{}(nil)).Tocollate(clt)

	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != nil {
		t.Errorf("expected %v got %v", nil, value)
	}
}

func TestVal2CollateTrue(t *testing.T) {
	ref := `F\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(true).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != true {
		t.Errorf("expected %v got %v", true, value)
	}
}

func TestVal2CollateFalse(t *testing.T) {
	ref := `<\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(false).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); value != false {
		t.Errorf("expected %v got %v", false, value)
	}
}

func TestVal2CollateNumber(t *testing.T) {
	// as float64 using FloatNumber configuration
	objf, ref := float64(10.2), `P>>2102-\x00`
	config := NewDefaultConfig().SetNumberKind(FloatNumber)
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(objf).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(objf, value) {
		t.Errorf("expected %v got %v", objf, value)
	}

	// as int64 using FloatNumber configuration
	obji, ref := int64(10), `P>>21-\x00`
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(obji).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, float64(10.0)) {
		t.Errorf("expected %v, got %v", obji, value)
	}

	// as float32 using FloatNumber configuration: FIXME
	objf32, ref := float32(10.2), `P>>210199999809265137-\x00`
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(objf32).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	value := clt.Tovalue().(float64)
	if !reflect.DeepEqual(value, 10.199999809265137) {
		t.Errorf("expected %v, got %v", 10.199999809265137, float64(value))
	}

	// as float64 using FloatNumber configuration
	objf, ref = float64(10.2), `P>>2102-\x00`
	objr := float64(10.2)
	config = config.SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(objf).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, objr) {
		t.Errorf("expected %v, got %v", objr, value)
	}

	// as int64 using SmartNumber configuration
	obji, ref = int64(10), `P>>21-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(obji).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, float64(obji)) {
		t.Errorf("expected %v, got %v", obji, value)
	}

	// as uint64 using SmartNumber configuration
	obju, ref := uint64(10), `P>>21-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(obju).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}
	value = clt.Tovalue().(float64)
	if !reflect.DeepEqual(uint64(value), obju) {
		t.Errorf("expected %v, got %v", obju, value)
	}

	// as float64 using SmartNumber configuration
	objf, ref = float64(0.2), `P>02-\x00`
	config = config.SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(objf).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, objf) {
		t.Errorf("expected %v, got %v", objf, value)
	}
}

func TestVal2CollateMissing(t *testing.T) {
	obj, ref := MissingLiteral, `1\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(obj).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}

	// expect panic when not configured for missing
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config = config.UseMissing(false)
		clt := config.NewCollate(make([]byte, 0, 1024))
		config.NewValue(MissingLiteral).Tocollate(clt)
	}()
}

func TestVal2CollateString(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	testcases := [][2]interface{}{
		{"", `Z\x00\x00`},
		{"hello world", `Zhello world\x00\x00`},
		{string(MissingLiteral), `1\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0].(string), tcase[1].(string)

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		}
		value := clt.Tovalue()
		if s, ok := value.(string); ok {
			if s != obj {
				t.Errorf("expected %v got %v", obj, value)
			}
		} else if s := string(value.(Missing)); s != obj {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// missing string without doMissing configuration
	obj, ref := string(MissingLiteral), `Z~[]{}falsenilNA~\x00\x00`
	config = config.UseMissing(false)

	config.NewValue(obj).Tocollate(clt.Reset(nil))
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}
}

func TestVal2CollateBytes(t *testing.T) {
	obj, ref := []byte("hello world"), `\x82hello world\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(obj).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
		t.Errorf("expected %v, got %v", obj, value)
	}
}

func TestVal2CollateArray(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	// without length prefix
	testcases := [][2]interface{}{
		{[]interface{}{nil, true, false, 10.0, "hello"},
			`n2\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`},
		{[]interface{}{},
			`n\x00`},
		{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`n2\x00F\x00P>>21-\x00P>>2102-\x00n\x00xd>1\x00Zkey\x00\x00xd0` +
				`\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0], tcase[1].(string)
		config.NewValue(obj).Tocollate(clt.Reset(nil))

		t.Logf("%v %v", tcase[0], clt.Bytes())

		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// with length prefix
	config = config.SortbyArrayLen(true)
	clt = config.NewCollate(make([]byte, 0, 1024))
	testcases = [][2]interface{}{
		{[]interface{}{nil, true, false, 10.0, "hello"},
			`nd>5\x002\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`},
		{[]interface{}{},
			`nd0\x00\x00`},
		{[]interface{}{
			nil, true, 10.0, 10.2, []interface{}{},
			map[string]interface{}{"key": map[string]interface{}{}}},
			`nd>6\x002\x00F\x00P>>21-\x00P>>2102-\x00nd0\x00\x00xd>1\x00` +
				`Zkey\x00\x00xd0\x00\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		obj, ref := tcase[0], tcase[1].(string)
		config.NewValue(obj).Tocollate(clt.Reset(nil))

		t.Logf("%v %v", tcase[0], clt.Bytes())

		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}
}

func TestVal2CollateMap(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	// with length prefix
	testcases := [][2]interface{}{
		{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`xd>5\x00Za\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00` +
				`Zd\x00\x00P>>21-\x00Ze\x00\x00Zhello\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}

	// without length prefix
	config = config.SortbyPropertyLen(false)
	testcases = [][2]interface{}{
		{
			map[string]interface{}{
				"a": nil, "b": true, "c": false, "d": 10.0, "e": "hello"},
			`xZa\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00Zd\x00\x00P` +
				`>>21-\x00Ze\x00\x00Zhello\x00\x00\x00`},
	}
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])
		obj, ref := tcase[0], tcase[1].(string)
		clt := config.NewCollate(make([]byte, 0, 1024))

		config.NewValue(obj).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != ref {
			t.Errorf("expected %v, got %v", ref, out)
		} else if value := clt.Tovalue(); !reflect.DeepEqual(value, obj) {
			t.Errorf("expected %v, got %v", obj, value)
		}
	}
}

func BenchmarkColl2ValNil(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(nil).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValTrue(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(interface{}(true)).Tocollate(clt)
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValFalse(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(interface{}(false)).Tocollate(clt)
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValF64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(float64(10.121312213123123)).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValI64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(int64(123456789)).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValMiss(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(MissingLiteral).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValStr(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue("hello world").Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(arr).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(obj).Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}

func BenchmarkColl2ValTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"))
	clt := config.NewCollate(make([]byte, 0, 10*1024))
	jsn.Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tovalue()
	}
}
