// Package camelcase converts dash/dot/underscore/space separated (and
// mixed-case) strings into camelCase or PascalCase, modeled on the npm
// "camelcase" package (sindresorhus/camelcase v9) that Express-adjacent tooling
// uses to normalize option keys, identifiers, and header-derived names. It
// exposes CamelCase for the default lower-first form, PascalCase for the
// upper-first form, and CamelCaseWith for callers that want to choose the style
// through an Options value, all built on only the Go standard library.
//
// Reach for this package whenever human-written or wire-format names have to be
// turned into program-friendly identifiers: converting "user-id" or "user_id"
// from a config file into "userId", mapping "Content-Type" style header names to
// struct-field-like tokens, or normalizing the several separator conventions a
// single input source might mix together. Keeping the word-splitting rules in one
// place means "foo-bar", "foo_bar", "foo bar", and "fooBar" all collapse to the
// same canonical output regardless of how the caller happened to write them.
//
// Word splitting is driven by two kinds of boundary, matching the npm original.
// Separators — and only these four characters: underscore, dot, hyphen, and
// space — separate words and are discarded. Any other non-letter character
// (punctuation such as "?", "@", "#", "$", ":", emoji, and so on) is NOT a
// separator: it is kept verbatim and does not start a new word, so "foo bar?"
// becomes "fooBar?" and "A::a" becomes "a::a". Case transitions are the second
// kind of boundary: a lower-case letter followed by an upper-case letter starts a
// new word (so "fooBar" splits into "foo" and "Bar"), and a run of upper-case
// letters immediately followed by a lower-case letter breaks just before that
// final upper-case letter (so "XMLHttpRequest" yields "xmlHttpRequest" and
// "FOOBar" yields "fooBar"). A run of digits also starts a new word for the
// letter that follows it, so "foo2bar" becomes "foo2Bar" and "hello1world"
// becomes "hello1World"; digits that are immediately followed by a separator do
// not force this boundary, so "b2b_registration" stays "b2bRegistration".
//
// The semantics around edges and degenerate input follow the npm package. Leading
// and trailing separators are ignored and consecutive separators collapse, so
// "--foo--bar--" yields "fooBar". Leading underscores and dollar signs are
// preserved because they carry semantic meaning (private/internal names, jQuery
// and observable conventions), so "_foo_bar" yields "_fooBar", "$foo_bar" yields
// "$fooBar", and an input consisting only of those characters is returned as-is,
// so "__" yields "__". An input that is empty or contains only separators yields
// "". Runs of consecutive upper-case letters are lower-cased by default rather
// than preserved (matching the non-preserveConsecutiveUppercase default), so
// "FOOBar" becomes "fooBar" and not "fOOBar". The transformation operates on
// runes, so multi-byte letters are handled without corrupting the encoding.
//
// The intentional differences from the Node original are idiomatic. The API is
// three exported Go functions plus an Options struct rather than a single
// JavaScript function with an options object, and this port implements only the
// package defaults of the more exotic npm flags: it does not offer the
// preserveConsecutiveUppercase toggle, locale-specific casing, a
// capitalizeAfterNumber:false mode, or array inputs. For plain string input with
// the default (or pascalCase) behavior, output matches the npm package.
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
	return camelCase(s, opts.Pascal)
}

// isSeparator reports whether r is one of the four characters the npm package
// treats as a word separator: underscore, dot, hyphen, or space.
func isSeparator(r rune) bool {
	return r == '_' || r == '.' || r == '-' || r == ' '
}

// isASCIIDigit reports whether r is an ASCII digit, matching JavaScript's \d.
func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// isIdentifier reports whether r matches the npm identifier class
// [\p{Alpha}\p{N}_]: a letter, a number, or an underscore.
func isIdentifier(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

func camelCase(input string, pascal bool) string {
	input = strings.TrimSpace(input)
	if len(input) == 0 {
		return ""
	}

	// Preserve leading _ and $ as they have semantic meaning.
	i := 0
	for i < len(input) && (input[i] == '_' || input[i] == '$') {
		i++
	}
	prefix := input[:i]
	input = input[i:]
	if len(input) == 0 {
		return prefix
	}

	runes := []rune(input)

	if len(runes) == 1 {
		r := runes[0]
		if isSeparator(r) {
			return prefix
		}
		if pascal {
			return prefix + string(unicode.ToUpper(r))
		}
		return prefix + string(unicode.ToLower(r))
	}

	// Insert separators at camelCase / ACRONYMWord boundaries.
	if input != strings.ToLower(input) {
		runes = preserveCamelCase(runes)
	}

	// Strip leading separators eagerly so they do not affect word detection.
	start := 0
	for start < len(runes) && isSeparator(runes[start]) {
		start++
	}
	runes = runes[start:]

	// Normalize base casing: lower-case everything (capitalizeAfterNumber is
	// true and preserveConsecutiveUppercase is false, matching the defaults).
	for idx := range runes {
		runes[idx] = unicode.ToLower(runes[idx])
	}

	if pascal && len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}

	return prefix + postProcess(runes)
}

