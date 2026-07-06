// Package deburr provides a faithful port of lodash's deburr utility, built on
// only the Go standard library. It exposes the single Deburr function, which
// converts accented Latin letters to their basic ASCII equivalents and removes
// combining diacritical marks, so that text such as "déjà vu" becomes "deja vu"
// and "Crème Brûlée" becomes "Creme Brulee".
//
// Deburring is most useful as a normalization step ahead of slugging, searching,
// sorting, or generating identifiers. Turning "Málaga" into "Malaga" lets a URL
// slug stay ASCII, lets a case-insensitive search match regardless of the
// accents a user typed, and lets names sort in a predictable order without
// pulling in a full Unicode collation library. It is a lossy, presentation-level
// transform, not a transliteration engine: it approximates the Latin letter a
// reader would recognize rather than attempting a linguistically correct
// romanization.
//
// The algorithm walks the input rune by rune and applies two tables borrowed
// directly from lodash. The first is a fixed map from the Latin-1 Supplement and
// Latin Extended-A blocks to basic Latin: single letters map to a single letter
// (é to e, ñ to n), and certain letters expand to a short digraph (æ to ae, ß to
// ss, Þ to Th, Œ to Oe, and so on), preserving the case of the original. The
// second is a set of combining-mark ranges — the Combining Diacritical Marks
// block and its Extended, Supplement, for-Symbols, and Half-Marks siblings —
// whose runes are dropped so that a base letter followed by a standalone
// combining accent survives as just the base letter.
//
// The edge cases follow from that design. A rune that is neither a known
// accented letter nor a combining mark is passed through unchanged, so ASCII
// text, digits, punctuation, emoji, and non-Latin scripts (Greek, Cyrillic, CJK)
// are all left exactly as they were. The empty string returns the empty string.
// Because the digraph expansions can make the result longer than the input, the
// output length is not tied to the input length even though the builder is
// pre-sized from it. Both precomposed characters (a single é code point) and
// decomposed sequences (e followed by a combining acute accent) are handled: the
// former via the letter table, the latter via combining-mark removal.
//
// Parity with the Node original is intentional and close: the letter-mapping
// table and the combining-mark ranges are the same ones lodash's deburr uses, so
// the ASCII output for Latin-1 and Latin Extended-A input matches. The
// differences are purely idiomatic — Deburr takes and returns a Go string and
// iterates runes rather than operating on a JavaScript string, and it relies on
// Go's native UTF-8 handling instead of a regular-expression replace.
package deburr

import "strings"

