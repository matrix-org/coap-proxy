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
	"net"

	"github.com/flynn/noise"
)

// InMemoryKeyStore is a struct containing remote and local Diffie Hellman keys
// implemented by the noise protocol library.
type InMemoryKeyStore struct {
	remoteKeys     map[net.Addr][]byte
	localStaticKey noise.DHKey
}

// NewKeyStore is a function that creates a new InMemoryKeyStore instance
func NewKeyStore() *InMemoryKeyStore {
	keyStore := &InMemoryKeyStore{}
	keyStore.remoteKeys = make(map[net.Addr][]byte)
	return keyStore
}

// GetLocalKey is a function that returns a static local key from the InMemoryKeyStore
func (ks *InMemoryKeyStore) GetLocalKey() (noise.DHKey, error) {
	return ks.localStaticKey, nil
}

// SetLocalKey is a function that takes in a DHKey and inserts it into the InMemoryKeyStore
func (ks *InMemoryKeyStore) SetLocalKey(key noise.DHKey) error {
	ks.localStaticKey = key
	return nil
}

// GetRemoteKey is a function that returns a remote key from the InMemoryKeyStore
func (ks *InMemoryKeyStore) GetRemoteKey(addr net.Addr) ([]byte, error) {
	return ks.remoteKeys[addr], nil
}

// SetRemoteKey is a function that takes in a remote key and the address it is
// associated with and inserts/updates it in the InMemoryKeyStore
func (ks *InMemoryKeyStore) SetRemoteKey(addr net.Addr, key []byte) error {
	ks.remoteKeys[addr] = key
	return nil
}
