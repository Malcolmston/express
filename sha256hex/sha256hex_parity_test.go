package sha256hex_test

// Parity vectors for the sha256hex port, drawn from the canonical NIST / RFC
// specifications and reference test suites (not invented). Sources:
//
//   - SHA-256 known-answer values: NIST FIPS 180-4 "Secure Hash Standard"
//     example messages ("abc" and the 448-bit / two-block strings), plus the
//     universally cited empty-string and one-million-"a" digests. These are the
//     same fixed values published in the NIST SHA test-vector documents.
//       sha256("")    = e3b0c442...b855
//       sha256("abc") = ba7816bf...15ad
//       sha256(56-byte "abcdb...nopq") = 248d6a61...06c1
//   - SHA-1 known-answer values: NIST FIPS 180-1/180-4 examples.
//   - MD5 known-answer values: RFC 1321 "The MD5 Message-Digest Algorithm",
//     Appendix A.5 test suite.
//   - HMAC-SHA256 known-answer values: RFC 4231 "Identifiers and Test Vectors
//     for HMAC-SHA-224, HMAC-SHA-256, HMAC-SHA-384, and HMAC-SHA-512",
//     Test Cases 1, 2, and 4.
//
// Every expected digest below was cross-checked against a reference
// implementation (Python hashlib/hmac) before being recorded here.

import (
	"strings"
	"testing"

	"github.com/malcolmston/express/sha256hex"
)

// TestParitySHA256NIST checks the SHA-256 helper against the NIST FIPS 180-4
// example messages and the classic empty / long inputs.
func TestParitySHA256NIST(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"abc", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{
			"fips-one-block",
			"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq",
			"248d6a61d20638b8e5c026930c3e6039a33ce45964ff2167f6ecedd419db06c1",
		},
		{
			"fips-two-block",
			"abcdefghbcdefghicdefghijdefghijkefghijklfghijklmghijklmnhijklmnoijklmnopjklmnopqklmnopqrlmnopqrsmnopqrstnopqrstu",
			"cf5b16a778af8380036ce59e7b0492370b249b11e8f07a51afac45037afee9d1",
		},
		{
			"one-million-a",
			strings.Repeat("a", 1000000),
			"cdc76e5c9914fb9281a1c7e284d73e67f1809a48a497200e046d39ccc7112cd0",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sha256hex.SHA256String(c.input); got != c.want {
				t.Errorf("SHA256String(%q) = %q, want %q", c.name, got, c.want)
			}
			if got := sha256hex.SHA256([]byte(c.input)); got != c.want {
				t.Errorf("SHA256(%q) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}

// TestParitySHA1NIST checks the SHA-1 helper against the NIST FIPS 180 examples.
func TestParitySHA1NIST(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", "da39a3ee5e6b4b0d3255bfef95601890afd80709"},
		{"abc", "abc", "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{
			"fips-one-block",
			"abcdbcdecdefdefgefghfghighijhijkijkljklmklmnlmnomnopnopq",
			"84983e441c3bd26ebaae4aa1f95129e5e54670f1",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sha256hex.SHA1String(c.input); got != c.want {
				t.Errorf("SHA1String(%q) = %q, want %q", c.name, got, c.want)
			}
		})
	}
}

// TestParityMD5RFC1321 checks the MD5 helper against the RFC 1321 Appendix A.5
// test suite.
func TestParityMD5RFC1321(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"a", "0cc175b9c0f1b6a831c399e269772661"},
		{"abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"message digest", "f96b697d7cb7938d525a2f31aaf161d0"},
		{"abcdefghijklmnopqrstuvwxyz", "c3fcd3d76192e4007dfb496cca67e13b"},
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789", "d174ab98d277d9f5a5611c2c9f419d9f"},
		{"12345678901234567890123456789012345678901234567890123456789012345678901234567890", "57edf4a22be3c955ac49da2e2107b67a"},
	}
	for _, c := range cases {
		if got := sha256hex.MD5String(c.input); got != c.want {
			t.Errorf("MD5String(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

// TestParityHMACSHA256RFC4231 checks the HMAC-SHA256 helper against RFC 4231
// Test Cases 1, 2, and 4.
func TestParityHMACSHA256RFC4231(t *testing.T) {
	cases := []struct {
		name string
		key  []byte
		data []byte
		want string
	}{
		{
			"tc1",
			repeatByte(0x0b, 20),
			[]byte("Hi There"),
			"b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7",
		},
		{
			"tc2",
			[]byte("Jefe"),
			[]byte("what do ya want for nothing?"),
			"5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843",
		},
		{
			"tc4",
			seqBytes(0x01, 25),
			repeatByte(0xcd, 50),
			"82558a389a443c0ea4cc819899f2083a85f0faa3e578f8077a2e3ff46729665b",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sha256hex.HMACSHA256(c.key, c.data); got != c.want {
				t.Errorf("HMACSHA256(%s) = %q, want %q", c.name, got, c.want)
			}
		})
	}

	// String variant against the js-sha256 documented "key" example.
	if got := sha256hex.HMACSHA256String("key", "The quick brown fox jumps over the lazy dog"); got != "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8" {
		t.Errorf("HMACSHA256String key/fox = %q", got)
	}
}

func repeatByte(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

func seqBytes(start byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = start + byte(i)
	}
	return out
}
