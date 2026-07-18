// Package changecase is a standard-library-only Go port of the popular npm
// package change-case (https://www.npmjs.com/package/change-case), which
// converts identifiers and phrases between the many casing conventions used in
// code and configuration. It exposes the full change-case family — CamelCase,
// PascalCase, SnakeCase, KebabCase (a.k.a. param case), ConstantCase, DotCase,
// PathCase, HeaderCase (train case), CapitalCase, SentenceCase and NoCase —
// built on a single word-splitting pass so every converter tokenises input the
// same way.
//
// Splitting follows change-case's rules: any run of separators (spaces,
// hyphens, underscores, dots and slashes) divides words, a lowercase- or
// digit-to-uppercase transition starts a new word (so "fooBar" splits into
// "foo" and "Bar"), and an uppercase run followed by a lowercase letter is
// treated as an acronym boundary (so "HTTPServer" splits into "HTTP" and
// "Server"). Digits attach to the preceding letters. The exported Words helper
// exposes this tokenisation directly.
//
// Each converter then lowercases the tokens and rejoins them with the
// appropriate delimiter and capitalisation, so ConstantCase("fooBar") is
// "FOO_BAR", KebabCase("fooBar") is "foo-bar" and PascalCase("foo_bar") is
// "FooBar". The functions are deterministic, ASCII- and Unicode-aware through
// the standard strings and unicode packages, and depend on no third-party code.
package changecase

import (
	"strings"
	"unicode"
)

// Words splits s into its constituent word tokens using change-case's rules:
// separator runs, camelCase humps and acronym boundaries all begin new words,
// and digits stay attached to the letters they follow. The returned tokens
// retain their original letter case.
func Words(s string) []string {
	var words []string
	runes := []rune(s)
	n := len(runes)
	start := -1
	flush := func(end int) {
		if start >= 0 && end > start {
			words = append(words, string(runes[start:end]))
		}
		start = -1
	}
	isSep := func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '-' || r == '_' || r == '.' || r == '/' || r == '\\'
	}
	for i := 0; i < n; i++ {
		r := runes[i]
		if isSep(r) {
			flush(i)
			continue
		}
		if start < 0 {
			start = i
			continue
		}
		prev := runes[i-1]
		// lower/digit -> Upper begins a new word
		if unicode.IsUpper(r) && (unicode.IsLower(prev) || unicode.IsDigit(prev)) {
			flush(i)
			start = i
			continue
		}
		// acronym boundary: Upper,Upper followed by lower -> break before this upper
		if unicode.IsUpper(prev) && unicode.IsUpper(r) && i+1 < n && unicode.IsLower(runes[i+1]) {
			flush(i)
			start = i
			continue
		}
	}
	flush(n)
	return words
}

func lowerWords(s string) []string {
	w := Words(s)
	for i, x := range w {
		w[i] = strings.ToLower(x)
	}
	return w
}

func title(w string) string {
	if w == "" {
		return w
	}
	r := []rune(w)
	r[0] = unicode.ToUpper(r[0])
	for i := 1; i < len(r); i++ {
		r[i] = unicode.ToLower(r[i])
	}
	return string(r)
}

// NoCase joins the lowercased words of s with single spaces, e.g.
// NoCase("fooBarBaz") is "foo bar baz".
func NoCase(s string) string {
	return strings.Join(lowerWords(s), " ")
}

// CamelCase converts s to camelCase (first word lowercase, subsequent words
// title-cased and joined without a delimiter), e.g. "foo-bar" -> "fooBar".
func CamelCase(s string) string {
	w := lowerWords(s)
	if len(w) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(w[0])
	for _, x := range w[1:] {
		b.WriteString(title(x))
	}
	return b.String()
}

// PascalCase converts s to PascalCase (every word title-cased, no delimiter),
// e.g. "foo_bar" -> "FooBar".
func PascalCase(s string) string {
	w := lowerWords(s)
	var b strings.Builder
	for _, x := range w {
		b.WriteString(title(x))
	}
	return b.String()
}

// SnakeCase converts s to snake_case (lowercase words joined by underscores).
func SnakeCase(s string) string {
	return strings.Join(lowerWords(s), "_")
}

// KebabCase converts s to kebab-case (lowercase words joined by hyphens). It is
// also known as param case.
func KebabCase(s string) string {
	return strings.Join(lowerWords(s), "-")
}

// ParamCase is an alias for KebabCase, matching change-case's paramCase name.
func ParamCase(s string) string { return KebabCase(s) }

// ConstantCase converts s to CONSTANT_CASE (uppercase words joined by
// underscores).
func ConstantCase(s string) string {
	w := lowerWords(s)
	for i, x := range w {
		w[i] = strings.ToUpper(x)
	}
	return strings.Join(w, "_")
}

// DotCase converts s to dot.case (lowercase words joined by dots).
func DotCase(s string) string {
	return strings.Join(lowerWords(s), ".")
}

// PathCase converts s to path/case (lowercase words joined by forward slashes).
func PathCase(s string) string {
	return strings.Join(lowerWords(s), "/")
}

// HeaderCase converts s to Header-Case (title-cased words joined by hyphens),
// also known as train case.
func HeaderCase(s string) string {
	w := lowerWords(s)
	for i, x := range w {
		w[i] = title(x)
	}
	return strings.Join(w, "-")
}

// CapitalCase converts s to Capital Case (title-cased words joined by spaces).
func CapitalCase(s string) string {
	w := lowerWords(s)
	for i, x := range w {
		w[i] = title(x)
	}
	return strings.Join(w, " ")
}

// SentenceCase converts s to Sentence case (only the first word capitalised,
// the rest lowercase, joined by spaces).
func SentenceCase(s string) string {
	w := lowerWords(s)
	if len(w) == 0 {
		return ""
	}
	w[0] = title(w[0])
	return strings.Join(w, " ")
}

// SwapCase returns s with the case of every letter inverted (upper becomes
// lower and vice versa), leaving non-letters untouched. It mirrors the npm
// swapcase package and does not tokenise the input.
func SwapCase(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsUpper(r):
			return unicode.ToLower(r)
		case unicode.IsLower(r):
			return unicode.ToUpper(r)
		default:
			return r
		}
	}, s)
}
