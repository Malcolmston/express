package otpauth

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// rfcSecretB32 is the ASCII secret "12345678901234567890" used by the RFC 4226
// and RFC 6238 test vectors, encoded as base32.
const rfcSecretB32 = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

// rfcSecret32B32 is "12345678901234567890123456789012" (32 bytes), the SHA256
// key from RFC 6238 Appendix B.
const rfcSecret32B32 = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZA"

// rfcSecret64B32 is the 64-byte SHA512 key from RFC 6238 Appendix B:
// "1234567890123456789012345678901234567890123456789012345678901234".
const rfcSecret64B32 = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQGEZDGNA"

// TestHOTPRFC4226Vectors checks GenerateCounter against every code in RFC 4226
// Appendix D.
func TestHOTPRFC4226Vectors(t *testing.T) {
	want := []string{
		"755224", "287082", "359152", "969429", "338314",
		"254676", "287922", "162583", "399871", "520489",
	}
	c := Config{Type: "hotp", Secret: rfcSecretB32, Digits: 6}
	for counter, w := range want {
		got, err := c.GenerateCounter(uint64(counter))
		if err != nil {
			t.Fatalf("counter %d: %v", counter, err)
		}
		if got != w {
			t.Errorf("counter %d: got %s, want %s", counter, got, w)
		}
	}
}

// TestTOTPRFC6238Vectors checks Generate against RFC 6238 Appendix B for all
// three hash algorithms. The RFC's vectors are 8 digits with a 30 second step.
func TestTOTPRFC6238Vectors(t *testing.T) {
	tests := []struct {
		unix int64
		alg  string
		key  string
		want string
	}{
		{59, "SHA1", rfcSecretB32, "94287082"},
		{59, "SHA256", rfcSecret32B32, "46119246"},
		{59, "SHA512", rfcSecret64B32, "90693936"},
		{1111111109, "SHA1", rfcSecretB32, "07081804"},
		{1111111111, "SHA1", rfcSecretB32, "14050471"},
		{1234567890, "SHA1", rfcSecretB32, "89005924"},
		{2000000000, "SHA1", rfcSecretB32, "69279037"},
		{20000000000, "SHA1", rfcSecretB32, "65353130"},
		{1111111109, "SHA256", rfcSecret32B32, "68084774"},
		{1111111109, "SHA512", rfcSecret64B32, "25091201"},
		{20000000000, "SHA256", rfcSecret32B32, "77737706"},
		{20000000000, "SHA512", rfcSecret64B32, "47863826"},
	}

	for _, tc := range tests {
		c := Config{Secret: tc.key, Algorithm: tc.alg, Digits: 8, Period: 30}
		got, err := c.Generate(time.Unix(tc.unix, 0))
		if err != nil {
			t.Fatalf("t=%d alg=%s: %v", tc.unix, tc.alg, err)
		}
		if got != tc.want {
			t.Errorf("t=%d alg=%s: got %s, want %s", tc.unix, tc.alg, got, tc.want)
		}
	}
}

func TestDefaultsMatchAuthenticatorApps(t *testing.T) {
	// Empty Type, Digits, Period and Algorithm mean TOTP / 6 / 30 / SHA1.
	bare := Config{Secret: rfcSecretB32}
	explicit := Config{Type: "totp", Secret: rfcSecretB32, Digits: 6, Period: 30, Algorithm: "SHA1"}

	at := time.Unix(1111111111, 0)
	a, err := bare.Generate(at)
	if err != nil {
		t.Fatalf("bare: %v", err)
	}
	b, err := explicit.Generate(at)
	if err != nil {
		t.Fatalf("explicit: %v", err)
	}
	if a != b {
		t.Fatalf("defaults diverge: %q vs %q", a, b)
	}
	if len(a) != DefaultDigits {
		t.Fatalf("default digits = %d, want %d", len(a), DefaultDigits)
	}
	// The RFC 6238 SHA1 vector for t=1111111111 is 14050471; a 6-digit code is
	// its last six digits.
	if a != "050471" {
		t.Fatalf("default TOTP = %s, want 050471", a)
	}
}

func TestDecodeSecret(t *testing.T) {
	ok := []string{
		"GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
		"gezdgnbvgy3tqojqgezdgnbvgy3tqojq",
		"GEZD GNBV GY3T QOJQ GEZD GNBV GY3T QOJQ",
		"JBSWY3DPEHPK3PXP",
		"MFRGG===",
	}
	for _, s := range ok {
		if _, err := DecodeSecret(s); err != nil {
			t.Errorf("DecodeSecret(%q): %v", s, err)
		}
	}
	bad := []string{"", "   ", "1", "not-base32!", "======"}
	for _, s := range bad {
		if _, err := DecodeSecret(s); !errors.Is(err, ErrInvalidSecret) {
			t.Errorf("DecodeSecret(%q): err = %v, want ErrInvalidSecret", s, err)
		}
	}
	// Whitespace and case must not change the decoded bytes.
	a, _ := DecodeSecret("JBSWY3DPEHPK3PXP")
	b, _ := DecodeSecret(" jbswy3dp ehpk3pxp ")
	if string(a) != string(b) {
		t.Errorf("case/whitespace changed the decoded secret")
	}
}

