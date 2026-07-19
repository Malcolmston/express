package hashids

// Upstream-parity tests for the "niieani/hashids.js" npm package.
//
// All vectors below are copied verbatim from the ORIGINAL library's own test
// suite (hashids v2.3.0), obtained from the published npm tarball
// (https://registry.npmjs.org/hashids/-/hashids-2.3.0.tgz) which ships the
// original `src/tests/*.test.ts` sources. Concrete sources:
//
//   src/tests/default-params.test.ts   (salt "", minLength 0, default alphabet)
//   src/tests/custom-params.test.ts    (salt "this is my salt", minLength 30,
//                                        alphabet "xzal86grmb4jhysfoqp3we7291kuct5iv0nd")
//   src/tests/custom-salt.test.ts      (round-trip under assorted salts)
//   src/tests/custom-alphabet.test.ts  (round-trip under assorted alphabets)
//   src/tests/min-length.test.ts       (minLength >= requested)
//   src/tests/bad-input.test.ts        (empty/negative -> empty, small alphabet
//                                        rejected, spaces allowed)
//
// The GitHub mirror is github.com/niieani/hashids.js (default branch "master");
// the same files live under src/tests/ there. bigint-only vectors that exceed
// int64 are intentionally excluded because this Go port's API is int64-based.

import (
	"reflect"
	"testing"
)

// TestParityDefaultParams mirrors src/tests/default-params.test.ts:
// new Hashids() -> salt "", minLength 0, default alphabet.
func TestParityDefaultParams(t *testing.T) {
	h, err := New("", 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cases := []struct {
		id   string
		nums []int64
	}{
		{"gY", []int64{0}},
		{"jR", []int64{1}},
		{"R8ZN0", []int64{928728}},
		{"o2fXhV", []int64{1, 2, 3}},
		{"jRfMcP", []int64{1, 0, 0}},
		{"jQcMcW", []int64{0, 0, 1}},
		{"gYcxcr", []int64{0, 0, 0}},
		{"gLpmopgO6", []int64{1000000000000}},
		{"lEW77X7g527", []int64{9007199254740991}},
		{"BrtltWt2tyt1tvt7tJt2t1tD", []int64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}},
		{"G6XOnGQgIpcVcXcqZ4B8Q8B9y", []int64{10000000000, 0, 0, 0, 999999999999999}},
		{"5KoLLVL49RLhYkppOplM6piwWNNANny8N", []int64{9007199254740991, 9007199254740991, 9007199254740991}},
		{"BPg3Qx5f8VrvQkS16wpmwIgj9Q4Jsr93gqx", []int64{1000000001, 1000000002, 1000000003, 1000000004, 1000000005}},
		{"1wfphpilsMtNumCRFRHXIDSqT2UPcWf1hZi3s7tN", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
	}
	for _, c := range cases {
		got, err := h.Encode(c.nums...)
		if err != nil {
			t.Fatalf("Encode(%v): %v", c.nums, err)
		}
		if got != c.id {
			t.Errorf("Encode(%v) = %q, want %q", c.nums, got, c.id)
		}
		back, err := h.Decode(c.id)
		if err != nil {
			t.Fatalf("Decode(%q): %v", c.id, err)
		}
		if !reflect.DeepEqual(back, c.nums) {
			t.Errorf("Decode(%q) = %v, want %v", c.id, back, c.nums)
		}
	}
}

// TestParityCustomParams mirrors src/tests/custom-params.test.ts:
// new Hashids('this is my salt', 30, 'xzal86grmb4jhysfoqp3we7291kuct5iv0nd').
func TestParityCustomParams(t *testing.T) {
	const minLength = 30
	h, err := NewWithAlphabet("this is my salt", minLength, "xzal86grmb4jhysfoqp3we7291kuct5iv0nd")
	if err != nil {
		t.Fatalf("NewWithAlphabet: %v", err)
	}
	cases := []struct {
		id   string
		nums []int64
	}{
		{"nej1m3d5a6yn875e7gr9kbwpqol02q", []int64{0}},
		{"dw1nqdp92yrajvl9v6k3gl5mb0o8ea", []int64{1}},
		{"onqr0bk58p642wldq14djmw21ygl39", []int64{928728}},
		{"18apy3wlqkjvd5h1id7mn5ore2d06b", []int64{1, 2, 3}},
		{"o60edky1ng3vl9hbfavwr5pa2q8mb9", []int64{1, 0, 0}},
		{"o60edky1ng3vlqfbfp4wr5pa2q8mb9", []int64{0, 0, 1}},
		{"qek2a08gpl575efrfd7yomj9dwbr63", []int64{0, 0, 0}},
		{"m3d5a6yn875rae8y81a94gr9kbwpqo", []int64{1000000000000}},
		{"1q3y98ln48w96kpo0wgk314w5mak2d", []int64{9007199254740991}},
		{"op7qrcdc3cgc2c0cbcrcoc5clce4d6", []int64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}},
		{"5430bd2jo0lxyfkfjfyojej5adqdy4", []int64{10000000000, 0, 0, 0, 999999999999999}},
		{"aa5kow86ano1pt3e1aqm239awkt9pk380w9l3q6", []int64{9007199254740991, 9007199254740991, 9007199254740991}},
		{"mmmykr5nuaabgwnohmml6dakt00jmo3ainnpy2mk", []int64{1000000001, 1000000002, 1000000003, 1000000004, 1000000005}},
		{"w1hwinuwt1cbs6xwzafmhdinuotpcosrxaz0fahl", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}},
	}
	for _, c := range cases {
		got, err := h.Encode(c.nums...)
		if err != nil {
			t.Fatalf("Encode(%v): %v", c.nums, err)
		}
		if got != c.id {
			t.Errorf("Encode(%v) = %q, want %q", c.nums, got, c.id)
		}
		if len([]rune(got)) < minLength {
			t.Errorf("Encode(%v) length %d < minLength %d", c.nums, len([]rune(got)), minLength)
		}
		back, err := h.Decode(c.id)
		if err != nil {
			t.Fatalf("Decode(%q): %v", c.id, err)
		}
		if !reflect.DeepEqual(back, c.nums) {
			t.Errorf("Decode(%q) = %v, want %v", c.id, back, c.nums)
		}
	}
}

