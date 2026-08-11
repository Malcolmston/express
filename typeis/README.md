# typeis — matches Content-Type header values against a set of type candidates

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/typeis.svg)](https://pkg.go.dev/github.com/malcolmston/express/typeis)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/typeis)

Package typeis matches Content-Type header values against a set of type
candidates. It is a standard-library-only Go port of the npm "type-is"
package, the content negotiation helper Express exposes as req.is and that
body-parsing middleware uses to decide whether a request body is JSON, form
data, or something else. The problem it solves is that a raw Content-Type
header such as "application/json; charset=utf-8" is awkward to test directly:
it carries parameters, is case-insensitive, and callers usually want to ask a
convenience question like "is this JSON?" rather than compare strings.

The high-level entry point is Is, which takes a concrete Content-Type value
and one or more candidate patterns and returns the matched candidate together
with a boolean. Candidates may be written in several convenient shapes: a full
type such as "application/json"; an extension-style shorthand such as "json",
"html", or "png"; the specials "urlencoded" and "multipart"; a wildcard such
as "*/*", "text/*", or "*/json"; or a structured "+suffix" match such as
"+json" that matches any subtype ending in that suffix. When no candidates are
supplied, Is reports whether the value is a non-empty, valid type and echoes
it back.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/typeis"
```

## Usage

This is the package's own `ExampleMatch`, so it compiles and its output is
asserted on every `go test ./typeis/`.

```go
fmt.Println(typeis.Match("application/*", "application/json"))
	fmt.Println(typeis.Match("*/*+json", "application/vnd.api+json"))
	fmt.Println(typeis.Match("text/*", "application/json"))
```

```
true
true
false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Is(contentType string, types ...string) (string, bool)` | Is matches a Content-Type value against one or more candidate types. |
| `func Match(expected, actual string) bool` | Match reports whether the expected mime type matches the actual mime type, with "*" wildcard and "*+suffix" support. |
| `func Normalize(t string) (string, bool)` | Normalize expands a single candidate type into its full, matchable form. |
| `func NormalizeType(value string) string` | NormalizeType strips any parameters (e.g. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/typeis).

## Measured parity

Compared case-for-case against **`type-is@1.6.18`** by the harness in
[`parity/express/nested/typeis`](https://github.com/malcolmston/go/tree/main/parity/express/nested/typeis):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 105 |
| Matching | 105 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/typeis/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
