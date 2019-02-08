package gson

import "testing"
import "bytes"
import "reflect"
import "encoding/json"
import "os"
import "strings"
import "compress/gzip"
import "io/ioutil"

func TestConfig(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	val := config.NewValue(10.2)
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
}

type testLocal byte

func TestUndefined(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	val := config.NewValue(CborUndefined(cborSimpleUndefined))
	val.Tocbor(cbr)
	if value := cbr.Tovalue(); !reflect.DeepEqual(val.data, value) {
		t.Errorf("expected %v got %v", val.data, value)
	}
	// test unknown type.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		config.NewValue(testLocal(10)).Tocbor(cbr.Reset(nil))
	}()
}

func TestJsonToValue(t *testing.T) {
	config := NewDefaultConfig().SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" `))
	jsnrmn, value := jsn.Tovalue()
	if string(jsnrmn.Bytes()) != `  "xyz" "10" ` {
		t.Errorf("expected %q, got %q", `  "xyz" "10" `, string(jsnrmn.Bytes()))
	}

	jsnback := config.NewJson(make([]byte, 0, 1024))
	config.NewValue(value).Tojson(jsnback)
	if ref := `"abcd"`; string(jsnback.Bytes()) != ref {
		t.Errorf("expected %v, got %v", ref, string(jsnback.Bytes()))
	}
}

func TestJsonToValues(t *testing.T) {
	var s string
	uniStr := `"汉语 / 漢語; Hàn\b \t\uef24yǔ "`
	json.Unmarshal([]byte(uniStr), &s)
	ref := []interface{}{"abcd", "xyz", "10", s}

	config := NewDefaultConfig().SetSpaceKind(AnsiSpace)
	jsn := config.NewJson([]byte(`"abcd"  "xyz" "10" ` + uniStr))
	if values := jsn.Tovalues(); !reflect.DeepEqual(values, ref) {
		t.Errorf("expected %v, got %v", ref, values)
	}
}

func TestParseJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	jptr := config.NewJsonpointer("/a/b")
	refsegs := [][]byte{[]byte("a"), []byte("b")}
	if segments := jptr.Segments(); !reflect.DeepEqual(segments, refsegs) {
		t.Errorf("expected %v, got %v", refsegs, segments)
	}
}

func TestToJsonPointer(t *testing.T) {
	config := NewDefaultConfig()
	refptr := config.NewJsonpointer("/a/b")
	jptr := config.NewJsonpointer("").ResetSegments([]string{"a", "b"})

	if bytes.Compare(jptr.path, refptr.path) != 0 {
		t.Errorf("expected %v, got %v", refptr.path, jptr.path)
	}
}

func TestGsonToCollate(t *testing.T) {
	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	clt := config.NewCollate(make([]byte, 0, 1024))
	config.NewValue(map[string]interface{}{"a": 10, "b": 20}).Tocollate(clt)
	ref := map[string]interface{}{"a": 10.0, "b": 20.0}
	if value := clt.Tovalue(); !reflect.DeepEqual(ref, value) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCborToCollate(t *testing.T) {
	config := NewDefaultConfig().SetNumberKind(SmartNumber)
	cbr := config.NewCbor(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 0, 1024))
	out := config.NewCbor(make([]byte, 0, 1024))

	o := [][2]interface{}{
		{"a", 10.0},
		{"b", 20.0},
	}
	refm := CborMap2golangMap(o)

	value := config.NewValue(o).Tocbor(cbr).Tocollate(clt).Tocbor(out).Tovalue()
	if !reflect.DeepEqual(refm, value) {
		t.Errorf("expected %v, got %v", refm, value)
	}
}

func TestResetPool(t *testing.T) {
	config := NewDefaultConfig().ResetPools(1, 2, 3, 4)
	if config.strlen != 1 {
		t.Errorf("expected %v, got %v", 1, config.strlen)
	} else if config.numkeys != 2 {
		t.Errorf("expected %v, got %v", 2, config.numkeys)
	} else if config.itemlen != 3 {
		t.Errorf("expected %v, got %v", 3, config.itemlen)
	} else if config.ptrlen != 4 {
		t.Errorf("expected %v, got %v", 4, config.ptrlen)
	}
}

func TestNewCollate(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 123))
	if clt.n != 123 {
		t.Errorf("expected %v, got %v", 123, clt.n)
	}
}

func TestConfigString(t *testing.T) {
	config := NewDefaultConfig()
	s := config.String()
	if !strings.Contains(s, "nk:") {
		t.Errorf("expected config string, %v", s)
	}
}

