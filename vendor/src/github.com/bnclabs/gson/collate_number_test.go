package gson

import "strconv"
import "testing"

func TestInteger(t *testing.T) {
	var samples = [][2]string{
		{"7", ">7"},
		{"+7", ">7"},
		{"+1234567890", ">>>2101234567890"},
		{"-10", "--789"},
		{"-11", "--788"},
		{"-1234567891", "---7898765432108"},
		{"-1234567890", "---7898765432109"},
		{"-1234567889", "---7898765432110"},
		{"0", "0"},
		{"+0", "0"},
		{"-0", "0"},
	}
	text, code := make([]byte, 1024), make([]byte, 1024)
	for _, tcase := range samples {
		t.Logf("%v", tcase)
		sample, ref := tcase[0], tcase[1]
		n := collateInt([]byte(sample), code)
		if out := string(code[:n]); out != ref {
			t.Errorf("expected %q, got %q", ref, out)
		}
		_, y := collated2Int(code, text)
		if out := string(text[:y]); atoi(out, t) != atoi(sample, t) {
			t.Errorf("expected %q, got %q", ref, out)
		}
	}
}

func TestSmallDecimal(t *testing.T) {
	var samples = [][2]string{
		{"-0.9995", "-0004>"},
		{"-0.999", "-000>"},
		{"-0.0123", "-9876>"},
		{"-0.00123", "-99876>"},
		{"-0.0001233", "-9998766>"},
		{"-0.000123", "-999876>"},
		{"+0.000123", ">000123-"},
		{"+0.0001233", ">0001233-"},
		{"+0.00123", ">00123-"},
		{"+0.0123", ">0123-"},
		{"+0.999", ">999-"},
		{"+0.9995", ">9995-"},
	}
	code, text := make([]byte, 1024), make([]byte, 1024)
	for _, tcase := range samples {
		t.Logf("%v", tcase)
		sample, ref := tcase[0], tcase[1]
		x := collateSD([]byte(sample), code)
		if out := string(code[:x]); out != ref {
			t.Errorf("expected %q, got %q", ref, out)
		}
		_, y := collated2SD(code[:x], text)
		outf, samplef := atof(string(text[:y]), t), atof(sample, t)
		if outf != samplef {
			t.Errorf("expected %v, got %v", samplef, outf)
		}
	}
}

func TestFloat(t *testing.T) {
	var samples = [][2]string{
		{"-10000000000.0", "---7888>"},
		{"-1000000000.0", "---7898>"},
		{"-1.4", "--885>"},
		{"-1.3", "--886>"},
		{"-1", "--88>"},
		{"-0.123", "-0876>"},
		{"-0.0123", "->1876>"},
		{"-0.001233", "->28766>"},
		{"-0.00123", "->2876>"},
		{"0", "0"},
		{"+0.00123", ">-7123-"},
		{"+0.001233", ">-71233-"},
		{"+0.0123", ">-8123-"},
		{"+0.123", ">0123-"},
		{"+1", ">>11-"},
		{"+1.3", ">>113-"},
		{"+1.4", ">>114-"},
		{"+1000000000.0", ">>>2101-"},
		{"+10000000000.0", ">>>2111-"},
	}
	code, text := make([]byte, 1024), make([]byte, 1024)
	scratch := make([]byte, 0, 1024)
	for _, tcase := range samples {
		t.Logf("%v", tcase)
		sample, ref := tcase[0], tcase[1]
		f, err := strconv.ParseFloat(sample, 64)
		if err != nil {
			t.Fatal(err)
		}
		scratchf := strconv.AppendFloat(scratch, f, 'e', -1, 64)
		n := collateFloat(scratchf, code)
		if out := string(code[:n]); out != ref {
			t.Errorf("expected %q, got %q", ref, out)
		}
		_, y := collated2Float(code[:n], text)
		outf, samplef := atof(string(text[:y]), t), atof(sample, t)
		if outf != samplef {
			t.Errorf("expected %v, got %v", samplef, outf)
		}
	}
}

