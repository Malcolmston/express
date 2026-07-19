package shortid_test

// Upstream-parity tests for the npm library "dylang/shortid".
//
// The concrete input -> expected-output vectors below are extracted verbatim
// from the original project's Mocha test suite:
//
//   https://raw.githubusercontent.com/dylang/shortid/master/test/is-valid.test.js
//   https://raw.githubusercontent.com/dylang/shortid/master/test/shortid.test.js
//
// Upstream algorithm reference (for the gaps documented at the bottom of this
// file):
//   https://raw.githubusercontent.com/dylang/shortid/master/test/alphabet.test.js
//   https://raw.githubusercontent.com/dylang/shortid/master/test/random/random-from-seed.test.js
//
// Only the vectors that map onto this port's exported surface (Generate,
// IsValid) are encoded as executable tests. The seed / shuffle / worker
// determinism vectors exercise upstream machinery this port deliberately does
// not implement; they are listed as gaps in the trailing comment rather than
// as compiling tests.

import (
	"testing"

	"github.com/malcolmston/express/shortid"
)

// TestParityIsValidTrue mirrors is-valid.test.js "should find these valid".
// The upstream default character set (0-9a-zA-Z_-) contains exactly the same
// 64 symbols as this port's default alphabet, so membership is identical even
// though the two orderings differ.
func TestParityIsValidTrue(t *testing.T) {
	valid := []string{
		"N1aGxE",
		"N1M6GeN",
		"41-aGe4",
		"7J--Cn7_Y",
		"7kb-y6XdF",
	}
	for _, id := range valid {
		if !shortid.IsValid(id) {
			t.Errorf("IsValid(%q) = false, want true", id)
		}
	}
}

// TestParityIsValidFalse mirrors is-valid.test.js "should find these invalid".
// Upstream rejects strings containing characters outside the alphabet and any
// string shorter than six characters ("abc"). The upstream number (1234) and
// undefined cases have no analog in Go's typed API; the empty string is the
// closest stand-in for undefined and is included.
func TestParityIsValidFalse(t *testing.T) {
	invalid := []string{
		"i have spaces",
		"i have \n breaks \n of \n the \n lines",
		"abc",
		"",
	}
	for _, id := range invalid {
		if shortid.IsValid(id) {
			t.Errorf("IsValid(%q) = true, want false", id)
		}
	}
}

// TestParityGenerateLengthBelow17 mirrors shortid.test.js, which asserts
// expect(id.length).to.be.below(17) for every generated id.
func TestParityGenerateLengthBelow17(t *testing.T) {
	for i := 0; i < 5000; i++ {
		id, err := shortid.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if n := len([]rune(id)); n >= 17 {
			t.Fatalf("generated id %q length %d, want below 17", id, n)
		}
	}
}

// TestParityGenerateNoDuplicates mirrors shortid.test.js "should run a bunch
// and never get duplicates" (5000 iterations, all ids distinct).
func TestParityGenerateNoDuplicates(t *testing.T) {
	seen := make(map[string]bool, 5000)
	for i := 0; i < 5000; i++ {
		id, err := shortid.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if seen[id] {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = true
	}
}

// TestParityGeneratedIsValid ties the two upstream surfaces together: every id
// produced by Generate must satisfy IsValid, an invariant relied on throughout
// the upstream suite.
func TestParityGeneratedIsValid(t *testing.T) {
	for i := 0; i < 1000; i++ {
		id, err := shortid.Generate()
		if err != nil {
			t.Fatal(err)
		}
		if !shortid.IsValid(id) {
			t.Fatalf("generated id %q failed IsValid", id)
		}
	}
}

// Documented upstream gaps not encoded above (this port intentionally omits the
// seeded, deterministic worker/cluster machinery):
//
//   test/random/random-from-seed.test.js — seeded PRNG determinism, e.g.
//     seed(0) => nextValue() == 0.21132115912208504
//     seed(1) => nextValue() == 0.2511917009602195
//     seed(2) => nextValue() == 0.2910622427983539
//   No exported seeded-random surface exists in this port.
//
//   test/alphabet.test.js — seeded shuffle determinism, e.g.
//     seed(1) => shuffled() == "ylZM7VHLvOFcohp01x-fXNr8P_tqin6RkgWGm4SIDdK5s2TAJebzQEBUwuY9j3aC"
//   And characters(newSet) returning the shuffled alphabet. This port's
//   SetAlphabet stores the alphabet verbatim (no shuffle) and returns an error
//   value rather than the shuffled string, so these vectors cannot be mapped.
