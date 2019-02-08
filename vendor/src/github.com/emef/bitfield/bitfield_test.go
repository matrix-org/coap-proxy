package bitfield

import (
	"math/big"
	"testing"
)

func TestTest(t *testing.T) {
	b := New(100)
	if b.Test(0) {
		t.Errorf("Test is broken!")
	}
	b[0] = 1
	if !b.Test(0) {
		t.Errorf("Test is broken!")
	}
}

func TestSet(t *testing.T) {
	b := New(100)
	for i := uint32(0); i < 100; i++ {
		if b.Test(i) {
			t.Errorf("%d shouldn't be set", i)
		}
		b.Set(i)
		if !b.Test(i) {
			t.Errorf("%d should be set", i)
		}
	}
}

func TestClear(t *testing.T) {
	b := New(100)
	for i := uint32(0); i < 100; i++ {
		b.Set(i)
	}

	for i := uint32(0); i < 100; i += 3 {
		b.Clear(i)
	}

	for i := uint32(0); i < 100; i++ {
		if b.Test(i) != (i%3 != 0) {
			t.Errorf("Clear is broken!")
		}
	}
}

func TestFlip(t *testing.T) {
	b := New(100)
	for i := uint32(0); i < 100; i++ {
		if b.Test(i) {
			t.Errorf("%d shouldn't be set", i)
		}
		b.Flip(i)
		if !b.Test(i) {
			t.Errorf("%d should be set", i)
		}
		b.Flip(i)
		if b.Test(i) {
			t.Errorf("%d shouldn't be set", i)
		}
	}
}

func TestClearAll(t *testing.T) {
	b := New(100)
	b[0] = 0xff
	b[2] = 0xff
	b[3] = 0xff
	b[4] = 0xff
	b.ClearAll()
	for i := uint32(0); i < 100; i++ {
		if b.Test(i) {
			t.Errorf("%d shouldn't be set", i)
		}
	}
}

func TestSetAll(t *testing.T) {
	b := New(100)
	b.SetAll()
	for i := uint32(0); i < 100; i++ {
		if !b.Test(i) {
			t.Errorf("%d should be set", i)
		}
	}
}

func TestFlipAll(t *testing.T) {
	b := New(100)
	b[0] = 0xff
	b[1] = 0xff
	b[2] = 0xff
	b[3] = 0xff
	b.FlipAll()
	for i := uint32(0); i < 100; i++ {
		if i < 32 {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		} else {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		}
	}
}

func TestANDMask(t *testing.T) {
	b := New(100)
	b[0] = 0xff
	b[1] = 0xff
	b[2] = 0xff
	b[3] = 0xff
	b2 := New(100)
	b2[2] = 0xff
	b2[3] = 0xff
	b2[4] = 0xff
	b2[5] = 0xff
	b.ANDMask(b2)
	for i := uint32(0); i < 100; i++ {
		if i >= 16 && i < 32 {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		} else {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		}
	}
}

func TestORMask(t *testing.T) {
	b := New(100)
	b[0] = 0xff
	b[1] = 0xff
	b[2] = 0xff
	b[3] = 0xff
	b2 := New(100)
	b2[2] = 0xff
	b2[3] = 0xff
	b2[4] = 0xff
	b2[5] = 0xff
	b.ORMask(b2)
	for i := uint32(0); i < 100; i++ {
		if i < 48 {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		} else {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		}
	}
}

func TestXORMask(t *testing.T) {
	b := New(100)
	b[0] = 0xff
	b[1] = 0xff
	b[2] = 0xff
	b[3] = 0xff
	b2 := New(100)
	b2[2] = 0xff
	b2[3] = 0xff
	b2[4] = 0xff
	b2[5] = 0xff
	b.XORMask(b2)
	for i := uint32(0); i < 100; i++ {
		if i < 16 || (i >= 32 && i < 48) {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		} else {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		}
	}
}

func TestFromUint32(t *testing.T) {
	b := NewFromUint32(uint32(0xff00))
	for i := uint32(0); i < 32; i++ {
		if i >= 8 && i < 16 {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		} else {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		}
	}
}

func TestFromUint64(t *testing.T) {
	b := NewFromUint64(uint64(0xff00))
	for i := uint32(0); i < 64; i++ {
		if i >= 8 && i < 16 {
			if !b.Test(i) {
				t.Errorf("%d should be set", i)
			}
		} else {
			if b.Test(i) {
				t.Errorf("%d shouldn't be set", i)
			}
		}
	}
}

func TestToUint32(t *testing.T) {
	baseUint := uint32(0xff00)
	b := NewFromUint32(baseUint)
	ui := b.ToUint32()
	if !(ui == baseUint) {
		t.Errorf("ToUint32 is broken!")
	}
}

func TestToUint32Safe(t *testing.T) {
	baseUint := uint32(0xff00)
	b := NewFromUint32(baseUint)
	ui := b.ToUint32Safe()
	if !(ui == baseUint) {
		t.Errorf("ToUint32Safe is broken!")
	}
}
func TestToUint64(t *testing.T) {
	baseUint := uint64(0xff00)
	b := NewFromUint64(baseUint)
	ui := b.ToUint64()
	if !(ui == baseUint) {
		t.Errorf("ToUint64 is broken!")
	}
}
func TestToUint64Safe(t *testing.T) {
	baseUint := uint64(0xff00)
	b := NewFromUint64(baseUint)
	ui := b.ToUint64Safe()
	if !(ui == baseUint) {
		t.Errorf("ToUint64Safe is broken!")
	}
}

func BenchmarkSet(b *testing.B) {
	l := uint32(1000)
	s := New(int(l))
	for n := 0; n < b.N; n++ {
		for i := uint32(0); i < l; i++ {
			s.Set(i)
		}
	}
}

func BenchmarkSetMathBig(b *testing.B) {
	l := 1000
	s := big.NewInt(0)
	for n := 0; n < b.N; n++ {
		for i := 0; i < l; i++ {
			s.SetBit(s, i, 1)
		}
	}
}
