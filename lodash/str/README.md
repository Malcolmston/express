# str — ports the "String" category of the npm lodash library to Go

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/str.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/str)

Package str ports the "String" category of the npm lodash library to Go. It
provides the familiar lodash string helpers: the case converters (CamelCase,
KebabCase, SnakeCase, StartCase, LowerCase, UpperCase, Capitalize, UpperFirst,
LowerFirst), the word splitter Words, the padding and repeating helpers (Pad,
PadStart, PadEnd, Repeat), the trimming family (Trim, TrimStart, TrimEnd),
predicates (StartsWith, EndsWith), HTML entity handling (Escape, Unescape),
diacritic folding (Deburr), plus Truncate, Replace and a JavaScript-style
ParseInt. Every function operates on plain Go strings and depends only on the
standard library.

Use this package when you are porting front-end JavaScript that relied on
lodash string utilities, or whenever you want their concise, well-defined
behavior in Go: turning arbitrary identifiers or user input into a consistent
case (slugs, config keys, display labels), normalizing accented text for
search or comparison, padding and truncating strings for fixed-width output,
or escaping text for safe inclusion in HTML. The helpers are Unicode-aware
where it matters — lengths, slicing and padding are measured in runes rather
than bytes — so multi-byte input is handled the way lodash measures strings
by code point.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/str"
```

## Usage

This is the package's own `ExampleCamelCase`, so it compiles and its output is
asserted on every `go test ./str/`.

```go
fmt.Println(str.CamelCase("Foo Bar"))
	fmt.Println(str.CamelCase("--foo-bar--"))
	fmt.Println(str.CamelCase("__FOO_BAR__"))
```

```
fooBar
fooBar
fooBar
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CamelCase(s string) string` | CamelCase converts string to camel case. |
| `func Capitalize(s string) string` | Capitalize converts the first character of string to upper case and the remaining to lower case. |
| `func Chars(s string) []string` | Chars splits s into a slice of its Unicode characters (runes as strings), so Chars("a€b") is ["a", "€", "b"]. |
| `func Deburr(s string) string` | Deburr converts the Latin-1 Supplement and common Latin Extended-A accented letters in string to their basic Latin equivalents and removes combining… |
| `func EndsWith(s, target string, position int) bool` | EndsWith reports whether string ends with the given target, testing up to the supplied position (in runes). |
| `func Escape(s string) string` | Escape converts the characters "&", "<", ">", '"' and "'" in string to their corresponding HTML entities. |
| `func KebabCase(s string) string` | KebabCase converts string to kebab case. |
| `func LowerCase(s string) string` | LowerCase converts string, as space separated words, to lower case. |
| `func LowerFirst(s string) string` | LowerFirst converts the first character of string to lower case, leaving the remainder untouched. |
| `func Pad(s string, length int, chars string) string` | Pad pads string on the left and right sides if it is shorter than length. |
| `func PadEnd(s string, length int, chars string) string` | PadEnd pads string on the right side if it is shorter than length. |
| `func PadStart(s string, length int, chars string) string` | PadStart pads string on the left side if it is shorter than length. |
| `func ParseInt(s string, radix int) int` | ParseInt converts string to an integer of the specified radix. |
| `func Repeat(s string, n int) string` | Repeat repeats the given string n times. |
| `func Replace(s, old, replacement string) string` | Replace replaces the first occurrence of old in string with replacement, matching JavaScript's String.prototype.replace when called with a string… |
| `func SnakeCase(s string) string` | SnakeCase converts string to snake case. |
| `func Split(s, separator string, limit int) []string` | Split splits s around each instance of separator. |
| `func StartCase(s string) string` | StartCase converts string to start case, capitalizing the first letter of every word while preserving the remaining letters of each word. |
| `func StartsWith(s, target string, position int) bool` | StartsWith reports whether string begins with the given target, testing from the supplied position (in runes). |
| `func ToLower(s string) string` | ToLower returns s with all Unicode letters mapped to lower case. |
| `func ToUpper(s string) string` | ToUpper returns s with all Unicode letters mapped to upper case. |
| `func Trim(s string, chars string) string` | Trim removes leading and trailing characters (whitespace by default, or any of the runes in chars when provided) from string. |
| `func TrimEnd(s string, chars string) string` | TrimEnd removes trailing characters (whitespace by default, or any of the runes in chars when provided) from string. |
| `func TrimStart(s string, chars string) string` | TrimStart removes leading characters (whitespace by default, or any of the runes in chars when provided) from string. |
| `func Truncate(s string, opts TruncateOptions) string` | Truncate truncates string if it is longer than the requested length. |
| `func Unescape(s string) string` | Unescape is the inverse of Escape; it converts the HTML entities "&amp;", "&lt;", "&gt;", "&quot;" and "&#39;" in string back to their characters. |
| `func UpperCase(s string) string` | UpperCase converts string, as space separated words, to upper case. |
| `func UpperFirst(s string) string` | UpperFirst converts the first character of string to upper case, leaving the remainder untouched. |
| `func Words(s string) []string` | Words splits string into an array of its words. |

### Types

| Type | What it is |
| --- | --- |
| `TruncateOptions` | TruncateOptions configures Truncate. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/str).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