func TestNumberKindString(t *testing.T) {
	if s := SmartNumber.String(); s != "SmartNumber" {
		t.Errorf("expected %v, got %v", "SmartNumber", s)
	} else if s := FloatNumber.String(); s != "FloatNumber" {
		t.Errorf("expected %v, got %v", "FloatNumber", s)
	}
}

func TestContainerEncodingString(t *testing.T) {
	if s := LengthPrefix.String(); s != "LengthPrefix" {
		t.Errorf("expected %v, got %v", "LengthPrefix", s)
	} else if s := Stream.String(); s != "Stream" {
		t.Errorf("expected %v, got %v", "Stream", s)
	}
}

func TestCheckSortedkeys(t *testing.T) {
	val := map[string]interface{}{
		"ddd": true, "ccc": false, "fff": true, "aaa": nil,
		"bbb": []interface{}{1, 2},
	}
	srtjson := `{"aaa":null,"bbb":[1,2],"ccc":false,"ddd":true,"fff":true}`

	config := NewDefaultConfig()
	jsn := config.NewJson(make([]byte, 0, 1024))
	cbr := config.NewCbor(make([]byte, 0, 1024))
	col := config.NewCollate(make([]byte, 0, 1024))

	json := config.NewValue(val).Tojson(jsn.Reset(nil)).Bytes()
	if string(json) != srtjson {
		t.Errorf("expected %v, got %v", srtjson, string(json))
	}

	cbr = config.NewValue(val).Tocbor(cbr.Reset(nil))
	json = cbr.Tojson(jsn.Reset(nil)).Bytes()
	if string(json) != srtjson {
		t.Errorf("expected %v, got %v", srtjson, string(json))
	}

	col = config.NewValue(val).Tocollate(col.Reset(nil))
	json = col.Tocbor(cbr.Reset(nil)).Tojson(jsn.Reset(nil)).Bytes()
	if string(json) != srtjson {
		t.Errorf("expected %v, got %v", srtjson, string(json))
	}
}

func testdataFile(filename string) []byte {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var data []byte
	if strings.HasSuffix(filename, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			panic(err)
		}
		data, err = ioutil.ReadAll(gz)
		if err != nil {
			panic(err)
		}
	} else {
		data, err = ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
	}
	return data
}

var allValueIndent, allValueCompact, pallValueIndent, pallValueCompact []byte
var mapValue []byte
var scanvalid []string
var scaninvalid []string

func init() {
	var value interface{}
	var err error

	allValueIndent, err = ioutil.ReadFile("testdata/allValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(allValueIndent, &value); err != nil {
		panic(err)
	}
	if allValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	pallValueIndent, err = ioutil.ReadFile("testdata/pallValueIndent")
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(pallValueIndent, &value); err != nil {
		panic(err)
	}
	if pallValueCompact, err = json.Marshal(value); err != nil {
		panic(err)
	}

	mapValue, err = ioutil.ReadFile("testdata/map")
	if err != nil {
		panic(err)
	}

	scanvalidb, err := ioutil.ReadFile("testdata/scan_valid")
	if err != nil {
		panic(err)
	}
	scanvalid = []string{}
	for _, s := range strings.Split(string(scanvalidb), "\n") {
		if strings.Trim(s, " ") != "" {
			scanvalid = append(scanvalid, s)
		}
	}
	scanvalid = append(scanvalid, []string{
		"\"hello\xffworld\"",
		"\"hello\xc2\xc2world\"",
		"\"hello\xc2\xffworld\"",
		"\"hello\xed\xa0\x80\xed\xb0\x80world\""}...)

	scaninvalidb, err := ioutil.ReadFile("testdata/scan_invalid")
	if err != nil {
		panic(err)
	}
	scaninvalid = []string{}
	for _, s := range strings.Split(string(scaninvalidb), "\n") {
		if strings.Trim(s, " ") != "" {
			scaninvalid = append(scaninvalid, s)
		}
	}
	scaninvalid = append(scaninvalid, []string{
		"\xed\xa0\x80", // RuneError
		"\xed\xbf\xbf", // RuneError
		// raw value errors
		"\x01 42",
		"\x01 true",
		"\x01 1.2",
		"\x01 \"string\"",
		// bad-utf8
		"hello\xffworld",
		"\xff",
		"\xff\xff",
		"a\xffb",
		"\xe6\x97\xa5\xe6\x9c\xac\xff\xaa\x9e"}...)
}

func fixFloats(val interface{}) interface{} {
	switch v := val.(type) {
	case float64:
		return float32(v)
	case []interface{}:
		for i, x := range v {
			v[i] = fixFloats(x)
		}
		return v
	case map[string]interface{}:
		for p, q := range v {
			v[p] = fixFloats(q)
		}
		return v
	}
	return val
}
