package gson

import "testing"
import "encoding/json"
import "reflect"

func TestCborGet(t *testing.T) {
	var ref map[string]interface{}

	txt := `{"a": 10, "arr": [1,2], "nestd":[[23]], ` +
		`"dict": {"a":10, "b":20, "": 2}, "": 1}`
	json.Unmarshal([]byte(txt), &ref)

	testcases := [][2]interface{}{
		{"/a", 10.0},
		{"/arr/0", 1.0},
		{"/arr/1", 2.0},
		{"/nestd", []interface{}{[]interface{}{23.0}}},
		{"/dict/a", 10.0},
		{"/dict/b", 20.0},
		{"/dict/b", 20.0},
		{"/", 1.0},
		{
			"/dict",
			map[string]interface{}{"a": 10.0, "b": 20.0, "": 2.0}},
		{"/dict/", 2.0},
		{"", ref},
	}
	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		item := config.NewCbor(make([]byte, 0, 1024))

		config.NewJson([]byte(txt)).Tocbor(cbr)
		ptr := config.NewJsonpointer("")

		t.Logf("%v", txt)
		t.Logf("%v", cbr.Bytes())

		for _, tcase := range testcases {
			t.Logf("%v", tcase[0].(string))

			ptr = ptr.ResetPath(tcase[0].(string))
			value := cbr.Get(ptr, item.Reset(nil)).Tovalue()
			if !reflect.DeepEqual(value, tcase[1]) {
				fmsg := "for %q expected %v, got %v"
				t.Errorf(fmsg, string(ptr.Path()), tcase[1], value)
			}
		}
	}
	dotest(NewDefaultConfig().SetContainerEncoding(LengthPrefix))
	dotest(NewDefaultConfig().SetContainerEncoding(Stream))
}

func TestCborSet(t *testing.T) {
	var refobj map[string]interface{}
	txt := `{"a": 10, "arr": [1,2], "-": [[1]], "nestd": [[23]], ` +
		`"dict": {"a":10, "b":20}}`
	ref := `{"b":1,"a":11,"arr":[10,30],"-":[[30]], "nestd":[[23]],` +
		`"dict":{"a":1,"b":2}}`
	json.Unmarshal([]byte(txt), &refobj)

	testcases := [][3]interface{}{
		{"/a", 11.0, 10.0},
		{"/b", 1.0, 1.0},
		{"/arr/0", 10.0, 1.0},
		{"/arr/1", 30.0, 2.0},
		{"/-/0/0", 30.0, 1.0},
		{"/dict/a", 1.0, 10.0},
		{"/dict/b", 2.0, 20.0},
	}

	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		item := config.NewCbor(make([]byte, 0, 1024))
		newcbr := config.NewCbor(make([]byte, 0, 1024))
		old := config.NewCbor(make([]byte, 0, 1024))

		config.NewJson([]byte(txt)).Tocbor(cbr)
		ptr := config.NewJsonpointer("")

		for _, tcase := range testcases {
			t.Logf("%v", tcase[0].(string))

			ptr.ResetPath(tcase[0].(string))
			config.NewValue(tcase[1]).Tocbor(item.Reset(nil))
			cbr.Set(ptr, item, newcbr.Reset(nil), old.Reset(nil))
			cbr, newcbr = newcbr, cbr

			if olditem := old.Tovalue(); !reflect.DeepEqual(olditem, tcase[2]) {
				fmsg := "for %q expected %v, got %v"
				t.Errorf(fmsg, string(ptr.Path()), tcase[2], olditem)
			}

			value := cbr.Get(ptr, old.Reset(nil)).Tovalue()
			if !reflect.DeepEqual(value, tcase[1]) {
				fmsg := "for %q expected %v, got %v"
				t.Errorf(fmsg, string(ptr.Path()), tcase[1], value)
			}
		}
		var refval interface{}
		if err := json.Unmarshal([]byte(ref), &refval); err != nil {
			t.Fatalf("unmarshal: %v", err)
		} else if value := cbr.Tovalue(); !reflect.DeepEqual(refval, value) {
			t.Errorf("expected %v, got %v", refval, value)
		}
	}
	dotest(NewDefaultConfig().SetContainerEncoding(LengthPrefix))
	dotest(NewDefaultConfig().SetContainerEncoding(Stream))
}

