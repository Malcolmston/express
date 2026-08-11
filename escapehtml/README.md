# escapehtml — escapes the HTML-significant characters in a string so it can be safely embedded

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/escapehtml.svg)](https://pkg.go.dev/github.com/malcolmston/express/escapehtml)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/escapehtml)

Package escapehtml escapes the HTML-significant characters in a string so it
can be safely embedded in HTML markup, a port of the npm "escape-html" package
that Express uses when rendering error pages and other user-influenced text.
It exposes the single Escape function and depends on only the Go standard
library.

Escaping is the baseline defense against HTML injection and cross-site
scripting. Any time untrusted or merely unpredictable text — a form value, a
filename, a database field, an error message — is placed into an HTML
document or an attribute value, the characters that HTML treats specially must
be turned into entities so a browser renders them as literal text instead of
interpreting them as tags or attribute delimiters. Escape gives you exactly
that conversion in one call.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/escapehtml"
```

## Usage

This is the package's own `ExampleEscape`, so it compiles and its output is
asserted on every `go test ./escapehtml/`.

```go
fmt.Println(escapehtml.Escape(`<a href="x">Tom & Jerry's</a>`))
```

```
&lt;a href=&quot;x&quot;&gt;Tom &amp; Jerry&#39;s&lt;/a&gt;
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Escape(s string) string` | Escape returns s with the HTML-significant characters &, <, >, ", and ' replaced by their entity equivalents: |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/escapehtml).

## Measured parity

Compared case-for-case against **`escape-html@1.0.3`** by the harness in
[`parity/express/nested/escapehtml`](https://github.com/malcolmston/go/tree/main/parity/express/nested/escapehtml):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 36 |
| Matching | 36 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/escapehtml/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
