// Package wordwrap wraps text to a fixed width, modeled on the npm "word-wrap"
// package.
//
// Text is broken on whitespace so that no line exceeds Width characters
// (measured excluding the indent). Each output line is prefixed with Indent.
// Existing newlines in the input are preserved as paragraph breaks. When Cut is
// true, words longer than Width are broken into Width-sized pieces instead of
// overflowing.
package wordwrap

import (
	"strings"
	"unicode/utf8"
)

// Default option values, matching the npm word-wrap defaults.
const (
	DefaultWidth   = 50
	DefaultIndent  = "  "
	DefaultNewline = "\n"
)

// Options configures Wrap.
//
// A zero-value Options is usable: Width <= 0 is treated as DefaultWidth and an
// empty Newline is treated as DefaultNewline. Note that a zero-value Indent is
// the empty string (no indentation); to obtain the two-space default indent,
// build options with NewOptions rather than a bare Options literal.
type Options struct {
	// Width is the maximum line length excluding the indent. Values <= 0 are
	// treated as DefaultWidth (50).
	Width int
	// Indent is prefixed to every output line. The zero value is no indent.
	Indent string
	// Newline is the separator placed between output lines. An empty value is
	// treated as DefaultNewline ("\n").
	Newline string
	// TrimTrailing removes trailing spaces and tabs from every output line when
	// true.
	TrimTrailing bool
	// Cut breaks words longer than Width into Width-sized pieces when true.
	Cut bool
}

// NewOptions returns an Options populated with the package defaults: Width 50,
// Indent two spaces, and Newline "\n".
func NewOptions() Options {
	return Options{
		Width:   DefaultWidth,
		Indent:  DefaultIndent,
		Newline: DefaultNewline,
	}
}

// Wrap wraps text according to opts and returns the wrapped result.
func Wrap(text string, opts Options) string {
	width := opts.Width
	if width <= 0 {
		width = DefaultWidth
	}
	newline := opts.Newline
	if newline == "" {
		newline = DefaultNewline
	}
	indent := opts.Indent

	var lines []string
	for _, para := range strings.Split(text, "\n") {
		lines = append(lines, wrapParagraph(para, width, indent, opts.Cut)...)
	}

	if opts.TrimTrailing {
		for i := range lines {
			lines[i] = strings.TrimRight(lines[i], " \t")
		}
	}
	return strings.Join(lines, newline)
}

// wrapParagraph greedily wraps a single paragraph (no embedded newlines) and
// returns its output lines, each already prefixed with indent.
func wrapParagraph(para string, width int, indent string, cut bool) []string {
	words := strings.Fields(para)
	if len(words) == 0 {
		// Preserve blank lines as an (indented) empty line.
		return []string{indent}
	}
	if cut {
		words = breakLongWords(words, width)
	}

	var lines []string
	cur := ""
	for _, w := range words {
		switch {
		case cur == "":
			cur = w
		case utf8.RuneCountInString(cur)+1+utf8.RuneCountInString(w) <= width:
			cur += " " + w
		default:
			lines = append(lines, indent+cur)
			cur = w
		}
	}
	lines = append(lines, indent+cur)
	return lines
}

// breakLongWords splits any word longer than width into width-sized chunks.
func breakLongWords(words []string, width int) []string {
	out := make([]string, 0, len(words))
	for _, w := range words {
		r := []rune(w)
		for len(r) > width {
			out = append(out, string(r[:width]))
			r = r[width:]
		}
		out = append(out, string(r))
	}
	return out
}
