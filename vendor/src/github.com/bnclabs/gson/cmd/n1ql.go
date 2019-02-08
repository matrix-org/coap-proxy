// +build n1ql

package main

import "sort"
import "strings"
import "io/ioutil"

import qv "github.com/couchbase/query/value"

func init() {
	n1qltag = true
}

func sortn1ql(filename string) []string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	s := string(data)
	items := strings.Split(s, "\n")
	for i := 0; i < options.repeat; i++ {
		list := &jsonList{vals: items, compares: 0}
		sort.Sort(list)
	}
	list := &jsonList{vals: items, compares: 0}
	sort.Sort(list)
	return list.vals
}

// sort type for n1ql

type jsonList struct {
	compares int
	vals     []string
}

func (jsons *jsonList) Len() int {
	return len(jsons.vals)
}

func (jsons *jsonList) Less(i, j int) bool {
	key1, key2 := jsons.vals[i], jsons.vals[j]
	jsons.compares++
	value1 := qv.NewValue([]byte(key1))
	value2 := qv.NewValue([]byte(key2))
	return value1.Collate(value2) < 0
}

func (jsons *jsonList) Swap(i, j int) {
	jsons.vals[i], jsons.vals[j] = jsons.vals[j], jsons.vals[i]
}
