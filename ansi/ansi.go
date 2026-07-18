// Package ansi is a standard-library-only Go port of the popular npm terminal
// styling library chalk (https://www.npmjs.com/package/chalk), which Node CLIs
// and Express dev tooling use to colour and format terminal output. It emits
// ANSI SGR (Select Graphic Rendition) escape sequences for the sixteen basic
// colours, the common text attributes (bold, dim, italic, underline, inverse,
// strikethrough), 24-bit truecolor via RGB/Hex, and provides Strip and
// VisibleWidth to measure or remove styling.
//
// Two styles of API are offered, matching chalk's ergonomics. The package-level
// convenience functions wrap a string in a single attribute or colour and reset
// afterwards, so ansi.Red("err") or ansi.Bold("go") returns a ready-to-print
// string. For chalk-style composition, Style is an immutable builder: each
// method returns a new Style with an added SGR code, and Apply renders a string
// with the accumulated codes, so New().Bold().Red().Apply("x") produces bold red
// text. Because Style values are immutable, a base style can be reused and
// extended without interfering across call sites.
//
// Every wrapper appends a full reset (ESC[0m) after the text, so styles never
// leak into later output, and Strip removes any CSI escape sequence so styled
// text can be logged or measured. VisibleWidth reports the number of runes that
// remain after stripping, which is the printed column width for ordinary text.
// The package depends only on fmt, regexp, strings and unicode/utf8 and is fully
// deterministic — it always emits escape codes regardless of whether the output
// is a terminal, leaving that decision to the caller.
package ansi

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const ansiReset = "\x1b[0m"

// SGR attribute codes.
const (
	codeBold          = 1
	codeDim           = 2
	codeItalic        = 3
	codeUnderline     = 4
	codeInverse       = 7
	codeStrikethrough = 9
)

// Style is an immutable set of ANSI SGR codes. The zero Style applies no
// styling. Build one with New and the chainable methods, then render text with
// Apply. Each method returns a new Style, leaving the receiver unchanged.
type Style struct {
	codes []int
}

// New returns an empty Style with no codes.
func New() Style { return Style{} }

func (s Style) with(c ...int) Style {
	next := make([]int, len(s.codes), len(s.codes)+len(c))
	copy(next, s.codes)
	next = append(next, c...)
	return Style{codes: next}
}

// Apply renders text with the style's accumulated SGR codes, appending a reset.
// A style with no codes returns text unchanged.
func (s Style) Apply(text string) string {
	if len(s.codes) == 0 {
		return text
	}
	parts := make([]string, len(s.codes))
	for i, c := range s.codes {
		parts[i] = strconv.Itoa(c)
	}
	return "\x1b[" + strings.Join(parts, ";") + "m" + text + ansiReset
}

// Bold returns a new Style with the bold attribute added.
func (s Style) Bold() Style { return s.with(codeBold) }

// Dim returns a new Style with the dim (faint) attribute added.
func (s Style) Dim() Style { return s.with(codeDim) }

// Italic returns a new Style with the italic attribute added.
func (s Style) Italic() Style { return s.with(codeItalic) }

// Underline returns a new Style with the underline attribute added.
func (s Style) Underline() Style { return s.with(codeUnderline) }

// Inverse returns a new Style with the inverse (reverse video) attribute added.
func (s Style) Inverse() Style { return s.with(codeInverse) }

// Strikethrough returns a new Style with the strikethrough attribute added.
func (s Style) Strikethrough() Style { return s.with(codeStrikethrough) }

// Black returns a new Style with a black foreground.
func (s Style) Black() Style { return s.with(30) }

// Red returns a new Style with a red foreground.
func (s Style) Red() Style { return s.with(31) }

// Green returns a new Style with a green foreground.
func (s Style) Green() Style { return s.with(32) }

// Yellow returns a new Style with a yellow foreground.
func (s Style) Yellow() Style { return s.with(33) }

// Blue returns a new Style with a blue foreground.
func (s Style) Blue() Style { return s.with(34) }

// Magenta returns a new Style with a magenta foreground.
func (s Style) Magenta() Style { return s.with(35) }

// Cyan returns a new Style with a cyan foreground.
func (s Style) Cyan() Style { return s.with(36) }

// White returns a new Style with a white foreground.
func (s Style) White() Style { return s.with(37) }

// Gray returns a new Style with a gray (bright black) foreground.
func (s Style) Gray() Style { return s.with(90) }

// BgBlack returns a new Style with a black background.
func (s Style) BgBlack() Style { return s.with(40) }

// BgRed returns a new Style with a red background.
func (s Style) BgRed() Style { return s.with(41) }

// BgGreen returns a new Style with a green background.
func (s Style) BgGreen() Style { return s.with(42) }

