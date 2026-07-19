package base32_test

// Upstream parity tests for the npm library emn178/hi-base32.
//
// Vectors below are copied verbatim from the library's own test suite:
//   https://raw.githubusercontent.com/emn178/hi-base32/master/tests/test.js
//   (driver: https://raw.githubusercontent.com/emn178/hi-base32/master/tests/node-test.js)
//
// In the upstream tests base32.encode(str, true) encodes the raw byte values of
// an ASCII string, base32.encode(str) encodes the UTF-8 bytes of a string, and
// base32.decode(str) returns the decoded UTF-8 string. The Go port operates on
// []byte, so ASCII/UTF-8 inputs are represented as []byte(literal) (Go string
// literals are already UTF-8) and decode results are compared as bytes.

import (
	"bytes"
	"testing"

	"github.com/malcolmston/express/base32"
)

// ASCII encode vectors: base32.encode(str, true).
var parityASCII = []struct {
	in   string
	want string
}{
	{"", ""},
	{"H", "JA======"},
	{"He", "JBSQ===="},
	{"Hel", "JBSWY==="},
	{"Hell", "JBSWY3A="},
	{"Hello", "JBSWY3DP"},
	{"Man is distinguished, not only by his reason, but by this singular passion from other animals, which is a lust of the mind, that by a perseverance of delight in the continued and indefatigable generation of knowledge, exceeds the short vehemence of any carnal pleasure.", "JVQW4IDJOMQGI2LTORUW4Z3VNFZWQZLEFQQG433UEBXW43DZEBRHSIDINFZSA4TFMFZW63RMEBRHK5BAMJ4SA5DINFZSA43JNZTXK3DBOIQHAYLTONUW63RAMZZG63JAN52GQZLSEBQW42LNMFWHGLBAO5UGSY3IEBUXGIDBEBWHK43UEBXWMIDUNBSSA3LJNZSCYIDUNBQXIIDCPEQGCIDQMVZHGZLWMVZGC3TDMUQG6ZRAMRSWY2LHNB2CA2LOEB2GQZJAMNXW45DJNZ2WKZBAMFXGIIDJNZSGKZTBORUWOYLCNRSSAZ3FNZSXEYLUNFXW4IDPMYQGW3TPO5WGKZDHMUWCAZLYMNSWKZDTEB2GQZJAONUG64TUEB3GK2DFNVSW4Y3FEBXWMIDBNZ4SAY3BOJXGC3BAOBWGKYLTOVZGKLQ="},
	{"Base64 is a group of similar binary-to-text encoding schemes that represent binary data in an ASCII string format by translating it into a radix-64 representation.", "IJQXGZJWGQQGS4ZAMEQGO4TPOVYCA33GEBZWS3LJNRQXEIDCNFXGC4TZFV2G6LLUMV4HIIDFNZRW6ZDJNZTSA43DNBSW2ZLTEB2GQYLUEBZGK4DSMVZWK3TUEBRGS3TBOJ4SAZDBORQSA2LOEBQW4ICBKNBUSSJAON2HE2LOM4QGM33SNVQXIIDCPEQHI4TBNZZWYYLUNFXGOIDJOQQGS3TUN4QGCIDSMFSGS6BNGY2CA4TFOBZGK43FNZ2GC5DJN5XC4==="},
}

// UTF-8 encode vectors: base32.encode(str).
var parityUTF8 = []struct {
	in   string
	want string
}{
	{"", ""},
	{"中文", "4S4K3ZUWQ4======"},
	{"中文1", "4S4K3ZUWQ4YQ===="},
	{"中文12", "4S4K3ZUWQ4YTE==="},
	{"aécio", "MHB2SY3JN4======"},
	{"𠜎", "6CQJZDQ="},
	{"Base64是一種基於64個可列印字元來表示二進制資料的表示方法", "IJQXGZJWGTTJRL7EXCAOPKFO4WP3VZUWXQ3DJZMARPSY7L7FRCL6LDNQ4WWZPZMFQPSL5BXIUGUOPJF24S5IZ2MAWLSYRNXIWOD6NFUZ46NIJ2FBVDT2JOXGS246NM4V"},
}

// Byte-array encode vectors: base32.encode([...]).
var parityBytes = []struct {
	in   []byte
	want string
}{
	{[]byte{72}, "JA======"},
	{[]byte{72, 101}, "JBSQ===="},
	{[]byte{72, 101, 108}, "JBSWY==="},
	{[]byte{72, 101, 108, 108}, "JBSWY3A="},
	{[]byte{72, 101, 108, 108, 111}, "JBSWY3DP"},
	{[]byte{0}, "AA======"}, // new ArrayBuffer(1)
}

func TestParityEncodeASCII(t *testing.T) {
	for _, c := range parityASCII {
		if got := base32.Encode([]byte(c.in)); got != c.want {
			t.Errorf("Encode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParityEncodeUTF8(t *testing.T) {
	for _, c := range parityUTF8 {
		if got := base32.Encode([]byte(c.in)); got != c.want {
			t.Errorf("Encode(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParityEncodeBytes(t *testing.T) {
	for _, c := range parityBytes {
		if got := base32.Encode(c.in); got != c.want {
			t.Errorf("Encode(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestParityDecodeASCII(t *testing.T) {
	for _, c := range parityASCII {
		got, err := base32.Decode(c.want)
		if err != nil {
			t.Errorf("Decode(%q) error: %v", c.want, err)
			continue
		}
		if !bytes.Equal(got, []byte(c.in)) {
			t.Errorf("Decode(%q) = %q, want %q", c.want, got, c.in)
		}
	}
}

func TestParityDecodeUTF8(t *testing.T) {
	for _, c := range parityUTF8 {
		got, err := base32.Decode(c.want)
		if err != nil {
			t.Errorf("Decode(%q) error: %v", c.want, err)
			continue
		}
		if !bytes.Equal(got, []byte(c.in)) {
			t.Errorf("Decode(%q) = %q, want %q", c.want, got, c.in)
		}
	}
}

// Upstream treats "1 ======" as an invalid string (it contains characters
// outside the base32 alphabet) and expects decoding to throw. The Go port
// reports this as a non-nil error.
func TestParityDecodeInvalid(t *testing.T) {
	if _, err := base32.Decode("1 ======"); err == nil {
		t.Errorf("Decode(%q) = nil error, want non-nil", "1 ======")
	}
}
