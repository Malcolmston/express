# qs — parses and serializes URL query strings with support for nested objects and arrays via

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/qs.svg)](https://pkg.go.dev/github.com/malcolmston/express/qs)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/qs)

Package qs parses and serializes URL query strings with support for nested
objects and arrays via bracket notation. It is a Go port of the npm module
"qs", the query-string library Express uses when its extended body/query
parser is enabled.

Where the standard library's net/url and this repository's querystringlite
treat a query string as a flat mapping of keys to values, qs understands the
bracket syntax that lets a single query string describe arbitrarily nested
structures. This is useful whenever an HTML form, an Express client, or an
AJAX request encodes an object or array into a URL, and the server needs to
reconstruct that shape rather than a flat list of strings.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/qs"
```

## Usage

This is the package's own `ExampleStringify`, so it compiles and its output is
asserted on every `go test ./qs/`.

```go
m := map[string]any{"a": map[string]any{"b": "1", "c": "2"}}
	fmt.Println(qs.Stringify(m))
```

```
a[b]=1&a[c]=2
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Parse(s string) map[string]any` | Parse parses a URL query string into a nested map using the npm qs defaults: depth 5, parameter limit 1000, array limit 20, and "&" as the delimiter. |
| `func ParseWith(s string, opts ParseOptions) map[string]any` | ParseWith parses a URL query string into a nested map using opts. |
| `func Stringify(m map[string]any) string` | Stringify serializes a nested map into a URL query string. |
| `func StringifyWith(m map[string]any, opts StringifyOptions) string` | StringifyWith serializes a nested map into a URL query string using opts. |

### Types

| Type | What it is |
| --- | --- |
| `ParseOptions` | ParseOptions controls ParseWith. |
| `StringifyOptions` | StringifyOptions controls StringifyWith. |

### Constants

`DefaultDepth`, `DefaultParameterLimit`, `DefaultArrayLimit`, `DefaultDelimiter`, `DuplicatesCombine`, `DuplicatesFirst`, `DuplicatesLast`, `ArrayFormatIndices`, `ArrayFormatBrackets`, `ArrayFormatRepeat`, `ArrayFormatComma`, `FormatRFC3986`, `FormatRFC1738`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/qs).

## Measured parity

Compared case-for-case against **`qs@6.15.3`** by the harness in
[`parity/express/nested/qs`](https://github.com/malcolmston/go/tree/main/parity/express/nested/qs):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 271 |
| Matching | 268 |
| Mismatching | 0 |
| Declared deviations | 3 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/qs/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