// deburredLetters maps accented Latin runes (Latin-1 Supplement and Latin
// Extended-A blocks) to their basic Latin equivalents. It mirrors the
// mapping table used by lodash's deburr implementation.
var deburredLetters = map[rune]string{
	// Latin-1 Supplement block.
	'À': "A", 'Á': "A", 'Â': "A", 'Ã': "A", 'Ä': "A", 'Å': "A",
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a",
	'Ç': "C", 'ç': "c",
	'Ð': "D", 'ð': "d",
	'È': "E", 'É': "E", 'Ê': "E", 'Ë': "E",
	'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'Ì': "I", 'Í': "I", 'Î': "I", 'Ï': "I",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i",
	'Ñ': "N", 'ñ': "n",
	'Ò': "O", 'Ó': "O", 'Ô': "O", 'Õ': "O", 'Ö': "O", 'Ø': "O",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o",
	'Ù': "U", 'Ú': "U", 'Û': "U", 'Ü': "U",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u",
	'Ý': "Y", 'ý': "y", 'ÿ': "y",
	'Æ': "Ae", 'æ': "ae",
	'Þ': "Th", 'þ': "th",
	'ß': "ss",
	// Latin Extended-A block.
	'Ā': "A", 'Ă': "A", 'Ą': "A",
	'ā': "a", 'ă': "a", 'ą': "a",
	'Ć': "C", 'Ĉ': "C", 'Ċ': "C", 'Č': "C",
	'ć': "c", 'ĉ': "c", 'ċ': "c", 'č': "c",
	'Ď': "D", 'Đ': "D", 'ď': "d", 'đ': "d",
	'Ē': "E", 'Ĕ': "E", 'Ė': "E", 'Ę': "E", 'Ě': "E",
	'ē': "e", 'ĕ': "e", 'ė': "e", 'ę': "e", 'ě': "e",
	'Ĝ': "G", 'Ğ': "G", 'Ġ': "G", 'Ģ': "G",
	'ĝ': "g", 'ğ': "g", 'ġ': "g", 'ģ': "g",
	'Ĥ': "H", 'Ħ': "H", 'ĥ': "h", 'ħ': "h",
	'Ĩ': "I", 'Ī': "I", 'Ĭ': "I", 'Į': "I", 'İ': "I",
	'ĩ': "i", 'ī': "i", 'ĭ': "i", 'į': "i", 'ı': "i",
	'Ĵ': "J", 'ĵ': "j",
	'Ķ': "K", 'ķ': "k", 'ĸ': "k",
	'Ĺ': "L", 'Ļ': "L", 'Ľ': "L", 'Ŀ': "L", 'Ł': "L",
	'ĺ': "l", 'ļ': "l", 'ľ': "l", 'ŀ': "l", 'ł': "l",
	'Ń': "N", 'Ņ': "N", 'Ň': "N", 'Ŋ': "N",
	'ń': "n", 'ņ': "n", 'ň': "n", 'ŋ': "n",
	'Ō': "O", 'Ŏ': "O", 'Ő': "O",
	'ō': "o", 'ŏ': "o", 'ő': "o",
	'Ŕ': "R", 'Ŗ': "R", 'Ř': "R",
	'ŕ': "r", 'ŗ': "r", 'ř': "r",
	'Ś': "S", 'Ŝ': "S", 'Ş': "S", 'Š': "S",
	'ś': "s", 'ŝ': "s", 'ş': "s", 'š': "s",
	'Ţ': "T", 'Ť': "T", 'Ŧ': "T",
	'ţ': "t", 'ť': "t", 'ŧ': "t",
	'Ũ': "U", 'Ū': "U", 'Ŭ': "U", 'Ů': "U", 'Ű': "U", 'Ų': "U",
	'ũ': "u", 'ū': "u", 'ŭ': "u", 'ů': "u", 'ű': "u", 'ų': "u",
	'Ŵ': "W", 'ŵ': "w",
	'Ŷ': "Y", 'ŷ': "y", 'Ÿ': "Y",
	'Ź': "Z", 'Ż': "Z", 'Ž': "Z",
	'ź': "z", 'ż': "z", 'ž': "z",
	'Ĳ': "IJ", 'ĳ': "ij",
	'Œ': "Oe", 'œ': "oe",
	'ŉ': "'n", 'ſ': "s",
}

// isComboMark reports whether r is a combining diacritical mark that should be
// stripped. It covers the same Unicode ranges as lodash's combining-mark
// regular expression.
func isComboMark(r rune) bool {
	switch {
	case r >= '̀' && r <= 'ͯ': // Combining Diacritical Marks
		return true
	case r >= '᪰' && r <= '᫿': // Combining Diacritical Marks Extended
		return true
	case r >= '᷀' && r <= '᷿': // Combining Diacritical Marks Supplement
		return true
	case r >= '⃐' && r <= '⃿': // Combining Diacritical Marks for Symbols
		return true
	case r >= '︠' && r <= '︯': // Combining Half Marks
		return true
	default:
		return false
	}
}

// Deburr converts Latin-1 Supplement and Latin Extended-A letters in s to
// their basic Latin equivalents and removes combining diacritical marks.
//
// For example, Deburr("Crème Brûlée") returns "Creme Brulee".
func Deburr(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if mapped, ok := deburredLetters[r]; ok {
			b.WriteString(mapped)
			continue
		}
		if isComboMark(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
