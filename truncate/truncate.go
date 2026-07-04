// Package truncate shortens a string to a maximum length, appending an
// ellipsis. It is a Go substitute inspired by the npm truncate package.
//
// The length is measured in runes and includes the length of the ellipsis, so
// the returned string never exceeds the requested length. Strings that are
// already short enough are returned unchanged.
package truncate

import "unicode"

// Options configures TruncateOpts.
type Options struct {
	// Ellipsis is the string appended to a truncated result. When empty,
	// no ellipsis is appended.
	Ellipsis string

	// WordBoundary, when true, trims the truncated content back to the last
	// word boundary (whitespace) so that words are not cut in half.
	WordBoundary bool
}

// Truncate shortens s to at most length runes, appending the default ellipsis
// "..." when truncation occurs. If s is already at most length runes long it is
// returned unchanged.
func Truncate(s string, length int) string {
	return TruncateOpts(s, length, Options{Ellipsis: "..."})
}

// TruncateOpts shortens s to at most length runes using the provided options.
// The ellipsis length counts toward the total length. If s already fits within
// length runes it is returned unchanged.
func TruncateOpts(s string, length int, opts Options) string {
	runes := []rune(s)
	if len(runes) <= length {
		return s
	}

	ellipsisRunes := []rune(opts.Ellipsis)
	keep := length - len(ellipsisRunes)
	if keep < 0 {
		keep = 0
	}

	content := runes[:keep]

	if opts.WordBoundary {
		idx := lastSpaceIndex(content)
		if idx >= 0 {
			content = content[:idx]
		}
	}

	return string(content) + opts.Ellipsis
}

// lastSpaceIndex returns the index of the last whitespace rune in rs, or -1 if
// there is none.
func lastSpaceIndex(rs []rune) int {
	for i := len(rs) - 1; i >= 0; i-- {
		if unicode.IsSpace(rs[i]) {
			return i
		}
	}
	return -1
}