func TestGenerateErrors(t *testing.T) {
	now := time.Unix(0, 0)
	tests := []struct {
		name string
		c    Config
		want error
	}{
		{"empty secret", Config{}, ErrInvalidSecret},
		{"bad secret", Config{Secret: "!!!!"}, ErrInvalidSecret},
		{"bad algorithm", Config{Secret: rfcSecretB32, Algorithm: "MD5"}, ErrUnsupportedAlgorithm},
		{"bad type", Config{Secret: rfcSecretB32, Type: "xotp"}, ErrUnsupportedType},
	}
	for _, tc := range tests {
		if _, err := tc.c.Generate(now); !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
		if _, ok := tc.c.Validate("000000", now, 1); ok {
			t.Errorf("%s: Validate should fail closed on a bad config", tc.name)
		}
	}
}

func TestValidateWindow(t *testing.T) {
	c := Config{Secret: rfcSecretB32, Digits: 6, Period: 30}
	now := time.Unix(1_700_000_000, 0)

	tests := []struct {
		name      string
		codeAt    time.Time
		window    int
		wantDelta int
		wantOK    bool
	}{
		{"current step", now, 1, 0, true},
		{"previous step", now.Add(-30 * time.Second), 1, -1, true},
		{"next step", now.Add(30 * time.Second), 1, 1, true},
		{"two steps back, window 1", now.Add(-60 * time.Second), 1, 0, false},
		{"two steps back, window 2", now.Add(-60 * time.Second), 2, -2, true},
		{"zero window rejects neighbours", now.Add(-30 * time.Second), 0, 0, false},
		{"negative window behaves as zero", now, -5, 0, true},
	}

	for _, tc := range tests {
		code, err := c.Generate(tc.codeAt)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		delta, ok := c.Validate(code, now, tc.window)
		if ok != tc.wantOK || (ok && delta != tc.wantDelta) {
			t.Errorf("%s: Validate = (%d, %v), want (%d, %v)", tc.name, delta, ok, tc.wantDelta, tc.wantOK)
		}
	}
}

func TestValidateRejectsWrongCode(t *testing.T) {
	c := Config{Secret: rfcSecretB32}
	now := time.Unix(1_700_000_000, 0)
	valid, _ := c.Generate(now)
	for _, code := range []string{"", "12345", "1234567", "abcdef"} {
		if _, ok := c.Validate(code, now, 3); ok {
			t.Errorf("Validate accepted %q", code)
		}
	}
	// Sanity: the real code is still accepted.
	if _, ok := c.Validate(valid, now, 0); !ok {
		t.Error("Validate rejected the correct code")
	}
}

func TestHOTPValidateUsesConfiguredCounter(t *testing.T) {
	c := Config{Type: "hotp", Secret: rfcSecretB32, Counter: 7}
	// Time is irrelevant for HOTP.
	code, err := c.Generate(time.Unix(999999, 0))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	want, _ := c.GenerateCounter(7)
	if code != want {
		t.Fatalf("HOTP Generate = %s, want the code for counter 7 (%s)", code, want)
	}
	if delta, ok := c.Validate(code, time.Now(), 0); !ok || delta != 0 {
		t.Fatalf("Validate = (%d, %v), want (0, true)", delta, ok)
	}
	// A code for a later counter is found by the look-ahead window.
	ahead, _ := c.GenerateCounter(9)
	if delta, ok := c.Validate(ahead, time.Now(), 3); !ok || delta != 2 {
		t.Fatalf("Validate(ahead) = (%d, %v), want (2, true)", delta, ok)
	}
}

func TestDigitsPaddedAndDefaulted(t *testing.T) {
	base := Config{Secret: rfcSecretB32, Period: 30}
	at := time.Unix(59, 0)

	for _, d := range []int{0, -1} {
		c := base
		c.Digits = d
		got, _ := c.Generate(at)
		if len(got) != DefaultDigits {
			t.Errorf("Digits=%d produced %q, want %d digits", d, got, DefaultDigits)
		}
	}
	// Beyond MaxDigits the 31-bit truncation cannot add information, so the
	// code is simply zero padded on the left — the same as the npm library.
	c := base
	c.Digits = MaxDigits + 10
	got, _ := c.Generate(at)
	if len(got) != MaxDigits+10 {
		t.Errorf("Digits above MaxDigits produced %q (len %d), want %d", got, len(got), MaxDigits+10)
	}
	if !strings.HasPrefix(got, strings.Repeat("0", 10)) {
		t.Errorf("Digits above MaxDigits produced %q, want ten leading zeros", got)
	}
}

// TestParseThenGenerate is the end-to-end enrollment path: an app scans the
// URI, parses it, and computes the same code the server expects.
func TestParseThenGenerate(t *testing.T) {
	uri := URL(Config{
		Issuer:    "ACME Co",
		Account:   "alice@example.com",
		Secret:    rfcSecretB32,
		Digits:    8,
		Period:    30,
		Algorithm: "SHA1",
	})
	parsed, err := Parse(uri)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := parsed.Generate(time.Unix(59, 0))
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if got != "94287082" {
		t.Fatalf("round-tripped config produced %s, want the RFC 6238 vector 94287082", got)
	}
}
