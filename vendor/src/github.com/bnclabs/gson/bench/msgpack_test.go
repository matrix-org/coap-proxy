package main

import "bytes"
import "testing"

import "github.com/vmihailenco/msgpack"

func BenchmarkMPEncNil(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(nil)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncInt64(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(9223372036854775807)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncUint64(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(9223372036854775807)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncFloat64(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(123456.123456)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncBool(b *testing.B) {
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(true)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncArr(b *testing.B) {
	val := []interface{}{5, 5.0, "hello world", true, nil}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(val)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncMap(b *testing.B) {
	val := map[string]interface{}{
		"key0": 5,
		"key1": 5.0,
		"key2": "hello world",
		"key3": true,
		"key4": nil,
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(val)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncBytes(b *testing.B) {
	val := []byte("hello world")
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(val)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPEncString(b *testing.B) {
	val := "hello world"
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(val)
	}
	b.SetBytes(int64(buf.Len()))
}

func BenchmarkMPDecNil(b *testing.B) {
	var val interface{}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(nil)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(val)
	}
}

func BenchmarkMPDecInt64(b *testing.B) {
	var val int64
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(9223372036854775807)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}

func BenchmarkMPDecUint64(b *testing.B) {
	var val uint64
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(9223372036854775807)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}

func BenchmarkMPDecFloat64(b *testing.B) {
	var val float64
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(123456.123456)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}

func BenchmarkMPDecBool(b *testing.B) {
	var val bool
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(true)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}

func BenchmarkMPDecArr(b *testing.B) {
	var val []interface{}
	inp := []interface{}{5, 5.0, "hello world", true, nil}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(inp)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
		val = nil
	}
}

func BenchmarkMPDecMap(b *testing.B) {
	var val map[string]interface{}
	inp := map[string]interface{}{
		"key0": 5,
		"key1": 5.0,
		"key2": "hello world",
		"key3": true,
		"key4": nil,
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(inp)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
		val = nil
	}
}

func BenchmarkMPDecBytes(b *testing.B) {
	var val []byte
	inp := []byte("hello world")
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(inp)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}

func BenchmarkMPDecString(b *testing.B) {
	var val string
	inp := "hello world"
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	enc := msgpack.NewEncoder(buf)
	enc.Encode(inp)

	out := make([]byte, 1024)
	out = out[:copy(out, buf.Bytes())]
	dec := msgpack.NewDecoder(buf)
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.Write(out)
		dec.Decode(&val)
	}
}
