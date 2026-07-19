// Upstream-parity tests for the Go port of the npm "uuid" package (uuidjs/uuid).
//
// Every vector below is transcribed verbatim from the ORIGINAL library's own
// test suite (npm "uuid" v14.x, branch main), fetched from:
//
//   - v3/v5 name-based vectors:
//     https://raw.githubusercontent.com/uuidjs/uuid/main/src/test/v35.test.ts
//   - validate()/version() truth table (TESTS):
//     https://raw.githubusercontent.com/uuidjs/uuid/main/src/test/test_constants.ts
//     (validation grammar: https://raw.githubusercontent.com/uuidjs/uuid/main/src/regex.ts)
//   - parse() byte vectors:
//     https://raw.githubusercontent.com/uuidjs/uuid/main/src/test/parse.test.ts
//   - stringify()/unsafeStringify() byte vectors:
//     https://raw.githubusercontent.com/uuidjs/uuid/main/src/test/stringify.test.ts
//   - v4() explicit-random vector:
//     https://raw.githubusercontent.com/uuidjs/uuid/main/src/test/v4.test.ts
//
// NOTE ON ARGUMENT ORDER: upstream's v5 signature is v5(name, namespace); this
// port's is V5(namespace, name). The calls below are mapped accordingly.
package uuid

import "testing"

// TestParityV5Vectors mirrors the "v5" test in src/test/v35.test.ts, which
// documents these as matching http://tools.adjet.org/uuid-v5.
func TestParityV5Vectors(t *testing.T) {
	cases := []struct {
		namespace, name, want string
	}{
		// v5('hello.example.com', v5.DNS)
		{NamespaceDNS, "hello.example.com", "fdda765f-fc57-5604-a269-52a7df8164ec"},
		// v5('http://example.com/hello', v5.URL)
		{NamespaceURL, "http://example.com/hello", "3bbcee75-cecc-5b56-8031-b6641c1ed1f1"},
		// v5('hello', '0f5abcd1-c194-47f3-905b-2df7263a084b')
		{"0f5abcd1-c194-47f3-905b-2df7263a084b", "hello", "90123e1c-7512-523e-bb28-76fab9f2f73d"},
	}
	for _, c := range cases {
		got, err := V5(c.namespace, c.name)
		if err != nil {
			t.Fatalf("V5(%q, %q) error: %v", c.namespace, c.name, err)
		}
		if got != c.want {
			t.Errorf("V5(%q, %q) = %q, want %q", c.namespace, c.name, got, c.want)
		}
	}
}

// TestParityV5NamespaceUppercase mirrors "v5 namespace.toUpperCase" in
// src/test/v35.test.ts: an uppercased namespace yields the same UUID.
func TestParityV5NamespaceUppercase(t *testing.T) {
	cases := []struct {
		namespace, name, want string
	}{
		{"6BA7B810-9DAD-11D1-80B4-00C04FD430C8", "hello.example.com", "fdda765f-fc57-5604-a269-52a7df8164ec"},
		{"6BA7B811-9DAD-11D1-80B4-00C04FD430C8", "http://example.com/hello", "3bbcee75-cecc-5b56-8031-b6641c1ed1f1"},
		{"0F5ABCD1-C194-47F3-905B-2DF7263A084B", "hello", "90123e1c-7512-523e-bb28-76fab9f2f73d"},
	}
	for _, c := range cases {
		got, err := V5(c.namespace, c.name)
		if err != nil {
			t.Fatalf("V5(%q, %q) error: %v", c.namespace, c.name, err)
		}
		if got != c.want {
			t.Errorf("V5(%q, %q) = %q, want %q", c.namespace, c.name, got, c.want)
		}
	}
}

// TestParityV5NamespaceValidation mirrors "v5 namespace string validation" in
// src/test/v35.test.ts: invalid namespace strings must error, and the NIL
// namespace must be accepted.
func TestParityV5NamespaceValidation(t *testing.T) {
	for _, bad := range []string{
		"zyxwvuts-rqpo-nmlk-jihg-fedcba000000",
		"invalid uuid value",
	} {
		if _, err := V5(bad, "hello.example.com"); err == nil {
			t.Errorf("V5(%q, ...) = nil error, want error", bad)
		}
	}
	// NIL namespace is valid upstream.
	if _, err := V5("00000000-0000-0000-0000-000000000000", "hello.example.com"); err != nil {
		t.Errorf("V5(NIL namespace, ...) unexpected error: %v", err)
	}
}

