// Full (multi-character and context-sensitive) Unicode case mapping.
//
// JavaScript's String.prototype.toLowerCase and toUpperCase implement the FULL
// case mappings of Unicode's SpecialCasing.txt: a single code point may map to
// several ("ß" uppercases to "SS", "ﬄ" to "FFL"), and Greek capital sigma
// lowercases to the final form "ς" at the end of a word. Go's standard library
// only offers the SIMPLE per-rune mapping (unicode.ToUpper/ToLower and their
// strings wrappers), which leaves "ß" and "ﬄ" unchanged and always produces the
// medial "σ".
//
// camelcase re-cases every word it splits, so those differences are directly
// observable in its output. This file supplies the missing full mappings so the
// port matches npm camelcase, which is a thin layer over the JavaScript
// built-ins, using only the standard library.
package camelcase

import "unicode"

// fullUpperExpansions holds every code point whose full uppercase mapping is
// longer than one code point. It is generated mechanically from the JavaScript
// engine itself, which is the behaviour the port has to match:
//
//	node -e 'for (let c = 0; c <= 0x2FFFF; c++) {
//	           const s = String.fromCodePoint(c), u = s.toUpperCase()
//	           if ([...u].length > 1) console.log(c, [...u].map(x => x.codePointAt(0)))
//	         }'
//
// The single multi-character LOWERCASE mapping (U+0130) and the context
// sensitive final-sigma rule are handled in code below rather than by a table.
var fullUpperExpansions = map[rune][]rune{
	0x00DF: {0x0053, 0x0053},         // U+00DF "ß" -> "SS"
	0x0149: {0x02BC, 0x004E},         // U+0149 "ŉ" -> "ʼN"
	0x01F0: {0x004A, 0x030C},         // U+01F0 "ǰ" -> "J̌"
	0x0390: {0x0399, 0x0308, 0x0301}, // U+0390 "ΐ" -> "Ϊ́"
	0x03B0: {0x03A5, 0x0308, 0x0301}, // U+03B0 "ΰ" -> "Ϋ́"
	0x0587: {0x0535, 0x0552},         // U+0587 "և" -> "ԵՒ"
	0x1E96: {0x0048, 0x0331},         // U+1E96 "ẖ" -> "H̱"
	0x1E97: {0x0054, 0x0308},         // U+1E97 "ẗ" -> "T̈"
	0x1E98: {0x0057, 0x030A},         // U+1E98 "ẘ" -> "W̊"
	0x1E99: {0x0059, 0x030A},         // U+1E99 "ẙ" -> "Y̊"
	0x1E9A: {0x0041, 0x02BE},         // U+1E9A "ẚ" -> "Aʾ"
	0x1F50: {0x03A5, 0x0313},         // U+1F50 "ὐ" -> "Υ̓"
	0x1F52: {0x03A5, 0x0313, 0x0300}, // U+1F52 "ὒ" -> "Υ̓̀"
	0x1F54: {0x03A5, 0x0313, 0x0301}, // U+1F54 "ὔ" -> "Υ̓́"
	0x1F56: {0x03A5, 0x0313, 0x0342}, // U+1F56 "ὖ" -> "Υ̓͂"
	0x1F80: {0x1F08, 0x0399},         // U+1F80 "ᾀ" -> "ἈΙ"
	0x1F81: {0x1F09, 0x0399},         // U+1F81 "ᾁ" -> "ἉΙ"
	0x1F82: {0x1F0A, 0x0399},         // U+1F82 "ᾂ" -> "ἊΙ"
	0x1F83: {0x1F0B, 0x0399},         // U+1F83 "ᾃ" -> "ἋΙ"
	0x1F84: {0x1F0C, 0x0399},         // U+1F84 "ᾄ" -> "ἌΙ"
	0x1F85: {0x1F0D, 0x0399},         // U+1F85 "ᾅ" -> "ἍΙ"
	0x1F86: {0x1F0E, 0x0399},         // U+1F86 "ᾆ" -> "ἎΙ"
	0x1F87: {0x1F0F, 0x0399},         // U+1F87 "ᾇ" -> "ἏΙ"
	0x1F88: {0x1F08, 0x0399},         // U+1F88 "ᾈ" -> "ἈΙ"
	0x1F89: {0x1F09, 0x0399},         // U+1F89 "ᾉ" -> "ἉΙ"
	0x1F8A: {0x1F0A, 0x0399},         // U+1F8A "ᾊ" -> "ἊΙ"
	0x1F8B: {0x1F0B, 0x0399},         // U+1F8B "ᾋ" -> "ἋΙ"
	0x1F8C: {0x1F0C, 0x0399},         // U+1F8C "ᾌ" -> "ἌΙ"
	0x1F8D: {0x1F0D, 0x0399},         // U+1F8D "ᾍ" -> "ἍΙ"
	0x1F8E: {0x1F0E, 0x0399},         // U+1F8E "ᾎ" -> "ἎΙ"
	0x1F8F: {0x1F0F, 0x0399},         // U+1F8F "ᾏ" -> "ἏΙ"
	0x1F90: {0x1F28, 0x0399},         // U+1F90 "ᾐ" -> "ἨΙ"
	0x1F91: {0x1F29, 0x0399},         // U+1F91 "ᾑ" -> "ἩΙ"
	0x1F92: {0x1F2A, 0x0399},         // U+1F92 "ᾒ" -> "ἪΙ"
	0x1F93: {0x1F2B, 0x0399},         // U+1F93 "ᾓ" -> "ἫΙ"
	0x1F94: {0x1F2C, 0x0399},         // U+1F94 "ᾔ" -> "ἬΙ"
	0x1F95: {0x1F2D, 0x0399},         // U+1F95 "ᾕ" -> "ἭΙ"
	0x1F96: {0x1F2E, 0x0399},         // U+1F96 "ᾖ" -> "ἮΙ"
	0x1F97: {0x1F2F, 0x0399},         // U+1F97 "ᾗ" -> "ἯΙ"
	0x1F98: {0x1F28, 0x0399},         // U+1F98 "ᾘ" -> "ἨΙ"
	0x1F99: {0x1F29, 0x0399},         // U+1F99 "ᾙ" -> "ἩΙ"
	0x1F9A: {0x1F2A, 0x0399},         // U+1F9A "ᾚ" -> "ἪΙ"
	0x1F9B: {0x1F2B, 0x0399},         // U+1F9B "ᾛ" -> "ἫΙ"
	0x1F9C: {0x1F2C, 0x0399},         // U+1F9C "ᾜ" -> "ἬΙ"
	0x1F9D: {0x1F2D, 0x0399},         // U+1F9D "ᾝ" -> "ἭΙ"
	0x1F9E: {0x1F2E, 0x0399},         // U+1F9E "ᾞ" -> "ἮΙ"
	0x1F9F: {0x1F2F, 0x0399},         // U+1F9F "ᾟ" -> "ἯΙ"
	0x1FA0: {0x1F68, 0x0399},         // U+1FA0 "ᾠ" -> "ὨΙ"
	0x1FA1: {0x1F69, 0x0399},         // U+1FA1 "ᾡ" -> "ὩΙ"
	0x1FA2: {0x1F6A, 0x0399},         // U+1FA2 "ᾢ" -> "ὪΙ"
	0x1FA3: {0x1F6B, 0x0399},         // U+1FA3 "ᾣ" -> "ὫΙ"
	0x1FA4: {0x1F6C, 0x0399},         // U+1FA4 "ᾤ" -> "ὬΙ"
	0x1FA5: {0x1F6D, 0x0399},         // U+1FA5 "ᾥ" -> "ὭΙ"
	0x1FA6: {0x1F6E, 0x0399},         // U+1FA6 "ᾦ" -> "ὮΙ"
	0x1FA7: {0x1F6F, 0x0399},         // U+1FA7 "ᾧ" -> "ὯΙ"
	0x1FA8: {0x1F68, 0x0399},         // U+1FA8 "ᾨ" -> "ὨΙ"
	0x1FA9: {0x1F69, 0x0399},         // U+1FA9 "ᾩ" -> "ὩΙ"
	0x1FAA: {0x1F6A, 0x0399},         // U+1FAA "ᾪ" -> "ὪΙ"
	0x1FAB: {0x1F6B, 0x0399},         // U+1FAB "ᾫ" -> "ὫΙ"
	0x1FAC: {0x1F6C, 0x0399},         // U+1FAC "ᾬ" -> "ὬΙ"
	0x1FAD: {0x1F6D, 0x0399},         // U+1FAD "ᾭ" -> "ὭΙ"
	0x1FAE: {0x1F6E, 0x0399},         // U+1FAE "ᾮ" -> "ὮΙ"
	0x1FAF: {0x1F6F, 0x0399},         // U+1FAF "ᾯ" -> "ὯΙ"
	0x1FB2: {0x1FBA, 0x0399},         // U+1FB2 "ᾲ" -> "ᾺΙ"
	0x1FB3: {0x0391, 0x0399},         // U+1FB3 "ᾳ" -> "ΑΙ"
	0x1FB4: {0x0386, 0x0399},         // U+1FB4 "ᾴ" -> "ΆΙ"
	0x1FB6: {0x0391, 0x0342},         // U+1FB6 "ᾶ" -> "Α͂"
	0x1FB7: {0x0391, 0x0342, 0x0399}, // U+1FB7 "ᾷ" -> "Α͂Ι"
	0x1FBC: {0x0391, 0x0399},         // U+1FBC "ᾼ" -> "ΑΙ"
	0x1FC2: {0x1FCA, 0x0399},         // U+1FC2 "ῂ" -> "ῊΙ"
	0x1FC3: {0x0397, 0x0399},         // U+1FC3 "ῃ" -> "ΗΙ"
	0x1FC4: {0x0389, 0x0399},         // U+1FC4 "ῄ" -> "ΉΙ"
	0x1FC6: {0x0397, 0x0342},         // U+1FC6 "ῆ" -> "Η͂"
	0x1FC7: {0x0397, 0x0342, 0x0399}, // U+1FC7 "ῇ" -> "Η͂Ι"
	0x1FCC: {0x0397, 0x0399},         // U+1FCC "ῌ" -> "ΗΙ"
	0x1FD2: {0x0399, 0x0308, 0x0300}, // U+1FD2 "ῒ" -> "Ϊ̀"
	0x1FD3: {0x0399, 0x0308, 0x0301}, // U+1FD3 "ΐ" -> "Ϊ́"
	0x1FD6: {0x0399, 0x0342},         // U+1FD6 "ῖ" -> "Ι͂"
	0x1FD7: {0x0399, 0x0308, 0x0342}, // U+1FD7 "ῗ" -> "Ϊ͂"
	0x1FE2: {0x03A5, 0x0308, 0x0300}, // U+1FE2 "ῢ" -> "Ϋ̀"
	0x1FE3: {0x03A5, 0x0308, 0x0301}, // U+1FE3 "ΰ" -> "Ϋ́"
	0x1FE4: {0x03A1, 0x0313},         // U+1FE4 "ῤ" -> "Ρ̓"
	0x1FE6: {0x03A5, 0x0342},         // U+1FE6 "ῦ" -> "Υ͂"
	0x1FE7: {0x03A5, 0x0308, 0x0342}, // U+1FE7 "ῧ" -> "Ϋ͂"
	0x1FF2: {0x1FFA, 0x0399},         // U+1FF2 "ῲ" -> "ῺΙ"
	0x1FF3: {0x03A9, 0x0399},         // U+1FF3 "ῳ" -> "ΩΙ"
	0x1FF4: {0x038F, 0x0399},         // U+1FF4 "ῴ" -> "ΏΙ"
	0x1FF6: {0x03A9, 0x0342},         // U+1FF6 "ῶ" -> "Ω͂"
	0x1FF7: {0x03A9, 0x0342, 0x0399}, // U+1FF7 "ῷ" -> "Ω͂Ι"
	0x1FFC: {0x03A9, 0x0399},         // U+1FFC "ῼ" -> "ΩΙ"
	0xFB00: {0x0046, 0x0046},         // U+FB00 "ﬀ" -> "FF"
	0xFB01: {0x0046, 0x0049},         // U+FB01 "ﬁ" -> "FI"
	0xFB02: {0x0046, 0x004C},         // U+FB02 "ﬂ" -> "FL"
	0xFB03: {0x0046, 0x0046, 0x0049}, // U+FB03 "ﬃ" -> "FFI"
	0xFB04: {0x0046, 0x0046, 0x004C}, // U+FB04 "ﬄ" -> "FFL"
	0xFB05: {0x0053, 0x0054},         // U+FB05 "ﬅ" -> "ST"
	0xFB06: {0x0053, 0x0054},         // U+FB06 "ﬆ" -> "ST"
	0xFB13: {0x0544, 0x0546},         // U+FB13 "ﬓ" -> "ՄՆ"
	0xFB14: {0x0544, 0x0535},         // U+FB14 "ﬔ" -> "ՄԵ"
	0xFB15: {0x0544, 0x053B},         // U+FB15 "ﬕ" -> "ՄԻ"
	0xFB16: {0x054E, 0x0546},         // U+FB16 "ﬖ" -> "ՎՆ"
	0xFB17: {0x0544, 0x053D},         // U+FB17 "ﬗ" -> "ՄԽ"
}

// upperRune returns the full uppercase mapping of r, which may be more than one
// rune. It mirrors JavaScript's String.prototype.toUpperCase applied to a single
// character.
func upperRune(r rune) []rune {
	if exp, ok := fullUpperExpansions[r]; ok {
		out := make([]rune, len(exp))
		copy(out, exp)
		return out
	}
	return []rune{unicode.ToUpper(r)}
}

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
