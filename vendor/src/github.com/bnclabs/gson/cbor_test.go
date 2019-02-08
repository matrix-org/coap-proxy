package gson

import "testing"
import "reflect"

func TestCborMajor(t *testing.T) {
	if typ := cborMajor(0xff); typ != 0xe0 {
		t.Errorf("fail major() got %v wanted 0xe0", typ)
	}
}

func TestCborSmallInt(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	for i := int8(-24); i < 24; i++ { // SmallInt is -24..23
		cbr.EncodeSmallint(i)
		item := cbr.Tovalue()
		if val1, ok := item.(int64); ok && val1 != int64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val1, i)
		} else if val2, ok := item.(uint64); ok && val2 != uint64(i) {
			t.Errorf("fail decode on SmallInt: %x, want %x", val2, i)
		}
		cbr.Reset(nil)
	}
}

func TestCborSimpleType(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	// test encoding type7/simpletype < 20
	for i := 0; i < 20; i++ {
		cbr.EncodeSimpletype(byte(i))
		item := cbr.Tovalue()
		if item.(byte) != byte(i) {
			t.Errorf("fail decode on simple-type: %v want %v", item, i)
		}
		cbr.Reset(nil)
	}

	// test decoding typ7/simpletype extended byte
	for i := 32; i < 255; i++ {
		cbr.EncodeSimpletype(byte(i))
		if item := cbr.Tovalue(); item.(byte) != byte(i) {
			t.Errorf("fail codec simpletype extended: %v", item)
		}
		cbr.Reset(nil)
	}
}

func TestCborMapslice(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))

	items := [][2]interface{}{
		{"first", true},
		{"second", 12.2},
		{"third", []interface{}{true, false, 10.2}},
		{
			"fourth",
			[][2]interface{}{
				{"a", 10.2},
				{"b", 11.2},
			},
		},
	}

	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[0] = cborType5 | cborIndefiniteLength
	cbr.n++
	cbr.EncodeMapslice(items)
	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[cbr.n] = cborType7 | cborItemBreak
	cbr.n++
	value, ref := cbr.Tovalue(), CborMap2golangMap(items)
	if !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCborItem(t *testing.T) {
	txt := `{"a": 10, "b": 1024, "c": 1048576, "d": 8589934592, ` +
		`"an": -10, "bn": -1024, "cn": -1048576, "dn": -8589934592, ` +
		`"arr": [1,2], "nestd":[[23]], ` +
		`"dict": {"a":10, "b":20, "": 2}, "": 1}`

	testcases := [][2]interface{}{
		{"/a", 10},
		{"/b", 1024},
		{"/c", 1048576},
		{"/d", 8589934592},
		{"/an", -10},
		{"/bn", -1024},
		{"/cn", -1048576},
		{"/dn", -8589934592},
		{"/arr", []interface{}{1, 2}},
		{
			"/dict",
			map[string]interface{}{"a": 10, "b": 20, "": 2},
		},
	}
	fn := func(ptr *Jsonpointer, ref interface{}, cbr, item *Cbor) {
		t.Logf("%v", string(ptr.Path()))

		cbr.Get(ptr, item)
		if value := item.Tovalue(); reflect.DeepEqual(value, ref) {
			t.Errorf("expected %v, got %v", ref, value)
		}
	}

	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	config = config.SetContainerEncoding(Stream)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	item := config.NewCbor(make([]byte, 0, 1024))
	config.NewJson([]byte(txt)).Tocbor(cbr)
	for _, tcase := range testcases {
		ptr := config.NewJsonpointer(tcase[0].(string))
		fn(ptr, tcase[1], cbr, item.Reset(nil))
	}

	config = NewDefaultConfig().SetNumberKind(SmartNumber)
	config = config.SetContainerEncoding(LengthPrefix)
	cbr = config.NewCbor(make([]byte, 0, 1024))
	item = config.NewCbor(make([]byte, 0, 1024))
	config.NewJson([]byte(txt)).Tocbor(cbr)
	for _, tcase := range testcases {
		ptr := config.NewJsonpointer(tcase[0].(string))
		fn(ptr, tcase[1], cbr, item.Reset(nil))
	}

	// special cases
	config = NewDefaultConfig().SetContainerEncoding(Stream)
	cbr = config.NewCbor(make([]byte, 0, 1024))
	item = config.NewCbor(make([]byte, 0, 1024))

	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[0] = cborType4 | cborIndefiniteLength
	cbr.n++
	cbr.EncodeSmallint(10).EncodeSmallint(-10).EncodeSimpletype(128)
	config.NewValue(uint8(100)).Tocbor(cbr)
	config.NewValue(float32(10.2)).Tocbor(cbr)
	config.NewValue(CborTagEpoch(1234567890)).Tocbor(cbr)
	cbr.EncodeBytechunks([][]byte{[]byte("hello"), []byte("world")})
	cbr.EncodeTextchunks([]string{"sound", "ok", "horn"})
	cbr.data = cbr.data[:cbr.n+1]
	cbr.data[cbr.n] = brkstp
	cbr.n++

	testcases = [][2]interface{}{
		{"/0", uint64(10)},
		{"/1", int64(-10)},
		{"/2", uint8(128)},
		{"/3", uint64(100)},
		{"/4", float32(10.2)},
		{"/5", CborTagEpoch(1234567890)},
		{"/6", [][]byte{[]byte("hello"), []byte("world")}},
		{"/7", []string{"sound", "ok", "horn"}},
	}

	t.Logf("%v", string(cbr.Bytes()))

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0].(string))

		ptr := config.NewJsonpointer(tcase[0].(string))
		value := cbr.Get(ptr, item.Reset(nil)).Tovalue()
		if !reflect.DeepEqual(value, tcase[1]) {
			t.Errorf("expected %T, got %T", tcase[1], value)
		}
	}
}

func TestCborEmpty(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	jsn := config.NewJson(nil)
	clt := config.NewCollate(nil)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.Tojson(jsn)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		cbr.Tocollate(clt)
	}()
}
