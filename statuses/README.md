# statuses — maps between HTTP status codes and their standard reason phrases and classifies codes

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/statuses.svg)](https://pkg.go.dev/github.com/malcolmston/express/statuses)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/statuses)

Package statuses maps between HTTP status codes and their standard reason
phrases and classifies codes by behavior. It is a Go port of the npm
"statuses" package, reimplemented using only the Go standard library. Where
the Node original exposes a callable module plus lookup tables, this package
exposes the same capabilities as ordinary functions and returns Go-idiomatic
values and errors.

A web framework or HTTP client reaches for this package whenever it needs the
canonical text for a numeric status code (for example, to build a default
response body or a log line reading "404 Not Found"), or the reverse: to turn
a human-written phrase such as "Not Found" back into the number 404. It also
answers the three behavioral questions that middleware most often asks about a
status: is it a redirect, may the request be retried, and must the response be
sent without a body.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/statuses"
```

## Usage

This is the package's own `ExampleCode`, so it compiles and its output is
asserted on every `go test ./statuses/`.

```go
code, err := statuses.Code("Not Found")
	fmt.Println(code, err)

	code, err = statuses.Code("not found")
	fmt.Println(code, err)

	code, err = statuses.Code("nonsense")
	fmt.Println(code, err)
```

```
404 <nil>
404 <nil>
0 invalid status message: "nonsense"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Code(message string) (int, error)` | Code returns the status code for the given reason phrase. |
| `func Codes() []int` | Codes returns all known status codes in ascending order. |
| `func IsEmpty(code int) bool` | IsEmpty reports whether responses with this status code must not carry a body. |
| `func IsRedirect(code int) bool` | IsRedirect reports whether the status code is a redirect. |
| `func IsRetry(code int) bool` | IsRetry reports whether a request that received this status code may be retried. |
| `func Message(code int) string` | Message returns the reason phrase for the given status code. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/statuses).

## Measured parity

Compared case-for-case against **`statuses@2.0.1`** by the harness in
[`parity/express/nested/statuses`](https://github.com/malcolmston/go/tree/main/parity/express/nested/statuses):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 142 |
| Matching | 141 |
| Mismatching | 0 |
| Declared deviations | 1 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/statuses/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
