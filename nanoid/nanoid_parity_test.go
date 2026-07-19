package nanoid

// Parity tests derived from the npm original "nanoid" (repo: ai/nanoid).
//
// Vectors and invariants were taken verbatim from the upstream test suite and
// source, pulled from:
//   https://raw.githubusercontent.com/ai/nanoid/main/test/index.test.js
//   https://raw.githubusercontent.com/ai/nanoid/main/url-alphabet/index.js
//   https://raw.githubusercontent.com/ai/nanoid/main/index.js
//   https://raw.githubusercontent.com/ai/nanoid/main/package.json  (name: nanoid)
//
// Mapping of upstream API onto this Go port:
//   nanoid()            -> New()          (default urlAlphabet, size 21)
//   nanoid(size)        -> NewSize(size)
//   customAlphabet(a,n) -> Custom(a, n)   (string fast path: mask & reject)
//
// The port implements the mask-and-reject sampler that upstream uses for
// nanoid()/customAlphabet(string). Upstream's separate customRandom() export
// (modulo + safeByteCutoff rejection, injectable RNG) is NOT ported, so its
// exact-value vectors ('cccc', 'dddc', 'acccccdcbacccccdcb') are recorded as a
// missing feature rather than tested here. Multi-byte and >256-symbol
// alphabets (upstream falls back to customRandom for those) are likewise not
// supported by this byte-indexed port.

import (
	"strings"
	"testing"
)

// urlAlphabet describe-block, test/index.test.js:
//
//	'is string', 'has 64 symbols', 'has no duplicates'.
//
// The constant itself is url-alphabet/index.js.
func TestParityURLAlphabet(t *testing.T) {
	const want = "useandom-26T198340PX75pxJACKVERYMINDBUSHWOLF_GQZbfghjklqvwyzrict"
	if DefaultAlphabet != want {
		t.Fatalf("DefaultAlphabet = %q, want upstream urlAlphabet %q", DefaultAlphabet, want)
	}
	// has 64 symbols
	if len(DefaultAlphabet) != 64 {
		t.Fatalf("urlAlphabet length = %d, want 64", len(DefaultAlphabet))
	}
	// has no duplicates (upstream: lastIndexOf(char) === firstIndex)
	for i := 0; i < len(DefaultAlphabet); i++ {
		if strings.LastIndexByte(DefaultAlphabet, DefaultAlphabet[i]) != i {
			t.Fatalf("urlAlphabet has duplicate char %q", DefaultAlphabet[i])
		}
	}
	// uses only A-Za-z0-9_- symbols
	for _, c := range DefaultAlphabet {
		ok := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-'
		if !ok {
			t.Fatalf("urlAlphabet char %q is not URL-safe", c)
		}
	}
}

// test('is ready for 0 size'): equal(nanoid(0), ”).
// Also customAlphabet('abc')(0) === ”, customAlphabet(”)(0) === ”,
// and customAlphabet('a'.repeat(300))(0) === ” (0 short-circuits before the
// alphabet is ever inspected).
func TestParityZeroSize(t *testing.T) {
	cases := []struct {
		alphabet string
		desc     string
	}{
		{DefaultAlphabet, "nanoid(0)"},
		{"abc", "customAlphabet('abc')(0)"},
		{"", "customAlphabet('')(0)"},
		{strings.Repeat("a", 300), "customAlphabet('a'.repeat(300))(0)"},
	}
	for _, c := range cases {
		got, err := Custom(c.alphabet, 0)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", c.desc, err)
		}
		if got != "" {
			t.Fatalf("%s = %q, want empty string", c.desc, got)
		}
	}
}

// customAlphabet describe-block, test/index.test.js:
//
//	test('has options'):  customAlphabet('a', 5)()   === 'aaaaa'
//	test('changes size'): customAlphabet('a')(10)    === 'aaaaaaaaaa'
//
// A single-symbol alphabet is deterministic under mask & reject.
func TestParitySingleCharAlphabet(t *testing.T) {
	got, err := Custom("a", 5)
	if err != nil {
		t.Fatalf("Custom(\"a\", 5) error: %v", err)
	}
	if got != "aaaaa" {
		t.Fatalf("Custom(\"a\", 5) = %q, want %q", got, "aaaaa")
	}

	got, err = Custom("a", 10)
	if err != nil {
		t.Fatalf("Custom(\"a\", 10) error: %v", err)
	}
	if got != "aaaaaaaaaa" {
		t.Fatalf("Custom(\"a\", 10) = %q, want %q", got, "aaaaaaaaaa")
	}
}

// test('generates URL-friendly IDs'): 100 ids, each length 21, every char in
// urlAlphabet. test('changes ID length'): nanoid(10).length === 10.
func TestParityURLFriendlyIDs(t *testing.T) {
	for i := 0; i < 100; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if len(id) != 21 {
			t.Fatalf("id length = %d, want 21", len(id))
		}
		if strings.Trim(id, DefaultAlphabet) != "" {
			t.Fatalf("id %q contains a char outside urlAlphabet", id)
		}
	}
	id, err := NewSize(10)
	if err != nil {
		t.Fatalf("NewSize(10) error: %v", err)
	}
	if len(id) != 10 {
		t.Fatalf("NewSize(10) length = %d, want 10", len(id))
	}
}

// test('generates large IDs'):          nanoid(1000).length === 1000
// test('generates IDs bigger than pool size'): nanoid(70000).length === 70000
// Every character must still be a member of urlAlphabet.
func TestParityLargeIDs(t *testing.T) {
	for _, n := range []int{1000, 70000} {
		id, err := NewSize(n)
		if err != nil {
			t.Fatalf("NewSize(%d) error: %v", n, err)
		}
		if len(id) != n {
			t.Fatalf("NewSize(%d) length = %d", n, len(id))
		}
		if strings.Trim(id, DefaultAlphabet) != "" {
			t.Fatalf("NewSize(%d) produced a char outside urlAlphabet", n)
		}
	}
}

// test('throws on negative or too big ID length'): nanoid(-10) throws
// /Wrong ID size/. The port returns an error instead of panicking.
func TestParityNegativeSize(t *testing.T) {
	if _, err := NewSize(-10); err == nil {
		t.Fatal("NewSize(-10): expected error, got nil")
	}
}

// test('has no collisions'): upstream runs 50*1000 ids and asserts uniqueness.
// A reduced count keeps the parity run fast while still exercising the property.
func TestParityNoCollisions(t *testing.T) {
	seen := make(map[string]struct{}, 20000)
	for i := 0; i < 20000; i++ {
		id, err := New()
		if err != nil {
			t.Fatalf("New error: %v", err)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("collision on id %q", id)
		}
		seen[id] = struct{}{}
	}
}

// customAlphabet test('has flat distribution'): with a 26-letter alphabet every
// symbol must appear across a batch of ids. We assert coverage (all 26 chars
// observed), a weaker but deterministic-enough form of the upstream property.
func TestParityCustomAlphabetCoverage(t *testing.T) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	seen := make(map[rune]bool, len(alphabet))
	for i := 0; i < 2000; i++ {
		id, err := Custom(alphabet, 30)
		if err != nil {
			t.Fatalf("Custom error: %v", err)
		}
		if len(id) != 30 {
			t.Fatalf("length = %d, want 30", len(id))
		}
		for _, c := range id {
			if !strings.ContainsRune(alphabet, c) {
				t.Fatalf("char %q not in alphabet", c)
			}
			seen[c] = true
		}
	}
	if len(seen) != len(alphabet) {
		t.Fatalf("observed %d distinct chars, want all %d", len(seen), len(alphabet))
	}
}
