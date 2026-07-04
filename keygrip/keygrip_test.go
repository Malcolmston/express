package keygrip

import "testing"

func TestNewEmptyPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for empty keys")
		}
	}()
	New(nil)
}

func TestSignVerify(t *testing.T) {
	kg := New([]string{"SEKRIT1"})
	sig := kg.Sign("hello")
	if !kg.Verify("hello", sig) {
		t.Fatal("expected verify to succeed")
	}
	if kg.Verify("world", sig) {
		t.Fatal("expected verify to fail for different data")
	}
}

func TestSignUsesFirstKey(t *testing.T) {
	kg := New([]string{"a", "b"})
	if kg.Sign("data") != sign("data", "a") {
		t.Fatal("Sign should use first key")
	}
}

func TestIndexRotation(t *testing.T) {
	old := New([]string{"SEKRIT2"})
	sig := old.Sign("hello")

	// New keygrip rotated the new key to the front, old key kept for verify.
	kg := New([]string{"SEKRIT3", "SEKRIT2"})
	if idx := kg.Index("hello", sig); idx != 1 {
		t.Fatalf("expected index 1, got %d", idx)
	}
	if kg.Index("hello", kg.Sign("hello")) != 0 {
		t.Fatal("expected index 0 for current key")
	}
}

func TestIndexNotFound(t *testing.T) {
	kg := New([]string{"a", "b"})
	if kg.Index("hello", "bogus") != -1 {
		t.Fatal("expected -1 for bogus digest")
	}
}

func TestVerifyFalse(t *testing.T) {
	kg := New([]string{"a"})
	if kg.Verify("hello", "nope") {
		t.Fatal("expected false")
	}
}

func TestDigestNoPadding(t *testing.T) {
	kg := New([]string{"key"})
	sig := kg.Sign("hello")
	for _, c := range sig {
		if c == '=' || c == '+' || c == '/' {
			t.Fatalf("digest must be url-safe unpadded, got %q", sig)
		}
	}
}

func TestKeysCopied(t *testing.T) {
	keys := []string{"a", "b"}
	kg := New(keys)
	keys[0] = "mutated"
	if kg.Sign("x") != sign("x", "a") {
		t.Fatal("Keygrip should not be affected by caller mutation")
	}
}
