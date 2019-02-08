package gson

import "testing"
import "fmt"

func TestJson2CollateNil(t *testing.T) {
	inp, ref := "null", `2\x00`

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out := fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}

	jsn := config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != inp {
		t.Errorf("expected %v, got %v", inp, s)
	}
}

func TestJson2CollateTrue(t *testing.T) {
	inp, ref := "true", `F\x00`

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out := fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}

	jsn := config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != inp {
		t.Errorf("expected %v, got %v", inp, s)
	}
}

func TestJson2CollateFalse(t *testing.T) {
	inp, ref := "false", `<\x00`

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out := fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != ref {
		t.Errorf("expected %v, got %v", ref, out)
	}

	jsn := config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != inp {
		t.Errorf("expected %v, got %v", inp, s)
	}
}

func TestJson2CollateNumber(t *testing.T) {
	// as float64 using FloatNumber configuration
	inp, refcode, reftxt := "10.2", `P>>2102-\x00`, "1.02e+01"
	config := NewDefaultConfig().SetNumberKind(FloatNumber)
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out := fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn := config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v, got %v", reftxt, s)
	}

	// as int64 using FloatNumber configuration
	inp, refcode, reftxt = "10", `P>>21-\x00`, "1e+01"
	config = NewDefaultConfig().SetNumberKind(FloatNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out = fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn = config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v got %v", reftxt, s)
	}

	// as float64 using IntNumber configuration
	inp, refcode, reftxt = "10.2", `P>>2102-\x00`, "1.02e+01"
	config = NewDefaultConfig().SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out = fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn = config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v, got %v", reftxt, s)
	}

	// as int64 using IntNumber configuration
	inp, refcode, reftxt = "10", `P>>21-\x00`, "1e+01"
	config = NewDefaultConfig().SetNumberKind(SmartNumber)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out = fmt.Sprintf("%q", string(clt.Bytes()))
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn = config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v, got %v", reftxt, s)
	}
}

func TestJson2CollateString(t *testing.T) {
	// empty string
	inp, refcode, reftxt := `""`, `Z\x00\x00`, `""`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out := fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn := config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v, got %v", reftxt, s)
	}

	// normal string
	inp, refcode = `"hello world"`, `Zhello world\x00\x00`
	dotest := func(config *Config) {
		clt = config.NewCollate(make([]byte, 0, 1024))

		config.NewJson([]byte(inp)).Tocollate(clt)
		out = fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}

		jsn = config.NewJson(make([]byte, 0, 1024))
		clt.Tojson(jsn)
		if s := string(jsn.Bytes()); s != inp {
			t.Errorf("expected %v, got %v", inp, s)
		}
	}
	dotest(NewDefaultConfig().SetStrict(true))
	dotest(NewDefaultConfig().SetStrict(false))

	// missing string
	inp, refcode = fmt.Sprintf(`"%s"`, MissingLiteral), `1\x00`
	reftxt = string(MissingLiteral)
	config = NewDefaultConfig()
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn = config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != reftxt {
		t.Errorf("expected %v, got %v", reftxt, s)
	}

	// missing string without doMissing configuration
	inp = fmt.Sprintf(`"%s"`, MissingLiteral)
	refcode = `Z~[]{}falsenilNA~\x00\x00`
	config = NewDefaultConfig().UseMissing(false)
	clt = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)
	out = fmt.Sprintf("%q", clt.Bytes())
	if out = out[1 : len(out)-1]; out != refcode {
		t.Errorf("expected %v, got %v", refcode, out)
	}

	jsn = config.NewJson(make([]byte, 0, 1024))
	clt.Tojson(jsn)
	if s := string(jsn.Bytes()); s != inp {
		t.Errorf("expected %v, got %v", inp, s)
	}
}

