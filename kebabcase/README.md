# kebabcase — Full (context-sensitive) Unicode lowercase mapping

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/kebabcase.svg)](https://pkg.go.dev/github.com/malcolmston/express/kebabcase)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/kebabcase)

Full (context-sensitive) Unicode lowercase mapping.

change-case lower-cases each split word with JavaScript's
String.prototype.toLowerCase, which implements Unicode's FULL lowercase
mapping: U+0130 (Latin capital I with dot above) expands to two code points,
and a Greek capital sigma at the end of a word becomes the final form
"\u03c2". Go's standard library offers only the SIMPLE per-rune mapping, which
collapses U+0130 to a bare "i" and always produces the medial "\u03c3". This
file supplies the two missing rules so KebabCase matches the Node original
using only the standard library.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/kebabcase"
```

## Usage

This is the package's own `ExampleKebabCase_separators`, so it compiles and its output is
asserted on every `go test ./kebabcase/`.

```go
fmt.Println(kebabcase.KebabCase("foo_bar"))
	fmt.Println(kebabcase.KebabCase("XMLHttpRequest"))
	fmt.Println(kebabcase.KebabCase("foo123Bar"))
```

```
foo-bar
xml-http-request
foo123-bar
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func KebabCase(s string) string` | KebabCase converts s to kebab-case. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/kebabcase).

## Measured parity

Compared case-for-case against **`change-case@5.4.4`** by the harness in
[`parity/express/nested/kebabcase`](https://github.com/malcolmston/go/tree/main/parity/express/nested/kebabcase):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 102 |
| Matching | 102 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/kebabcase/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
