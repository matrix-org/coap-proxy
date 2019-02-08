package gson

import "sort"
import "reflect"
import "testing"
import "encoding/json"

// All test cases are folded into collate_value_test.go, contains only few
// missing testcases (if any) and benchmarks.

func TestValNumber2Coll(t *testing.T) {
	testcases := [][2]interface{}{
		{json.Number("0"), float64(0)},
		{json.Number("-9223372036854775808"), int64(-9223372036854775808)},
		{json.Number("24"), float64(24)},
		{json.Number("9223372036854775808"), uint64(9223372036854775808)},
		{json.Number("-24"), float64(-24)},
	}

	var items ByteSlices

	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	col := config.NewCollate(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		nums := tcase[0].(json.Number)
		config.NewValue(nums).Tocollate(col.Reset(nil))
		value := col.Tovalue()
		if reflect.DeepEqual(value, tcase[1]) == false {
			t.Errorf("expected %v, got %v", tcase[1], value)
		}
		items = append(items, append(make([]byte, 0), col.Bytes()...))
	}

	// do a sort
	sort.Sort(items)
	refs := []interface{}{testcases[1][1], testcases[4][1], testcases[0][1],
		testcases[2][1], testcases[3][1]}
	outs := []interface{}{}
	for _, item := range items {
		outs = append(outs, config.NewCollate(item).Tovalue())
	}
	if reflect.DeepEqual(outs, refs) == false {
		t.Errorf("expected %v, got %v", refs, outs)
	}
}

func BenchmarkVal2CollNil(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(nil)

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollTrue(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(true))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollFalse(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(false))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollF64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(float64(10.121312213123123)))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollI64(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(int64(123456789)))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollINum(b *testing.B) {
	config := NewDefaultConfig()
	col := config.NewCollate(nil)
	val := config.NewValue(json.Number("-9223372036854775808"))

	for i := 0; i < b.N; i++ {
		val.Tocollate(col.Reset(nil))
	}
}

func BenchmarkVal2CollUNum(b *testing.B) {
	config := NewDefaultConfig()
	col := config.NewCollate(nil)
	val := config.NewValue(json.Number("9223372036854775808"))

	for i := 0; i < b.N; i++ {
		val.Tocollate(col.Reset(nil))
	}
}

func BenchmarkVal2CollMiss(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(MissingLiteral))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollStr(b *testing.B) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}("hello world"))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollArr(b *testing.B) {
	arr := []interface{}{nil, true, false, "hello world", 10.23122312}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(arr))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollMap(b *testing.B) {
	obj := map[string]interface{}{
		"key1": nil, "key2": true, "key3": false, "key4": "hello world",
		"key5": 10.23122312,
	}
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	val := config.NewValue(interface{}(obj))

	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}

func BenchmarkVal2CollTyp(b *testing.B) {
	config := NewDefaultConfig()
	jsn := config.NewJson(testdataFile("testdata/typical.json"))
	clt := config.NewCollate(make([]byte, 0, 10*1024))
	_, value := jsn.Tovalue()
	val := config.NewValue(value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tocollate(clt.Reset(nil))
	}
}