func TestLargeDecimal(t *testing.T) {
	var samples = [][2]string{
		{"+1", ">1-"},
		{"+1.01", ">101-"},
		{"+3.14", ">314-"},
		{"+3.145", ">3145-"},
		{"+10.5", ">>2105-"},
		{"+100.5", ">>31005-"},
		{"0", ">0-"},
		{"0.0", ">0-"},
		{"+0", ">0-"},
		{"+0.0", ">0-"},
		{"-0", "-9>"},
		{"-0.0", "-9>"},
		{"-100.5", "--68994>"},
		{"-10.5", "--7894>"},
		{"-3.145", "-6854>"},
		{"-3.14", "-685>"},
		{"-1.01", "-898>"},
		{"-1", "-8>"},
	}
	code, text := make([]byte, 1024), make([]byte, 1024)
	for _, tcase := range samples {
		t.Logf("%v", tcase)
		sample, ref := tcase[0], tcase[1]
		x := collateLD([]byte(sample), code)
		if out := string(code[:x]); out != ref {
			t.Errorf("expected %q, got %q", ref, out)
		}
		_, y := collated2LD(code[:x], text)
		outf, samplef := atof(string(text[:y]), t), atof(sample, t)
		if outf != samplef {
			t.Errorf("expected %v, got %v", samplef, outf)
		}
	}
}

func BenchmarkEncodeInt(b *testing.B) {
	bEncodeInt(b, []byte("+1234567890"))
}

func BenchmarkEncIntNeg(b *testing.B) {
	bEncodeInt(b, []byte("-1234567890"))
}

func BenchmarkDecodeInt(b *testing.B) {
	bDecodeInt(b, []byte(">>>2101234567890"))
}

func BenchmarkDecIntNeg(b *testing.B) {
	bDecodeInt(b, []byte("---7898765432108"))
}

func BenchmarkEncIntTypical(b *testing.B) {
	var samples = [][]byte{
		[]byte("+7"),
		[]byte("+123"),
		[]byte("+1234567890"),
		[]byte("-1234567891"),
		[]byte("-1234567890"),
		[]byte("-1234567889"),
		[]byte("0"),
		[]byte("0"),
	}
	ln := len(samples)
	var code [1024]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collateInt(samples[i%ln], code[:])
	}
}

func BenchmarkDecIntTypical(b *testing.B) {
	var samples = [][]byte{
		[]byte(">7"),
		[]byte(">>3123"),
		[]byte(">>>2101234567890"),
		[]byte("---7898765432108"),
		[]byte("---7898765432109"),
		[]byte("---7898765432110"),
		[]byte("0"),
	}
	var text [1024]byte
	ln := len(samples)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collated2Int(samples[i%ln], text[:])
	}
}

func BenchmarkEncodeFloat(b *testing.B) {
	bEncodeFloat(b, "+0.001233")
}

func BenchmarkEncFloatNeg(b *testing.B) {
	bEncodeFloat(b, "-10000000000")
}

func BenchmarkDecodeFloat(b *testing.B) {
	bDecodeFloat(b, "->28766>")
}

func BenchmarkDecFloatNeg(b *testing.B) {
	bDecodeFloat(b, "---7888>")
}

func BenchmarkEncFltTypical(b *testing.B) {
	var samples = [][]byte{
		[]byte("-10000000000.0"),
		[]byte("-1000000000.0"),
		[]byte("-1.4"),
		[]byte("-1.3"),
		[]byte("-1"),
		[]byte("-0.123"),
		[]byte("-0.0123"),
		[]byte("-0.001233"),
		[]byte("-0.00123"),
		[]byte("0"),
		[]byte("+0.00123"),
		[]byte("+0.001233"),
		[]byte("+0.0123"),
		[]byte("+0.123"),
		[]byte("+1"),
		[]byte("+1.3"),
		[]byte("+1.4"),
		[]byte("+1000000000.0"),
		[]byte("+10000000000.0"),
	}
	ln := len(samples)
	for i := 0; i < ln; i++ {
		f, _ := strconv.ParseFloat(string(samples[i]), 64)
		samples[i] = []byte(strconv.FormatFloat(f, 'e', -1, 64))
	}
	var code [1024]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collateFloat(samples[i%ln], code[:])
	}
}