// preserveCamelCase inserts '-' separators before the boundaries that a
// camelCase or ACRONYMWord run implies, mirroring the npm package's
// preserveCamelCase helper with preserveConsecutiveUppercase disabled.
func preserveCamelCase(s []rune) []rune {
	isLastCharLower := false
	isLastCharUpper := false
	isLastLastCharUpper := false

	for index := 0; index < len(s); index++ {
		ch := s[index]

		// Was the character three positions back an inserted separator? For the
		// first few positions this is treated as true, which suppresses the
		// ACRONYMWord split guard exactly as the npm implementation does.
		isLastLastCharPreserved := true
		if index > 2 {
			isLastLastCharPreserved = s[index-3] == '-'
		}

		switch {
		case isLastCharLower && unicode.IsUpper(ch):
			// fooBar -> foo-Bar
			s = insertRune(s, index, '-')
			isLastCharLower = false
			isLastLastCharUpper = isLastCharUpper
			isLastCharUpper = true
			index++
		case isLastCharUpper && isLastLastCharUpper && unicode.IsLower(ch) && !isLastLastCharPreserved:
			// FOOBar -> FOO-Bar
			s = insertRune(s, index-1, '-')
			isLastLastCharUpper = isLastCharUpper
			isLastCharUpper = false
			isLastCharLower = true
		default:
			lower := unicode.ToLower(ch)
			upper := unicode.ToUpper(ch)
			isLastCharLower = lower == ch && upper != ch
			isLastLastCharUpper = isLastCharUpper
			isLastCharUpper = upper == ch && lower != ch
		}
	}

	return s
}

// insertRune inserts r into s immediately before position pos.
func insertRune(s []rune, pos int, r rune) []rune {
	s = append(s, 0)
	copy(s[pos+1:], s[pos:])
	s[pos] = r
	return s
}

// postProcess collapses separators and applies the digit-boundary rule, matching
// the two-pass replace in the npm package's postProcess (capitalizeAfterNumber
// true): first uppercase the letter following a digit run, then drop separator
// runs and uppercase the following identifier character.
func postProcess(runes []rune) string {
	runes = numbersPass(runes)
	return separatorsPass(runes)
}

// numbersPass implements the NUMBERS_AND_IDENTIFIER replace: for every run of
// digits followed by a single identifier character, upper-case that character —
// unless the run is immediately followed by a separator, or the digits end the
// string.
func numbersPass(runes []rune) []rune {
	i := 0
	for i < len(runes) {
		if !isASCIIDigit(runes[i]) {
			i++
			continue
		}
		j := i
		for j < len(runes) && isASCIIDigit(runes[j]) {
			j++
		}
		if j < len(runes) && isIdentifier(runes[j]) {
			// match spans runes[i:j+1]; look at the char after the match.
			after := j + 1
			if after < len(runes) && isSeparator(runes[after]) {
				// Continued token: do not force a new word.
			} else {
				runes[j] = unicode.ToUpper(runes[j])
			}
			i = j + 1
		} else {
			// Digits at end (match is digits + empty) or followed by a
			// non-identifier: nothing to capitalize.
			i = j
		}
	}
	return runes
}

// separatorsPass implements the SEPARATORS_AND_IDENTIFIER replace: drop each run
// of separators and upper-case the identifier character that follows it. A
// separator run at the end of the string is dropped; a separator run followed by
// a non-identifier character is left in place.
func separatorsPass(runes []rune) string {
	var b strings.Builder
	i := 0
	for i < len(runes) {
		if !isSeparator(runes[i]) {
			b.WriteRune(runes[i])
			i++
			continue
		}
		j := i
		for j < len(runes) && isSeparator(runes[j]) {
			j++
		}
		switch {
		case j < len(runes) && isIdentifier(runes[j]):
			b.WriteRune(unicode.ToUpper(runes[j]))
			i = j + 1
		case j >= len(runes):
			// Trailing separators: dropped.
			i = j
		default:
			// Followed by a non-identifier: no match, keep separators as-is.
			for k := i; k < j; k++ {
				b.WriteRune(runes[k])
			}
			i = j
		}
	}
	return b.String()
}
