// Package gson provide a toolkit for JSON representation, collation
// and transformation.
//
// Package provides APIs to convert data representation from one format
// to another. Supported formats are:
//   * JSON
//   * Golang value
//   * CBOR - Concise Binary Object Representation
//   * Binary-collation
//
// CBOR:
//
// Concise Binary Object Representation, CBOR, is based on RFC-7049
// specification to encode golang data into machine friendly format.
// Following golang native types are supported:
//   * nil, true, false.
//   * native integer types, and its alias, of all width.
//   * float32, float64.
//   * slice of bytes.
//   * native string.
//   * slice of interface - []interface{}.
//   * map of string to interface{} - map[string]interface{}.
//
// Types from golang's standard library and custom types provided
// by this package that can be encoded using CBOR:
//   * CborTagBytes: a cbor encoded []bytes treated as value.
//   * CborUndefined: encode a data-item as undefined.
//   * CborIndefinite: encode bytes, string, array and map of unspecified length.
//   * CborBreakStop: to encode end of CborIndefinite length item.
//   * CborTagEpoch: in seconds since epoch.
//   * CborTagEpochMicro: in micro-seconds epoch.
//   * CborTagFraction: m*(10**e)
//   * CborTagFloat: m*(2**e)
//   * CborTagPrefix: to self identify a binary blog as CBOR.
//
// Package also provides an implementation for encoding JSON to CBOR
// and vice-versa:
//   * Number can be encoded as integer or float.
//   * Arrays and maps are encoded using indefinite encoding.
//   * Byte-string encoding is not used.
//
// Json-Pointer:
//
// Package also provides a RFC-6901 (JSON-pointers) implementation.
//
// NOTE: Buffer supplied to APIs NewJson(), NewCbor(), NewCollate()
// should atleast be 128 bytes in size.
package gson