func BenchmarkDecFltTypical(b *testing.B) {
	var samples = [][]byte{
		[]byte("-10000000000.0"),
		[]byte("-1000000000.0"),
		[]byte("-1.4"),
		[]byte("-1.3"),
		[]byte("-1"),
		[]byte("-0.123"),
		[]byte("-0.0123"),
		[]byte("-0.001233"),
		[]byte("-0.00123"),
		[]byte("0"),
		[]byte("+0.00123"),
		[]byte("+0.001233"),
		[]byte("+0.0123"),
		[]byte("+0.123"),
		[]byte("+1"),
		[]byte("+1.3"),
		[]byte("+1.4"),
		[]byte("+1000000000.0"),
		[]byte("+10000000000.0"),
	}
	ln := len(samples)
	for i := 0; i < ln; i++ {
		code := make([]byte, 128)
		f, _ := strconv.ParseFloat(string(samples[i]), 64)
		n := collateFloat([]byte(strconv.FormatFloat(f, 'e', -1, 64)), code[:])
		samples[i] = code[:n]
	}

	var text [1024]byte
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collated2Float(samples[i%ln], text[:])
	}
}

func BenchmarkEncodeSD(b *testing.B) {
	var samples = [][]byte{
		[]byte("-0.9995"),
		[]byte("-0.999"),
		[]byte("-0.0123"),
		[]byte("-0.00123"),
		[]byte("-0.0001233"),
		[]byte("-0.000123"),
		[]byte("+0.000123"),
		[]byte("+0.0001233"),
		[]byte("+0.00123"),
		[]byte("+0.0123"),
		[]byte("+0.999"),
		[]byte("+0.9995"),
	}
	var code [1024]byte
	ln := len(samples)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collateSD(samples[i%ln], code[:])
	}
}

func BenchmarkDecodeSD(b *testing.B) {
	var samples = [][]byte{
		[]byte("-0004>"),
		[]byte("-000>"),
		[]byte("-9876>"),
		[]byte("-99876>"),
		[]byte("-9998766>"),
		[]byte("-999876>"),
		[]byte(">000123-"),
		[]byte(">0001233-"),
		[]byte(">00123-"),
		[]byte(">0123-"),
		[]byte(">999-"),
		[]byte(">9995-"),
	}
	var text [1024]byte
	ln := len(samples)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collated2SD(samples[i%ln], text[:])
	}
}

func BenchmarkEncodeLD(b *testing.B) {
	var samples = [][]byte{
		[]byte("-100.5"),
		[]byte("-10.5"),
		[]byte("-3.145"),
		[]byte("-3.14"),
		[]byte("-1.01"),
		[]byte("+1.01"),
		[]byte("+3.14"),
		[]byte("+3.145"),
		[]byte("+10.5"),
		[]byte("+100.5"),
	}
	var code [1024]byte
	ln := len(samples)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collateLD(samples[i%ln], code[:])
	}
}

func BenchmarkDecodeLD(b *testing.B) {
	var samples = [][]byte{
		[]byte("--68994>"),
		[]byte("--7894>"),
		[]byte("-6854>"),
		[]byte("-685>"),
		[]byte("-898>"),
		[]byte("-89>"),
		[]byte(">1-"),
		[]byte(">101-"),
		[]byte(">314-"),
		[]byte(">3145-"),
		[]byte(">>2105-"),
		[]byte(">>31005-"),
	}
	var text [1024]byte
	ln := len(samples)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collated2LD(samples[i%ln], text[:])
	}
}

func bEncodeInt(b *testing.B, in []byte) {
	var code [1024]byte
	for i := 0; i < b.N; i++ {
		collateInt(in, code[:])
	}
}

func bDecodeInt(b *testing.B, in []byte) {
	var text [1024]byte
	for i := 0; i < b.N; i++ {
		collated2Int(in, text[:])
	}
}

func bEncodeFloat(b *testing.B, in string) {
	var code [1024]byte
	f, _ := strconv.ParseFloat(in, 64)
	inb := str2bytes(strconv.FormatFloat(f, 'e', -1, 64))
	for i := 0; i < b.N; i++ {
		collateFloat(inb, code[:])
	}
}

func bDecodeFloat(b *testing.B, in string) {
	var text [1024]byte
	inb := str2bytes(in)
	for i := 0; i < b.N; i++ {
		collated2Float(inb, text[:])
	}
}

func atoi(text string, t *testing.T) int {
	if val, err := strconv.Atoi(text); err == nil {
		return val
	}
	t.Errorf("atoi: Unable to convert %v", text)
	return 0
}

func atof(text string, t *testing.T) float64 {
	val, err := strconv.ParseFloat(text, 64)
	if err == nil {
		return val
	}
	t.Error("atof: Unable to convert", text, err)
	return 0.0
}
