package gson

import "encoding/json"
import "testing"

func BenchmarkMarshalJson(b *testing.B) {
	config := NewDefaultConfig()
	config = config.SetNumberKind(FloatNumber).SetSpaceKind(AnsiSpace)

	jsn := config.NewJson(testdataFile("testdata/typical.json"))
	_, val := jsn.Tovalue()
	b.SetBytes(int64(len(jsn.Bytes())))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(val)
	}
}
