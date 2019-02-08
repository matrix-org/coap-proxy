package types

import (
	"bytes"

	"coap-proxy/common"

	"github.com/ugorji/go/codec"
)

// CBOR is a type that allows for encoding and decoding of arbitrary types to
// and from CBOR
type CBOR struct{}

// Encode encodes any golang struct into CBOR
func (c *CBOR) Encode(val interface{}) []byte {
	common.DumpPayload("Encoding CBOR", val)

	var b bytes.Buffer

	var cborH codec.CborHandle
	cborH.Canonical = true
	enc := codec.NewEncoder(&b, &cborH)
	err := enc.Encode(val)
	if err != nil {
		panic(err)
	}

	return b.Bytes()

	// TODO: Same as above. For some reason it also blows up in some cases on
	// the JSON->CBOR way if the allocated buffer is smaller than the payload.
	// cbr = cbr.Reset(make([]byte, 0, len(pl)*2))
	// return jsn.Reset(pl).Tocbor(cbr)
}

// Decode decodes a CBOR byte array to a golang struct
func (c *CBOR) Decode(pl []byte) interface{} {
	common.DumpPayload("Decoding CBOR", pl)

	var val interface{}

	var cborH codec.Handle = new(codec.CborHandle)
	dec := codec.NewDecoderBytes(pl, cborH)
	err := dec.Decode(&val)
	if err != nil {
		panic(err)
	}

	return val
}
