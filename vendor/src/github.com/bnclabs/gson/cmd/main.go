package main

import "flag"
import "strings"
import "strconv"
import "os"
import "log"
import "fmt"
import "reflect"
import "unsafe"
import "sort"
import "bytes"
import "path"
import "compress/gzip"
import "io/ioutil"
import "runtime/pprof"

import "github.com/bnclabs/gson"

var options struct {
	// generat options
	inpfile   string
	inptxt    string
	mprof     string
	outfile   string
	overheads bool
	quote     bool
	repeat    int
	nk        string
	ws        string

	// convert from JSON
	json2value   bool
	json2cbor    bool
	json2collate bool
	pointers     bool

	// convert from CBOR
	cbor2json    bool
	cbor2collate bool
	cbor2value   bool
	ct           string

	// convert from Collate
	collate2cbor      bool
	collate2json      bool
	collate2value     bool
	arrayLenPrefix    bool
	propertyLenPrefix bool
	doMissing         bool
	collatesort       string
	n1qlsort          string
	checkdir          string

	// convert from Value
	value2json    bool
	value2cbor    bool
	value2collate bool
}

var n1qltag = false

func argParse() []string {
	// general options
	flag.StringVar(&options.inpfile, "inpfile", "",
		"file containing one or more json docs based on the context")
	flag.StringVar(&options.inptxt, "inptxt", "",
		"use input text for the operation.")
	flag.StringVar(&options.mprof, "mprof", "",
		"take memory profile for testdata/code.json.gz")
	flag.StringVar(&options.outfile, "outfile", "",
		"write output to file")
	flag.BoolVar(&options.overheads, "overheads", false,
		"compute overheads on cbor and collation encoding")
	flag.BoolVar(&options.quote, "quote", false,
		"use strconv.Unquote on inptxt/inpfile")
	flag.IntVar(&options.repeat, "repeat", 0,
		"repeat count")
	flag.StringVar(&options.nk, "nk", "float",
		"interpret number as smart | float "+
			"smart - either treat number as int64 or fall back to float64 "+
			"float - treat number only as float64 ")
	flag.StringVar(&options.ws, "ws", "ansi",
		"interpret space as ansi (ansi whitespace) | unicode (unicode space).")

	// convert from JSON
	flag.BoolVar(&options.json2value, "json2value", false,
		"convert inptxt or content in inpfile to golang value")
	flag.BoolVar(&options.json2cbor, "json2cbor", false,
		"convert inptxt or content in inpfile to cbor output")
	flag.BoolVar(&options.json2collate, "json2collate", false,
		"convert inptxt or content in inpfile to collated output")
	flag.BoolVar(&options.pointers, "pointers", false,
		"list of json-pointers from input file")

	// convert from CBOR
	flag.BoolVar(&options.cbor2json, "cbor2json", false,
		"convert inptxt or content in inpfile to json output")
	flag.BoolVar(&options.cbor2collate, "cbor2collate", false,
		"convert inptxt or content in inpfile to collated output")
	flag.BoolVar(&options.cbor2value, "cbor2value", false,
		"convert inptxt or content in inpfile to golang value")
	flag.StringVar(&options.ct, "ct", "stream",
		"container encoding for cbor stream | lenprefix.")

	// convert from collate
	flag.BoolVar(&options.collate2cbor, "collate2cbor", false,
		"convert inptxt or content in inpfile to cbor output")
	flag.BoolVar(&options.collate2json, "collate2json", false,
		"convert inptxt or content in inpfile to json output")
	flag.BoolVar(&options.collate2value, "collate2value", false,
		"convert inptxt or content in inpfile to json output")
	flag.BoolVar(&options.arrayLenPrefix, "arrlenprefix", false,
		"set SortbyArrayLen for collation ordering")
	flag.BoolVar(&options.propertyLenPrefix, "maplenprefix", true,
		"set SortbyPropertyLen for collation ordering")
	flag.BoolVar(&options.doMissing, "domissing", true,
		"use missing type for collation")
	flag.StringVar(&options.collatesort, "collatesort", "",
		"sort inpfile, with one or more JSON terms, using collation algorithm")
	flag.StringVar(&options.n1qlsort, "n1qlsort", "",
		"sort inpfile, with one or more JSON terms, using n1ql algorithm")
	flag.StringVar(&options.checkdir, "checkdir", "",
		"check test files for collation order")

	// convert from value
	flag.BoolVar(&options.value2json, "value2json", false,
		"convert inptxt json to value and then back to json")
	flag.BoolVar(&options.value2cbor, "value2cbor", false,
		"convert inptxt json to value and then to cbor")
	flag.BoolVar(&options.value2collate, "value2collate", false,
		"convert inptxt json to value and then to collate")

	flag.Parse()

	return flag.Args()
}