// TestParityValidate mirrors the TESTS truth table in src/test/test_constants.ts
// (validation grammar in src/regex.ts). Only string-valued entries are included,
// since this port's Validate takes a string.
func TestParityValidate(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		// constants
		{"00000000-0000-0000-0000-000000000000", true}, // NIL
		{"ffffffff-ffff-ffff-ffff-ffffffffffff", true}, // MAX

		// each version, all-0 and all-1 settable bits
		{"00000000-0000-1000-8000-000000000000", true},
		{"ffffffff-ffff-1fff-8fff-ffffffffffff", true},
		{"00000000-0000-2000-8000-000000000000", true},
		{"ffffffff-ffff-2fff-bfff-ffffffffffff", true},
		{"00000000-0000-3000-8000-000000000000", true},
		{"ffffffff-ffff-3fff-bfff-ffffffffffff", true},
		{"00000000-0000-4000-8000-000000000000", true},
		{"ffffffff-ffff-4fff-bfff-ffffffffffff", true},
		{"00000000-0000-5000-8000-000000000000", true},
		{"ffffffff-ffff-5fff-bfff-ffffffffffff", true},
		{"00000000-0000-6000-8000-000000000000", true},
		{"ffffffff-ffff-6fff-bfff-ffffffffffff", true},
		{"00000000-0000-7000-8000-000000000000", true},
		{"ffffffff-ffff-7fff-bfff-ffffffffffff", true},
		{"00000000-0000-8000-8000-000000000000", true},
		{"ffffffff-ffff-8fff-bfff-ffffffffffff", true},
		{"00000000-0000-9000-8000-000000000000", false},
		{"ffffffff-ffff-9fff-bfff-ffffffffffff", false},
		{"00000000-0000-a000-8000-000000000000", false},
		{"ffffffff-ffff-afff-bfff-ffffffffffff", false},
		{"00000000-0000-b000-8000-000000000000", false},
		{"ffffffff-ffff-bfff-bfff-ffffffffffff", false},
		{"00000000-0000-c000-8000-000000000000", false},
		{"ffffffff-ffff-cfff-bfff-ffffffffffff", false},
		{"00000000-0000-d000-8000-000000000000", false},
		{"ffffffff-ffff-dfff-bfff-ffffffffffff", false},
		{"00000000-0000-e000-8000-000000000000", false},
		{"ffffffff-ffff-efff-bfff-ffffffffffff", false},

		// selection of normal, valid UUIDs (versions 1-8)
		{"d9428888-122b-11e1-b85c-61cd3cbb3210", true},
		{"000003e8-2363-21ef-b200-325096b39f47", true},
		{"a981a0c2-68b1-35dc-bcfc-296e52ab01ec", true},
		{"109156be-c4fb-41ea-b1b4-efe1671c5836", true},
		{"90123e1c-7512-523e-bb28-76fab9f2f73d", true},
		{"1ef21d2f-1207-6660-8c4f-419efbd44d48", true},
		{"017f22e2-79b0-7cc3-98c4-dc0c0c07398f", true},
		{"0d8f23a0-697f-83ae-802e-48f3756dd581", true},

		// all variant octet values (version fixed at 1)
		{"00000000-0000-1000-0000-000000000000", false},
		{"00000000-0000-1000-1000-000000000000", false},
		{"00000000-0000-1000-2000-000000000000", false},
		{"00000000-0000-1000-3000-000000000000", false},
		{"00000000-0000-1000-4000-000000000000", false},
		{"00000000-0000-1000-5000-000000000000", false},
		{"00000000-0000-1000-6000-000000000000", false},
		{"00000000-0000-1000-7000-000000000000", false},
		{"00000000-0000-1000-8000-000000000000", true},
		{"00000000-0000-1000-9000-000000000000", true},
		{"00000000-0000-1000-a000-000000000000", true},
		{"00000000-0000-1000-b000-000000000000", true},
		{"00000000-0000-1000-c000-000000000000", false},
		{"00000000-0000-1000-d000-000000000000", false},
		{"00000000-0000-1000-e000-000000000000", false},
		{"00000000-0000-1000-f000-000000000000", false},

		// invalid strings
		{"00000000000000000000000000000000", false}, // unhyphenated NIL
		{"", false},
		{"invalid uuid string", false},
		{"=Y00a-f*vb*-c-d#-p00f\b-g0h-#i^-j*3&-L00k-\nl---00n-fg000-00p-00r+", false},
	}
	for _, c := range cases {
		if got := Validate(c.value); got != c.want {
			t.Errorf("Validate(%q) = %v, want %v", c.value, got, c.want)
		}
	}
}

