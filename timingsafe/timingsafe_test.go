package timingsafe

import "testing"

func TestEqual(t *testing.T) {
	if !Equal([]byte("secret"), []byte("secret")) {
		t.Error("expected equal")
	}
	if Equal([]byte("secret"), []byte("secreT")) {
		t.Error("expected not equal")
	}
}

func TestDifferentLength(t *testing.T) {
	if Equal([]byte("short"), []byte("longervalue")) {
		t.Error("expected not equal for different lengths")
	}
	if Equal([]byte(""), []byte("x")) {
		t.Error("expected not equal for different lengths")
	}
}

func TestEqualString(t *testing.T) {
	if !EqualString("abc", "abc") {
		t.Error("expected equal")
	}
	if EqualString("abc", "abd") {
		t.Error("expected not equal")
	}
	if EqualString("abc", "abcd") {
		t.Error("expected not equal")
	}
}
