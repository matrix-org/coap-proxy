package gson

import "reflect"
import "testing"
import "encoding/json"

// All test cases are folded into cbor_value_test.go, contains only few
// missing testcases (if any) and benchmarks.

func TestValNumber2Cbor(t *testing.T) {
	testcases := [][2]interface{}{
		{json.Number("9223372036854775808"), uint64(9223372036854775808)},
		{json.Number("-9223372036854775808"), int64(-9223372036854775808)},
	}

	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	cbr := config.NewCbor(make([]byte, 0, 1024))

	for _, tcase := range testcases {
		nums := tcase[0].(json.Number)
		config.NewValue(nums).Tocbor(cbr.Reset(nil))
		value := cbr.Tovalue()
		if reflect.DeepEqual(value, tcase[1]) == false {
			t.Errorf("expected %v, got %v", tcase[1], value)
		}
	}
}

func BenchmarkVal2CborNull(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(nil)

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborTrue(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(true)

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborFalse(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(false)
	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborUint8(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(uint8(255))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborInt8(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(int8(-128))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborUint16(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(uint16(65535))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborInt16(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(int16(-32768))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborUint32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(uint32(4294967295))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborInt32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(int32(-2147483648))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborUint64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(uint64(18446744073709551615))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborInt64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(int64(-2147483648))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborFlt32(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(float32(10.2))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborFlt64(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(float64(10.2))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborINum(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(json.Number("-9223372036854775808"))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborUNum(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(json.Number("9223372036854775808"))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborTBytes(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue([]byte("hello world"))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborText(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue("hello world")

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborArr0(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(make([]interface{}, 0))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborArr5(b *testing.B) {
	value := interface{}([]interface{}{5, 5.0, "hello world", true, nil})
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(value)

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborMap0(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(nil)
	val := config.NewValue(make([][2]interface{}, 0))

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborMap5(b *testing.B) {
	value := map[string]interface{}{
		"key0": 5,
		"key1": 5.0,
		"key2": "hello world",
		"key3": true,
		"key4": nil,
	}
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	val := config.NewValue(value)

	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}

func BenchmarkVal2CborTyp(b *testing.B) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	jsn := config.NewJson(testdataFile("testdata/typical.json"))
	_, value := jsn.Tovalue()
	val := config.NewValue(value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Tocbor(cbr.Reset(nil))
	}
	b.SetBytes(int64(len(cbr.Bytes())))
}
