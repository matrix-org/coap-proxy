package gson

import "testing"

func TestJsonEmpty(t *testing.T) {
	config := NewDefaultConfig()
	cbr := config.NewCbor(make([]byte, 0, 128))
	jsn := config.NewJson(make([]byte, 0, 128))
	clt := config.NewCollate(nil)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Tovalue()
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Tocbor(cbr)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		jsn.Tocollate(clt)
	}()
}
