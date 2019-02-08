package gson

import "sort"
import "bytes"
import "strings"
import "testing"
import "reflect"

import "golang.org/x/text/collate"
import "golang.org/x/text/language"

func TestCollateReset(t *testing.T) {
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	cltr := config.NewCollate(make([]byte, 0, 1024))

	ref := []interface{}{"sound", "ok", "horn"}
	config.NewValue(ref).Tocollate(clt)
	cltr.Reset(clt.Bytes())
	if value := cltr.Tovalue(); !reflect.DeepEqual(value, ref) {
		t.Errorf("expected %v, got %v", ref, value)
	}
}

func TestCollateEmpty(t *testing.T) {
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
		clt.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tojson(jsn)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		clt.Tocbor(cbr)
	}()
}

func TestAlternateSortTypes(t *testing.T) {
	testCases := []struct {
		lang string
		in   testtxtclts
		want []string
	}{{
		lang: "zh,cmn,zh-Hant-u-co-pinyin,zh-HK-u-co-pinyin,zh-pinyin",
		in: testtxtclts{
			&testtxtclt{in: "爸爸"}, &testtxtclt{in: "妈妈"},
			&testtxtclt{in: "儿子"}, &testtxtclt{in: "女儿"},
		},
		want: []string{"爸爸", "儿子", "妈妈", "女儿"},
	}, {
		lang: "zh-Hant,zh-u-co-stroke,zh-Hant-u-co-stroke",
		in: testtxtclts{
			&testtxtclt{in: "爸爸"}, &testtxtclt{in: "妈妈"},
			&testtxtclt{in: "儿子"}, &testtxtclt{in: "女儿"},
		},
		want: []string{"儿子", "女儿", "妈妈", "爸爸"},
	}}

	for _, tc := range testCases {
		for _, tag := range strings.Split(tc.lang, ",") {
			collator := collate.New(language.MustParse(tag))
			config := NewDefaultConfig().SetTextCollator(collator)
			for _, item := range tc.in {
				item.collate(config)
			}
			sort.Sort(tc.in)
			got := []string{}
			for _, item := range tc.in {
				got = append(got, item.in)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%v %v expected %v; got %v", tag, tc.in, tc.want, got)
			}
		}
	}
}

func TestTextNocase(t *testing.T) {
	testCases := []struct {
		lang string
		in   testtxtclts
		want []string
	}{{
		lang: "en",
		in: testtxtclts{
			&testtxtclt{in: "B"}, &testtxtclt{in: "b"},
			&testtxtclt{in: "a"}, &testtxtclt{in: "A"},
		},
		want: []string{"a", "A", "B", "b"},
	}}

	for _, tc := range testCases {
		for _, tag := range strings.Split(tc.lang, ",") {
			collator := collate.New(language.MustParse(tag), collate.IgnoreCase)
			config := NewDefaultConfig().SetTextCollator(collator)
			for _, item := range tc.in {
				item.collate(config)
			}
			sort.Sort(tc.in)
			got := []string{}
			for _, item := range tc.in {
				got = append(got, item.in)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%v %v expected %v; got %v", tag, tc.in, tc.want, got)
			}
		}
	}
}

func TestTextGermanSwedish(t *testing.T) {
	testCases := []struct {
		lang string
		in   testtxtclts
		want []string
	}{{
		lang: "de",
		in: testtxtclts{
			&testtxtclt{in: "a"}, &testtxtclt{in: "z"}, &testtxtclt{in: "ä"},
		},
		want: []string{"a", "ä", "z"},
	}, {
		lang: "sv",
		in: testtxtclts{
			&testtxtclt{in: "a"}, &testtxtclt{in: "z"}, &testtxtclt{in: "ä"},
		},
		want: []string{"a", "z", "ä"},
	}}

	for _, tc := range testCases {
		for _, tag := range strings.Split(tc.lang, ",") {
			collator := collate.New(language.MustParse(tag))
			config := NewDefaultConfig().SetTextCollator(collator)
			for _, item := range tc.in {
				item.collate(config)
			}
			sort.Sort(tc.in)
			got := []string{}
			for _, item := range tc.in {
				got = append(got, item.in)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%v %v expected %v; got %v", tag, tc.in, tc.want, got)
			}
		}
	}
}

// sort type for slice of []byte

type ByteSlices [][]byte

func (bs ByteSlices) Len() int {
	return len(bs)
}

func (bs ByteSlices) Less(i, j int) bool {
	return bytes.Compare(bs[i], bs[j]) < 0
}

func (bs ByteSlices) Swap(i, j int) {
	bs[i], bs[j] = bs[j], bs[i]
}

type testtxtclt struct {
	in  string
	clt []byte
}

func (item *testtxtclt) collate(config *Config) {
	val := config.NewValue(item.in)
	item.clt = val.Tocollate(config.NewCollate(nil)).Bytes()
}

type testtxtclts []*testtxtclt

func (items testtxtclts) Len() int {
	return len(items)
}

func (items testtxtclts) Less(i, j int) bool {
	return bytes.Compare(items[i].clt, items[j].clt) < 0
}

func (items testtxtclts) Swap(i, j int) {
	items[i], items[j] = items[j], items[i]
}
