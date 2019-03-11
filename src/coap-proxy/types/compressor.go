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
	"compress/flate"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emef/bitfield"
	"github.com/ugorji/go/codec"
)

// destTable is a struct that contains a list of servers and a distance cost. It
// is used in fanout routing for determining which servers to relay federation
// traffic through.
type destTable struct {
	cost    uint
	servers []uint
}

// Compressor implements go-coap.Compressor
type Compressor struct {
	dict []byte // dict is a dictionary of common string values to flate data
	cbor *CBOR  // cbor is an instance of a cbor struct for de/encoding CBOR data
}

// NewCompressor returns a new instance of the Compressor struct with its
// dictionary initialised from the given files.
// Returns an error if the files couldn't be parsed or read.
func NewCompressor(mapsDir string, mapFiles []string, cborStruct *CBOR) (*Compressor, error) {
	c := new(Compressor)

	var d = ""

	var buf []string
	j := new(JSON)
	for _, f := range mapFiles {
		buf = make([]string, 0)

		if err := j.ParseFile(
			filepath.Join(mapsDir, f),
			&buf,
		); err != nil {
			return nil, err
		}

		for _, el := range buf {
			d += el
		}
	}

	b, err := ioutil.ReadFile(filepath.Join(mapsDir, "extra_flate_data"))
	if err != nil {
		return nil, err
	}

	var parsedBytes []byte
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		val, err := strconv.Unquote(line)
		if err == nil {
			line = val
		}
		parsedBytes = append(parsedBytes, []byte(line)...)
	}

	c.dict = append([]byte(d), parsedBytes...)
	c.cbor = cborStruct

	return c, nil
}

// CompressPayload compresses a given byte array
func (c *Compressor) CompressPayload(j []byte) ([]byte, error) {
	var b bytes.Buffer

	// Compress the data using the specially crafted dictionary.
	zw, err := flate.NewWriterDict(&b, flate.BestCompression, c.dict)
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(zw, bytes.NewReader(j)); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// DecompressPayload decompresses a given byte array
func (c *Compressor) DecompressPayload(j []byte) ([]byte, error) {
	var b bytes.Buffer

	zr := flate.NewReaderDict(bytes.NewReader(j), c.dict)

	if _, err := io.Copy(&b, zr); err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// CompressTransaction is a function that compresses PDU bodies and destination
// tables held inside a federation transaction.
func (c *Compressor) CompressTransaction(val interface{}) interface{} {
	bodyMap, ok := val.(map[interface{}]interface{})
	if !ok {
		return val
	}

	pduSlice, ok := bodyMap["pdus"].([]interface{})
	if !ok {
		return val
	}

	for i := range pduSlice {
		pdu, ok := pduSlice[i].(map[interface{}]interface{})
		if !ok {
			continue
		}

		pduUnsigned, ok := pdu["unsigned"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		pduUnsigned["dtab"] = c.compressDestTable(pduUnsigned["dtab"])
	}

	return val
}

// DecompressTransaction is a function that compresses a transaction for
// federation traffic.
func (c *Compressor) DecompressTransaction(val interface{}) interface{} {
	bodyMap, ok := val.(map[interface{}]interface{})
	if !ok {
		return val
	}

	pduSlice, ok := bodyMap["pdus"].([]interface{})
	if !ok {
		return val
	}

	for i := range pduSlice {
		pdu, ok := pduSlice[i].(map[interface{}]interface{})
		if !ok {
			continue
		}

		pduUnsigned, ok := pdu["unsigned"].(map[interface{}]interface{})
		if !ok {
			continue
		}

		pduUnsigned["dtab"] = c.decompressDestTable(pduUnsigned["dtab"])
	}

	return val
}

// compressDestTable is a function that encodes the destinations in a
// destTable into an integer.
func (c *Compressor) compressDestTable(val interface{}) interface{} {
	destTableSlice, ok := val.([]interface{})
	if !ok {
		return val
	}

	var entries []destTable
	for _, entry := range destTableSlice {
		entrySlice, ok := entry.([]interface{})
		if !ok {
			return val
		}

		if len(entrySlice) != 2 {
			return val
		}

		cost, ok := entrySlice[0].(uint64)
		if !ok {
			return val
		}

		var servers []uint
		if slice, ok := entrySlice[1].([]interface{}); ok {
			for _, s := range slice {
				if k, ok := s.(string); ok {
					// TODO: fs-specific stuff
					k = strings.Replace(k, "synapse", "", 1)
					u, err := strconv.ParseUint(k, 10, 64)
					if err == nil {
						servers = append(servers, uint(u))
					}
				}
			}
		}

		if len(servers) == 0 {
			continue
		}

		entries = append(entries, destTable{
			cost:    uint(cost),
			servers: servers,
		})
	}

	finalMap := make(map[uint]bitfield.BitField)

	if len(entries) == 0 {
		return finalMap
	}

	for _, entry := range entries {
		var maxServer uint
		for _, s := range entry.servers {
			if maxServer < s {
				maxServer = s
			}
		}

		common.Debug("Max server is:", maxServer)

		field := bitfield.New(int(maxServer) + 1)

		for _, s := range entry.servers {
			field.Set(uint32(s))
		}

		finalMap[entry.cost] = field
	}

	return finalMap
}

// decompressDestTable is a function that decodes the destinations in a
// destTable from an integer back into their original addresses.
func (c *Compressor) decompressDestTable(val interface{}) interface{} {
	// Cheekily just serialize/deserialize to save having to manually parse stuff
	bytes := c.cbor.Encode(val)

	var destTableMap map[uint][]byte

	var cborH codec.Handle = new(codec.CborHandle)
	dec := codec.NewDecoderBytes(bytes, cborH)
	err := dec.Decode(&destTableMap)
	if err != nil {
		return val
	}

	var finalSlice []interface{}
	for cost, bits := range destTableMap {
		var entry []interface{}
		entry = append(entry, cost)

		parsedBitfield := bitfield.BitField(bits)

		var servers []string
		for idx := 0; idx < len(bits)*8; idx++ {
			if parsedBitfield.Test(uint32(idx)) {
				servers = append(servers, fmt.Sprintf("synapse%d", idx))
			}
		}

		entry = append(entry, servers)
		finalSlice = append(finalSlice, entry)
	}

	return finalSlice
}
