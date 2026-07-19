package cuid

// Upstream-parity tests for the npm library "paralleldrive/cuid".
//
// Vectors are derived from the original library's own source and test suite:
//   - Implementation: https://raw.githubusercontent.com/paralleldrive/cuid/master/index.js
//   - Tests:          https://raw.githubusercontent.com/paralleldrive/cuid/master/test/test.js
//   - Slug/pad libs:  https://raw.githubusercontent.com/paralleldrive/cuid/master/lib/pad.js
//
// The upstream values (cuid, slug, fingerprint) are random, so — exactly as the
// upstream tape suite does — these tests assert structural invariants and the
// isCuid/isSlug contracts rather than fixed literal strings.
//
// Two upstream behaviours are DELIBERATELY not asserted here because this Go
// package is a cuid2-style port and reconciling them would contradict the
// port's documented design and break its existing tests (see notes):
//   1. Upstream cuid() hard-codes a leading 'c'; this port uses a random
//      leading letter a..z.
//   2. Upstream isCuid() returns true iff the argument is a string starting
//      with 'c' (so isCuid("abcdefghijklmnopqrstuvwxy") === false); this port's
//      IsCuid is a stricter cuid2 syntactic validator.

import "testing"

// Upstream test: `t.ok(typeof cuid() === 'string', ...)`.
// Port analogue: New() returns a non-empty string and no error.
func TestParityCuidReturnsString(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("New error: %v", err)
	}
	if id == "" {
		t.Fatal("New() returned an empty string")
	}
}

// Upstream cuid() always begins with a lowercase letter (upstream: hard-coded
// 'c'; this port: a random letter). The reconcilable, shared invariant is that
// the first character is a lowercase letter a..z, which keeps the id safe as an
// HTML element id, exactly as the upstream comment intends.
func TestParityCuidStartsWithLetter(t *testing.T) {
	for i := 0; i < 200; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if c := id[0]; c < 'a' || c > 'z' {
			t.Fatalf("first char %q is not a lowercase letter: %q", c, id)
		}
	}
}

// Upstream test: `cuid.isCuid(id) === true` for an id produced by cuid().
// Port analogue: IsCuid accepts the ids this package generates.
func TestParityIsCuidAcceptsGenerated(t *testing.T) {
	for i := 0; i < 200; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if !IsCuid(id) {
			t.Fatalf("IsCuid(%q) = false, want true", id)
		}
	}
}

// Upstream test: `cuid.isCuid(null) === false` and `cuid.isCuid(undefined)
// === false`. Go's IsCuid is typed to string, so the closest representable
// analogue of the absent/empty value is the empty string, which both upstream
// (""" does not start with 'c') and this port reject.
func TestParityIsCuidRejectsEmpty(t *testing.T) {
	if IsCuid("") {
		t.Fatal(`IsCuid("") = true, want false`)
	}
}

// Upstream test: `cuid.slug()` returns a string, and slugs pass isSlug.
// Port analogue: Slug() returns a non-empty, IsSlug-valid string.
func TestParitySlugReturnsValidSlug(t *testing.T) {
	s, err := Slug()
	if err != nil {
		t.Fatalf("Slug error: %v", err)
	}
	if s == "" {
		t.Fatal("Slug() returned an empty string")
	}
	if !IsSlug(s) {
		t.Fatalf("IsSlug(Slug()=%q) = false, want true (len=%d)", s, len(s))
	}
	// The upstream slug alphabet is base-36; verify URL-safe lowercase chars.
	for j := 0; j < len(s); j++ {
		c := s[j]
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')) {
			t.Fatalf("slug %q contains non-base36 char %q", s, c)
		}
	}
}

// Upstream test: `cuid.isSlug(slug) === true` for a valid slug, plus rejection
// of non-slugs. Upstream isSlug is purely a length check: 7 <= len <= 10.
// These vectors exercise the exact boundaries (6 reject, 7 accept, 10 accept,
// 11 reject) and the empty string (the analogue of null/undefined → false).
func TestParityIsSlugLengthContract(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},                // null/undefined analogue
		{"abcde", false},           // len 5
		{"abcdef", false},          // len 6 (just below the lower bound)
		{"abcdefg", true},          // len 7 (lower bound)
		{"abcdefgh", true},         // len 8
		{"abcdefghij", true},       // len 10 (upper bound)
		{"abcdefghijk", false},     // len 11 (just above the upper bound)
		{"abcdefghijklmno", false}, // len 15
	}
	for _, c := range cases {
		if got := IsSlug(c.in); got != c.want {
			t.Fatalf("IsSlug(%q) = %v, want %v (len=%d)", c.in, got, c.want, len(c.in))
		}
	}
}

// Upstream relies on a monotone per-process counter so same-millisecond slugs
// differ; the upstream suite runs a large collision test. Assert the port's
// slugs and cuids do not collide across many rapid calls.
func TestParityNoCollisions(t *testing.T) {
	seen := make(map[string]struct{}, 20000)
	for i := 0; i < 10000; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("cuid collision: %q", id)
		}
		seen[id] = struct{}{}
	}
	for i := 0; i < 10000; i++ {
		s, err := Slug()
		if err != nil {
			t.Fatalf("Slug error: %v", err)
		}
		if _, dup := seen[s]; dup {
			t.Fatalf("slug collision: %q", s)
		}
		seen[s] = struct{}{}
	}
}
