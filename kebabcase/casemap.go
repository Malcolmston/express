// Full (context-sensitive) Unicode lowercase mapping.
//
// change-case lower-cases each split word with JavaScript's
// String.prototype.toLowerCase, which implements Unicode's FULL lowercase
// mapping: U+0130 (Latin capital I with dot above) expands to two code points,
// and a Greek capital sigma at the end of a word becomes the final form "\u03c2".
// Go's standard library offers only the SIMPLE per-rune mapping, which collapses
// U+0130 to a bare "i" and always produces the medial "\u03c3". This file
// supplies the two missing rules so KebabCase matches the Node original using
// only the standard library.
package kebabcase

import "unicode"

// isCased reports whether r is Cased in the Unicode sense: an upper-, lower- or
// title-case letter, or a character in the Other_Lowercase/Other_Uppercase sets
// approximated here by the modifier letters that carry case.
func isCased(r rune) bool {
	return unicode.IsUpper(r) || unicode.IsLower(r) || unicode.IsTitle(r)
}

// isCaseIgnorable approximates the Unicode Case_Ignorable property, which the
// Final_Sigma condition skips over when looking for the surrounding cased
// letters: combining marks, format characters, modifier letters and symbols,
// plus the apostrophes and middle dots called out by the derivation.
func isCaseIgnorable(r rune) bool {
	switch r {
	case '\'', '’', '·', '·', '״', '‧', ':':
		return true
	}
	return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) ||
		unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Lm, r) ||
		unicode.Is(unicode.Sk, r)
}

// finalSigma implements Unicode's Final_Sigma condition for the rune at index i
// of runes: some cased letter precedes it (ignoring case-ignorable runes) and no
// cased letter follows it (likewise). Under that condition U+03A3 lowercases to
// the final form "ς" instead of "σ".
func finalSigma(runes []rune, i int) bool {
	before := false
	for j := i - 1; j >= 0; j-- {
		if isCaseIgnorable(runes[j]) {
			continue
		}
		before = isCased(runes[j])
		break
	}
	if !before {
		return false
	}
	for j := i + 1; j < len(runes); j++ {
		if isCaseIgnorable(runes[j]) {
			continue
		}
		return !isCased(runes[j])
	}
	return true
}

// lowerString returns s with the full lowercase mapping applied, matching
// JavaScript's String.prototype.toLowerCase: U+0130 (Latin capital I with dot
// above) expands to "i" plus a combining dot above, Greek capital sigma becomes
// the final form at the end of a word, and every other rune uses the simple
// mapping.
func lowerString(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes)+1)
	for i, r := range runes {
		switch {
		case r == 'İ':
			out = append(out, 'i', '̇')
		case r == 'Σ' && finalSigma(runes, i):
			out = append(out, 'ς')
		default:
			out = append(out, unicode.ToLower(r))
		}
	}
	return string(out)
}