func TestCborPrepend(t *testing.T) {
	var ref interface{}

	txt := `{"a": 10, "-": [1], "arr": [1,2], "nestd": [[3]]}`
	reftxt := `{"a": 10, "-": [10,1], "arr": [10,1,2], "nestd": [[10, 3]]}`
	json.Unmarshal([]byte(reftxt), &ref)
	testcases := []string{"/-", "/arr", "/nestd/0"}

	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		newcbr := config.NewCbor(make([]byte, 0, 1024))
		item := config.NewCbor(make([]byte, 0, 1024))

		config.NewJson([]byte(txt)).Tocbor(cbr)
		config.NewValue(10.0).Tocbor(item)
		ptr := config.NewJsonpointer("")

		for _, tcase := range testcases {
			t.Logf("%v", tcase)
			ptr.ResetPath(tcase)
			cbr.Prepend(ptr, item, newcbr.Reset(nil))
			cbr, newcbr = newcbr, cbr
		}

		if value := cbr.Tovalue(); !reflect.DeepEqual(value, ref) {
			t.Errorf("expected %v, got %v", ref, value)
		}
	}
	dotest(NewDefaultConfig().SetContainerEncoding(Stream))
	dotest(NewDefaultConfig().SetContainerEncoding(LengthPrefix))

	// Prepend an array
	txt, reftxt = `[]`, `[20,10]`
	json.Unmarshal([]byte(reftxt), &ref)

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	newcbr := config.NewCbor(make([]byte, 0, 1024))
	item := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson([]byte(txt)).Tocbor(cbr)
	ptr := config.NewJsonpointer("")

	cbr.Prepend(ptr, config.NewValue(10.0).Tocbor(item.Reset(nil)), newcbr)
	newcbr.Prepend(ptr, config.NewValue(20.0).Tocbor(item.Reset(nil)), cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}

	// panic case
	config = NewDefaultConfig()
	cbr = config.NewCbor(make([]byte, 0, 1024))
	newcbr = config.NewCbor(make([]byte, 0, 1024))
	item = config.NewCbor(make([]byte, 0, 1024))

	config.NewJson([]byte(`{"a": 10}`)).Tocbor(cbr)
	fn := func(jptr *Jsonpointer, v interface{}) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.NewValue(v).Tocbor(item.Reset(nil))
		cbr.Prepend(ptr, item, newcbr.Reset(nil))
	}
	fn(config.NewJsonpointer("/a"), 10)
	fn(config.NewJsonpointer(""), 10)
}

func TestCborAppend(t *testing.T) {
	var ref interface{}

	txt := `{"a": 10, "-": [1], "arr": [1,2], "nestd": [[3]]}`
	reftxt := `{"a": 10, "-": [1,10], "arr": [1,2,10], "nestd": [[3,10]]}`
	json.Unmarshal([]byte(reftxt), &ref)
	testcases := []string{"/-", "/arr", "/nestd/0"}

	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		newcbr := config.NewCbor(make([]byte, 0, 1024))
		item := config.NewCbor(make([]byte, 0, 1024))

		config.NewJson([]byte(txt)).Tocbor(cbr)
		config.NewValue(10.0).Tocbor(item)
		ptr := config.NewJsonpointer("")

		for _, tcase := range testcases {
			t.Logf("%v", tcase)
			ptr.ResetPath(tcase)
			cbr.Append(ptr, item, newcbr.Reset(nil))
			cbr, newcbr = newcbr, cbr
		}

		if value := cbr.Tovalue(); !reflect.DeepEqual(value, ref) {
			t.Errorf("expected %v, got %v", ref, value)
		}
	}
	dotest(NewDefaultConfig().SetContainerEncoding(Stream))
	dotest(NewDefaultConfig().SetContainerEncoding(LengthPrefix))

	// Append an array
	txt, reftxt = `[]`, `[10,20]`
	json.Unmarshal([]byte(reftxt), &ref)

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	newcbr := config.NewCbor(make([]byte, 0, 1024))
	item := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson([]byte(txt)).Tocbor(cbr)
	ptr := config.NewJsonpointer("")

	cbr.Append(ptr, config.NewValue(10.0).Tocbor(item.Reset(nil)), newcbr)
	newcbr.Append(ptr, config.NewValue(20.0).Tocbor(item.Reset(nil)), cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}

	// panic case
	config = NewDefaultConfig()
	cbr = config.NewCbor(make([]byte, 0, 1024))
	newcbr = config.NewCbor(make([]byte, 0, 1024))
	item = config.NewCbor(make([]byte, 0, 1024))

	config.NewJson([]byte(`{"a": 10}`)).Tocbor(cbr)
	fn := func(jptr *Jsonpointer, v interface{}) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.NewValue(v).Tocbor(item.Reset(nil))
		cbr.Prepend(ptr, item, newcbr.Reset(nil))
	}
	fn(config.NewJsonpointer("/a"), 10)
	fn(config.NewJsonpointer(""), 10)
}

