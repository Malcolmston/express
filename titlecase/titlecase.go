// Package titlecase converts strings to Title Case.
//
// It is inspired by the npm title-case package from the change-case family:
// the input is split into words on whitespace, punctuation, and camelCase
// boundaries, and each resulting word has its first letter capitalized and the
// remaining letters lowercased. Words are then rejoined with single spaces.
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
