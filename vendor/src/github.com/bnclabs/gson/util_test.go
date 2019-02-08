package gson

import "testing"
import "encoding/json"
import "sort"
import "reflect"
import "fmt"

func TestBytes2Str(t *testing.T) {
	if bytes2str(nil) != "" {
		t.Errorf("fail bytes2str(nil)")
	}
}

func TestStr2Bytes(t *testing.T) {
	if str2bytes("") != nil {
		t.Errorf(`fail str2bytes("")`)
	}
}

func TestCborMap2Golang(t *testing.T) {
	ref := `{"a":10,"b":[true,false,null]}`
	config := NewDefaultConfig()
	jsn := config.NewJson([]byte(ref))
	cbr := config.NewCbor(make([]byte, 0, 1024))

	_, val1 := jsn.Tovalue()
	value := config.NewValue(GolangMap2cborMap(val1))
	value.Tocbor(cbr)
	val2 := cbr.Tovalue()
	data, err := json.Marshal(CborMap2golangMap(val2))
	if err != nil {
		t.Fatalf("json parsing: %v\n	%v", val2, err)
	}
	if s := string(data); s != ref {
		t.Errorf("expected %q, got %q", ref, s)
	}
}

func TestKvrefs(t *testing.T) {
	items := make(kvrefs, 4)
	items[0] = kvref{"1", []byte("1")}
	items[2] = kvref{"3", []byte("3")}
	items[1] = kvref{"2", []byte("2")}
	items[3] = kvref{"0", []byte("0")}
	sort.Sort(items)
	ref := kvrefs{
		kvref{"0", []byte("0")},
		kvref{"1", []byte("1")},
		kvref{"2", []byte("2")},
		kvref{"3", []byte("3")},
	}
	if !reflect.DeepEqual(ref, items) {
		t.Errorf("expected %v, got %v", ref, items)
	}
}

func TestCollated2Number(t *testing.T) {
	tcases := [][]interface{}{
		{
			float64(0),
			"0",
			float64(0), float64(0),
		},
		{
			uint64(0),
			"0",
			float64(0), float64(0),
		},
		{
			int64(0),
			"0",
			float64(0), float64(0),
		},
		{
			float64(10.1231131311),
			">>2101231131311-",
			float64(10.1231131311), float64(10.1231131311),
		},
		{
			float64(-10.1231131311),
			"--7898768868688>",
			float64(-10.1231131311), float64(-10.1231131311),
		},
		{
			float64(9007199254740992),
			">>>2169007199254740992-",
			float64(9.007199254740992e+15), uint64(9007199254740992),
		},
		{
			float64(-9007199254740992),
			"---7830992800745259007>",
			float64(-9.007199254740992e+15), int64(-9.007199254740992e+15),
		},
		{
			int64(9223372036854775807),
			">>>2199223372036854775807-",
			float64(9.223372036854776e+18), uint64(9223372036854775807),
		},
		{
			int64(-9223372036854775808),
			"---7800776627963145224191>",
			float64(-9.223372036854776e+18), int64(-9223372036854775808),
		},
		{
			uint64(9223372036854775808),
			">>>2199223372036854775808-",
			float64(9.223372036854776e+18), uint64(9223372036854775808),
		},
		{
			uint64(18446744073709551615),
			">>>22018446744073709551615-",
			float64(1.8446744073709552e+19), uint64(18446744073709551615),
		},
	}

	var code [64]byte
	var n int
	config := NewDefaultConfig()

	for _, tcase := range tcases {
		t.Logf("%T %v", tcase[0], tcase[0])
		switch val := tcase[0].(type) {
		case float64:
			n = collateFloat64(val, code[:])
		case uint64:
			n = collateUint64(val, code[:], config)
		case int64:
			n = collateInt64(val, code[:], config)
		}
		if ref := tcase[1].(string); ref != string(code[:n]) {
			t.Errorf("expected %v, got %v", ref, string(code[:n]))
		}

		ui, i, f, what := collated2Number(code[:n], FloatNumber)
		switch what {
		case 1:
			if x := uint64(tcase[2].(float64)); ui != x {
				t.Errorf("expected uint64 %v, got %v", x, ui)
			}
		case 2:
			if x := int64(tcase[2].(float64)); i != x {
				t.Errorf("expected int64 %v, got %v", x, i)
			}
		case 3:
			if x := tcase[2].(float64); f != x {
				t.Errorf("expected float64 %v, got %v", x, f)
			}
		default:
			panic("what is unknown")
		}

		ui, i, f, what = collated2Number(code[:n], SmartNumber)
		switch what {
		case 1:
			if x := tcase[3].(uint64); ui != x {
				t.Errorf("expected uint64 %v, got %v", x, ui)
			}
		case 2:
			if x := tcase[3].(int64); i != x {
				t.Errorf("expected int64 %v, got %v", x, i)
			}
		case 3:
			if x := tcase[3].(float64); f != x {
				t.Errorf("expected float64 %v, got %v", x, f)
			}
		default:
			panic("what is unknown")
		}
	}
}

func BenchmarkBytes2Str(b *testing.B) {
	bs := []byte("hello world")
	for i := 0; i < b.N; i++ {
		bytes2str(bs)
	}
}

func BenchmarkStr2Bytes(b *testing.B) {
	s := "hello world"
	for i := 0; i < b.N; i++ {
		str2bytes(s)
	}
}

func compareJSONs(t *testing.T, json1, json2 string) error {
	var m1, m2 interface{}
	err := json.Unmarshal(str2bytes(json1), &m1)
	if err != nil {
		return fmt.Errorf("parsing %v: %v", json1, err)
	}
	err = json.Unmarshal(str2bytes(json2), &m2)
	if err != nil {
		return fmt.Errorf("parsing %v: %v", json2, err)
	}
	if !reflect.DeepEqual(m1, m2) {
		return fmt.Errorf("expected %v, got %v", m1, m2)
	}
	return nil
}