// TestParityCustomSalt mirrors src/tests/custom-salt.test.ts: [1,2,3] must
// round-trip under each salt (empty, spaces, normal, and long weird salts).
func TestParityCustomSalt(t *testing.T) {
	salts := []string{
		"",
		"   ",
		"this is my salt",
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890`~!@#$%^&*()-_=+\\|'\";:/?.>,<{[}]",
		"`~!@#$%^&*()-_=+\\|'\";:/?.>,<{[}]",
	}
	want := []int64{1, 2, 3}
	for _, salt := range salts {
		h, err := New(salt, 0)
		if err != nil {
			t.Fatalf("New(salt=%q): %v", salt, err)
		}
		id, err := h.Encode(want...)
		if err != nil {
			t.Fatalf("Encode(salt=%q): %v", salt, err)
		}
		got, err := h.Decode(id)
		if err != nil {
			t.Fatalf("Decode(salt=%q): %v", salt, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("salt=%q: round-trip %v -> %q -> %v", salt, want, id, got)
		}
	}
}

// TestParityCustomAlphabet mirrors src/tests/custom-alphabet.test.ts: [1,2,3]
// must round-trip under each alphabet. Upstream explicitly supports alphabets
// containing spaces and separator-heavy alphabets.
func TestParityCustomAlphabet(t *testing.T) {
	alphabets := []string{
		"cCsSfFhHuUiItT01",  // worst alphabet
		"cCsSfFhH uUiItT01", // contains a space
		"abdegjklCFHISTUc",  // half separators
		"abdegjklmnopqrSF",  // exactly 2 separators
		"abdegjklmnopqrvwxyzABDEGJKLMNOPQRVWXYZ1234567890",                                                 // no separators
		"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890`~!@#$%^&*()-_=+\\|'\";:/?.>,<{[}]", // super long
		"`~!@#$%^&*()-_=+\\|'\";:/?.>,<{[}]",                                                               // weird
	}
	want := []int64{1, 2, 3}
	for _, alphabet := range alphabets {
		h, err := NewWithAlphabet("", 0, alphabet)
		if err != nil {
			t.Fatalf("NewWithAlphabet(alphabet=%q): %v", alphabet, err)
		}
		id, err := h.Encode(want...)
		if err != nil {
			t.Fatalf("Encode(alphabet=%q): %v", alphabet, err)
		}
		got, err := h.Decode(id)
		if err != nil {
			t.Fatalf("Decode(alphabet=%q): %v", alphabet, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("alphabet=%q: round-trip %v -> %q -> %v", alphabet, want, id, got)
		}
	}
}

// TestParityMinLength mirrors src/tests/min-length.test.ts: [1,2,3] round-trips
// and the id length is at least minLength for each requested minimum.
func TestParityMinLength(t *testing.T) {
	want := []int64{1, 2, 3}
	for _, minLength := range []int{0, 1, 10, 999, 1000} {
		h, err := New("", minLength)
		if err != nil {
			t.Fatalf("New(minLength=%d): %v", minLength, err)
		}
		id, err := h.Encode(want...)
		if err != nil {
			t.Fatalf("Encode(minLength=%d): %v", minLength, err)
		}
		if len([]rune(id)) < minLength {
			t.Errorf("minLength=%d: id %q length %d < %d", minLength, id, len([]rune(id)), minLength)
		}
		got, err := h.Decode(id)
		if err != nil {
			t.Fatalf("Decode(minLength=%d): %v", minLength, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("minLength=%d: round-trip %v -> %q -> %v", minLength, want, id, got)
		}
	}
}

// TestParityBadInput mirrors src/tests/bad-input.test.ts for the int64 API:
// encoding nothing yields "", encoding a negative number yields "" (upstream
// returns an empty string rather than throwing), an alphabet under 16 unique
// characters is rejected, and an alphabet containing a space is accepted.
func TestParityBadInput(t *testing.T) {
	h, err := New("", 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if id, err := h.Encode(); err != nil || id != "" {
		t.Errorf("Encode() = %q, %v; want \"\", nil", id, err)
	}

	// Upstream: encode(-1) -> "". This port returns an error instead; accept
	// either an error or an empty string, but never a non-empty encoding.
	if id, err := h.Encode(-1); err == nil && id != "" {
		t.Errorf("Encode(-1) = %q; want empty", id)
	}

	if nums, err := h.Decode(""); err != nil || len(nums) != 0 {
		t.Errorf("Decode(\"\") = %v, %v; want empty", nums, err)
	}

	// Alphabet with fewer than 16 unique characters must be rejected.
	if _, err := NewWithAlphabet("", 0, "1234567890"); err == nil {
		t.Errorf("NewWithAlphabet with 10-char alphabet: want error, got nil")
	}

	// Alphabet containing a space must be accepted (upstream does not throw).
	if _, err := NewWithAlphabet("", 0, "a cdefghijklmnopqrstuvwxyz"); err != nil {
		t.Errorf("NewWithAlphabet with spaced alphabet: unexpected error %v", err)
	}
}
