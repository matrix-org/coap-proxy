package gson

import "strconv"
import "testing"
import "math/big"

func BenchmarkAtoi(b *testing.B) {
	for i := 0; i < b.N; i++ {
		strconv.Atoi("6744073709551615")
	}
}

func BenchmarkBigParseFloat(b *testing.B) {
	bf := big.NewFloat(0)
	bf.SetPrec(64)
	for i := 0; i < b.N; i++ {
		bf.Parse("18446744073709551615", 0)
	}
}

func BenchmarkBigUint64Fmt(b *testing.B) {
	bf := big.NewFloat(0)
	out := make([]byte, 0, 1024)
	num := uint64(18446744073709551615)
	bf.SetPrec(64)
	bf.SetUint64(num)

	for i := 0; i < b.N; i++ {
		bf.Append(out, 'e', -1)
	}
}

func BenchmarkBigInt64Fmt(b *testing.B) {
	bf := big.NewFloat(0)
	out := make([]byte, 0, 128)
	num := int64(446744073709551615)
	bf.SetPrec(64)
	bf.SetInt64(num)

	for i := 0; i < b.N; i++ {
		bf.Append(out, 'e', -1)
	}
}

func BenchmarkBigFloatFmt(b *testing.B) {
	bf := big.NewFloat(0)
	out := make([]byte, 0, 128)
	num := float64(18446744073709551615.123)
	bf.SetPrec(64)
	bf.SetFloat64(num)

	for i := 0; i < b.N; i++ {
		bf.Append(out, 'e', -1)
	}
}
