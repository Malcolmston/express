# negotiator — HTTP content negotiation helper

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/negotiator.svg)](https://pkg.go.dev/github.com/malcolmston/express/negotiator)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/negotiator)

Package negotiator is an HTTP content negotiation helper, a stdlib-only Go
port of the npm "negotiator" package
(https://www.npmjs.com/package/negotiator). It inspects request headers such
as Accept, Accept-Language, Accept-Charset and Accept-Encoding and reports
which of a set of server-side representations best match the client's stated
preferences. Like the Node original, it is a low-level building block: it
makes no assumptions about your routes and only answers the question "given
these headers and these available options, which options are acceptable and in
what order?".

Content negotiation matters whenever a single URL can be served in more than
one form. A client advertises what it can handle by sending weighted lists in
its request headers, for example "Accept: application/json;q=0.9, text/html"
or "Accept-Encoding: gzip, br;q=0.8". The server holds a fixed menu of things
it is able to produce (JSON or HTML, gzip or brotli, English or French) and
must pick the best mutually agreeable choice. This package performs that pick,
mirroring the behavior that Express, the accepts package, and many other Node
middlewares rely on internally.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/negotiator"
```

## Usage

This is the package's own `ExampleNegotiator_Charset`, so it compiles and its output is
asserted on every `go test ./negotiator/`.

```go
h := http.Header{}
	h.Set("Accept-Charset", "utf-8, iso-8859-1;q=0.5")

	n := negotiator.New(h)
	fmt.Println(n.Charset("iso-8859-1", "utf-8"))
```

```
utf-8
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Negotiator` | Negotiator performs content negotiation against a set of request headers. |

<details>
<summary><code>Negotiator</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(header http.Header) *Negotiator` | New constructs a Negotiator from the given request headers. |
| `func (n *Negotiator) Charset(available ...string) string` | Charset returns the single best acceptable charset, or "". |
| `func (n *Negotiator) Charsets(available ...string) []string` | Charsets returns the acceptable charsets given the request's Accept-Charset header, restricted to the provided available charsets. |
| `func (n *Negotiator) Encoding(available ...string) string` | Encoding returns the single best acceptable encoding, or "". |
| `func (n *Negotiator) Encodings(available ...string) []string` | Encodings returns the acceptable encodings given the request's Accept-Encoding header, restricted to the provided available encodings. |
| `func (n *Negotiator) Language(available ...string) string` | Language returns the single best acceptable language, or "". |
| `func (n *Negotiator) Languages(available ...string) []string` | Languages returns the acceptable languages given the request's Accept-Language header, restricted to the provided available languages. |
| `func (n *Negotiator) MediaType(available ...string) string` | MediaType returns the single best acceptable media type, or "" if none of the available media types are acceptable. |
| `func (n *Negotiator) MediaTypes(available ...string) []string` | MediaTypes returns the media types that are acceptable given the request's Accept header, restricted to the provided available media types. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/negotiator).

## Measured parity

Compared case-for-case against **`negotiator@0.6.3`** by the harness in
[`parity/express/nested/negotiator`](https://github.com/malcolmston/go/tree/main/parity/express/nested/negotiator):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 142 |
| Matching | 142 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/negotiator/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
