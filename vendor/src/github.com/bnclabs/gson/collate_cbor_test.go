package gson

import "testing"
import "bytes"
import "fmt"

func TestCbor2CollateNil(t *testing.T) {
	inp, ref, config := "null", `2\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %q, got %q", ref, seqn)
	}
}

func TestCbor2CollateTrue(t *testing.T) {
	inp, ref, config := "true", `F\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateFalse(t *testing.T) {
	inp, ref, config := "false", `<\x00`, NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	config.NewJson([]byte(inp)).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != ref {
		t.Errorf("expected %v, got %v", ref, seqn)
	}
}

func TestCbor2CollateNumber(t *testing.T) {
	testcases := [][3]interface{}{
		{"10.2", `P>>2102-\x00`, FloatNumber},
		{"10", `P>>21-\x00`, FloatNumber},
		{"10.2", `P>>2102-\x00`, SmartNumber},
		{"10", `P>>21-\x00`, SmartNumber},
		{"10", `P>>21-\x00`, FloatNumber},
		{"-10", `P--78>\x00`, SmartNumber},
		{"-10", `P--78>\x00`, FloatNumber},
		{"200", `P>>32-\x00`, SmartNumber},
		{"200", `P>>32-\x00`, FloatNumber},
		{"-200", `P--67>\x00`, SmartNumber},
		{"-200", `P--67>\x00`, FloatNumber},
		{
			"4294967297", `P>>>2104294967297-\x00`, FloatNumber},
		{
			"-4294967297", `P---7895705032702>\x00`, FloatNumber},
		{
			"4294967297", `P>>>2104294967297-\x00`, SmartNumber},
		{
			"-4294967297", `P---7895705032702>\x00`, SmartNumber},
		{
			"9007199254740992", `P>>>2169007199254740992-\x00`, FloatNumber},
		{
			"-9007199254740993", `P---7830992800745259007>\x00`, FloatNumber},
		{
			"9007199254740992", `P>>>2169007199254740992-\x00`, SmartNumber},
		{
			"-9007199254740993", `P---7830992800745259006>\x00`, SmartNumber},
	}

	for _, tcase := range testcases {
		inp, refcode := tcase[0].(string), tcase[1].(string)
		nk := tcase[2].(NumberKind)

		t.Logf("%v", inp)

		config := NewDefaultConfig().SetNumberKind(nk)
		clt := config.NewCollate(make([]byte, 0, 1024))
		cbr := config.NewCbor(make([]byte, 0, 1024))
		cltb := config.NewCollate(make([]byte, 0, 1024))

		config.NewJson([]byte(inp)).Tocollate(clt).Tocbor(cbr).Tocollate(cltb)

		seqn := fmt.Sprintf("%q", cltb.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateString(t *testing.T) {
	testcases := [][2]string{
		{`""`, `Z\x00\x00`},
		{`"hello world"`, `Zhello world\x00\x00`},
		{fmt.Sprintf(`"%s"`, MissingLiteral), `1\x00`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		t.Logf("%v", inp)

		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// missing string without doMissing configuration
	inp := []byte(fmt.Sprintf(`"%s"`, MissingLiteral))
	refcode := `Z~[]{}falsenilNA~\x00\x00`
	config = NewDefaultConfig().UseMissing(false)
	clt = config.NewCollate(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	cltback = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson(inp).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)
	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}

	// utf8 string
	inp = []byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`)

	config = NewDefaultConfig()
	clt = config.NewCollate(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	cltback = config.NewCollate(make([]byte, 0, 1024))

	config.NewJson(inp).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)

	if bytes.Compare(clt.Bytes(), cltback.Bytes()) != 0 {
		t.Errorf("expected %v, got %v", clt.Bytes(), cltback.Bytes())
	}
}

func TestCbor2CollateBytes(t *testing.T) {
	inp, refcode := []byte("hello world"), `\x82hello world\x00`
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue(inp).Tocollate(clt).Tocbor(cbr).Tocollate(cltback)
	seqn := fmt.Sprintf("%q", cltback.Bytes())
	if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
		t.Errorf("expected %v, got %v", refcode, seqn)
	}
}

func TestCbor2CollateArray(t *testing.T) {

	// without length prefix
	testcases := [][4]string{
		{`[]`,
			`n\x00`,
			`nd0\x00\x00`,
			`[]`},
		{`[null,true,false,10.0,"hello"]`,
			`n2\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`,
			`nd>5\x002\x00F\x00<\x00P>>21-\x00Zhello\x00\x00\x00`,
			`[null,true,false,+0.1e+2,"hello"]`},
		{`[null,true,10.0,10.2,[],{"key":{}}]`,
			`n2\x00F\x00P>>21-\x00P>>2102-\x00n\x00xd>1\x00Zkey\x00` +
				`\x00xd0\x00\x00\x00\x00`,
			`nd>6\x002\x00F\x00P>>21-\x00P>>2102-\x00nd0\x00\x00xd>1` +
				`\x00Zkey\x00\x00xd0\x00\x00\x00\x00`,
			`[null,true,+0.1e+2,+0.102e+2,[],{"key":{}}]`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]

		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// with sort by length and length prefix
	config = config.SortbyArrayLen(true).SetContainerEncoding(LengthPrefix)
	clt = config.NewCollate(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	cltback = config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// with sort by length and stream encoding
	config = config.SortbyArrayLen(true).SetContainerEncoding(Stream)
	clt = config.NewCollate(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	cltback = config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func TestCbor2CollateMap(t *testing.T) {
	// with length prefix
	testcases := [][4]string{
		{
			`{}`,
			`xd0\x00\x00`,
			`x\x00`,
			`{}`},
		{
			`{"a":null,"b":true,"c":false,"d":10.0,"e":"hello","f":["wo"]}`,
			`xd>6\x00Za\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00Zd\x00` +
				`\x00P>>21-\x00Ze\x00\x00Zhello\x00\x00Zf\x00\x00nZwo\x00` +
				`\x00\x00\x00`,
			`xZa\x00\x002\x00Zb\x00\x00F\x00Zc\x00\x00<\x00Zd\x00\x00P` +
				`>>21-\x00Ze\x00\x00Zhello\x00\x00Zf\x00\x00nZwo\x00\x00` +
				`\x00\x00`,
			`{"a":null,"b":true,"c":false,"d":+0.1e+2,"e":"hello","f":["wo"]}`},
	}

	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	cltback := config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[1]
		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}

	// without length prefix, and different length for keys
	config = NewDefaultConfig().SetMaxkeys(10).SortbyPropertyLen(false)
	config = config.SetContainerEncoding(LengthPrefix)
	clt = config.NewCollate(make([]byte, 0, 1024))
	cbr = config.NewCbor(make([]byte, 0, 1024))
	cltback = config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		inp, refcode := tcase[0], tcase[2]
		config.NewJson(
			[]byte(inp)).Tocollate(
			clt.Reset(nil)).Tocbor(
			cbr.Reset(nil)).Tocollate(cltback.Reset(nil))

		seqn := fmt.Sprintf("%q", cltback.Bytes())
		if seqn = seqn[1 : len(seqn)-1]; seqn != refcode {
			t.Logf("%v", tcase[0])
			t.Errorf("expected %v, got %v", refcode, seqn)
		}
	}
}

func BenchmarkColl2CborNil(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborTrue(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("true"))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborFalse(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborF64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("10.121312213123123"))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborI64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456789"))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborMiss(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(fmt.Sprintf(`"%s"`, MissingLiteral)))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"hello world"`))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborArr(b *testing.B) {
	in := []byte(`[null,true,false,"hello world",10.23122312]`)

	config := NewDefaultConfig()
	jsn := config.NewJson(in)
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`

	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp))
	clt := config.NewCollate(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}

func BenchmarkColl2CborTyp(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig().SetMaxkeys(100)
	jsn := config.NewJson(data)
	clt := config.NewCollate(make([]byte, 0, 10*1024))
	cbr := config.NewCbor(make([]byte, 0, 10*1024))

	jsn.Tocollate(clt)

	for i := 0; i < b.N; i++ {
		clt.Tocbor(cbr.Reset(nil))
	}
}
