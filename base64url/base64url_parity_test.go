package base64url

import (
	"bytes"
	"strings"
	"testing"
)

// Upstream parity tests for the npm package brianloveswords/base64url (v3.0.1).
// Vectors are transcribed from the library's own test suite and sources:
//   https://raw.githubusercontent.com/brianloveswords/base64url/master/test/base64url.test.js
//   https://raw.githubusercontent.com/brianloveswords/base64url/master/src/base64url.ts
//   https://raw.githubusercontent.com/brianloveswords/base64url/master/src/pad-string.ts
//
// The 'example from readme' test in base64url.test.js pins a concrete
// encode/decode vector; the property tests (from string to base64url, from
// base64url to base64, from base64 to base64url, from base64url to string) run
// over a binary fixture (test/test.jpg) and assert url-safety plus round-trip
// interconversion between standard base64 and base64url. Those properties are
// reproduced here over concrete binary inputs whose standard base64 contains
// both '+' and '/'.

// readmeOriginal / readmeEncoded are the exact values asserted by the upstream
// 'example from readme' test.
const (
	readmeOriginal = "ladies and gentlemen, we are floating in space"
	readmeEncoded  = "bGFkaWVzIGFuZCBnZW50bGVtZW4sIHdlIGFyZSBmbG9hdGluZyBpbiBzcGFjZQ"
	// standard base64 of readmeOriginal, i.e. base64url with '=' padding restored.
	readmeStdB64 = "bGFkaWVzIGFuZCBnZW50bGVtZW4sIHdlIGFyZSBmbG9hdGluZyBpbiBzcGFjZQ=="
)

// TestParityReadmeEncode mirrors upstream:
//
//	base64url.encode('ladies and gentlemen, we are floating in space')
//	  === 'bGFkaWVzIGFuZCBnZW50bGVtZW4sIHdlIGFyZSBmbG9hdGluZyBpbiBzcGFjZQ'
func TestParityReadmeEncode(t *testing.T) {
	got := EncodeString(readmeOriginal)
	if got != readmeEncoded {
		t.Errorf("EncodeString(%q) = %q, want %q", readmeOriginal, got, readmeEncoded)
	}
}

// TestParityReadmeDecode mirrors upstream:
//
//	base64url.decode('bGFkaWVzIGFuZCBnZW50bGVtZW4sIHdlIGFyZSBmbG9hdGluZyBpbiBzcGFjZQ')
//	  === 'ladies and gentlemen, we are floating in space'
func TestParityReadmeDecode(t *testing.T) {
	got, err := DecodeString(readmeEncoded)
	if err != nil {
		t.Fatalf("DecodeString(%q) error: %v", readmeEncoded, err)
	}
	if got != readmeOriginal {
		t.Errorf("DecodeString(%q) = %q, want %q", readmeEncoded, got, readmeOriginal)
	}
}

// TestParityFromBase64 mirrors the upstream 'from base64 to base64url' test:
// fromBase64(standardB64) strips '=' padding and swaps '+'->'-', '/'->'_'.
func TestParityFromBase64(t *testing.T) {
	got := FromBase64(readmeStdB64)
	if got != readmeEncoded {
		t.Errorf("FromBase64(%q) = %q, want %q", readmeStdB64, got, readmeEncoded)
	}
}

// TestParityToBase64 mirrors the upstream 'from base64url to base64' test:
// toBase64(base64url) pads to a multiple of 4 with '=' and swaps '-'->'+',
// '_'->'/'. For readmeEncoded that restores the two '=' pads.
func TestParityToBase64(t *testing.T) {
	got := ToBase64(readmeEncoded)
	if got != readmeStdB64 {
		t.Errorf("ToBase64(%q) = %q, want %q", readmeEncoded, got, readmeStdB64)
	}
}

// TestParityUrlSafetyAndInterconvert reproduces the upstream property tests
// ('from string to base64url', 'from base64url to base64', 'from base64 to
// base64url') over binary inputs. The upstream fixture's standard base64
// contains '+' and '/'; these inputs are chosen so it does too.
func TestParityUrlSafetyAndInterconvert(t *testing.T) {
	// std base64 (with padding) and its expected url-safe form, computed
	// independently from the raw bytes.
	cases := []struct {
		data []byte
		std  string
		url  string
	}{
		{
			data: []byte{0xfb, 0xff, 0xbf, 0x00, 0x10, 0x83, 0xfb, 0xef},
			std:  "+/+/ABCD++8=",
			url:  "-_-_ABCD--8",
		},
	}
	for _, c := range cases {
		url := Encode(c.data)
		if url != c.url {
			t.Errorf("Encode(% x) = %q, want %q", c.data, url, c.url)
		}
		// url-safe output must not contain '+', '/', or '='.
		if strings.ContainsAny(url, "+/=") {
			t.Errorf("Encode(% x) = %q contains +, / or =", c.data, url)
		}
		// The position of '+' in std matches the position of '-' in url, and
		// '/' matches '_' (the upstream indexOf assertions).
		if strings.IndexByte(c.std, '+') != strings.IndexByte(url, '-') {
			t.Errorf("index of + in std != index of - in url for %q / %q", c.std, url)
		}
		if strings.IndexByte(c.std, '/') != strings.IndexByte(url, '_') {
			t.Errorf("index of / in std != index of _ in url for %q / %q", c.std, url)
		}
		// fromBase64(std) == url
		if got := FromBase64(c.std); got != c.url {
			t.Errorf("FromBase64(%q) = %q, want %q", c.std, got, c.url)
		}
		// toBase64(url) == std
		if got := ToBase64(c.url); got != c.std {
			t.Errorf("ToBase64(%q) = %q, want %q", c.url, got, c.std)
		}
		// decode(url) round-trips to the original bytes ('from base64url to
		// buffer' / 'from base64url to string').
		dec, err := Decode(url)
		if err != nil {
			t.Fatalf("Decode(%q) error: %v", url, err)
		}
		if !bytes.Equal(dec, c.data) {
			t.Errorf("Decode(%q) = % x, want % x", url, dec, c.data)
		}
	}
}
