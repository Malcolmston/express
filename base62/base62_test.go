package base62

import (
	"bytes"
	"testing"
)

func TestEncodeInt(t *testing.T) {
	tests := []struct {
		n    uint64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{61, "z"},
		{62, "10"},
		{3844, "100"},
		{12345, "3D7"},
	}
	for _, tt := range tests {
		if got := EncodeInt(tt.n); got != tt.want {
			t.Errorf("EncodeInt(%d) = %q, want %q", tt.n, got, tt.want)
		}
		back, err := DecodeInt(tt.want)
		if err != nil || back != tt.n {
			t.Errorf("DecodeInt(%q) = %d,%v want %d", tt.want, back, err, tt.n)
		}
	}
}

func TestEncodeDecodeBytes(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello"),
		{0, 0, 1, 2, 3},
		{255, 254},
		{},
	}
	for _, in := range inputs {
		enc := Encode(in)
		dec, err := Decode(enc)
		if err != nil {
			t.Fatalf("Decode(%q): %v", enc, err)
		}
		if !bytes.Equal(dec, in) {
			t.Errorf("round trip %x -> %q -> %x", in, enc, dec)
		}
	}
}

func TestDecodeInvalid(t *testing.T) {
	if _, err := Decode("abc!"); err != ErrInvalidCharacter {
		t.Errorf("expected ErrInvalidCharacter, got %v", err)
	}
	if _, err := DecodeInt("!!"); err != ErrInvalidCharacter {
		t.Errorf("expected ErrInvalidCharacter, got %v", err)
	}
}

func BenchmarkEncodeInt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = EncodeInt(9999999999)
	}
}
