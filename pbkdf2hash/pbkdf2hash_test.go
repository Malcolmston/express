package pbkdf2hash

import (
	"encoding/hex"
	"testing"
)

func TestDeriveKeyVector(t *testing.T) {
	got := hex.EncodeToString(DeriveKey([]byte("password"), []byte("salt"), 1, 32))
	want := "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHashVerifyRoundTrip(t *testing.T) {
	encoded, err := Hash("s3cret", 1000)
	if err != nil {
		t.Fatal(err)
	}
	if !Verify("s3cret", encoded) {
		t.Error("expected verify to succeed for correct password")
	}
	if Verify("wrong", encoded) {
		t.Error("expected verify to fail for wrong password")
	}
}

func TestVerifyBadEncoding(t *testing.T) {
	if Verify("x", "not-a-valid-hash") {
		t.Error("expected verify to fail for malformed encoding")
	}
}