func TestJson2CollateArray(t *testing.T) {
	config := NewDefaultConfig()

	// without length prefix
	testcases := [][4]string{
		{`[null,true,false,10.0,"hello"]`,
			`n2\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`,
			`nd>5\x002\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`,
			`[null,true,false,1e+01,"hello"]`},
		{`[]`,
			`n\x00`,
			`nd0\x00\x00`,
			`[]`},
		{`[null,true,10.0,10.2,[],{"key":{}}]`,
			`n2\x00F\x00P>>21-\x00P>>2102-\x00n\x00xd>1\x00Zkey\x00\x00xd0` +
				`\x00\x00\x00\x00`,
			`nd>6\x002\x00F\x00P>>21-\x00P>>2102-\x00nd0\x00\x00xd>1\x00` +
				`Zkey\x00\x00xd0\x00\x00\x00\x00`,
			`[null,true,1e+01,1.02e+01,[],{"key":{}}]`},
	}

	config = NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])

		inp, refcode, reftxt := tcase[0], tcase[1], tcase[3]

		config.NewJson([]byte(inp)).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}

		clt.Tojson(jsn.Reset(nil))
		if s := string(jsn.Bytes()); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}

	// with length prefix
	config = NewDefaultConfig().SortbyArrayLen(true)
	clt = config.NewCollate(make([]byte, 0, 1024))
	jsn = config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])

		inp, refcode, reftxt := tcase[0], tcase[2], tcase[3]
		config.NewJson([]byte(inp)).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}

		clt.Tojson(jsn.Reset(nil))
		if s := string(jsn.Bytes()); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}
}

func TestJson2CollateMap(t *testing.T) {
	// with length prefix
	testcases := [][4]string{
		{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["world"]}`,
			`xd>6\x00Za\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00Zd\x00` +
				`\x00P>>21-\x00Ze\x00\x00Zhello\x00\x00Zf\x00\x00nZworld\x00` +
				`\x00\x00\x00`,
			`xZa\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00Zd\x00\x00P` +
				`>>21-\x00Ze\x00\x00Zhello\x00\x00Zf\x00\x00nZworld\x00` +
				`\x00\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":1e+01,"e":"hello","f":["world"]}`,
		},
	}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])

		inp, refcode, reftxt := tcase[0], tcase[1], tcase[3]
		config.NewJson([]byte(inp)).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}

		clt.Tojson(jsn.Reset(nil))
		if s := string(jsn.Bytes()); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}

	// without length prefix, and different length for keys
	config = NewDefaultConfig().SetMaxkeys(10).SortbyPropertyLen(false)
	clt = config.NewCollate(make([]byte, 0, 1024))
	jsn = config.NewJson(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0])

		inp, refcode, reftxt := tcase[0], tcase[2], tcase[3]
		config.NewJson([]byte(inp)).Tocollate(clt.Reset(nil))
		out := fmt.Sprintf("%q", clt.Bytes())
		if out = out[1 : len(out)-1]; out != refcode {
			t.Errorf("expected %v, got %v", refcode, out)
		}

		clt.Tojson(jsn.Reset(nil))
		if s := string(jsn.Bytes()); s != reftxt {
			t.Errorf("expected %v, got %v", reftxt, s)
		}
	}
}

func BenchmarkColl2JsonNil(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte("null")).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonTrue(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte("true")).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonFalse(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte("false")).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonF64(b *testing.B) {
	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte("10.121312213123123")).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonI64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte("123456789")).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonMiss(b *testing.B) {
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonStr(b *testing.B) {
	config := NewDefaultConfig().SetStrict(false)
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte(`"hello world"`)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonStrS(b *testing.B) {
	config := NewDefaultConfig().SetStrict(true)
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte(`"hello world"`)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonArr(b *testing.B) {
	inp := `[null,true,false,"hello world",10.23122312]`

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}

func BenchmarkColl2JsonTyp(b *testing.B) {
	inp := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 10*1024))
	jsn := config.NewJson(make([]byte, 0, 10*1024))

	config.NewJson([]byte(inp)).Tocollate(clt)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clt.Tojson(jsn.Reset(nil))
	}
}
