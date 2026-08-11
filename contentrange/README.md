# contentrange — formats and parses HTTP Content-Range header values using only the Go standard library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/contentrange.svg)](https://pkg.go.dev/github.com/malcolmston/express/contentrange)

Package contentrange formats and parses HTTP Content-Range header values using
only the Go standard library. It provides Format to build a header string from
its numeric components, Parse to decode a header string into a ContentRange
struct, and a String method to render a ContentRange back into a header value.

A Content-Range header has the form "<unit> <range>/<size>", for example
"bytes 0-499/1234". The range may be "*" for an unsatisfied range and the size
may be "*" when the total length is unknown. Servers send this header on 206
Partial Content responses to describe which slice of a representation the body
carries, and on 416 Range Not Satisfiable responses to report the total size;
you use this package on the server to construct that header and on a client to
interpret one.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/contentrange"
```

## Usage

This is the package's own `ExampleContentRange_String`, so it compiles and its output is
asserted on every `go test ./contentrange/`.

```go
cr := contentrange.ContentRange{
		Unit:     "bytes",
		Start:    0,
		End:      499,
		HasRange: true,
		HasSize:  false,
	}
	fmt.Println(cr.String())
```

```
bytes 0-499/*
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(unit string, start, end, size int64) string` | Format builds a Content-Range header value. |

### Types

| Type | What it is |
| --- | --- |
| `ContentRange` | ContentRange represents the parsed components of a Content-Range header. |

<details>
<summary><code>ContentRange</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(s string) (ContentRange, error)` | Parse parses a Content-Range header value such as "bytes 0-499/1234", "bytes */1234", or "bytes 0-499/*". |
| `func (cr ContentRange) String() string` | String renders the ContentRange back into a header value. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/contentrange).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