// BgYellow returns a new Style with a yellow background.
func (s Style) BgYellow() Style { return s.with(43) }

// BgBlue returns a new Style with a blue background.
func (s Style) BgBlue() Style { return s.with(44) }

// BgMagenta returns a new Style with a magenta background.
func (s Style) BgMagenta() Style { return s.with(45) }

// BgCyan returns a new Style with a cyan background.
func (s Style) BgCyan() Style { return s.with(46) }

// BgWhite returns a new Style with a white background.
func (s Style) BgWhite() Style { return s.with(47) }

// RGB returns a new Style with a 24-bit truecolor foreground.
func (s Style) RGB(r, g, b uint8) Style {
	return s.with(38, 2, int(r), int(g), int(b))
}

// BgRGB returns a new Style with a 24-bit truecolor background.
func (s Style) BgRGB(r, g, b uint8) Style {
	return s.with(48, 2, int(r), int(g), int(b))
}

// --- package-level convenience wrappers -------------------------------------

// Bold wraps text in the bold attribute.
func Bold(text string) string { return New().Bold().Apply(text) }

// Dim wraps text in the dim attribute.
func Dim(text string) string { return New().Dim().Apply(text) }

// Italic wraps text in the italic attribute.
func Italic(text string) string { return New().Italic().Apply(text) }

// Underline wraps text in the underline attribute.
func Underline(text string) string { return New().Underline().Apply(text) }

// Inverse wraps text in the inverse attribute.
func Inverse(text string) string { return New().Inverse().Apply(text) }

// Strikethrough wraps text in the strikethrough attribute.
func Strikethrough(text string) string { return New().Strikethrough().Apply(text) }

// Black colours text with a black foreground.
func Black(text string) string { return New().Black().Apply(text) }

// Red colours text with a red foreground.
func Red(text string) string { return New().Red().Apply(text) }

// Green colours text with a green foreground.
func Green(text string) string { return New().Green().Apply(text) }

// Yellow colours text with a yellow foreground.
func Yellow(text string) string { return New().Yellow().Apply(text) }

// Blue colours text with a blue foreground.
func Blue(text string) string { return New().Blue().Apply(text) }

// Magenta colours text with a magenta foreground.
func Magenta(text string) string { return New().Magenta().Apply(text) }

// Cyan colours text with a cyan foreground.
func Cyan(text string) string { return New().Cyan().Apply(text) }

// White colours text with a white foreground.
func White(text string) string { return New().White().Apply(text) }

// Gray colours text with a gray foreground.
func Gray(text string) string { return New().Gray().Apply(text) }

// BgRed colours text with a red background.
func BgRed(text string) string { return New().BgRed().Apply(text) }

// BgGreen colours text with a green background.
func BgGreen(text string) string { return New().BgGreen().Apply(text) }

// BgBlue colours text with a blue background.
func BgBlue(text string) string { return New().BgBlue().Apply(text) }

// BgYellow colours text with a yellow background.
func BgYellow(text string) string { return New().BgYellow().Apply(text) }

// RGB colours text with a 24-bit truecolor foreground.
func RGB(r, g, b uint8, text string) string { return New().RGB(r, g, b).Apply(text) }

// BgRGB colours text with a 24-bit truecolor background.
func BgRGB(r, g, b uint8, text string) string { return New().BgRGB(r, g, b).Apply(text) }

// Hex colours text with a 24-bit truecolor foreground given a "#rrggbb" or
// "#rgb" hex string. It returns text unchanged if the hex is malformed.
func Hex(hex, text string) string {
	r, g, b, ok := ansiParseHex(hex)
	if !ok {
		return text
	}
	return RGB(r, g, b, text)
}

// BgHex colours text with a 24-bit truecolor background given a hex string.
// It returns text unchanged if the hex is malformed.
func BgHex(hex, text string) string {
	r, g, b, ok := ansiParseHex(hex)
	if !ok {
		return text
	}
	return BgRGB(r, g, b, text)
}

var ansiEscapeRE = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

// Strip removes all ANSI CSI escape sequences from s, returning the plain text.
func Strip(s string) string {
	return ansiEscapeRE.ReplaceAllString(s, "")
}

// VisibleWidth returns the number of runes in s after removing ANSI escape
// sequences, i.e. the printed column width for ordinary (single-width) text.
func VisibleWidth(s string) int {
	return utf8.RuneCountInString(Strip(s))
}

func ansiParseHex(hex string) (r, g, b uint8, ok bool) {
	s := strings.TrimPrefix(strings.TrimSpace(hex), "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0, false
	}
	var v uint64
	if _, err := fmt.Sscanf(s, "%06x", &v); err != nil {
		return 0, 0, 0, false
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v), true
}
