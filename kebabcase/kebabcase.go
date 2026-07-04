// Package kebabcase converts strings to kebab-case, modeled on the npm
// "kebab-case" package.
//
// A hyphen is inserted before any uppercase letter that follows a lowercase
// letter or a digit, everything is lower-cased, spaces and underscores become
// hyphens, repeated hyphens collapse into one, and leading/trailing hyphens are
// trimmed. For example "fooBar Baz" becomes "foo-bar-baz".
package kebabcase

import (
	"strings"
	"unicode"
)

// KebabCase converts s to kebab-case.
func KebabCase(s string) string {
	runes := []rune(s)
	var b strings.Builder
	for i, r := range runes {
		switch {
		case r == ' ' || r == '_' || r == '-':
			b.WriteRune('-')
		case unicode.IsUpper(r):
			if i > 0 {
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteRune('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return trimAndCollapse(b.String())
}

// trimAndCollapse collapses runs of hyphens into a single hyphen and removes any
// leading or trailing hyphens.
func trimAndCollapse(s string) string {
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}
