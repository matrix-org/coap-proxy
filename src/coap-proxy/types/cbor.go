// Copyright 2019 New Vector Ltd
//
// This file is part of coap-proxy.
//
// coap-proxy is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// coap-proxy is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with coap-proxy.  If not, see <https://www.gnu.org/licenses/>.

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
