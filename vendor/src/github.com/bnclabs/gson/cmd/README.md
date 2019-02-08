General options
---------------

* `-inpfile` process file containing JSON doc(s) based on other options.
* `-inptxt` process input text based on their options.
* `-mprof` take memory profile for testdata/code.json.gz.
* `-outfile` write output to file.
* `-overheads` compute overheads on CBOR and collation encoding.
* `-quote` use strconv.Unquote on inptxt/inpfile.
* `-repeat` repeat count.
* `-nk` can be `smart`, treat number as int64 or fall back to float64, or,
  `float`, treat number only as float64 (default "float").
* `-ws` can be `ansi` white space, or, `unicode` white space, default
  "ANSI".

To include [n1ql](https://www.couchbase.com/products/n1ql), compile it with
`-tags n1ql`.

Convert from JSON
-----------------

* `-json2cbor` convert inptxt or content in inpfile to CBOR output.
* `-json2collate` convert inptxt or content in inpfile to collated output.
* `-json2value` convert inptxt or content in inpfile to golang value.

**options for JSON**

* `-pointers` list of json-pointers for doc specified by input-file.

Convert from CBOR
-----------------

* `-cbor2collate` convert inptxt or content in inpfile to collated output.
* `-cbor2json` convert inptxt or content in inpfile to JSON output.
* `-cbor2value` convert inptxt or content in inpfile to golang value.

**options for CBOR**

* `-ct` container encoding for CBOR, allowed `stream` (default), or,
  `lenprefix`.

Convert from Collate
--------------------

* `-collate2cbor` convert inptxt or content in inpfile to CBOR output.
* `-collate2json` convert inptxt or content in inpfile to JSON output.
* `-collate2value` convert inptxt or content in inpfile to value.


**options for collation**

* `-arrlenprefix` set SortbyArrayLen for collation ordering.
* `-maplenprefix` SortbyPropertyLen for collation ordering (default true)
* `-domissing` consider missing type while collation (default true).
* `-collatesort` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* `-n1qlsort` sort inpfile, with one or more JSON terms, using
  collation algorithm.
* `-checkdir` test files for collation order in specified directory. For
  every input file `<checkdir>/filename`, there should be reference file
  `<checkdir>/filename.ref`.

Convert from value
------------------

* `-value2cbor` convert inptxt JSON to value and then to CBOR.
* `-value2json` convert inptxt JSON to value and then back to JSON.
* `-value2collate` convert inptxt JSON to value and then back to binary.


Examples
--------

**Transformations**

```bash
$ gson -inpfile example.json -json2value
Json: "hello world"
Valu: hello world
```

```bash
$ gson -inptxt '"hello world"' -json2value
Json: "hello world"
Valu: hello world
```

```bash
$ gson -inptxt '"hello world"' -json2cbor
Json: "hello world"
Cbor: [107 104 101 108 108 111 32 119 111 114 108 100]
Json: "hello world"
```

```bash
$ gson -inptxt '"hello world"' -json2collate
Json: "hello world"
Coll: "\x06hello world\x00\x00"
Coll: [6 104 101 108 108 111 32 119 111 114 108 100 0 0]
```

Similarly to transform from CBOR:

```bash
$ gson -inptxt "khello world" -cbor2value
$ gson -inptxt "khello world" -cbor2json
$ gson -inptxt "khello world" -cbor2collate
```

Specifying container type for transforming to CBOR:

```bash
$ go build -o gson; gson -inptxt "[10,20]" -ct lenprefix -json2cbor
Json: [10,20]
Cbor: [130 251 64 36 0 0 0 0 0 0 251 64 52 0 0 0 0 0 0]
Cbor: "\x82\xfb@$\x00\x00\x00\x00\x00\x00\xfb@4\x00\x00\x00\x00\x00\x00"
Json: [10,20]
$ go build -o gson; gson -inptxt "[10,20]" -ct stream -json2cbor
Json: [10,20]
Cbor: [159 251 64 36 0 0 0 0 0 0 251 64 52 0 0 0 0 0 0 255]
Cbor: "\x9f\xfb@$\x00\x00\x00\x00\x00\x00\xfb@4\x00\x00\x00\x00\x00\x00\xff"
Json: [10,20]
```

Similarly to transform from collate:

```bash
$ gson -inpfile example.coll -collate2value
$ gson -inpfile example.coll -collate2json
$ gson -inpfile example.coll -collate2cbor
```

Similarly to transform from value:

```bash
$ gson -inptxt '"hello world"' -value2json
$ gson -inptxt '"hello world"' -value2cbor
$ gson -inptxt '"hello world"' -value2collate
```

**possible list of json-pointer from a given doc**

```bash
$ gson -inpfile typical.json -pointers
```
