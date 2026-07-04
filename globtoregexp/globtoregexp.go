// Package globtoregexp converts glob patterns into regular expressions.
//
// It is a faithful port of the npm package glob-to-regexp. The default
// behavior matches that library: extended globs are OFF and globstar is OFF,
// which means any run of "*" is translated to ".*" and therefore matches
// everything including path separators. Enable GlobStar in Options to make a
// single "*" match only a single path segment and "**" match across segments.
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
