package str

import (
	"reflect"
	"testing"
)

func TestToLowerUpper(t *testing.T) {
	if ToLower("FooBar") != "foobar" {
		t.Error("ToLower")
	}
	if ToUpper("FooBar") != "FOOBAR" {
		t.Error("ToUpper")
	}
}

func TestSplit(t *testing.T) {
	if got := Split("a-b-c", "-", -1); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("Split all = %v", got)
	}
	if got := Split("a-b-c", "-", 2); !reflect.DeepEqual(got, []string{"a", "b-c"}) {
		t.Errorf("Split limit = %v", got)
	}
	if got := Split("a-b-c", "-", 0); len(got) != 0 {
		t.Errorf("Split zero = %v", got)
	}
}

func TestChars(t *testing.T) {
	if got := Chars("a€b"); !reflect.DeepEqual(got, []string{"a", "€", "b"}) {
		t.Errorf("Chars = %v", got)
	}
	if got := Chars(""); len(got) != 0 {
		t.Errorf("Chars empty = %v", got)
	}
}
