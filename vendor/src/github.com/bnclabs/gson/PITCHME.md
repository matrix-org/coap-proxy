@title[gson]

@snap[midpoint slide1]
<h1>gson</h1>
@size[80%](Data formats)
@snapend

@snap[south-east author-box]
@fa[envelope](prataprc@gmail.com - R Pratap Chakravarthy) <br/>
@fa[github](https://github.com/bnclabs/gson) <br/>
@snapend

---

Goals
=====

@ul
* High performance algorithms for data transformation, serialization and manipulation.
* Based on well established standards.
* ZERO allocation when transforming from one format to another, except for APIs creating golang values from encoded data.
* [JSON](http://json.org) for web.
* [CBOR](http://cbor.io) for machine.
* [Binary-Collation](https://prataprc.github.io/jsonsort.io) for crazy fast comparison/sorting.
* [JSON-pointer](https://tools.ietf.org/html/rfc6901) for field lookup within document.
@ulend

---

Data Transformation
===================

@ul
* Golang native values are convenient for programming.
* Binary representation, like CBOR, is convenient for Machine and avoid parsing data before every computation.
* Collation format, crazy fast memcmp() on complex data shape.
* JSON for web / human interface.
@ulend

---

Transformations
===============

Convert from any format to any other format, without loss of information.

![Transforms](assets/transforms.png)

---

Data serialization
==================

@ul[para]
- **CBOR** is suggested for network serialization when nodes belong to the same cluster, especially when consumer of that data is another machine.
@ulend

@ul[para]
- **JSON** is suggested when data is serialized for web consumption or when the consumer of that data is a Human.
@ulend

@ul[para]
- During development JSON can be used for ease of debugging, and for production CBOR can be used. Note that data from one format can be seamlessly transformed to another format.
@ulend

---

Document processing
===================

@ul[para]
- To lookup a field within a document and to perform expression evaluation on the field, we recommend [JSON-pointer](https://tools.ietf.org/html/rfc6901).
@ulend

@ul[para]
- We aim to support ZERO allocation for field lookup and if it makes sense we can add expression evaluation as part of lookup.
@ulend

@ul[para]
- If document lookup and manipulations are not available in the expected format, please raise an [issue](http://github.com/bnclabs/gson/issues).
@ulend

---

<br/>
<br/>
<br/>
<br/>
**All formats are based on RFC standards and Collation format is
precisely defined.**

---

JSON for web
============

```json
{
  "firstName": "John",
  "lastName": "Smith",
  "age": 25,
  "city": "New York",
  "state": "NY",
  "phoneNumber": "212 555-1234",
}
```

JSON is a good trade-off between:

@ul
* Human readability.
* Representation of complex data.
* Parsing to machine value, which becomes unavoidable at some point.
@ulend

---

CBOR for machine
================

@ul
* Fast data serialization / de-serialization
* CBOR o/p can be compact when compared to JSON; due to variable length encoding.
* Extensible encoding, that can support future value types.
* Better fit for in-place evaluation, since values are stored in machine native formats.
* JSON specification is not precise about Number. CBOR can encode int64, uint64 even Big-numbers.
@ulend

---

CBOR Vs JSON
============

```bash
BenchmarkVal2CborTrue  100000000     12.7 ns/op  0 B/op   0 allocs/op
BenchmarkVal2CborFlt64 100000000     15.4 ns/op  0 B/op   0 allocs/op
BenchmarkCbor2ValTrue  20000000     62.3 ns/op   1 B/op   1 allocs/op
BenchmarkCbor2ValFlt64 20000000     68.4 ns/op   8 B/op   1 allocs/op

BenchmarkJson2ValBool  30000000     43.8 ns/op   1 B/op   1 allocs/op
BenchmarkJson2ValNum   20000000     88.1 ns/op   8 B/op   1 allocs/op
BenchmarkVal2JsonBool 100000000     15.4 ns/op   0 B/op   0 allocs/op
BenchmarkVal2JsonNum   10000000    135   ns/op   0 B/op   0 allocs/op
```

@ul[para]
- We get improvements between  2x to 4x. Note that if encoding/json is used, instead of gson.Json, degradation will be > 4x.
@ulend

---

Binary-Collation
================

Let us say a composite key looks like @color[blue](["paris", 35]),
which is basically an index on ``{city,age}``, the collated output
shall look like - @color[red]("\x12\x10paris\x00\x00\x0f>>235-\x00\x00")

```bash
# sorting a list of 10000 JSON entires for different shapes
# repeat 100 times
41.10s user 0.84s system 112% cpu 37.265 total
# sorting the same after collation encoding (using memcmp)
# repeat 100 times
2.93s user 0.09s system 115% cpu 2.622 total
```

<br/>

@ul[para]
- **An order of magnitude faster, 10x** @color[blue](CRAZY FAST COMPARISION).
@ulend

---

Ease of use
===========

Always start with the config object. Apply desired configuration,
note that config objects are immutable so make sure to receive a new
config object from each config API.

```go
config := NewDefaultConfig()
config = config.SetNumberKind(FloatNumber).SetContainerEncoding(Stream)
```

---

Format factories
================

Use the config object to create any number of buffer types: JSON,
CBOR, Value, Collate. Reuse the objects, avoid GC pressure.
**buffer objects are not thread-safe**

```go
val := config.NewValue("any golang value")
jsn := config.NewCbor(json_byteslice)
cbr := config.NewCbor(cbor_byteslice)
clt := config.NewCbor(collated_byteslice)

jsn.Reset(nil)
cbr.Reset(nil)
clt.Reset(nil)
```

@[1](create a value instance)
@[1](create a JSON instance)
@[1](create a CBOR instance)
@[1](create a collate instance)

@[6-8](instances and buffers can be reused)

---

Transform examples
==================

An example transformation from one data format to another, using
the config object and buffer types.

```go
val := config.NewValue(jsn.Tovalue())
cbr = jsn.Tocbor(cbr)
clt = jsn.Tocollate(clt)

val := config.NewValue(jsn.Tocollate(cbr).Tocbor(clt).Tovalue())

jsn.Bytes()
cbr.Bytes()
clt.Bytes()
```

@[1](create a new value instance, after parsing JSON to golang value)
@[2](convert json input to cbor, ``cbr`` contains CBOR encoded o/p)
@[3](convert json input to collate, ``clt`` contains collate encoded o/p)

@[5](APIs are convenient for function chaining)
@[7-9](To get the underlying data from each buffer)

---

Thank you
=========

If gson sounds useful please check out the following links.

<br/>

@fa[book] [Project README](https://github.com/bnclabs/gson). <br/>
@fa[code] [API doc](https://godoc.org/github.com/bnclabs/gson). <br/>
@fa[github] [Please contribute](https://github.com/bnclabs/gson/issues). <br/>
