// Package truncate shortens a string to a maximum length, appending an
// ellipsis when content is cut. It is a Go substitute inspired by the npm
// truncate package and the closely related lodash _.truncate helper, and is
// meant for producing fixed-width previews such as list summaries, table cells,
// or meta descriptions where an overly long value must be clipped.
//
// Length is measured in runes rather than bytes, so multibyte characters count
// as one unit and a truncated string is never split in the middle of a rune.
// The length budget is inclusive of the ellipsis: the ellipsis length is
// subtracted from the requested length before content is taken, so the returned
// string never exceeds the requested rune count. This matches the lodash
// convention where the omission string occupies part of the total length rather
// than being added on top of it.
//
// The algorithm is straightforward. If the input already fits within length
// runes it is returned unchanged, with no ellipsis appended. Otherwise the
// string is cut to (length - len(ellipsis)) runes; if the ellipsis alone is as
// long as or longer than the requested length the kept-content count clamps to
// zero, yielding just the ellipsis (or as much of the budget as remains). The
// kept content and the ellipsis are then concatenated.
//
// Options tunes two aspects. Ellipsis sets the string appended on truncation
// and may be empty to append nothing, in which case the string is simply cut to
// length runes. WordBoundary, when true, trims the cut content back to the last
// whitespace rune so a word is not sliced in half, mirroring lodash's
// separator behavior for whitespace; if there is no whitespace in the kept
// span the content is left as-is. The convenience Truncate function applies the
// default ellipsis "..." with word-boundary trimming disabled.
//
// Parity with the Node libraries covers the length, omission, and
// word-boundary features that callers most often use. It does not implement
// lodash's arbitrary regular-expression separator, HTML-aware truncation, or
// byte-length modes; the whitespace word boundary is the only boundary rule
// provided, and length is always counted in runes.
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
