package gson

import "testing"

// test case to sort/preserv-sort map-keys when transforming from
// one format to another.
func TestSortedkeys(t *testing.T) {
	json := []byte(`{"three":10, "two":20, "one":30}`)
	config := NewDefaultConfig()
	clt := config.NewCollate(make([]byte, 0, 1024))
	jsn := config.NewJson(json)

	ref := "xd>3\x00Zone\x00\x00P>>23-\x00Zthree\x00\x00P>>21-\x00Ztwo\x00\x00P>>22-\x00\x00"

	_, value := jsn.Tovalue()
	clt1 := config.NewValue(value).Tocollate(clt)
	if out := string(clt1.Bytes()); out != ref {
		t.Errorf("expected %s", ref)
		t.Errorf("got %s", out)
	}
	clt1 = jsn.Tocollate(clt.Reset(nil))
	if out := string(clt1.Bytes()); out != ref {
		t.Errorf("expected %s", ref)
		t.Errorf("got %s", out)
	}
}
