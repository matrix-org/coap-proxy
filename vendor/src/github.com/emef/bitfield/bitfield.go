package bitfield

// BitField type
type BitField []byte

// New returns a new BitField of at least n bits, all 0s
func New(n int) BitField {
	n = 1 + ((n - 1) / 8) // Ceiling of the division
	return BitField(make([]byte, n))
}

// NewFromUint32 returns a new BitField of 4 bytes, with n initial value
func NewFromUint32(n uint32) BitField {
	b := BitField(make([]byte, 4))
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	b[2] = byte(n >> 16)
	b[3] = byte(n >> 24)
	return b
}

// NewFromUint64 returns a new BitField of 8 bytes, with n initial value
func NewFromUint64(n uint64) BitField {
	b := BitField(make([]byte, 8))
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	b[2] = byte(n >> 16)
	b[3] = byte(n >> 24)
	b[4] = byte(n >> 32)
	b[5] = byte(n >> 40)
	b[6] = byte(n >> 48)
	b[7] = byte(n >> 56)
	return b
}

// Size returns BitField size in bytes (not bits)
func (b BitField) Size() int {
	return len(b)
}

// Set sets bit i to 1
func (b BitField) Set(i uint32) {
	idx, offset := (i / 8), (i % 8)
	b[idx] |= (1 << uint(offset))
}

// Clear sets bit i to 0
func (b BitField) Clear(i uint32) {
	idx, offset := (i / 8), (i % 8)
	b[idx] &= ^(1 << uint(offset))
}

// Flip toggles the value of bit i
func (b BitField) Flip(i uint32) {
	idx, offset := (i / 8), (i % 8)
	b[idx] ^= (1 << uint(offset))
}

// Test returns true/false on bit i value
func (b BitField) Test(i uint32) bool {
	idx, offset := (i / 8), (i % 8)
	return (b[idx] & (1 << uint(offset))) != 0
}

// ClearAll sets all BitField values to 0
func (b BitField) ClearAll() {
	for idx := range b {
		b[idx] = 0
	}
}

// SetAll sets all BitField bits to 1
func (b BitField) SetAll() {
	for idx := range b {
		b[idx] = 0xff
	}
}

// FlipAll flips all the BitField bits (1's compliment)
func (b BitField) FlipAll() {
	for idx := range b {
		b[idx] = ^b[idx]
	}
}

// ANDMask performs an AND operation between b and m, storing result in b.
// If b is smaller than m, the extra bits of m are ignored (b isn't enlarged).
func (b BitField) ANDMask(m BitField) {
	maxidx := len(m)
	for idx := range b {
		// B is longer than mask, everything else should be 0 on AND
		if idx > maxidx {
			b[idx] = 0
			continue
		}
		b[idx] &= m[idx]
	}
}

// ORMask performs an OR operation between b and m, storing result in b.
// If b is smaller than m, the extra bits of m are ignored (b isn't enlarged).
func (b BitField) ORMask(m BitField) {
	maxidx := len(m)
	for idx := range b {
		// B is longer than mask, everything else should be b on OR
		if idx > maxidx {
			break
		}
		b[idx] |= m[idx]
	}
}

// XORMask performs an XOR operation between b and m, storing result in b.
// If b is smaller than m, the extra bits of m are ignored (b isn't enlarged).
func (b BitField) XORMask(m BitField) {
	maxidx := len(m)
	for idx := range b {
		// B is longer than mask, everything else should be b on XOR
		if idx > maxidx {
			break
		}
		b[idx] ^= m[idx]
	}
}

// ToUint32 returns the lowest 4 bytes as a uint32
// NO BOUNDS CHECKING, ENSURE BitField is at least 4 bytes long
func (b BitField) ToUint32() uint32 {
	var r uint32
	r |= uint32(b[0])
	r |= uint32(b[1]) << 8
	r |= uint32(b[2]) << 16
	r |= uint32(b[3]) << 24
	return r
}

// ToUint32Safe returns the lowest 4 bytes as a uint32
func (b BitField) ToUint32Safe() uint32 {
	var r uint32
	for idx := range b {
		r |= uint32(b[idx]) << uint32(idx*8)
		if idx == 3 {
			break
		}
	}
	return r
}

// ToUint64 returns the lowest 8 bytes as a uint64
// NO BOUNDS CHECKING, ENSURE BitField is at least 8 bytes long
func (b BitField) ToUint64() uint64 {
	var r uint64
	r |= uint64(b[0])
	r |= uint64(b[1]) << 8
	r |= uint64(b[2]) << 16
	r |= uint64(b[3]) << 24
	r |= uint64(b[4]) << 32
	r |= uint64(b[5]) << 40
	r |= uint64(b[6]) << 48
	r |= uint64(b[7]) << 56
	return r
}

// ToUint64Safe returns the lowest 8 bytes as a uint64
func (b BitField) ToUint64Safe() uint64 {
	var r uint64
	for idx := range b {
		r |= uint64(b[idx]) << uint64(idx*8)
		if idx == 7 {
			break
		}
	}
	return r
}
