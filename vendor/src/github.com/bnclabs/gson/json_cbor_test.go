package gson

import "testing"

// All test cases are folded into cbor_json_test.go, contains only few
// missing testcases (if any) and benchmarks.

func TestStrictFloat(t *testing.T) {
	config := NewDefaultConfig().SetNumberKind(FloatNumber)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson([]byte("10.2"))
	jsn.Tocbor(cbr)
	if value := cbr.Tovalue(); value != 10.2 {
		t.Errorf("expected %v, got %v", 10.2, value)
	}
}

func BenchmarkJson2CborNull(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}

	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborInt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456567"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborFlt(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("1234.12312"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborBool(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}

	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"汉语 / 漢語; Hàn\b \t\uef24yǔ "`))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborArr(b *testing.B) {
	in := ` [null,true,false,10,"tru\"e"]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborMap(b *testing.B) {
	in := `{"a":null,"b":true,"c":false,"d\"":10,"e":"tru\"e", "f":[1,2]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(in))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}

func BenchmarkJson2CborTyp(b *testing.B) {
	in := testdataFile("testdata/typical.json")
	config := NewDefaultConfig()
	jsn := config.NewJson(in)
	cbr := config.NewCbor(make([]byte, 0, 10*1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(jsn.Bytes())))
}
