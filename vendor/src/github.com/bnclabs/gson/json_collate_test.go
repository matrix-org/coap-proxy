package gson

import "fmt"
import "testing"

// All test cases are folded into collate_json_test.go, contains only few
// missing testcases (if any) and benchmarks.

func BenchmarkJson2CollNil(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("null"))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollTrue(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("true"))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollFalse(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("false"))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollF64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("10.121312213123123"))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollI64(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte("123456789"))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollMiss(b *testing.B) {
	inp := fmt.Sprintf(`"%s"`, MissingLiteral)

	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp))
	clt := config.NewCollate(make([]byte, 0, 1024))

	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollStr(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(`"hello world"`))
	clt := config.NewCollate(make([]byte, 0, 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollArr(b *testing.B) {
	inp := `[null,true,false,"hello world",10.23122312]`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(inp))
	clt := config.NewCollate(make([]byte, 0, 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollMap(b *testing.B) {
	inp := `{"key1":null,"key2":true,"key3":false,"key4":"hello world",` +
		`"key5":10.23122312}`
	config := NewDefaultConfig().SetMaxkeys(10)
	jsn := config.NewJson([]byte(inp))
	clt := config.NewCollate(make([]byte, 0, 1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkJson2CollTyp(b *testing.B) {
	inp := testdataFile("testdata/typical.json")
	config := NewDefaultConfig().SetMaxkeys(100)
	jsn := config.NewJson(inp)
	clt := config.NewCollate(make([]byte, 0, 10*1024))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsn.Tocollate(clt.Reset(nil))
	}
}
