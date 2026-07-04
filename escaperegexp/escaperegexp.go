// Package escaperegexp escapes regular-expression metacharacters in a string
// so the string can be used as a literal inside a regular expression.
//
// It is a faithful port of the npm package escape-string-regexp.
package escaperegexp

import "strings"

// EscapeRegExp escapes the characters |\{}()[]^$+*?. by prefixing them with a
// backslash, and escapes - as \x2d, so that the returned string matches the
// input literally when used inside a regular expression.
//
// This mirrors the behavior of the npm escape-string-regexp package, which
// replaces the set [|\\{}()[\]^$+*?.] with "\\$&" and replaces - with "\\x2d".
func EscapeRegExp(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch r {
		case '|', '\\', '{', '}', '(', ')', '[', ']', '^', '$', '+', '*', '?', '.':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '-':
			b.WriteString(`\x2d`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
