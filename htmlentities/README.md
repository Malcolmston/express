# htmlentities — encodes and decodes HTML entities, a port of the npm "html-entities" library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/htmlentities.svg)](https://pkg.go.dev/github.com/malcolmston/express/htmlentities)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/htmlentities)

Package htmlentities encodes and decodes HTML entities, a port of the npm
"html-entities" library. It exists so that Go code ported from a Node codebase
can keep calling a familiar encode/decode pair when rendering user-supplied
text into HTML or when reading entity-laden text back out again, without
pulling in a third-party dependency.

Escaping HTML entities is the front line of defense against cross-site
scripting: turning the characters & < > " ' into their named forms prevents
user input from being interpreted as markup or from breaking out of an
attribute value. The npm html-entities library grew popular precisely because
it makes this safe by default while still allowing stricter modes that also
escape non-ASCII or non-printable text for transports that are not UTF-8
clean. This port keeps the same option surface and shares the library's own
reference tables, which are generated straight out of the installed package
into tables.go, so the two agree entity for entity rather than approximately.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/htmlentities"
```

## Usage

This is the package's own `ExampleDecodeEntity`, so it compiles and its output is
asserted on every `go test ./htmlentities/`.

```go
fmt.Println(htmlentities.DecodeEntity("&copy;"))
	fmt.Println(htmlentities.DecodeEntity("&#233;"))
	fmt.Println(htmlentities.DecodeEntity("&nosuchentity;"))
```

```
©
é
&nosuchentity;
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Decode(s string, opts ...DecodeOptions) string` | Decode converts every HTML entity reference in s back to the text it stands for. |
| `func DecodeEntity(entity string, opts ...DecodeOptions) string` | DecodeEntity resolves a single entity reference, such as "&copy;" or "&#233;", and returns it unchanged when it is not recognized. |
| `func Encode(s string, opts ...EncodeOptions) string` | Encode converts characters in s to HTML entities according to opts. |

### Types

| Type | What it is |
| --- | --- |
| `DecodeOptions` | DecodeOptions configures Decode and DecodeEntity. |
| `EncodeOptions` | EncodeOptions configures Encode. |

### Constants

`ModeSpecialChars`, `ModeNonASCII`, `ModeNonASCIIPrintable`, `ModeNonASCIIPrintableOnly`, `ModeExtensive`, `LevelXML`, `LevelHTML4`, `LevelHTML5`, `LevelAll`, `NumericDecimal`, `NumericHexadecimal`, `ScopeStrict`, `ScopeBody`, `ScopeAttribute`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/htmlentities).

## Measured parity

Compared case-for-case against **`html-entities@2.6.0`** by the harness in
[`parity/express/nested/htmlentities`](https://github.com/malcolmston/go/tree/main/parity/express/nested/htmlentities):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 113 |
| Matching | 110 |
| Mismatching | 0 |
| Declared deviations | 3 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/htmlentities/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
