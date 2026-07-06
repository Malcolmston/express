// Package wordwrap wraps text to a fixed width, modeled on the npm "word-wrap"
// package. It is a standard-library-only Go port that reflows a paragraph of
// text so that no output line exceeds a chosen width, which is useful for
// rendering help text, log messages, email bodies, terminal output, and any
// other place where a long run of prose must fit inside a bounded column.
//
// Wrapping is greedy and whitespace-based. The input is split into words on
// runs of whitespace, and words are packed onto the current line one at a time
// until adding the next word (plus the single space that would join it) would
// push the line past Width; at that point the current line is emitted and the
// word starts a new line. Width is measured in runes and counts only the text,
// excluding the per-line Indent, so multibyte and accented characters count as
// one column each and the indent does not eat into the usable width.
//
// Line structure is controlled by three fields. Indent is prefixed to every
// output line, which makes it easy to produce block-quoted or nested text.
// Newline is the separator placed between lines, defaulting to "\n". Existing
// newlines in the input are treated as hard paragraph breaks: the text is split
// on "\n" first and each paragraph is wrapped independently, so intentional line
// breaks in the source survive while over-long lines within a paragraph are
// reflowed. A blank paragraph is preserved as an indented empty line.
//
// Long words are handled by the Cut option. Normally a single word longer than
// Width is left intact and simply overflows its line, because breaking a word
// changes its meaning; when Cut is true such words are instead chopped into
// Width-sized pieces so the hard width limit is never exceeded. The TrimTrailing
// option removes trailing spaces and tabs from every emitted line, which is
// handy when the wrapped text is compared byte-for-byte or embedded where
// trailing whitespace is undesirable.
//
// Configuration is via the Options struct, and the zero value is usable: a Width
// of zero or less is treated as DefaultWidth (50) and an empty Newline as
// DefaultNewline ("\n"). Note the one subtlety with defaults — a zero-value
// Options has an empty Indent, whereas the npm library defaults to a two-space
// indent, so NewOptions is provided to build an Options carrying all three
// package defaults (Width 50, Indent two spaces, Newline "\n"). Parity with the
// Node original covers the greedy whitespace wrapping, indent, newline, trailing
// trim, and cut behavior; callers wanting the JavaScript defaults exactly should
// start from NewOptions rather than a bare struct literal.
package wordwrap

import (
	"strings"
	"unicode/utf8"
)

// Default option values, matching the npm word-wrap defaults.
const (
	// DefaultWidth is the default maximum line width in characters.
	DefaultWidth = 50
	// DefaultIndent is the default per-line indent prefix.
	DefaultIndent = "  "
	// DefaultNewline is the default line separator used in output.
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
