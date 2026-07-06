// Package escaperegexp escapes regular-expression metacharacters in a string so
// the string can be embedded as a literal inside a larger pattern. It is a
// faithful, standard-library-only Go port of the npm package
// escape-string-regexp, exposing a single function, EscapeRegExp, that takes an
// arbitrary string and returns a version in which every character that would
// otherwise carry special meaning to a regex engine has been neutralized with a
// backslash.
//
// The problem it solves appears whenever user- or data-derived text has to be
// spliced into a pattern: a search term typed into a box, a filename, a
// delimiter read from configuration. Concatenating such a value directly into a
// regular expression is both a correctness bug (a "." in the input would match
// any character, "(" would open an unbalanced group and fail to compile) and a
// denial-of-service risk (a crafted value can build a catastrophically
// backtracking pattern). Escaping the value first guarantees it is matched byte
// for byte, exactly as written.
//
// EscapeRegExp works by scanning the input rune by rune and emitting each
// character unchanged unless it is one of the regex metacharacters, in which
// case a leading backslash is written before it. Because the routine iterates
// over runes and copies non-special characters verbatim, multibyte UTF-8 text
// passes through untouched; only the ASCII metacharacters are ever rewritten.
// The result is assembled with a strings.Builder pre-sized to the input length
// to avoid repeated allocations for the common case where little or no escaping
// is needed.
//
// The set of escaped characters is |\{}()[]^$+*?. — each prefixed with a
// backslash — with one special case: the hyphen "-" is rendered as the escape
// sequence \x2d rather than \-. That mirrors the upstream package exactly and
// exists because an unescaped "-" is harmless outside a character class but
// meaningful inside one; emitting the hexadecimal escape keeps the output safe
// to interpolate into either position. The transformation is purely textual and
// never validates or compiles the result, so callers remain responsible for
// assembling and compiling the final pattern.
//
// Compared with the Node original the behavior is intended to be identical: the
// same metacharacter set is escaped and "-" receives the same \x2d treatment,
// so a string escaped here and one escaped by escape-string-regexp produce the
// same literal match. The chief differences are idiomatic rather than semantic:
// this port operates on Go strings and runes instead of JavaScript strings, and
// it performs no Unicode normalization, leaving the input's byte content
// otherwise unchanged.
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
