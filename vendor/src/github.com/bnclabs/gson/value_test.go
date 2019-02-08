package gson

import "testing"
import "bytes"
import "sort"
import "io/ioutil"
import "strings"
import "path"

func TestValueCompare(t *testing.T) {
	config := NewDefaultConfig()
	testcases := [][3]interface{}{
		// numbers
		{uint64(10), float64(10), 0},
		{uint64(10), float64(10.1), -1},
		{uint64(10), int64(-10), 1},
		{uint64(10), int64(10), 0},
		{uint64(11), int64(10), 1},
		{uint64(9), uint64(10), -1},
		{uint64(10), uint64(10), 0},
		{int64(10), int(11), -1},
		{int64(11), int(10), 1},
		{int64(10), int64(10), 0},
		{int64(10), uint64(11), -1},
		{int64(10), uint64(10), 0},
		{int64(10), uint64(9), 1},
		{float64(10), uint64(10), 0},
		{float64(10.1), uint64(10), 1},
		{float64(10), uint64(9), 1},
		// all others
		{nil, nil, 0},
		{nil, false, -1},
		{nil, true, -1},
		{nil, 10, -1},
		{nil, "hello", -1},
		{nil, []interface{}{10, 20}, -1},
		{nil, map[string]interface{}{"key1": 10}, -1},
		{false, nil, 1},
		{false, false, 0},
		{false, true, -1},
		{false, 10, -1},
		{false, "hello", -1},
		{false, []interface{}{10, 20}, -1},
		{false, map[string]interface{}{"key1": 10}, -1},
		{true, nil, 1},
		{true, false, 1},
		{true, true, 0},
		{true, 10, -1},
		{true, "hello", -1},
		{true, []interface{}{10, 20}, -1},
		{true, map[string]interface{}{"key1": 10}, -1},
		{10, nil, 1},
		{10, false, 1},
		{10, true, 1},
		{10, 10, 0},
		{10, "hello", -1},
		{10, []interface{}{10, 20}, -1},
		{10, map[string]interface{}{"key1": 10}, -1},
		{[]interface{}{10}, nil, 1},
		{[]interface{}{10}, false, 1},
		{[]interface{}{10}, true, 1},
		{[]interface{}{10}, 10, 1},
		{[]interface{}{10}, "hello", 1},
		{[]interface{}{10}, []interface{}{10}, 0},
		{[]interface{}{10}, map[string]interface{}{"key1": 10}, -1},
		{map[string]interface{}{"key1": 10}, nil, 1},
		{map[string]interface{}{"key1": 10}, false, 1},
		{map[string]interface{}{"key1": 10}, true, 1},
		{map[string]interface{}{"key1": 10}, 10, 1},
		{map[string]interface{}{"key1": 10}, "hello", 1},
		{map[string]interface{}{"key1": 10}, []interface{}{10}, 1},
		{map[string]interface{}{"key1": 10},
			map[string]interface{}{"key1": 10}, 0},
	}
	for _, tcase := range testcases {
		val1 := config.NewValue(tcase[0])
		val2 := config.NewValue(tcase[1])
		ref, cmp := tcase[2].(int), val1.Compare(val2)
		if cmp != ref {
			t.Errorf("for nil expected %v, got %v", ref, cmp)
		}
	}
}

func TestValueCollate(t *testing.T) {
	dirname := "testdata/collate"
	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		file := path.Join(dirname, entry.Name())
		if !strings.HasSuffix(file, ".ref") {
			out := strings.Join(collatefile(file), "\n")
			ref, err := ioutil.ReadFile(file + ".ref")
			if err != nil {
				t.Fatal(err)
			}
			if strings.Trim(string(ref), "\n") != out {
				//fmt.Println(string(ref))
				//fmt.Println(string(out))
				t.Fatalf("sort mismatch in %v", file)
			}
		}
	}
}

func collatefile(filename string) (outs []string) {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	config := NewDefaultConfig()
	if strings.Contains(filename, "numbers") {
		config = config.SetNumberKind(SmartNumber)
	}
	return collateLines(config, s)
}

func collateLines(config *Config, s []byte) []string {
	texts, values := lines(s), make(valueList, 0)
	for i, text := range texts {
		jsn := config.NewJson(text)
		_, val := jsn.Tovalue()
		values = append(values, valObj{i, config.NewValue(val)})
	}
	outs := doSort(texts, values)
	return outs
}

func doSort(texts [][]byte, values valueList) (outs []string) {
	sort.Sort(values)
	outs = make([]string, 0)
	for _, value := range values {
		outs = append(outs, string(texts[value.off]))
	}
	return
}

func lines(content []byte) [][]byte {
	content = bytes.Trim(content, "\r\n")
	return bytes.Split(content, []byte("\n"))
}

type valObj struct {
	off int
	val *Value
}

type valueList []valObj

func (values valueList) Len() int {
	return len(values)
}

func (values valueList) Less(i, j int) bool {
	return values[i].val.Compare(values[j].val) < 0
}

func (values valueList) Swap(i, j int) {
	values[i], values[j] = values[j], values[i]
}
