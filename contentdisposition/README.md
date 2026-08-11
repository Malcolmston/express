# contentdisposition — creates and parses HTTP Content-Disposition headers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/contentdisposition.svg)](https://pkg.go.dev/github.com/malcolmston/express/contentdisposition)

Package contentdisposition creates and parses HTTP Content-Disposition
headers. It is a port of the npm content-disposition package using only the Go
standard library, and it exposes the two operations that package offers:
Format builds a header value from a filename, and Parse decodes a header value
back into its type and parameters.

The Content-Disposition header tells a client how a response body should be
handled: "inline" to display it in place, or "attachment" to save it, most
often under a suggested filename. You use Format on the server side when
streaming a download so the browser offers a sensible "Save As" name, and you
use Parse on the client side, or in multipart/form-data handling, to recover
the disposition type and filename a peer sent.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/contentdisposition"
```

## Usage

This is the package's own `ExampleParse`, so it compiles and its output is
asserted on every `go test ./contentdisposition/`.

```go
cd, err := contentdisposition.Parse(`attachment; filename="report.pdf"`)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cd.Type, cd.Filename)
```

```
attachment report.pdf
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Format(filename string, opts ...Option) string` | Format builds a Content-Disposition header value for the given filename. |
| `func IsValidFallbackName(name string) bool` | IsValidFallbackName reports whether name is acceptable as an explicit fallback filename, i.e. |

### Types

| Type | What it is |
| --- | --- |
| `ContentDisposition` | ContentDisposition represents a parsed Content-Disposition header. |
| `Option` | Option configures how a Content-Disposition header is formatted. |

<details>
<summary><code>ContentDisposition</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Parse(s string) (ContentDisposition, error)` | Parse parses a Content-Disposition header value. |

</details>

<details>
<summary><code>Option</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func WithFallback(fallback bool) Option` | WithFallback controls whether an ASCII fallback filename parameter is emitted when the filename contains non-ASCII characters. |
| `func WithFallbackName(name string) Option` | WithFallbackName supplies an explicit US-ASCII fallback filename, mirroring upstream's string form of the fallback option. |
| `func WithType(t string) Option` | WithType sets the disposition type (default "attachment"). |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/contentdisposition).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
