# accepts — performs HTTP content negotiation based on the Accept, Accept-Language

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/accepts.svg)](https://pkg.go.dev/github.com/malcolmston/express/accepts)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/accepts)

Package accepts performs HTTP content negotiation based on the Accept,
Accept-Language, Accept-Charset and Accept-Encoding request headers. It is a
port of the npm accepts package (the higher-level wrapper around negotiator
that Express uses for req.accepts, req.acceptsLanguages, req.acceptsCharsets
and req.acceptsEncodings) built with only the Go standard library.

You reach for this package on the server side whenever a single endpoint can
respond in more than one representation and you want the client's stated
preferences to pick which one to send. Typical uses are choosing between JSON
and HTML for the same resource, selecting a UI language, agreeing on a
character set, or deciding whether to compress a response with gzip. Rather
than parsing the raw header strings yourself, you hand the request headers to
New and then ask, for each dimension, which of the representations you are
able to produce ("offers") the client prefers.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/accepts"
```

## Usage

This is the package's own `ExampleAccepts_Encoding`, so it compiles and its output is
asserted on every `go test ./accepts/`.

```go
h := http.Header{}
	h.Set("Accept-Encoding", "gzip, deflate;q=0.5")

	a := accepts.New(h)
	fmt.Println(a.Encoding("gzip", "deflate"))
```

```
gzip
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Accepts` | Accepts negotiates acceptable response representations for a request based on its Accept* headers. |

<details>
<summary><code>Accepts</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(header http.Header) *Accepts` | New creates an Accepts from the given request headers, reading the Accept, Accept-Language, Accept-Charset and Accept-Encoding fields. |
| `func (a *Accepts) Charset(offers ...string) string` | Charset returns the single best charset from offers, or "". |
| `func (a *Accepts) Charsets(offers ...string) []string` | Charsets returns the acceptable charsets from offers in preference order, or the header's own charsets when no offers are given. |
| `func (a *Accepts) Encoding(offers ...string) string` | Encoding returns the single best encoding from offers, or "". |
| `func (a *Accepts) Encodings(offers ...string) []string` | Encodings returns the acceptable encodings from offers in preference order, or the header's own encodings when no offers are given. |
| `func (a *Accepts) Language(offers ...string) string` | Language returns the single best language from offers, or "". |
| `func (a *Accepts) Languages(offers ...string) []string` | Languages returns the acceptable languages from offers in preference order, or the header's own languages when no offers are given. |
| `func (a *Accepts) Type(offers ...string) string` | Type returns the single best media type from offers, or "" if none are acceptable. |
| `func (a *Accepts) Types(offers ...string) []string` | Types returns the offers that are acceptable per the Accept header, ordered by preference. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/accepts).

## Measured parity

Compared case-for-case against **`accepts@1.3.8`** by the harness in
[`parity/express/nested/accepts`](https://github.com/malcolmston/go/tree/main/parity/express/nested/accepts):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 117 |
| Matching | 117 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/accepts/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