func main() {
	argParse()

	if options.mprof != "" {
		if options.inpfile == "" {
			options.inpfile = "testdata/code.json.gz"
		}
		if options.repeat == 0 {
			options.repeat = 10
		}
	}

	if options.pointers {
		listpointers(readinput())

	} else if options.checkdir != "" {
		checkdir(options.checkdir)

	} else if options.collatesort != "" {
		fmt.Println(strings.Join(collatefile(options.collatesort), "\n"))

	} else if options.n1qlsort != "" {
		fmt.Println(strings.Join(sortn1ql(options.n1qlsort), "\n"))

	} else if options.value2json {
		value2json(readinput())

	} else if options.json2value {
		json2value(readinput())

	} else if options.json2cbor {
		json2cbor(readinput())

	} else if options.cbor2json {
		cbor2json(readinput())

	} else if options.cbor2collate {
		cbor2collate(readinput())

	} else if options.collate2cbor {
		collate2cbor(readinput())

	} else if options.collate2value {
		collate2value(readinput())

	} else if options.value2cbor {
		value2cbor(readinput())

	} else if options.value2collate {
		value2collate(readinput())

	} else if options.cbor2value {
		cbor2value(readinput())

	} else if options.json2collate {
		json2collate(readinput())

	} else if options.collate2json {
		collate2json(readinput())

	} else if options.overheads {
		computeOverheads()
	}

	if options.mprof != "" {
		domprof()
	}
}

func listpointers(inp []byte) {
	config := gson.NewDefaultConfig()
	jsn := config.NewJson([]byte(inp))
	_, value := jsn.Tovalue()
	val := config.NewValue(value)

	pointers := val.ListPointers([]string{})
	sort.Strings(pointers)
	for _, pointer := range pointers {
		fmt.Println(pointer)
	}
}

func domprof() {
	fmsg := "used %q as input (repeat:%v) ...\n"
	log.Printf(fmsg, options.inpfile, options.repeat)
	log.Printf("dumping profile data to %q ...", options.mprof)
	fd, err := os.Create(options.mprof)
	mf(err)
	pprof.WriteHeapProfile(fd)
	fd.Close()
}

func checkdir(dirname string) {
	entries, err := ioutil.ReadDir(dirname)
	mf(err)
	for _, entry := range entries {
		file := path.Join(dirname, entry.Name())
		if !strings.HasSuffix(file, ".ref") {
			log.Println("Checking", file, "...")
			out := strings.Join(collatefile(file), "\n")
			ref, err := ioutil.ReadFile(file + ".ref")
			mf(err)
			if strings.Trim(string(ref), "\n") != out {
				fmt.Println("expected\n", string(ref))
				fmt.Println("got\n", out)
				panic(fmt.Errorf("sort mismatch in %v", file))
			}
		}
	}
}

func collatefile(filename string) (outs []string) {
	s, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err.Error())
	}
	config := gson.NewDefaultConfig()
	config = config.SortbyArrayLen(options.arrayLenPrefix)
	config = config.SortbyPropertyLen(options.propertyLenPrefix)
	config = config.SetNumberKind(gson.SmartNumber)

	out := collateLines(config, s)
	for i := 0; i < options.repeat; i++ {
		collateLines(config, s)
	}
	return out
}

func collateLines(config *gson.Config, s []byte) []string {
	texts, codes := lines(s), make(codeList, 0)
	for i, text := range texts {
		jsn := config.NewJson(text)
		clt := config.NewCollate(make([]byte, 1024))
		jsn.Tocollate(clt.Reset(nil))
		codes = append(codes, codeObj{i, clt.Bytes()})
	}
	outs := doSort(texts, codes)
	return outs
}

func doSort(texts [][]byte, codes codeList) (outs []string) {
	sort.Sort(codes)
	for _, code := range codes {
		outs = append(outs, string(texts[code.off]))
	}
	return
}

