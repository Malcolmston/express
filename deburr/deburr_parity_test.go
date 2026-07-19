package deburr

import "testing"

// Upstream parity tests for lodash's `deburr`.
//
// Vectors are taken verbatim from the lodash test suite:
//   https://raw.githubusercontent.com/lodash/lodash/master/test/test.js
//     - burredLetters / deburredLetters arrays (lines ~114-162)
//     - QUnit.module('lodash.deburr') block (lines ~4494-4526)
//
// The upstream suite asserts three things:
//  1. Each Latin-1 Supplement and Latin Extended-A letter in `burredLetters`
//     deburrs to the corresponding entry in `deburredLetters`.
//  2. The Latin mathematical operators U+00D7 (×) and U+00F7 (÷) are left
//     unchanged (they sit inside the Latin-1 block but are not letters).
//  3. Every combining diacritical mark, when placed between "e" and "i", is
//     stripped so that Deburr("e"+mark+"i") == "ei".

// parityBurredLetters is the ordered list of source letters from lodash's
// `burredLetters` array: Latin-1 Supplement 0xC0-0xFF (excluding the math
// operators 0xD7 and 0xF7), followed by Latin Extended-A 0x100-0x17F.
func parityBurredLetters() []rune {
	var rs []rune
	add := func(lo, hi rune) {
		for r := lo; r <= hi; r++ {
			rs = append(rs, r)
		}
	}
	add(0xC0, 0xD6) // À..Ö
	add(0xD8, 0xDF) // Ø..ß (skips × at 0xD7)
	add(0xE0, 0xF6) // à..ö
	add(0xF8, 0xFF) // ø..ÿ (skips ÷ at 0xF7)
	add(0x100, 0x17F)
	return rs
}

// parityDeburredLetters is lodash's `deburredLetters` array verbatim, aligned
// positionally with parityBurredLetters.
var parityDeburredLetters = []string{
	"A", "A", "A", "A", "A", "A", "Ae", "C", "E", "E", "E", "E", "I", "I", "I", "I",
	"D", "N", "O", "O", "O", "O", "O", "O", "U", "U", "U", "U", "Y", "Th", "ss", "a",
	"a", "a", "a", "a", "a", "ae", "c", "e", "e", "e", "e", "i", "i", "i", "i", "d",
	"n", "o", "o", "o", "o", "o", "o", "u", "u", "u", "u", "y", "th", "y", "A", "a",
	"A", "a", "A", "a", "C", "c", "C", "c", "C", "c", "C", "c", "D", "d", "D", "d",
	"E", "e", "E", "e", "E", "e", "E", "e", "E", "e", "G", "g", "G", "g", "G", "g",
	"G", "g", "H", "h", "H", "h", "I", "i", "I", "i", "I", "i", "I", "i", "I", "i",
	"IJ", "ij", "J", "j", "K", "k", "k", "L", "l", "L", "l", "L", "l", "L", "l", "L",
	"l", "N", "n", "N", "n", "N", "n", "'n", "N", "n", "O", "o", "O", "o", "O", "o",
	"Oe", "oe", "R", "r", "R", "r", "R", "r", "S", "s", "S", "s", "S", "s", "S", "s",
	"T", "t", "T", "t", "T", "t", "U", "u", "U", "u", "U", "u", "U", "u", "U", "u",
	"U", "u", "W", "w", "Y", "y", "Y", "Z", "z", "Z", "z", "Z", "z", "s",
}

// TestParityBurredLetters mirrors lodash's
// "should convert Latin Unicode letters to basic Latin".
func TestParityBurredLetters(t *testing.T) {
	burred := parityBurredLetters()
	if len(burred) != len(parityDeburredLetters) {
		t.Fatalf("array length mismatch: %d burred vs %d deburred", len(burred), len(parityDeburredLetters))
	}
	for i, r := range burred {
		want := parityDeburredLetters[i]
		if got := Deburr(string(r)); got != want {
			t.Errorf("Deburr(%q) [U+%04X] = %q, want %q", string(r), r, got, want)
		}
	}
}

// TestParityMathOperators mirrors lodash's
// "should not deburr Latin mathematical operators".
func TestParityMathOperators(t *testing.T) {
	for _, op := range []string{"×", "÷"} { // × ÷
		if got := Deburr(op); got != op {
			t.Errorf("Deburr(%q) = %q, want %q (unchanged)", op, got, op)
		}
	}
}

// TestParityComboMarks mirrors lodash's
// "should deburr combining diacritical marks": every combining mark between
// "e" and "i" must be stripped, leaving "ei".
func TestParityComboMarks(t *testing.T) {
	var marks []rune
	for r := rune(0x0300); r <= 0x036F; r++ {
		marks = append(marks, r)
	}
	marks = append(marks, 0xFE20, 0xFE21, 0xFE22, 0xFE23)
	for _, m := range marks {
		in := "e" + string(m) + "i"
		if got := Deburr(in); got != "ei" {
			t.Errorf("Deburr(e+U+%04X+i) = %q, want %q", m, got, "ei")
		}
	}
}
