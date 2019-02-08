package gson

import "testing"
import "reflect"
import "encoding/json"

func TestValueGet(t *testing.T) {
	var ref map[string]interface{}

	txt := `{"a": 10, "arr": [1,2], "nestd":[[23]], ` +
		`"dict": {"a":10, "b":20, "": 2}, "": 1}`
	json.Unmarshal([]byte(txt), &ref)

	testcases := [][2]interface{}{
		{"/a", 10.0},
		{"/arr/0", 1.0},
		{"/arr/1", 2.0},
		{"/arr/-", 2.0},
		{"/nestd", []interface{}{[]interface{}{23.0}}},
		{"/nestd/-", []interface{}{23.0}},
		{"/nestd/-/0", 23.0},
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
	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt)).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	t.Logf("%v", txt)

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0].(string))

		ptr = ptr.ResetPath(tcase[0].(string))
		if item := val.Get(ptr); !reflect.DeepEqual(item, tcase[1]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[1], item)
		}
	}
}

func TestValueSet(t *testing.T) {
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
		{"/arr/1", 20.0, 2.0},
		{"/arr/-", 30.0, 20.0},
		{"/-/-/0", 30.0, 1.0},
		{"/dict/a", 1.0, 10.0},
		{"/dict/b", 2.0, 20.0},
	}
	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt)).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase[0].(string))

		ptr.ResetPath(tcase[0].(string))
		value, oitem := val.Set(ptr, tcase[1])
		val = config.NewValue(value)
		if !reflect.DeepEqual(oitem, tcase[2]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[2], oitem)
		}

		item := val.Get(ptr)
		if !reflect.DeepEqual(item, tcase[1]) {
			fmsg := "for %q expected %v, got %v"
			t.Errorf(fmsg, string(ptr.Path()), tcase[1], item)
		}
	}

	var refval interface{}
	if err := json.Unmarshal([]byte(ref), &refval); err != nil {
		t.Fatalf("unmarshal: %v", err)
	} else if !reflect.DeepEqual(refval, val.data) {
		t.Errorf("expected %v, got %v", refval, val.data)
	}

	// corner case
	config = NewDefaultConfig()
	val = config.NewValue([]interface{}{"hello"})
	ptr = config.NewJsonpointer("/-")
	nval, oitem := val.Set(ptr, "world")
	if !reflect.DeepEqual(oitem, "hello") {
		t.Errorf("for %v expected %v, got %v", ptr.Path(), "hello", oitem)
	} else if v := nval.([]interface{}); !reflect.DeepEqual(v[0], "world") {
		t.Errorf("for %v expected %v, got %v", ptr.Path(), "world", v[0])
	}
}

func TestValuePrepend(t *testing.T) {
	var ref interface{}

	txt := `{"a": 10, "-": [1], "arr": [1,2]}`
	reftxt := `{"a": 10, "-": [10,1], "arr": [10,1,2]}`
	json.Unmarshal([]byte(reftxt), &ref)
	testcases := []string{"/-", "/arr"}

	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt)).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		ptr.ResetPath(tcase)
		val = config.NewValue(val.Prepend(ptr.ResetPath(tcase), 10.0))
	}

	if !reflect.DeepEqual(val.data, ref) {
		t.Errorf("expected %v, got %v", ref, val.data)
	}

	// Prepend an array
	txt, reftxt = `[]`, `[20,10]`
	json.Unmarshal([]byte(reftxt), &ref)

	config = NewDefaultConfig()
	_, value = config.NewJson([]byte(txt)).Tovalue()
	val = config.NewValue(value)
	ptr = config.NewJsonpointer("")
	val = config.NewValue(val.Prepend(ptr, 10.0))
	val = config.NewValue(val.Prepend(ptr, 20.0))

	if !reflect.DeepEqual(val.data, ref) {
		t.Errorf("expected %v, got %v", ref, val.data)
	}

	// panic case
	config = NewDefaultConfig()
	_, value = config.NewJson([]byte(`{"a": 10}`)).Tovalue()
	val = config.NewValue(value)
	fn := func(jptr *Jsonpointer, v interface{}) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		val.Prepend(ptr, v)
	}
	fn(config.NewJsonpointer("/a"), 10)
	fn(config.NewJsonpointer(""), 10)
}

