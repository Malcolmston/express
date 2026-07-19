package stringdistance

import "testing"

// Upstream parity vectors transcribed verbatim from the real test suite of
// sindresorhus/leven (the npm library this port reimplements):
//
//	https://raw.githubusercontent.com/sindresorhus/leven/main/test.js
//	https://raw.githubusercontent.com/sindresorhus/leven/main/index.js
//
// In upstream, `leven(a, b)` returns the Levenshtein edit distance as an
// integer. Those calls map directly onto this port's Levenshtein function.
// Only vectors whose asserted upstream value is the *true* (uncapped) edit
// distance are encoded here; upstream's {maxDistance} capping is a feature the
// port does not expose (see notes) and its capped assertions are excluded.

// TestParityLevenshtein covers every uncapped leven(a, b) assertion from
// upstream's `main` test block plus the plain (option-less) calls in its
// `maxDistance option` block.
func TestParityLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		// upstream test('main')
		{"a", "b", 1},
		{"ab", "ac", 1},
		{"ac", "bc", 1},
		{"abc", "axc", 1},
		{"kitten", "sitting", 3},
		{"xabxcdxxefxgx", "1ab2cd34ef5g6", 6},
		{"cat", "cow", 2},
		{"xabxcdxxefxgx", "abcdefg", 6},
		{"javawasneat", "scalaisgreat", 7},
		{"example", "samples", 3},
		{"sturgeon", "urgently", 6},
		{"levenshtein", "frankenstein", 6},
		{"distance", "difference", 5},
		{"因為我是中國人所以我會說中文", "因為我是英國人所以我會說英文", 2},

		// upstream test('maxDistance option') — calls made WITHOUT the option,
		// i.e. the true edit distance is asserted.
		{"abcdef", "123456", 6},
		{"abcdef", "abcdefg", 1},
		{"foo", "bar", 3},

		// upstream test('maxDistance option') — capped calls whose cap is >= the
		// true distance, so the asserted value IS the true edit distance.
		{"cat", "cow", 2},   // {maxDistance:5} -> 2
		{"same", "same", 0}, // {maxDistance:1} -> 0 (identical)
		{"", "abc", 3},      // {maxDistance:10} -> 3
		{"abc", "abc", 0},   // {maxDistance:0} -> 0 (identical)
	}
	for _, c := range cases {
		if got := Levenshtein(c.a, c.b); got != c.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d (upstream leven)", c.a, c.b, got, c.want)
		}
	}
}

// TestParityClosestMatch covers upstream closestMatch(target, candidates)
// assertions from test('closestMatch'). Upstream ranks candidates by
// Levenshtein distance; this port's ClosestMatch ranks by Sørensen–Dice
// coefficient. Only the subset of upstream cases whose winner is identical
// under both metrics (and whose signature is expressible here) is encoded, so a
// passing test confirms the port agrees with upstream on those inputs.
func TestParityClosestMatch(t *testing.T) {
	cases := []struct {
		target     string
		candidates []string
		want       string
	}{
		{"hello", []string{"jello", "yellow", "bellow"}, "jello"},
		{"foo", []string{"bar", "foo", "baz"}, "foo"},
		{"test", []string{"testing"}, "testing"},
		{"café", []string{"cafe", "caffè", "café"}, "café"},
		{"test", []string{"a", "b", "c", "test", "d", "e"}, "test"},
	}
	for _, c := range cases {
		if got, _ := ClosestMatch(c.target, c.candidates); got != c.want {
			t.Errorf("ClosestMatch(%q, %v) = %q, want %q (upstream leven closestMatch)", c.target, c.candidates, got, c.want)
		}
	}
}