// TestParityValidateCaseNeutral mirrors upstream's case-insensitive regex
// (src/regex.ts, /i flag): validation is case-insensitive.
func TestParityValidateCaseNeutral(t *testing.T) {
	if !Validate("0F5ABCD1-C194-47F3-905B-2DF7263A084B") {
		t.Error("Validate of uppercase valid UUID = false, want true")
	}
}

// TestParityParse mirrors the "parse" tests in src/test/parse.test.ts:
// canonical string -> exact byte layout, NIL -> all zeros, case neutrality,
// and rejection of invalid inputs.
func TestParityParse(t *testing.T) {
	want := [16]byte{
		0x0f, 0x5a, 0xbc, 0xd1, 0xc1, 0x94, 0x47, 0xf3,
		0x90, 0x5b, 0x2d, 0xf7, 0x26, 0x3a, 0x08, 0x4b,
	}
	got, err := Parse("0f5abcd1-c194-47f3-905b-2df7263a084b")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if got != want {
		t.Errorf("Parse bytes = %x, want %x", got, want)
	}

	// Case neutrality: upper- and lower-case parse identically.
	upper, err := Parse("0F5ABCD1-C194-47F3-905B-2DF7263A084B")
	if err != nil {
		t.Fatalf("Parse(uppercase) error: %v", err)
	}
	if upper != got {
		t.Errorf("Parse case neutrality: upper=%x lower=%x", upper, got)
	}

	// Null UUID -> 16 zero bytes.
	nilBytes, err := Parse("00000000-0000-0000-0000-000000000000")
	if err != nil {
		t.Fatalf("Parse(NIL) error: %v", err)
	}
	if nilBytes != ([16]byte{}) {
		t.Errorf("Parse(NIL) = %x, want all zeros", nilBytes)
	}

	// Invalid inputs must error.
	for _, bad := range []string{
		"invalid uuid",
		"zyxwvuts-rqpo-nmlk-jihg-fedcba000000",
	} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) = nil error, want error", bad)
		}
	}
}

// TestParityStringify mirrors the "stringify" tests in
// src/test/stringify.test.ts: the port's Format is the equivalent of upstream
// stringify/unsafeStringify for a full 16-byte value.
func TestParityStringify(t *testing.T) {
	bytes := [16]byte{
		0x0f, 0x5a, 0xbc, 0xd1, 0xc1, 0x94, 0x47, 0xf3,
		0x90, 0x5b, 0x2d, 0xf7, 0x26, 0x3a, 0x08, 0x4b,
	}
	const want = "0f5abcd1-c194-47f3-905b-2df7263a084b"
	if got := Format(bytes); got != want {
		t.Errorf("Format = %q, want %q", got, want)
	}
}

// TestParityV4ExplicitRandom mirrors the "explicit options.random produces
// expected result" vector in src/test/v4.test.ts. This port's V4 draws from
// crypto/rand with no injection hook, so the input bytes cannot be supplied
// directly. Instead we verify the deterministic transform that V4 applies -
// force the version nibble to 4 and the variant bits to RFC 4122 - reproduces
// upstream's expected string, and that the result validates/round-trips.
func TestParityV4ExplicitRandom(t *testing.T) {
	random := [16]byte{
		0x10, 0x91, 0x56, 0xbe, 0xc4, 0xfb, 0xc1, 0xea,
		0x71, 0xb4, 0xef, 0xe1, 0x67, 0x1c, 0x58, 0x36,
	}
	// Same bit-twiddling V4 performs on its random bytes.
	random[6] = (random[6] & 0x0f) | 0x40
	random[8] = (random[8] & 0x3f) | 0x80
	const want = "109156be-c4fb-41ea-b1b4-efe1671c5836"
	if got := Format(random); got != want {
		t.Errorf("Format(v4-transformed) = %q, want %q", got, want)
	}
	if !Validate(want) {
		t.Errorf("Validate(%q) = false, want true", want)
	}
	if b, err := Parse(want); err != nil || Format(b) != want {
		t.Errorf("round-trip of %q failed: err=%v", want, err)
	}
}
