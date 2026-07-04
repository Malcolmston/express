package totp

import (
	"testing"
	"time"
)

const rfcSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestGenerateAtRFC6238(t *testing.T) {
	opts := &Options{Digits: 8, Period: 30, Algorithm: "SHA1"}
	cases := []struct {
		unix int64
		want string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
	}
	for _, c := range cases {
		got, err := GenerateAt(rfcSecret, time.Unix(c.unix, 0).UTC(), opts)
		if err != nil {
			t.Fatalf("unix %d: unexpected error: %v", c.unix, err)
		}
		if got != c.want {
			t.Errorf("unix %d: got %q, want %q", c.unix, got, c.want)
		}
	}
}

func TestGenerateUsesTimeNow(t *testing.T) {
	old := timeNow
	defer func() { timeNow = old }()
	timeNow = func() time.Time { return time.Unix(59, 0).UTC() }

	opts := &Options{Digits: 8, Period: 30, Algorithm: "SHA1"}
	got, err := Generate(rfcSecret, opts)
	if err != nil {
		t.Fatal(err)
	}
	if got != "94287082" {
		t.Errorf("got %q, want %q", got, "94287082")
	}
}

func TestVerifyWindow(t *testing.T) {
	old := timeNow
	defer func() { timeNow = old }()
	timeNow = func() time.Time { return time.Unix(1111111111, 0).UTC() }

	opts := &Options{Digits: 8, Period: 30, Algorithm: "SHA1"}
	if !Verify(rfcSecret, "14050471", opts, 1) {
		t.Error("expected verify to succeed for current step")
	}
	if Verify(rfcSecret, "00000000", opts, 1) {
		t.Error("expected verify to fail for wrong code")
	}
}

func TestDecodeLowercaseNoPadding(t *testing.T) {
	// Default options (SHA1, 6 digits, period 30) should decode a lowercase secret.
	_, err := GenerateAt("gezdgnbvgy3tqojqgezdgnbvgy3tqojq", time.Unix(59, 0).UTC(), nil)
	if err != nil {
		t.Fatalf("unexpected error decoding lowercase secret: %v", err)
	}
}
