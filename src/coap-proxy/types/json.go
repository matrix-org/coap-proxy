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
	"encoding/json"
	"io/ioutil"

	"coap-proxy/common"

	"github.com/ugorji/go/codec"
)

// JSON is a struct that contains JSON encoding and decoding methods
type JSON struct{}

// Encode takes an arbitrary golang object and encodes it to JSON
func (j *JSON) Encode(val interface{}) []byte {
	var b bytes.Buffer

	var jsonH codec.Handle = new(codec.JsonHandle)
	enc := codec.NewEncoder(&b, jsonH)
	err := enc.Encode(val)
	if err != nil {
		panic(err)
	}

	common.DumpPayload("Encoding JSON", b.String())

	return b.Bytes()

	// TODO: We allocate a buffer with a len(pl)*2 capacity to ensure our buffer
	// can contain all of the JSON data (as len(jsonData) > len(cborData)). This
	// is far from being the most optimised way to do it, and a more efficient
	// way of computing the maximum size of the buffer to should be
	// investigated.
	// jsn = jsn.Reset(make([]byte, 0, len(pl)*2))
	// return cbr.Reset(pl).Tojson(jsn).Bytes()
}

// Decode takes a JSON byte array and produces a golang object
func (j *JSON) Decode(pl []byte) interface{} {
	var val interface{}

	common.DumpPayload("Decoding JSON", pl)

	var jsonH codec.Handle = new(codec.JsonHandle)
	dec := codec.NewDecoderBytes(pl, jsonH)
	err := dec.Decode(&val)
	if err != nil {
		panic(err)
	}

	return val

	// TODO: Same as above. For some reason it also blows up in some cases on
	// the JSON->CBOR way if the allocated buffer is smaller than the payload.
	// cbr = cbr.Reset(make([]byte, 0, len(pl)*2))
	// return jsn.Reset(pl).Tocbor(cbr)
}

// ParseFile takes in a filepath and a target struct and fills it with the
// contents of the JSON file
func (j *JSON) ParseFile(filepath string, target interface{}) error {
	common.Debugf("Parsing file %s", filepath)

	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, target); err != nil {
		return err
	}

	return nil
}
