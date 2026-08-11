# ansi — standard-library-only Go port of the popular npm terminal styling library chalk

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/ansi.svg)](https://pkg.go.dev/github.com/malcolmston/express/ansi)

Package ansi is a standard-library-only Go port of the popular npm terminal
styling library chalk (https://www.npmjs.com/package/chalk), which Node CLIs
and Express dev tooling use to colour and format terminal output. It emits
ANSI SGR (Select Graphic Rendition) escape sequences for the sixteen basic
colours, the common text attributes (bold, dim, italic, underline, inverse,
strikethrough), 24-bit truecolor via RGB/Hex, and provides Strip and
VisibleWidth to measure or remove styling.

Two styles of API are offered, matching chalk's ergonomics. The package-level
convenience functions wrap a string in a single attribute or colour and reset
afterwards, so ansi.Red("err") or ansi.Bold("go") returns a ready-to-print
string. For chalk-style composition, Style is an immutable builder: each
method returns a new Style with an added SGR code, and Apply renders a string
with the accumulated codes, so New().Bold().Red().Apply("x") produces bold red
text. Because Style values are immutable, a base style can be reused and
extended without interfering across call sites.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/ansi"
```

## Usage

This is the package's own `ExampleHex`, so it compiles and its output is
asserted on every `go test ./ansi/`.

```go
fmt.Printf("%q\n", ansi.Hex("#ff8800", "warn"))
	fmt.Printf("%q\n", ansi.Hex("#f80", "warn"))
	fmt.Printf("%q\n", ansi.Hex("not-a-color", "warn"))
```

```
"\x1b[38;2;255;136;0mwarn\x1b[0m"
"\x1b[38;2;255;136;0mwarn\x1b[0m"
"warn"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func BgBlue(text string) string` | BgBlue colours text with a blue background. |
| `func BgGreen(text string) string` | BgGreen colours text with a green background. |
| `func BgHex(hex, text string) string` | BgHex colours text with a 24-bit truecolor background given a hex string. |
| `func BgRGB(r, g, b uint8, text string) string` | BgRGB colours text with a 24-bit truecolor background. |
| `func BgRed(text string) string` | BgRed colours text with a red background. |
| `func BgYellow(text string) string` | BgYellow colours text with a yellow background. |
| `func Black(text string) string` | Black colours text with a black foreground. |
| `func Blue(text string) string` | Blue colours text with a blue foreground. |
| `func Bold(text string) string` | Bold wraps text in the bold attribute. |
| `func Cyan(text string) string` | Cyan colours text with a cyan foreground. |
| `func Dim(text string) string` | Dim wraps text in the dim attribute. |
| `func Gray(text string) string` | Gray colours text with a gray foreground. |
| `func Green(text string) string` | Green colours text with a green foreground. |
| `func Hex(hex, text string) string` | Hex colours text with a 24-bit truecolor foreground given a "#rrggbb" or "#rgb" hex string. |
| `func Inverse(text string) string` | Inverse wraps text in the inverse attribute. |
| `func Italic(text string) string` | Italic wraps text in the italic attribute. |
| `func Magenta(text string) string` | Magenta colours text with a magenta foreground. |
| `func RGB(r, g, b uint8, text string) string` | RGB colours text with a 24-bit truecolor foreground. |
| `func Red(text string) string` | Red colours text with a red foreground. |
| `func Strikethrough(text string) string` | Strikethrough wraps text in the strikethrough attribute. |
| `func Strip(s string) string` | Strip removes all ANSI CSI escape sequences from s, returning the plain text. |
| `func Underline(text string) string` | Underline wraps text in the underline attribute. |
| `func VisibleWidth(s string) int` | VisibleWidth returns the number of runes in s after removing ANSI escape sequences, i.e. |
| `func White(text string) string` | White colours text with a white foreground. |
| `func Yellow(text string) string` | Yellow colours text with a yellow foreground. |

### Types

| Type | What it is |
| --- | --- |
| `Style` | Style is an immutable set of ANSI SGR codes. |

<details>
<summary><code>Style</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New() Style` | New returns an empty Style with no codes. |
| `func (s Style) Apply(text string) string` | Apply renders text with the style's accumulated SGR codes, appending a reset. |
| `func (s Style) BgBlack() Style` | BgBlack returns a new Style with a black background. |
| `func (s Style) BgBlue() Style` | BgBlue returns a new Style with a blue background. |
| `func (s Style) BgCyan() Style` | BgCyan returns a new Style with a cyan background. |
| `func (s Style) BgGreen() Style` | BgGreen returns a new Style with a green background. |
| `func (s Style) BgMagenta() Style` | BgMagenta returns a new Style with a magenta background. |
| `func (s Style) BgRGB(r, g, b uint8) Style` | BgRGB returns a new Style with a 24-bit truecolor background. |
| `func (s Style) BgRed() Style` | BgRed returns a new Style with a red background. |
| `func (s Style) BgWhite() Style` | BgWhite returns a new Style with a white background. |
| `func (s Style) BgYellow() Style` | BgYellow returns a new Style with a yellow background. |
| `func (s Style) Black() Style` | Black returns a new Style with a black foreground. |
| `func (s Style) Blue() Style` | Blue returns a new Style with a blue foreground. |
| `func (s Style) Bold() Style` | Bold returns a new Style with the bold attribute added. |
| `func (s Style) Cyan() Style` | Cyan returns a new Style with a cyan foreground. |
| `func (s Style) Dim() Style` | Dim returns a new Style with the dim (faint) attribute added. |
| `func (s Style) Gray() Style` | Gray returns a new Style with a gray (bright black) foreground. |
| `func (s Style) Green() Style` | Green returns a new Style with a green foreground. |
| `func (s Style) Inverse() Style` | Inverse returns a new Style with the inverse (reverse video) attribute added. |
| `func (s Style) Italic() Style` | Italic returns a new Style with the italic attribute added. |
| `func (s Style) Magenta() Style` | Magenta returns a new Style with a magenta foreground. |
| `func (s Style) RGB(r, g, b uint8) Style` | RGB returns a new Style with a 24-bit truecolor foreground. |
| `func (s Style) Red() Style` | Red returns a new Style with a red foreground. |
| `func (s Style) Strikethrough() Style` | Strikethrough returns a new Style with the strikethrough attribute added. |
| `func (s Style) Underline() Style` | Underline returns a new Style with the underline attribute added. |
| `func (s Style) White() Style` | White returns a new Style with a white foreground. |
| `func (s Style) Yellow() Style` | Yellow returns a new Style with a yellow foreground. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/ansi).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
