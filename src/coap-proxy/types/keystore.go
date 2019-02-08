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
