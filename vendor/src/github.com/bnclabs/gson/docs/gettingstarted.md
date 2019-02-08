Getting started
===============

Gson is essentially a data representation package, transforming them from
one format to another. Added to that, some common operation on data, in
its different representation, can be supported as well.

And, all formats are specified as RFC standard ...

To start using the Gson package, start with configuration:

```go
config := NewDefaultConfig()
```

NewDefaultConfig() creates a new Config{} object with default set of
configuration. Note that fields within the Config{} object are intentionally
hidden, and it shall remain that way. Also, note that config objects are
neither re-entrant and nor thread-safe. It is advised to not to share
config object between go-routines unless method calls are globally
serialized.

Once we have the Config{} object its configuration settings can be modified
using the supplied APIs. For example:

```go
config := NewDefaultConfig()
config = config.SetNumberKind(FloatNumber).SetContainerEncoding(Stream)
```

Config{} objects are immutable. Call to one of the config API will return
a new config instance with new settings. This also allows config-apis to be
chained.

Use the config instance to create CBOR, JSON, Value and Collate instances.
It is not encouraged to create these instances directly. The created instances
will adhere to the configuration settings available from the config object
that was used to instantiate them. Once configured try to re-use these
instances as much as possible, to avoid memory pressure on GC.

**GSON objects:**

```go
val := config.NewValue("any golang value")  // *gson.Value
jsn := config.NewJson(json_byteslice)       // *gson.Json
cbr := config.NewCbor(cbor_byteslice)       // *gson.Cbor
clt := config.NewCollate(collated_byteslice)   // *gson.Collate
```

JSON, CBOR, and Collate can also be initialized with empty buffer,
that have adequate capacity:

```go
jsn := config.NewJson(make([]byte, 0, 1024))
cbr := config.NewCbor(make([]byte, 0, 1024))
clt := config.NewCollate(make([]byte, 0, 1024))
```

To reuse JSON, CBOR, or Collate objects:

```go
jsn.Reset(nil)
cbr.Reset(nil)
clt.Reset(nil)
```

To reset JSON, CBOR, or Collate objects with another buffer.

```go
jsn.Reset(another_json)
cbr.Reset(another_cbor)
clt.Reset(another_collate)
```

To get the underlying buffer as byte-slice.

```go
jsn.Bytes()
cbr.Bytes()
clt.Bytes()
```

An example transformation from one data format to another:

```go
val := config.NewValue(jsn.Tovalue()) // json -> value
cbr = jsn.Tocbor(cbr)                 // json -> cbor
clt = jsn.Tocollate(collate)          // json -> collate

// json -> collate -> cbor -> value
val := config.NewValue(jsn.Tocollate(cbr).Tocbor(clt).Tovalue())
```
