package base32

import (
	"bytes"
	"testing"
)

func TestKnownVector(t *testing.T) {
	got := Encode([]byte("foobar"))
	want := "MZXW6YTBOI======"
	if got != want {
		t.Errorf("Encode(foobar) = %q want %q", got, want)
	}
}

func TestNoPadding(t *testing.T) {
	got := EncodeNoPadding([]byte("foobar"))
	want := "MZXW6YTBOI"
	if got != want {
		t.Errorf("EncodeNoPadding(foobar) = %q want %q", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	inputs := []string{"", "f", "fo", "foo", "foob", "fooba", "foobar"}
	for _, in := range inputs {
		enc := Encode([]byte(in))
		dec, err := Decode(enc)
		if err != nil {
			t.Fatalf("decode %q: %v", enc, err)
		}
		if !bytes.Equal(dec, []byte(in)) {
			t.Errorf("round trip: got %q want %q", dec, in)
		}
	}
}

func TestDecodeCaseInsensitiveAndNoPad(t *testing.T) {
	got, err := Decode("mzxw6ytboi")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("foobar")) {
		t.Errorf("got %q want foobar", got)
	}
}