func lines(content []byte) [][]byte {
	content = bytes.Trim(content, "\r\n")
	return bytes.Split(content, []byte("\n"))
}

func value2json(inp []byte) { // catch: input comes as json-str
	// json->value->json
	config := makeConfig()
	jsn := config.NewJson(inp)
	jsnback := config.NewJson(make([]byte, 0, (len(inp)*2 + 128)))

	_, value := jsn.Tovalue()
	val := config.NewValue(value)

	if options.mprof == "" {
		val.Tojson(jsnback.Reset(nil))
		fmt.Printf("Valu: %v\n", value)
		fmt.Printf("Json: %v\n", bytes2str(inp))
	} else {
		repeat(func() { val.Tojson(jsnback.Reset(nil)) }, options.repeat)
	}
}

func json2value(inp []byte) {
	inp = bytes.TrimRight(inp, "\n")
	config := makeConfig()
	jsn := config.NewJson(inp)

	if options.mprof == "" {
		_, value := jsn.Tovalue()
		fmt.Printf("Json: %v\n", bytes2str(inp))
		fmt.Printf("Valu: %v\n", value)
	} else {
		repeat(func() { jsn.Tovalue() }, options.repeat)
	}
}

func json2cbor(inp []byte) { // catch: input comes as json-str
	inp = bytes.TrimRight(inp, "\n")

	config := makeConfig()
	jsn := config.NewJson(inp)
	cbr := config.NewCbor(make([]byte, 0, (len(inp)*2)+128))

	if options.mprof == "" {
		jsn.Tocbor(cbr.Reset(nil))
		if options.outfile == "" {
			fmt.Printf("Json: %v\n", bytes2str(inp))
			fmt.Printf("Cbor: %v\n", cbr.Bytes())
			fmt.Printf("Cbor: %q\n", bytes2str(cbr.Bytes()))
			cbr.Tojson(jsn.Reset(nil))
			fmt.Printf("Json: %v\n", bytes2str(jsn.Bytes()))

		} else {
			err := ioutil.WriteFile(options.outfile, cbr.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}

	} else {
		repeat(func() { jsn.Tocbor(cbr.Reset(nil)) }, options.repeat)
	}
}

func cbor2json(inp []byte) { // catch: input comes as cbor-str
	config := makeConfig()
	cbr := config.NewCbor(inp)
	jsn := config.NewJson(make([]byte, 0, (len(inp)*4 + 128)))

	if options.mprof == "" {
		cbr.Tojson(jsn.Reset(nil))
		if options.outfile == "" {
			fmt.Printf("Cbor: %q\n", bytes2str(inp))
			fmt.Printf("Json: %v\n", bytes2str(jsn.Bytes()))
		} else {
			err := ioutil.WriteFile(options.outfile, jsn.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}
	} else {
		repeat(func() { cbr.Tojson(jsn.Reset(nil)) }, options.repeat)
	}
}

func cbor2collate(inp []byte) { // catch: input comes as cbor-str
	strlen, numkeys, itemlen, ptrlen := 1024, 1024, len(inp)*2, 1024

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	cbr := config.NewCbor(inp)
	clt := config.NewCollate(make([]byte, len(inp)*2))

	if options.mprof == "" {
		cbr.Tocollate(clt.Reset(nil))
		if options.outfile == "" {
			fmt.Printf("Cbor: %q\n", bytes2str(inp))
			fmt.Printf("Coll: %q\n", bytes2str(clt.Bytes()))
			fmt.Printf("Coll: %v\n", clt.Bytes())
		} else {
			err := ioutil.WriteFile(options.outfile, clt.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}
	} else {
		repeat(func() { cbr.Tocollate(clt.Reset(nil)) }, options.repeat)
	}
}

func collate2cbor(inp []byte) { // catch: input comes as collate-str
	strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, len(inp)*2, 1024

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	clt := config.NewCollate(inp)
	cbr := config.NewCbor(make([]byte, 0, (len(inp)*2 + 128)))

	if options.mprof == "" {
		clt.Tocbor(cbr.Reset(nil))
		if options.outfile == "" {
			fmt.Printf("Coll: %q\n", bytes2str(inp))
			fmt.Printf("Cbor: %q\n", bytes2str(cbr.Bytes()))
			fmt.Printf("Cbor: %v\n", cbr.Bytes())
		} else {
			err := ioutil.WriteFile(options.outfile, cbr.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}
	} else {
		repeat(func() { clt.Tocbor(cbr.Reset(nil)) }, options.repeat)
	}
}

func value2collate(inp []byte) { // catch: input comes as json-str
	// json->value->collate
	strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, len(inp)*2, 1024

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	_, value := config.NewJson(inp).Tovalue()
	val := config.NewValue(value)
	clt := config.NewCollate(make([]byte, (len(inp)*2 + 128)))

	if options.mprof == "" {
		val.Tocollate(clt.Reset(nil))
		fmt.Printf("Valu: %v\n", value)
		fmt.Printf("Coll: %q\n", bytes2str(clt.Bytes()))
		fmt.Printf("Coll: %v\n", clt.Bytes())
	} else {
		repeat(func() { val.Tocollate(clt.Reset(nil)) }, options.repeat)
	}
}

func collate2value(inp []byte) { // catch: input comes as collate-str
	strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, len(inp)*2, 1024

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	clt := config.NewCollate(inp)

	if options.mprof == "" {
		value := clt.Tovalue()
		fmt.Printf("Coll: %q\n", bytes2str(inp))
		fmt.Printf("Valu: %v\n", value)
	} else {
		repeat(func() { clt.Tovalue() }, options.repeat)
	}
}

func value2cbor(inp []byte) { // catch: input comes as json-str
	// json->value->cbor

	config := makeConfig()
	jsn := config.NewJson(inp)
	cbr := config.NewCbor(make([]byte, 0, (len(inp)*2 + 128)))

	_, value := jsn.Tovalue()
	val := config.NewValue(value)
	if options.mprof == "" {
		val.Tocbor(cbr.Reset(nil))
		fmt.Printf("Valu: %v\n", value)
		fmt.Printf("Cbor: %q\n", bytes2str(cbr.Bytes()))
		fmt.Printf("Cbor: %v\n", cbr.Bytes())
	} else {
		repeat(func() { val.Tocbor(cbr.Reset(nil)) }, options.repeat)
	}
}

func cbor2value(inp []byte) { // catch: input comes as cbor-str
	var value interface{}

	config := makeConfig()
	cbr := config.NewCbor(inp)

	if options.mprof == "" {
		value = cbr.Tovalue()
		fmt.Printf("Cbor: %q\n", bytes2str(inp))
		fmt.Printf("Cbor: %v\n", inp)
		fmt.Printf("Valu: %v\n", value)
	} else {
		repeat(func() { cbr.Tovalue() }, options.repeat)
	}
}

func json2collate(inp []byte) { // catch: input comes as json-str
	strlen, numkeys, itemlen, ptrlen := 1024, 1024, len(inp)*2, 1024

	inp = bytes.TrimRight(inp, "\n")

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	jsn := config.NewJson(inp)
	clt := config.NewCollate(make([]byte, (len(inp)*2)+128))

	fn := func() {
		jsn.Tocollate(clt.Reset(nil))
	}

	if options.mprof == "" {
		fn()
		if options.outfile == "" {
			fmt.Printf("Json: %v\n", bytes2str(inp))
			fmt.Printf("Coll: %q\n", bytes2str(clt.Bytes()))
			fmt.Printf("Coll: %v\n", clt.Bytes())
		} else {
			err := ioutil.WriteFile(options.outfile, clt.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}
	} else {
		repeat(fn, options.repeat)
	}
}

func collate2json(inp []byte) { // catch: input comes as collate-str
	strlen, numkeys, itemlen, ptrlen := 1024*1024, 1024, len(inp)*2, 1024

	config := makeConfig()
	config = config.ResetPools(strlen, numkeys, itemlen, ptrlen)
	clt := config.NewCollate(inp)
	jsn := config.NewJson(make([]byte, 0, (len(inp)*4 + 128)))

	if options.mprof == "" {
		clt.Tojson(jsn.Reset(nil))
		if options.outfile == "" {
			fmt.Printf("Coll: %q\n", bytes2str(inp))
			fmt.Printf("Json: %v\n", bytes2str(jsn.Bytes()))
		} else {
			err := ioutil.WriteFile(options.outfile, jsn.Bytes(), 0666)
			if err != nil {
				fmt.Printf("error writing to %s: %v\n", options.outfile, err)
			}
		}
	} else {
		repeat(func() { clt.Tojson(jsn.Reset(nil)) }, options.repeat)
	}
}

func computeOverheads() {
	items := []string{
		"10",
		"10000",
		"1000000000",
		"100000000000000000.0",
		"123456789123565670",
		"10.2",
		"10.23456789012",
		"null",
		"true",
		"false",
		`"hello world"`,
		`[]`,
		`{}`,
		`[10,10000,1000000000,10.2,10.23456789012,null,true,false,"hello world"]`,
		`{"a":10000,"b":10.23456789012,"c":null,"d":true,"e":false,"f":"hello world"}`,
	}
	config := makeConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 1024))
	for _, item := range items {
		fmt.Printf("item: %v\n", item)

		jsn := config.NewJson([]byte(item))
		jsn.Tocbor(cbr.Reset(nil))
		jsn.Tocollate(clt.Reset(nil))
		fmsg := "Json: %v bytes, Cbor: %v bytes, Collated: %v bytes\n"
		fmt.Printf(fmsg, len(item), len(cbr.Bytes()), len(clt.Bytes()))

		clt.Tojson(jsn.Reset(nil))
		clt.Tocbor(cbr.Reset(nil))
		fmsg = "Json: %v bytes, Cbor: %v bytes, Collated: %v bytes\n"
		fmt.Printf(fmsg, len(jsn.Bytes()), len(cbr.Bytes()), len(clt.Bytes()))
	}
}

func readfile(filename string) []byte {
	f, err := os.Open(filename)
	mf(err)
	defer f.Close()

	var data []byte

	if strings.HasSuffix(filename, ".gz") {
		gz, err := gzip.NewReader(f)
		mf(err)
		data, err = ioutil.ReadAll(gz)
		mf(err)
	} else {
		data, err = ioutil.ReadAll(f)
		mf(err)
	}
	return data
}

func readinput() []byte {
	var input string
	if options.inptxt != "" {
		input = options.inptxt
	} else if options.inpfile != "" {
		input = bytes2str(readfile(options.inpfile))
	} else {
		log.Fatalf("provide -inptxt or -inpfile")
	}
	if options.quote {
		var err error
		input, err = strconv.Unquote(input)
		mf(err)
	}
	return str2bytes(input)
}

func makeConfig() *gson.Config {
	config := gson.NewDefaultConfig()
	switch options.nk {
	case "smart":
		config = config.SetNumberKind(gson.SmartNumber)
	case "float":
		config = config.SetNumberKind(gson.FloatNumber)
	default:
		log.Fatalf("unknown number kind %v\n", options.nk)
	}

	switch options.ws {
	case "ansi":
		config = config.SetSpaceKind(gson.AnsiSpace)
	case "unicode":
		config = config.SetSpaceKind(gson.UnicodeSpace)
	}

	switch options.ct {
	case "lenprefix":
		config = config.SetContainerEncoding(gson.LengthPrefix)
	case "stream":
		config = config.SetContainerEncoding(gson.Stream)
	}

	config.SortbyArrayLen(options.arrayLenPrefix)
	config.SortbyPropertyLen(options.propertyLenPrefix)
	config.UseMissing(options.doMissing)
	return config
}

func str2bytes(str string) []byte {
	if str == "" {
		return nil
	}
	st := (*reflect.StringHeader)(unsafe.Pointer(&str))
	sl := &reflect.SliceHeader{Data: st.Data, Len: st.Len, Cap: st.Len}
	return *(*[]byte)(unsafe.Pointer(sl))
}

func bytes2str(bytes []byte) string {
	if bytes == nil {
		return ""
	}
	sl := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	st := &reflect.StringHeader{Data: sl.Data, Len: sl.Len}
	return *(*string)(unsafe.Pointer(st))
}

func repeat(fn func(), repeat int) {
	for i := 0; i < repeat; i++ {
		fn()
	}
}

func mf(err error) {
	if err != nil {
		panic(err)
	}
}

// collated objects

type codeObj struct {
	off  int
	code []byte
}

type codeList []codeObj

func (codes codeList) Len() int {
	return len(codes)
}

func (codes codeList) Less(i, j int) bool {
	return bytes.Compare(codes[i].code, codes[j].code) < 0
}

func (codes codeList) Swap(i, j int) {
	codes[i], codes[j] = codes[j], codes[i]
}
