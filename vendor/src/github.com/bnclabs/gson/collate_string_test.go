package gson

import "testing"
import "bytes"

func TestSuffixCoding(t *testing.T) {
	testcases := [][]byte{
		[]byte("hello\x00wo\xffrld\x00"),
		[]byte("hello\x00wo\xffrld\x00ln"),
	}
	for _, bs := range testcases {
		code, text := make([]byte, 1024), make([]byte, 1024)
		n := suffixEncodeString(bs, code)
		code[n] = Terminator
		n++
		x, y := suffixDecodeString(code[:n], text)
		if bytes.Compare(bs, text[:y]) != 0 {
			t.Error("Suffix coding for strings failed")
		}
		if l := len(code[x:n]); l != 0 {
			t.Errorf("Suffix coding for strings, residue found %v", l)
		}
	}
}

func BenchmarkSuffixEncode(b *testing.B) {
	var code [1024]byte
	bs := []byte("hello\x00wo\xffrld\x00")
	for i := 0; i < b.N; i++ {
		suffixEncodeString(bs, code[:])
	}
}

func BenchmarkSuffixDecode(b *testing.B) {
	var code, text [1024]byte
	bs := []byte("hello\x00wo\xffrld\x00")
	n := suffixEncodeString(bs, code[:])
	code[n] = Terminator
	n++
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		suffixDecodeString(code[:n], text[:])
	}
}
