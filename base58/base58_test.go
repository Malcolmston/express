package base58

import (
	"bytes"
	"testing"
)

func TestEncodeKnown(t *testing.T) {
	tests := []struct {
		in   []byte
		want string
	}{
		{[]byte("hello world"), "StV1DL6CwTryKyV"},
		{[]byte{0x00, 0x00, 0x28, 0x7f, 0xb4, 0xcd}, "11233QC4"},
		{[]byte{}, ""},
		{[]byte{0}, "1"},
	}
	for _, tt := range tests {
		if got := Encode(tt.in); got != tt.want {
			t.Errorf("Encode(%x) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDecodeRoundTrip(t *testing.T) {
	inputs := [][]byte{
		[]byte("hello world"),
		{0x00, 0x00, 0x28, 0x7f, 0xb4, 0xcd},
		{0, 0, 0, 1, 2, 3},
		{255, 254, 253},
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
	if _, err := Decode("0OIl"); err != ErrInvalidCharacter {
		t.Errorf("expected ErrInvalidCharacter, got %v", err)
	}
}

func TestCheckEncodeDecode(t *testing.T) {
	payload := []byte{0xde, 0xad, 0xbe, 0xef}
	enc := CheckEncode(payload, 0x00)
	version, got, err := CheckDecode(enc)
	if err != nil {
		t.Fatalf("CheckDecode: %v", err)
	}
	if version != 0x00 {
		t.Errorf("version = %d", version)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("payload = %x, want %x", got, payload)
	}
}

func TestCheckDecodeBadChecksum(t *testing.T) {
	enc := CheckEncode([]byte{1, 2, 3}, 0x05)
	// corrupt the last character
	bad := enc[:len(enc)-1] + string(swapChar(enc[len(enc)-1]))
	if _, _, err := CheckDecode(bad); err == nil {
		t.Error("expected checksum error")
	}
	if _, _, err := CheckDecode("1"); err != ErrTooShort {
		t.Errorf("expected ErrTooShort, got %v", err)
	}
}

func swapChar(c byte) byte {
	if c == alphabet[0] {
		return alphabet[1]
	}
	return alphabet[0]
}

func BenchmarkEncode(b *testing.B) {
	data := []byte("the quick brown fox jumps over the lazy dog")
	for i := 0; i < b.N; i++ {
		_ = Encode(data)
	}
}
