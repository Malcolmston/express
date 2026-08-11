# contenttype — parses and formats HTTP Content-Type header values

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/contenttype.svg)](https://pkg.go.dev/github.com/malcolmston/express/contenttype)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/contenttype)

Package contenttype parses and formats HTTP Content-Type header values,
modeled on the npm "content-type" package and RFC 7231, using only the Go
standard library. It provides Parse to decode a header value into a
ContentType struct and Format to serialize a ContentType back into a header
value.

The Content-Type header names the media type of a message body and carries
optional parameters such as the character set. You use Parse on the receiving
side to learn the media type and, for example, the charset of a request or
response body, and Format on the sending side to build a well-formed header
from a media type and a parameter map. Both operations are pure functions, so
they are convenient to test and to reuse outside of a live HTTP handler.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/contenttype"
```

## Usage

This is the package's own `ExampleFormat`, so it compiles and its output is
asserted on every `go test ./contenttype/`.

```go
s, err := contenttype.Format(contenttype.ContentType{
		Type:       "application/json",
		Parameters: map[string]string{"charset": "utf-8"},
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
```

```
application/json; charset=utf-8
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(ct ContentType) (string, error)` | Format serializes a ContentType back into a header value such as "text/html; charset=utf-8". |

### Types

| Type | What it is |
| --- | --- |
| `ContentType` | ContentType represents a parsed media type and its parameters. |

<details>
<summary><code>ContentType</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(s string) (ContentType, error)` | Parse parses a Content-Type header value such as "text/html; charset=utf-8". |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/contenttype).

## Measured parity

Compared case-for-case against **`content-type@1.0.5`** by the harness in
[`parity/express/nested/contenttype`](https://github.com/malcolmston/go/tree/main/parity/express/nested/contenttype):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 76 |
| Matching | 76 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/contenttype/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
