// Package camelcase converts dash/dot/underscore/space separated (and
// mixed-case) strings into camelCase or PascalCase, modeled on the npm
// "camelcase" package.
//
// Delimiters (spaces, hyphens, underscores, and any other non-alphanumeric
// characters) separate words. Case boundaries are also treated as word
// boundaries so that "fooBar" splits into "foo" and "Bar". Runs of consecutive
// uppercase letters are lowercased by default (for example "FOOBar" becomes
// "fooBar"). Leading and trailing separators are ignored, sequences of
// separators collapse, and embedded numbers are preserved.
package camelcase

import (
	"strings"
	"unicode"
)

// Options configures how CamelCaseWith transforms a string.
type Options struct {
	// Pascal, when true, upper-cases the first character of the result so the
	// output is PascalCase (for example "FooBarBaz") instead of camelCase.
	Pascal bool
}

// CamelCase converts s to camelCase. The first word is lower-cased and every
// subsequent word is title-cased, for example "foo-bar_baz", "foo bar" and
// "Foo Bar" all become "fooBar"/"fooBarBaz".
func CamelCase(s string) string {
	return CamelCaseWith(s, Options{})
}

// PascalCase converts s to PascalCase (also called UpperCamelCase). Every word,
// including the first, is title-cased, for example "foo-bar" becomes "FooBar".
func PascalCase(s string) string {
	return CamelCaseWith(s, Options{Pascal: true})
}

// CamelCaseWith converts s to camelCase or PascalCase according to opts.
func CamelCaseWith(s string, opts Options) string {
	words := splitWords(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	for i, w := range words {
		if i == 0 && !opts.Pascal {
			b.WriteString(strings.ToLower(w))
		} else {
			b.WriteString(capitalize(w))
		}
	}
	return b.String()
}

// splitWords breaks s into words on delimiters and case boundaries, preserving
// the original case of each word for later normalization.
func splitWords(s string) []string {
	runes := []rune(s)
	var words []string
	var cur []rune

	flush := func() {
		if len(cur) > 0 {
			words = append(words, string(cur))
			cur = cur[:0]
		}
	}

	isDelim := func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if isDelim(r) {
			flush()
			continue
		}
		if len(cur) > 0 {
			prev := cur[len(cur)-1]
			switch {
			case unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev)):
				// lower/digit -> upper marks a new word: "fooBar" -> foo, Bar
				flush()
			case unicode.IsUpper(prev) && unicode.IsUpper(r) &&
				i+1 < len(runes) && unicode.IsLower(runes[i+1]):
				// UPPER UPPER lower: the last upper starts the next word,
				// e.g. "HTTPServer" -> HTTP, Server and "FOOBar" -> FOO, Bar.
				flush()
			}
		}
		cur = append(cur, r)
	}
	flush()
	return words
}

// capitalize lower-cases the whole word and then upper-cases its first letter.
func capitalize(w string) string {
	r := []rune(strings.ToLower(w))
	if len(r) == 0 {
		return ""
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