func TestValueAppend(t *testing.T) {
	var ref interface{}

	txt := `{"a": 10, "-": [1], "arr": [1,2]}`
	reftxt := `{"a": 10, "-": [1,10], "arr": [1,2,10]}`
	json.Unmarshal([]byte(reftxt), &ref)
	testcases := []string{"/-", "/arr"}

	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt)).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase)
		val = config.NewValue(val.Append(ptr.ResetPath(tcase), 10.0))
	}

	if !reflect.DeepEqual(val.data, ref) {
		t.Errorf("expected %v, got %v", ref, val.data)
	}

	// Append an array
	txt, reftxt = `[]`, `[10,20]`
	json.Unmarshal([]byte(reftxt), &ref)

	config = NewDefaultConfig()
	_, value = config.NewJson([]byte(txt)).Tovalue()
	val = config.NewValue(value)
	ptr = config.NewJsonpointer("")
	val = config.NewValue(val.Append(ptr, 10.0))
	val = config.NewValue(val.Append(ptr, 20.0))

	if !reflect.DeepEqual(val.data, ref) {
		t.Errorf("expected %v, got %v", ref, val.data)
	}

	// panic case
	config = NewDefaultConfig()
	_, value = config.NewJson([]byte(`{"a": 10}`)).Tovalue()
	val = config.NewValue(value)
	fn := func(jptr *Jsonpointer, v interface{}) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		val.Append(ptr, v)
	}
	fn(config.NewJsonpointer("/a"), 10)
	fn(config.NewJsonpointer(""), 10)
}

func TestValueDelete(t *testing.T) {
	txt := `{"a": 10, "-": [1], "arr": [1,2],  "nestd": [[23]], ` +
		`"dict": {"a":10, "b":20, "": 30}}`
	testcases := [][2]interface{}{
		{"/a", 10.0},
		{"/arr/1", 2.0},
		{"/arr/0", 1.0},
		{"/-/0", 1.0},
		{"/nestd/-/0", 23.0},
		{"/dict/a", 10.0},
		{"/dict/b", 20.0},
		{"/dict/", 30.0},
	}

	config := NewDefaultConfig()
	_, value := config.NewJson([]byte(txt)).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("")

	for _, tcase := range testcases {
		t.Logf("%v", tcase)

		ptr.ResetPath(tcase[0].(string))
		value, deleted := val.Delete(ptr)
		if !reflect.DeepEqual(deleted, tcase[1]) {
			t.Errorf("for %v expected %v, got %v", tcase[0], tcase[1], deleted)
		}
		val = config.NewValue(value)
	}

	remtxt := `{"arr": [], "-": [], "nestd": [[]], "dict":{}}"`
	_, remvalue := config.NewJson([]byte(remtxt)).Tovalue()
	if !reflect.DeepEqual(val.data, remvalue) {
		t.Errorf("expected %v, got %v", remvalue, val.data)
	}

	// corner case
	config = NewDefaultConfig()
	val = config.NewValue([]interface{}{"hello", "world"})
	ptr = config.NewJsonpointer("/1")

	value, deleted := val.Delete(ptr)
	if !reflect.DeepEqual(deleted, "world") {
		t.Errorf("for %v expected %v, got %v", ptr, "world", deleted)
	} else if v := value.([]interface{}); len(v) != 1 {
		t.Errorf("for %v expected length %v, got %v", ptr, 1, len(v))
	} else if !reflect.DeepEqual(v[0], "hello") {
		t.Errorf("for %v expected %v, got %v", ptr, "hello", v[0])
	}
}

func BenchmarkValueGet(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Get(ptr)
	}
}

func BenchmarkValueSet(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)
	ptr := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Set(ptr, 10)
	}
}

func BenchmarkValuePrepend(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)
	ptrp := config.NewJsonpointer("/projects/Sherri/members")
	ptrd := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Prepend(ptrp, "bench")
		val.Delete(ptrd)
	}
}

func BenchmarkValueAppend(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)
	ptra := config.NewJsonpointer("/projects/Sherri/members")
	ptrd := config.NewJsonpointer("/projects/Sherri/members/0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Append(ptra, "bench")
		val.Delete(ptrd)
	}
}

func BenchmarkValueDelete(b *testing.B) {
	config := NewDefaultConfig()
	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)

	ptrd := config.NewJsonpointer("/projects/Sherri/members/0")
	ptra := config.NewJsonpointer("/projects/Sherri/members")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		val.Delete(ptrd)
		val.Append(ptra, "delete")
	}
}
