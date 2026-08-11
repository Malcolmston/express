// Package kebabcase converts strings to kebab-case, a stdlib-only Go port of
// the `kebabCase` function of the npm package "change-case"
// (blakeembrey/change-case). Kebab-case is the lower-cased, hyphen-delimited
// word form popular in URL slugs, CSS class and custom-property names, HTML
// data attributes, CLI flags, and file names, where a single ASCII hyphen
// visually separates words like the skewers of a kebab. The package exposes a
// single entry point, KebabCase, so it can be dropped in wherever the Node
// original was used without carrying any third-party dependencies.
//
// A caller reaches for this package whenever a human-authored or programmatic
// identifier arrives in some other casing convention (camelCase, PascalCase,
// snake_case, "Title Case", or a mix of these) and a canonical kebab-case form
// is required. Typical uses are deriving a stable slug from a page title,
// normalizing component names, or turning a struct field name into a
// configuration key. Because the transformation is deterministic and depends
// only on the input runes, the same string always maps to the same result.
//
// The algorithm is change-case's `split` followed by a lower-case join, and it
// runs in four stages. First the input is trimmed. Then two word boundaries are
// marked: one between a lower-case letter or ASCII digit and a following
// upper-case letter (so "fooBar" splits into "foo" and "Bar"), and one between
// an upper-case letter and a following upper-case-then-lower-case pair (the
// acronym rule, so "XMLHttpRequest" splits into "XML" and "HttpRequest" rather
// than "XMLHttp" and "Request"). Both are applied left to right without
// overlapping, exactly as the upstream global regular-expression replaces are.
// Third, every run of characters that is neither a Unicode letter nor an ASCII
// digit is collapsed into a single boundary marker, and markers at the two ends
// are dropped, so the result never begins, ends, or contains a doubled
// delimiter. Finally each word is lower-cased and the words are joined with
// "-".
//
// The case rule gives well-defined behaviour on the awkward inputs. A run of
// adjacent capitals with nothing after it is one word, so "FOO" stays "foo",
// while "ABCdef" splits before the last capital and yields "ab-cdef". A digit
// that precedes a capital counts as a boundary, so "foo123Bar" becomes
// "foo123-bar", but only ASCII digits do: upstream's \d is ASCII-only, so
// Arabic-Indic digits are stripped as non-word characters. Empty or
// whitespace-only input, and input consisting solely of separators such as
// "__foo__bar__" or "--foo--bar--", reduce cleanly to "" or to the trimmed
// "foo-bar" respectively. Lower-casing uses the full Unicode mapping that
// JavaScript's String.prototype.toLowerCase implements, so the Turkish dotted
// capital I expands to two code points and a Greek capital sigma at the end of a
// word becomes the final form "ς" (see casemap.go).
//
// Only the default options of the Node original are implemented: the delimiter
// is always "-", and the locale, prefixCharacters, suffixCharacters,
// separateNumbers and custom split/strip options are not exposed.
package kebabcase

import (
	"strings"
	"unicode"
)

// marker is the sentinel change-case injects at a word boundary before
// splitting. Upstream uses NUL for the same purpose.
const marker = '\x00'

// isWord reports whether r is a word character in upstream's sense: the
// DEFAULT_STRIP_REGEXP is /[^\p{L}\d]+/, and JavaScript's \d is ASCII-only even
// under the /u flag, so only Unicode letters and ASCII digits survive.
func isWord(r rune) bool {
	return unicode.IsLetter(r) || (r >= '0' && r <= '9')
}

// isASCIIDigit reports whether r matches JavaScript's \d.
func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// splitLowerUpper marks the boundary of upstream's
// /([\p{Ll}\d])(\p{Lu})/gu replace: a lower-case letter or ASCII digit followed
// by an upper-case letter. Matches consume both runes, so they cannot overlap —
// "aBC" marks only between "a" and "B".
func splitLowerUpper(runes []rune) []rune {
	out := make([]rune, 0, len(runes)+8)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if i+1 < len(runes) &&
			(unicode.IsLower(r) || isASCIIDigit(r)) &&
			unicode.IsUpper(runes[i+1]) {
			out = append(out, r, marker, runes[i+1])
			i++
			continue
		}
		out = append(out, r)
	}
	return out
}

// splitUpperUpper marks the boundary of upstream's
// /(\p{Lu})([\p{Lu}][\p{Ll}])/gu replace: an upper-case letter followed by an
// upper-case letter and then a lower-case letter, which is how an acronym is
// separated from the word that follows it. Matches consume three runes.
func splitUpperUpper(runes []rune) []rune {
	out := make([]rune, 0, len(runes)+8)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if i+2 < len(runes) &&
			unicode.IsUpper(r) &&
			unicode.IsUpper(runes[i+1]) &&
			unicode.IsLower(runes[i+2]) {
			out = append(out, r, marker, runes[i+1], runes[i+2])
			i += 2
			continue
		}
		out = append(out, r)
	}
	return out
}

// splitWords reproduces change-case's `split`: mark the two case boundaries,
// collapse every non-word run (markers included) into one boundary, drop the
// boundaries at the two ends, and return the remaining words.
func splitWords(s string) []string {
	runes := splitUpperUpper(splitLowerUpper([]rune(strings.TrimSpace(s))))

	// Collapse non-word runs, including the markers just inserted, so a
	// boundary next to punctuation does not produce an empty word.
	collapsed := make([]rune, 0, len(runes))
	for i := 0; i < len(runes); i++ {
		if isWord(runes[i]) {
			collapsed = append(collapsed, runes[i])
			continue
		}
		for i+1 < len(runes) && !isWord(runes[i+1]) {
			i++
		}
		collapsed = append(collapsed, marker)
	}

	start, end := 0, len(collapsed)
	for start < end && collapsed[start] == marker {
		start++
	}
	for end > start && collapsed[end-1] == marker {
		end--
	}
	if start == end {
		return nil
	}
	return strings.Split(string(collapsed[start:end]), string(marker))
}

// KebabCase converts s to kebab-case.
func KebabCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		words[i] = lowerString(w)
	}
	return strings.Join(words, "-")
}
