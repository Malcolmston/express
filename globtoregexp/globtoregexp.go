// Package globtoregexp converts glob patterns into regular expressions. It is a
// faithful port of the npm package glob-to-regexp, translating shell-style glob
// syntax into a compiled *regexp.Regexp that can be used for matching file
// paths, keys, or any other strings.
//
// Reach for this package when you have glob-style patterns (from configuration,
// user input, or a manifest) and want to match strings against them using Go's
// standard regexp engine. Rather than implementing a bespoke matcher, you
// convert the glob once and reuse the resulting *regexp.Regexp. GlobToRegExp
// uses the library's default options; GlobToRegExpOpts accepts an Options value
// for finer control over extended syntax, globstar behavior, and RegExp flags.
//
// Conversion walks the pattern rune by rune. Regular expression metacharacters
// such as ".", "(", ")", "+", "^", "$", "|", and "/" are escaped so they match
// literally. A run of "*" is the interesting case: with GlobStar off (the
// default) any run of stars becomes ".*", which matches everything including
// path separators; with GlobStar on, a single "*" becomes "([^/]*)" and matches
// within a single path segment, while a "**" segment bounded by "/" or the ends
// of the string becomes "((?:[^/]*(?:/|$))*)" and matches across zero or more
// segments. Unless the "g" flag is present the result is anchored with "^" and
// "$" so the pattern must match the whole input.
//
// The Extended option unlocks additional glob features that are otherwise
// treated literally: "?" matches any single character, "[...]" character
// classes are passed through to the regexp unchanged, and "{a,b}" groups become
// "(a|b)" alternations. With Extended off, those characters are all escaped and
// match themselves. The Flags field carries JavaScript-style RegExp flags:
// "i" produces a case-insensitive pattern (via a "(?i)" prefix) and "g"
// suppresses the "^...$" anchoring so the pattern may match a substring. Any
// other flag characters are ignored. Both functions return the compiled regexp
// and any error from regexp.Compile.
//
// Behavior mirrors the npm glob-to-regexp module, including its defaults
// (Extended off, GlobStar off), its globstar boundary rules, and its treatment
// of the "i" and "g" flags. The chief adaptation to Go is that anchoring and
// the "i" flag are expressed with Go's "^", "$", and "(?i)" regexp syntax
// instead of JavaScript's native RegExp flag object, and errors are surfaced as
// Go error values rather than thrown exceptions. The generated pattern strings
// are otherwise identical to those the JavaScript library produces.
package globtoregexp

import (
	"regexp"
	"strings"
)

// Options configures how a glob pattern is converted to a regular expression.
type Options struct {
	// Extended enables extended glob features: "?" matches a single
	// character, "[...]" character classes are passed through, and
	// "{a,b}" groups become alternations.
	Extended bool

	// GlobStar controls how "*" is translated. When false (the default,
	// matching the npm library), any run of "*" is translated to ".*" and
	// matches everything including path separators. When true, a single
	// "*" matches any run of non-separator characters ([^/]*) and a "**"
	// segment matches across separators.
	GlobStar bool

	// Flags holds JavaScript-style RegExp flags. "i" enables
	// case-insensitive matching and "g" disables ^...$ anchoring. Other
	// flags are ignored.
	Flags string
}

// GlobToRegExp converts a glob pattern to a *regexp.Regexp using the default
// options (Extended off, GlobStar off), matching the default behavior of the
// npm glob-to-regexp library.
func GlobToRegExp(glob string) (*regexp.Regexp, error) {
	return GlobToRegExpOpts(glob, Options{})
}

// GlobToRegExpOpts converts a glob pattern to a *regexp.Regexp using the
// provided options.
func GlobToRegExpOpts(glob string, opts Options) (*regexp.Regexp, error) {
	extended := opts.Extended
	globstar := opts.GlobStar
	flags := opts.Flags
	inGroup := false

	var reStr strings.Builder
	runes := []rune(glob)

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '/', '$', '^', '+', '.', '(', ')', '=', '!', '|':
			reStr.WriteByte('\\')
			reStr.WriteRune(c)

		case '?':
			if extended {
				reStr.WriteByte('.')
			} else {
				reStr.WriteByte('\\')
				reStr.WriteRune(c)
			}

		case '[', ']':
			if extended {
				reStr.WriteRune(c)
			} else {
				reStr.WriteByte('\\')
				reStr.WriteRune(c)
			}

		case '{':
			if extended {
				inGroup = true
				reStr.WriteByte('(')
			} else {
				reStr.WriteByte('\\')
				reStr.WriteRune(c)
			}

		case '}':
			if extended {
				inGroup = false
				reStr.WriteByte(')')
			} else {
				reStr.WriteByte('\\')
				reStr.WriteRune(c)
			}

		case ',':
			if inGroup {
				reStr.WriteByte('|')
			} else {
				reStr.WriteByte('\\')
				reStr.WriteRune(c)
			}

		case '*':
			// Capture the character preceding this run of stars.
			prevChar := rune(-1)
			if i-1 >= 0 {
				prevChar = runes[i-1]
			}
			// Consume all consecutive stars.
			starCount := 1
			for i+1 < len(runes) && runes[i+1] == '*' {
				starCount++
				i++
			}
			nextChar := rune(-1)
			if i+1 < len(runes) {
				nextChar = runes[i+1]
			}

			if !globstar {
				// GlobStar disabled: any number of stars becomes ".*".
				reStr.WriteString(".*")
			} else {
				isGlobstar := starCount > 1 &&
					(prevChar == '/' || prevChar == rune(-1)) &&
					(nextChar == '/' || nextChar == rune(-1))
				if isGlobstar {
					// Match zero or more path segments.
					reStr.WriteString("((?:[^/]*(?:/|$))*)")
					i++ // move over the trailing "/"
				} else {
					// Match a single path segment.
					reStr.WriteString("([^/]*)")
				}
			}

		default:
			reStr.WriteRune(c)
		}
	}

	result := reStr.String()
	if flags == "" || !strings.Contains(flags, "g") {
		result = "^" + result + "$"
	}
	if strings.Contains(flags, "i") {
		result = "(?i)" + result
	}
	return regexp.Compile(result)
}
