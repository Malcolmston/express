package escaperegexp_test

import (
	"fmt"
	"regexp"

	"github.com/malcolmston/express/escaperegexp"
)

// ExampleEscapeRegExp shows how a string containing regex metacharacters is
// turned into a pattern that matches those characters literally. The input here
// contains ".", "*", and parentheses, every one of which carries special meaning
// to a regular-expression engine and would otherwise change what the pattern
// matches (or fail to compile). After escaping, each metacharacter is prefixed
// with a backslash so the value matches itself byte for byte. The escaped form
// is what you would interpolate into a larger pattern to search for the literal
// text safely.
func ExampleEscapeRegExp() {
	fmt.Println(escaperegexp.EscapeRegExp("a.b*c(1)"))
	// Output: a\.b\*c\(1\)
}

// ExampleEscapeRegExp_hyphen demonstrates the one special case in the escaping
// rules: the hyphen "-" is rendered as the hexadecimal escape \x2d rather than
// as "\-". This mirrors the upstream escape-string-regexp package exactly and
// keeps the output safe to interpolate inside a character class, where a bare
// hyphen would denote a range. Every other character in the input passes through
// unchanged. The result is a literal-matching fragment for the original text.
func ExampleEscapeRegExp_hyphen() {
	fmt.Println(escaperegexp.EscapeRegExp("foo-bar"))
	// Output: foo\x2dbar
}

// ExampleEscapeRegExp_compile ties the escaped output back to its purpose by
// compiling it into a real regexp and matching against literal input. A raw user
// term like "1+1=2?" contains "+" and "?" quantifiers that would break a naive
// pattern; escaping it first yields a fragment that regexp.Compile accepts and
// that matches only the exact string. The example confirms the compiled pattern
// matches the original text. This is the intended end-to-end use of the package:
// escape untrusted text, then embed it in a pattern.
func ExampleEscapeRegExp_compile() {
	re := regexp.MustCompile(escaperegexp.EscapeRegExp("1+1=2?"))
	fmt.Println(re.MatchString("1+1=2?"))
	// Output: true
}
