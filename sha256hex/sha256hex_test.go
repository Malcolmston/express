package sha256hex

import "testing"

func TestSHA256String(t *testing.T) {
	got := SHA256String("abc")
	want := "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestMD5String(t *testing.T) {
	got := MD5String("abc")
	want := "900150983cd24fb0d6963f7d28e17f72"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSHA1String(t *testing.T) {
	got := SHA1String("abc")
	want := "a9993e364706816aba3e25717850c26c9cd0d89d"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestHMACSHA256String(t *testing.T) {
	got := HMACSHA256String("key", "The quick brown fox jumps over the lazy dog")
	want := "f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
