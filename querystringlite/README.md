# querystringlite — faithful port of Node.js's built-in querystring module for flat (non-nested) query

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/querystringlite.svg)](https://pkg.go.dev/github.com/malcolmston/express/querystringlite)

Package querystringlite is a faithful port of Node.js's built-in querystring
module for flat (non-nested) query strings. It mirrors the querystring.parse,
querystring.stringify, querystring.escape, and querystring.unescape functions
that Node exposes for reading and writing application/x-www-form-urlencoded
data.

Use this package when you want Node-compatible query-string handling without
the nested bracket semantics of the qs package. It treats a query string as a
flat multimap: every key maps to the ordered list of values it was given, so
repeated keys such as "a=1&a=3" are preserved rather than collapsed. This is
the right tool for straightforward form submissions and API query parameters
where "a`b`" is just a literal key and not a request to build a nested object.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/querystringlite"
```

## Usage

This is the package's own `ExampleStringify`, so it compiles and its output is
asserted on every `go test ./querystringlite/`.

```go
v := map[string][]string{"a": {"1", "3"}, "b": {"2"}}
	fmt.Println(querystringlite.Stringify(v))
```

```
a=1&a=3&b=2
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Escape(s string) string` | Escape percent-encodes s the way Node's querystring.escape does. |
| `func Parse(s string) map[string][]string` | Parse parses a URL query string into a map of keys to their values. |
| `func ParseSingle(s string) map[string]string` | ParseSingle is a convenience wrapper around Parse that keeps only the first value seen for each key. |
| `func Stringify(v map[string][]string) string` | Stringify serializes a map of keys to values into a query string. |
| `func StringifySingle(v map[string]string) string` | StringifySingle is a convenience wrapper around Stringify for maps with a single value per key. |
| `func Unescape(s string) string` | Unescape decodes s the way Node's querystring.unescape does: "+" becomes a space and valid "%XX" sequences are percent-decoded. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/querystringlite).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
