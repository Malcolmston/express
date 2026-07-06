// Package camelcase converts dash/dot/underscore/space separated (and
// mixed-case) strings into camelCase or PascalCase, modeled on the npm
// "camelcase" package that Express-adjacent tooling uses to normalize option
// keys, identifiers, and header-derived names. It exposes CamelCase for the
// default lower-first form, PascalCase for the upper-first form, and
// CamelCaseWith for callers that want to choose the style through an Options
// value, all built on only the Go standard library.
//
// Reach for this package whenever human-written or wire-format names have to be
// turned into program-friendly identifiers: converting "user-id" or "user_id"
// from a config file into "userId", mapping "Content-Type" style header names to
// struct-field-like tokens, or normalizing the many separator conventions a
// single input source might mix together. Keeping the word-splitting rules in one
// place means "foo-bar", "foo_bar", "foo bar", and "fooBar" all collapse to the
// same canonical output regardless of how the caller happened to write them.
//
// Word splitting is driven by two kinds of boundary. Delimiters — spaces,
// hyphens, underscores, and every other character that is neither a letter nor a
// digit — separate words and are discarded. Case transitions are also boundaries:
// a lower-case or digit followed by an upper-case letter starts a new word (so
// "fooBar" splits into "foo" and "Bar"), and a run of upper-case letters
// immediately followed by a lower-case letter breaks just before that final
// upper-case letter (so "HTTPServer" splits into "HTTP" and "Server", and
// "FOOBar" into "FOO" and "Bar"). Each recovered word is then normalized by
// lower-casing it entirely and, for every word after the first in camelCase or
// every word in PascalCase, upper-casing its leading rune.
//
// The semantics around edges and degenerate input are deliberate. Leading and
// trailing separators are ignored and consecutive separators collapse, so
// "__foo__bar__" yields "fooBar"; an input that contains no letters or digits
// (or the empty string) yields "". Runs of consecutive upper-case letters are
// lower-cased by default rather than preserved, matching the npm package's
// non-preserveConsecutiveUppercase default, so "FOOBar" becomes "fooBar" and not
// "fOOBar". Embedded digits are preserved and treated as ordinary in-word
// characters, so "foo2bar" stays "foo2bar" rather than gaining a spurious word
// boundary at the digit. The transformation operates on runes, so multi-byte
// letters are handled without corrupting the encoding.
//
// Parity with the Node original covers the common cases: the same separator set,
// the same case-boundary detection, and the same lower-first/upper-first output
// for typical identifiers. The intentional differences are idiomatic. The API is
// three exported Go functions plus an Options struct rather than a single
// JavaScript function with an options object, and this port implements only the
// default behavior of the more exotic npm flags (it does not offer
// locale-specific casing or a preserveConsecutiveUppercase toggle). Where a
// choice had to be made the default lodash/camelcase behavior is followed so that
// results line up with what Express developers expect for ordinary ASCII input.
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
