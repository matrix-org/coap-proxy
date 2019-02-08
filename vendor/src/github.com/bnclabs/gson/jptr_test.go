package gson

import "testing"
import "strings"
import "sort"

func TestParsePointer(t *testing.T) {
	var tcasesJSONPointers = [][3]interface{}{
		{``, []string{}, ``},
		{`/`, []string{""}, `/`},
		{"/foo", []string{"foo"}, "/foo"},
		{"/foo/0", []string{"foo", "0"}, "/foo/0"},
		{"/a~1b", []string{"a/b"}, "/a~1b"},
		{"/c%d", []string{"c%d"}, "/c%d"},
		{"/e^f", []string{"e^f"}, "/e^f"},
		{"/g|h", []string{"g|h"}, "/g|h"},
		{`/i\j`, []string{`i\j`}, `/i\\j`},
		{`/k\"l`, []string{`k\"l`}, `/k\\\"l`},
		{"/ ", []string{" "}, "/ "},
		{"/m~0n", []string{"m~n"}, "/m~0n"},
		{"/g~1n~1r", []string{"g/n/r"}, "/g~1n~1r"},
		{"/g/汉语/r", []string{"g", "汉语", "r"}, "/g/汉语/r"},
		{"/dict/", []string{"dict", ""}, "/dict/"},
	}

	// test ParseJsonPointer
	config := NewDefaultConfig()
	jptr := config.NewJsonpointer("")
	for _, tcase := range tcasesJSONPointers {
		t.Logf("input pointer %q", tcase[0].(string))
		ref := tcase[1].([]string)

		jptr.ResetPath(tcase[0].(string))
		segments := jptr.Segments()
		if len(segments) != len(ref) {
			t.Errorf("expected %v, got %v", len(ref), len(segments))
		} else {
			for i, x := range ref {
				if string(segments[i]) != x {
					t.Errorf("expected %v, got %v", x, string(segments[i]))
				}
			}
		}

		// test encode pointers
		jptr.ResetPath("").ResetSegments(ref)
		if path := string(jptr.Path()); path != tcase[2].(string) {
			t.Errorf("expected %v, got %v", tcase[2].(string), path)
		}
	}
}

func TestTypicalPointers(t *testing.T) {
	refs := strings.Split(string(testdataFile("testdata/typical_pointers")), "\n")
	refs = refs[:len(refs)-1] // skip the last empty line
	sort.Strings(refs)
	config := NewDefaultConfig()

	data := testdataFile("testdata/typical.json")
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)

	// test list pointers
	pointers := val.ListPointers(make([]string, 0, 1024))
	sort.Strings(pointers)
	if len(refs) != len(pointers) {
		t.Errorf("expected %v, got %v", len(refs), len(pointers))
	}
	for i, r := range refs {
		if r != pointers[i] {
			t.Errorf("expected %v, got %v", r, pointers[i])
		}
	}

	// test list pointers for document using [][2]interface{} for map.
	value = GolangMap2cborMap(value)
	val = config.NewValue(value)

	pointers = val.ListPointers(make([]string, 0, 1024))
	sort.Strings(pointers)

	if len(refs) != len(pointers) {
		t.Errorf("expected %v, got %v", len(refs), len(pointers))
	}
	for i, r := range refs {
		if r != pointers[i] {
			t.Errorf("expected %v, got %v", r, pointers[i])
		}
	}
}

func BenchmarkParseJsonPtr3(b *testing.B) {
	path := "/foo/g/0"
	jptr := NewDefaultConfig().NewJsonpointer(path)

	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		jptr.ResetPath(path).Segments()
	}
}

func BenchmarkParseJsonPtr4(b *testing.B) {
	path := "/foo/g~1n~1r/0/hello"
	jptr := NewDefaultConfig().NewJsonpointer(path)

	b.SetBytes(int64(len(path)))
	for i := 0; i < b.N; i++ {
		jptr.ResetPath(path).Segments()
	}
}

func BenchmarkParseJsonPtr5(b *testing.B) {
	segments := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
	jptr := NewDefaultConfig().NewJsonpointer("").ResetSegments(segments)
	path := string(jptr.Path())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jptr.ResetPath(path).Segments()
	}
	b.SetBytes(int64(len(jptr.Path())))
}

func BenchmarkToJsonPtr5(b *testing.B) {
	segments := []string{"a", "ab", "a~b", "a/b", "a~/~/b"}
	jptr := NewDefaultConfig().NewJsonpointer("")

	b.SetBytes(15)
	for i := 0; i < b.N; i++ {
		jptr.ResetSegments(segments)
	}
}

func BenchmarkListPtrsTyp(b *testing.B) {
	data := testdataFile("testdata/typical.json")
	config := NewDefaultConfig()
	_, value := config.NewJson(data).Tovalue()
	val := config.NewValue(value)

	pointers := []string{}

	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pointers = val.ListPointers(pointers[:0])
	}
}
