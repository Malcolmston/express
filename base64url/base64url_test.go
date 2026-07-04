package base64url

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	inputs := []string{"", "f", "fo", "foo", "foob", "fooba", "foobar", "hello world?>"}
	for _, in := range inputs {
		enc := EncodeString(in)
		if bytes.ContainsAny([]byte(enc), "=+/") {
			t.Errorf("encoding %q contains padding or non-url chars: %q", in, enc)
		}
		dec, err := DecodeString(enc)
		if err != nil {
			t.Fatalf("decode %q: %v", enc, err)
		}
		if dec != in {
			t.Errorf("round trip failed: got %q want %q", dec, in)
		}
	}
}

func TestEncodeBytes(t *testing.T) {
	data := []byte{0xfb, 0xff, 0xbf}
	enc := Encode(data)
	dec, err := Decode(enc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(dec, data) {
		t.Errorf("got %v want %v", dec, data)
	}
}

func TestFromToBase64(t *testing.T) {
	data := []byte{0xfb, 0xff, 0xbf, 0x00, 0x11}
	std := base64.StdEncoding.EncodeToString(data)
	url := FromBase64(std)
	if url != Encode(data) {
		t.Errorf("FromBase64 = %q want %q", url, Encode(data))
	}
	back := ToBase64(url)
	if back != std {
		t.Errorf("ToBase64 = %q want %q", back, std)
	}
	decoded, err := base64.StdEncoding.DecodeString(back)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decoded, data) {
		t.Errorf("std decode got %v want %v", decoded, data)
	}
}

func TestDecodeTolerateMissingPadding(t *testing.T) {
	// "foob" -> base64url with padding would be "Zm9vYg==" -> raw "Zm9vYg"
	got, err := DecodeString("Zm9vYg==")
	if err != nil {
		t.Fatal(err)
	}
	if got != "foob" {
		t.Errorf("got %q want foob", got)
	}
}
