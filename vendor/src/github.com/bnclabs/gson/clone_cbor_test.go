package gson

import "testing"
import "reflect"

func TestCborClone(t *testing.T) {
	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		cbr.data = cbr.data[:cbr.n+1]
		cbr.data[0] = hdrIndefiniteArray
		cbr.n++
		config.NewValue(nil).Tocbor(cbr)
		config.NewValue(true).Tocbor(config.NewValue(false).Tocbor(cbr))
		config.NewValue(int8(-1)).Tocbor(config.NewValue(uint8(1)).Tocbor(cbr))
		config.NewValue(int8(-100)).Tocbor(cbr)
		config.NewValue(uint8(100)).Tocbor(cbr)
		config.NewValue(uint16(1024)).Tocbor(cbr)
		config.NewValue(int16(-1024)).Tocbor(cbr)
		config.NewValue(uint32(1048576)).Tocbor(cbr)
		config.NewValue(int32(-1048576)).Tocbor(cbr)
		config.NewValue(uint64(1099511627776)).Tocbor(cbr)
		config.NewValue(int64(-1099511627776)).Tocbor(cbr)
		config.NewValue(float32(10.2)).Tocbor(cbr)
		config.NewValue(float64(-10.2)).Tocbor(cbr)
		config.NewValue([]byte("hello world")).Tocbor(cbr)
		config.NewValue("hello world").Tocbor(cbr)
		config.NewValue(CborUndefined(1)).Tocbor(cbr)
		config.NewValue(CborTagEpoch(1234567890)).Tocbor(cbr)
		cbr.EncodeSimpletype(100)
		config.NewValue(
			[]interface{}{12.0, nil, true, false, "hello world"}).Tocbor(cbr)
		config.NewValue(
			map[string]interface{}{
				"a": 12.0, "b": nil, "c": true, "d": false, "e": "hello world",
			}).Tocbor(cbr)
		cbr.EncodeBytechunks([][]byte{[]byte("hello"), []byte("world")})
		cbr.EncodeTextchunks([]string{"hello", "world"})
		cbr.data = cbr.data[:cbr.n+1]
		cbr.data[cbr.n] = brkstp
		cbr.n++

		out := make([]byte, 1024)
		n := cborclone(cbr.Bytes(), out, config)
		cloned := config.NewCbor(out[:n])

		ref := cbr.Tovalue()
		value := cloned.Tovalue()
		if !reflect.DeepEqual(value, ref) {
			t.Errorf("expected %v, got %v", ref, value)
		}
	}

	// with LengthPrefix
	config := NewDefaultConfig().SetContainerEncoding(LengthPrefix)
	dotest(config)
	// with Stream
	config = NewDefaultConfig().SetContainerEncoding(Stream)
	dotest(config)
}
