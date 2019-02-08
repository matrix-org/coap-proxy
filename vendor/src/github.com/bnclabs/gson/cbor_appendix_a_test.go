package gson

import "testing"
import "reflect"
import "encoding/json"
import hexcodec "encoding/hex"

func TestCborAppendixA(t *testing.T) {
	t.Skip("skipping CborAppendixA")
	var testcases []interface{}

	data := testdataFile("testdata/cbor.appendix.a.json")
	json.Unmarshal(data, &testcases)

	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 10))

	for _, tcase := range testcases {
		tcmap := tcase.(map[string]interface{})
		hexinp, refval := tcmap["hex"].(string), tcmap["decoded"].(interface{})
		roundtrip := tcmap["roundtrip"].(bool)
		cborbuf, err := hexcodec.DecodeString(hexinp)
		if err != nil {
			t.Log(tcase)
			t.Error(err)
		}
		val := cbr.Reset(cborbuf).Tovalue()
		if reflect.DeepEqual(val, refval) == false {
			t.Log(tcase)
			t.Errorf("expected %v, got %v", refval, val)
		}
		if roundtrip {
			v := config.NewValue(val)
			cborbytes := v.Tocbor(cbr.Reset(make([]byte, 1024*1024))).Bytes()
			s := hexcodec.EncodeToString(cborbytes)
			if s != hexinp {
				t.Logf("roundtrip %v", tcase)
				t.Errorf("expected %q, got %q", hexinp, s)
			}
		}
	}
}