func TestCborDelete(t *testing.T) {
	txt := `{"a": 10, "-": [1], "arr": [1,2],  "nestd": [[23]], ` +
		`"dict": {"a":10, "b":20, "": 30}}`
	testcases := [][2]interface{}{
		{"/a", 10.0},
		{"/arr/1", 2.0},
		{"/arr/0", 1.0},
		{"/-/0", 1.0},
		{"/nestd/0/0", 23.0},
		{"/dict/a", 10.0},
		{"/dict/b", 20.0},
		{"/dict/", 30.0},
	}

	dotest := func(config *Config) {
		cbr := config.NewCbor(make([]byte, 0, 1024))
		newcbr := config.NewCbor(make([]byte, 0, 1024))
		deleted := config.NewCbor(make([]byte, 0, 1024))

		config.NewJson([]byte(txt)).Tocbor(cbr)
		ptr := config.NewJsonpointer("")

		for _, tcase := range testcases {
			t.Logf("%v", tcase)

			ptr.ResetPath(tcase[0].(string))
			cbr.Delete(ptr, newcbr.Reset(nil), deleted.Reset(nil))
			if value := deleted.Tovalue(); !reflect.DeepEqual(value, tcase[1]) {
				t.Errorf("for %v expected %v, got %v", tcase[0], tcase[1], value)
			}
			cbr, newcbr = newcbr, cbr
		}
		remtxt := `{"arr": [], "-": [], "nestd": [[]], "dict":{}}"`
		_, remvalue := config.NewJson([]byte(remtxt)).Tovalue()
		if value := cbr.Tovalue(); !reflect.DeepEqual(remvalue, value) {
			t.Errorf("expected %v, got %v", remvalue, value)
		}
	}
	dotest(NewDefaultConfig().SetContainerEncoding(LengthPrefix))
	dotest(NewDefaultConfig().SetContainerEncoding(Stream))

	// corner case
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	newcbr := config.NewCbor(make([]byte, 0, 1024))
	deleted := config.NewCbor(make([]byte, 0, 1024))

	config.NewValue([]interface{}{"hello", "world"}).Tocbor(cbr)
	ptr := config.NewJsonpointer("/1")

	remvalue := cbr.Delete(ptr, newcbr, deleted).Tovalue()
	if value := deleted.Tovalue(); !reflect.DeepEqual(value, "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", deleted)
	} else if v := remvalue.([]interface{}); len(v) != 1 {
		t.Errorf("for %v expected length %v, got %v", ptr, 1, len(v))
	} else if !reflect.DeepEqual(v[0], "hello") {
		t.Errorf("for %v expected %v, got %v", ptr, "hello", v[0])
	}
}

func BenchmarkCborGet(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	item := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson(data).Tocbor(cbr)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Get(ptr, item.Reset(nil))
	}
}

func BenchmarkCborSet(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	newcbr := config.NewCbor(make([]byte, 0, 10*1024))
	item := config.NewCbor(make([]byte, 0, 1024))
	olditem := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson(data).Tocbor(cbr)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Set(ptr, item.Reset(nil), newcbr.Reset(nil), olditem.Reset(nil))
	}
}

func BenchmarkCborAppend(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	newcbr := config.NewCbor(make([]byte, 0, 10*1024))
	item := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson(data).Tocbor(cbr)
	config.NewValue("bench").Tocbor(item)
	ptra := config.NewJsonpointer("/projects/Sherri/members")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Append(ptra, item, newcbr.Reset(nil))
	}
}

func BenchmarkCborPrepend(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	newcbr := config.NewCbor(make([]byte, 0, 10*1024))
	item := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson(data).Tocbor(cbr)
	config.NewValue("bench").Tocbor(item)
	ptrp := config.NewJsonpointer("/projects/Sherri/members")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Prepend(ptrp, item, newcbr.Reset(nil))
	}
}

func BenchmarkCborDelete(b *testing.B) {
	data := testdataFile("testdata/typical.json")

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10*1024))
	newcbr := config.NewCbor(make([]byte, 0, 10*1024))
	deleted := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson(data).Tocbor(cbr)
	ptrd := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbr.Delete(ptrd, newcbr.Reset(nil), deleted.Reset(nil))
	}
}
