package gson

import "fmt"

func ExampleConfig() {
	config := NewDefaultConfig()
	// override default configuration options.
	// IMPORTANT: config objects are immutable, so assign back
	// the new object returned by each of the settings method.
	config = config.SetContainerEncoding(LengthPrefix)
	config = config.SetJptrlen(1024).SetMaxkeys(10000)
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)
	config = config.SortbyArrayLen(true).SortbyPropertyLen(true)
	config = config.UseMissing(false)
	fmt.Println(config)
	// Output:
	// nk:FloatNumber, ws:AnsiSpace, ct:LengthPrefix, arrayLenPrefix:true, propertyLenPrefix:true, doMissing:false
}

func ExampleCbor() {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))

	val1 := [][]byte{[]byte("hello"), []byte("world")}
	cbr.EncodeBytechunks(val1)
	fmt.Printf("value: %v cbor: %q\n", val1, cbr.Bytes())
	cbr.Reset(nil)

	val2 := [][2]interface{}{{"first", true}}
	cbr.EncodeMapslice(val2)
	fmt.Printf("value: %v cbor: %q\n", val2, cbr.Bytes())
	cbr.Reset(nil)

	val3, val4 := byte(128), int8(-10)
	cbr.EncodeSimpletype(val3).EncodeSmallint(val4)
	fmt.Printf("value: {%v,%v} cbor: %q\n", val3, val4, cbr.Bytes())
	cbr.Reset(nil)

	val5 := []string{"sound", "ok", "horn"}
	cbr.EncodeTextchunks(val5)
	fmt.Printf("value: %v cbor: %q\n", val5, cbr.Bytes())
	cbr.Reset(nil)

	config = NewDefaultConfig()
	cbr = config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue([]interface{}{10, 20, 100}).Tocbor(cbr)
	fmt.Printf("to json   : %q\n", cbr.Tojson(jsn).Bytes())
	fmt.Printf("to collate: %q\n", cbr.Tocollate(clt).Bytes())
	fmt.Printf("to value  : %v\n", cbr.Tovalue())

	// Output:
	// value: [[104 101 108 108 111] [119 111 114 108 100]] cbor: "_EhelloEworld\xff"
	// value: [[first true]] cbor: "efirst\xf5"
	// value: {128,-10} cbor: "\xf8\x80)"
	// value: [sound ok horn] cbor: "\u007fesoundbokdhorn\xff"
	// to json   : "[10,20,100]"
	// to collate: "nP>>21-\x00P>>22-\x00P>>31-\x00\x00"
	// to value  : [10 20 100]
}

func ExampleCbor_Append() {
	config := NewDefaultConfig()
	cbr1 := config.NewCbor(make([]byte, 0, 1024))
	cbr2 := config.NewCbor(make([]byte, 0, 1024))
	cbritem := config.NewCbor(make([]byte, 0, 1024))
	olditem := config.NewCbor(make([]byte, 0, 1024))

	config.NewJson([]byte(`[]`)).Tocbor(cbr1)
	fmt.Println("start with  `[]`")

	ptr := config.NewJsonpointer("")
	config.NewValue(10.0).Tocbor(cbritem.Reset(nil))
	cbr1.Append(ptr, cbritem, cbr2)
	fmt.Printf("after appending 10 %v\n", cbr2.Tovalue())
	cbr1.Reset(nil)

	ptr = config.NewJsonpointer("")
	config.NewValue(20.0).Tocbor(cbritem.Reset(nil))
	cbr2.Prepend(ptr, cbritem, cbr1)
	fmt.Printf("after prepending 20 %v\n", cbr1.Tovalue())
	cbr2.Reset(nil)

	ptr = config.NewJsonpointer("/1")
	config.NewValue(30.0).Tocbor(cbritem.Reset(nil))
	cbr1.Set(ptr, cbritem, cbr2, olditem)
	fmsg := "after setting 30 to second item %v old value %v\n"
	fmt.Printf(fmsg, cbr2.Tovalue(), olditem.Tovalue())
	cbr1.Reset(nil)

	ptr = config.NewJsonpointer("/1")
	cbr2.Get(ptr, cbritem.Reset(nil))
	fmt.Printf("get second item %v\n", cbritem.Tovalue())

	ptr = config.NewJsonpointer("/0")
	cbr2.Delete(ptr, cbr1, cbritem.Reset(nil))
	fmsg = "after deleting first item %v, deleted value %v\n"
	fmt.Printf(fmsg, cbr1.Tovalue(), cbritem.Tovalue())
	cbr2.Reset(nil)

	// Output:
	// start with  `[]`
	// after appending 10 [10]
	// after prepending 20 [20 10]
	// after setting 30 to second item [20 30] old value 10
	// get second item 30
	// after deleting first item [30], deleted value 20
}

func ExampleCollate() {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue([]interface{}{10, 20, 100}).Tocollate(clt)
	fmt.Printf("json : %q\n", clt.Tojson(jsn).Bytes())
	fmt.Printf("cbor : %q\n", clt.Tocbor(cbr).Bytes())
	fmt.Printf("value: %v\n", clt.Tovalue())
	// Output:
	// json : "[1e+01,2e+01,1e+02]"
	// cbor : "\x9f\xfb@$\x00\x00\x00\x00\x00\x00\xfb@4\x00\x00\x00\x00\x00\x00\xfb@Y\x00\x00\x00\x00\x00\x00\xff"
	// value: [10 20 100]
}

func ExampleJson() {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 1024))
	jsn := config.NewJson(make([]byte, 0, 1024))
	clt := config.NewCollate(make([]byte, 0, 1024))

	config.NewValue([]interface{}{10, 20, 100}).Tojson(jsn)
	fmt.Printf("collate : %q\n", jsn.Tocollate(clt).Bytes())
	fmt.Printf("cbor : %q\n", jsn.Tocbor(cbr).Bytes())
	_, value := jsn.Tovalue()
	fmt.Printf("value: %v\n", value)
	// Output:
	// collate : "nP>>21-\x00P>>22-\x00P>>31-\x00\x00"
	// cbor : "\x9f\xfb@$\x00\x00\x00\x00\x00\x00\xfb@4\x00\x00\x00\x00\x00\x00\xfb@Y\x00\x00\x00\x00\x00\x00\xff"
	// value: [10 20 100]
}
