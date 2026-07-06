// Package titlecase converts strings to Title Case, capitalizing the first
// letter of each word. It is useful for normalizing headings, labels, names,
// and other short display strings into a consistent, human-friendly form, for
// example turning "hello world" or "helloWorld" into "Hello World".
//
// The package is inspired by the npm title-case package from the change-case
// family. Like that library it treats the input as a sequence of words rather
// than performing a naive character replacement: the string is first split into
// words, each word is recased independently, and the words are rejoined. This
// makes the behavior robust across the many separators and casing styles that
// real-world identifiers and prose use.
//
// Word splitting happens in splitWords and recognizes three kinds of
// boundaries. Runs of non-alphanumeric characters (whitespace, punctuation,
// underscores, hyphens) separate words, so "foo_bar-baz" yields three words. A
// lower-case or digit followed by an upper-case letter is a camelCase boundary,
// so "helloWorld" splits into "hello" and "World". An upper-case letter
// followed by an upper-case then lower-case letter is an acronym boundary, so
// "XMLHttp" splits into "XML" and "Http". Unicode letters and digits are
// recognized via the unicode package, so accented and non-ASCII words are
// handled.
//
// Each word is then normalized by capitalize: the first rune is upper-cased and
// every remaining rune is lower-cased. The recased words are joined with single
// spaces regardless of the original separators, so mixed or repeated separators
// collapse to one space. A consequence worth noting is that this package does
// not implement small-word handling: unlike some English title-case styles that
// leave short function words such as "a", "an", "of", or "the" in lower case,
// every word here is capitalized, and acronyms are folded to initial-capital
// form ("XML" becomes "Xml") because trailing letters are always lower-cased.
//
// Parity with the Node library is at the level of the split-and-recase model
// and the common camelCase, PascalCase, and separator cases. The exact acronym
// and small-word treatment of individual change-case releases is not
// replicated, so callers who need a specific style guide's rules for short
// words or preserved acronyms should post-process the result.
package titlecase

import (
	"strings"
	"unicode"
)

// TitleCase converts s to Title Case. The string is split into words on
// whitespace, punctuation, and camelCase/PascalCase boundaries; each word is
// then capitalized (first letter upper, rest lower) and the words are joined
// with single spaces.
//
// For example:
//
//	TitleCase("a simple test")   == "A Simple Test"
//	TitleCase("hello world")     == "Hello World"
//	TitleCase("helloWorld")      == "Hello World"
//	TitleCase("foo_bar-baz")     == "Foo Bar Baz"
func TitleCase(s string) string {
	words := splitWords(s)
	for i, w := range words {
		words[i] = capitalize(w)
	}
	return strings.Join(words, " ")
}

// splitWords splits s into words on non-alphanumeric separators and on
// camelCase/PascalCase boundaries.
func splitWords(s string) []string {
	var words []string
	var current []rune
	runes := []rune(s)

	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0:0]
		}
	}

	isAlnum := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}

	for i, r := range runes {
		if !isAlnum(r) {
			flush()
			continue
		}
		if len(current) > 0 {
			prev := current[len(current)-1]
			switch {
			case (unicode.IsLower(prev) || unicode.IsDigit(prev)) && unicode.IsUpper(r):
				// lower/digit followed by upper: "fooBar" -> "foo" | "Bar"
				flush()
			case unicode.IsUpper(prev) && unicode.IsUpper(r) &&
				i+1 < len(runes) && unicode.IsLower(runes[i+1]):
				// upper followed by upper+lower: "XMLHttp" -> "XML" | "Http"
				flush()
			}
		}
		current = append(current, r)
	}
	flush()
	return words
}

// capitalize returns w with its first rune uppercased and the rest lowercased.
func capitalize(w string) string {
	rs := []rune(w)
	if len(rs) == 0 {
		return w
	}
	rs[0] = unicode.ToUpper(rs[0])
	for i := 1; i < len(rs); i++ {
		rs[i] = unicode.ToLower(rs[i])
	}
	return string(rs)
}
