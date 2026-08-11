# sanitizehtml — sanitizes untrusted HTML by removing disallowed tags and attributes

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/sanitizehtml.svg)](https://pkg.go.dev/github.com/malcolmston/express/sanitizehtml)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/sanitizehtml)

Package sanitizehtml sanitizes untrusted HTML by removing disallowed tags and
attributes, mirroring a practical subset of the npm "sanitize-html" library.
It is intended for the common case of accepting rich text from a user
(comments, profile bios, CMS fields) and rendering it back into a page without
opening the door to cross-site scripting or layout-breaking markup.

Sanitization is driven by an allowlist rather than a blocklist, which is the
safe default: only tags that appear in Options.AllowedTags survive, and on
those surviving tags only the attributes named in Options.AllowedAttributes
are kept. Everything else is discarded. Because the policy enumerates what is
permitted instead of what is forbidden, a novel or misspelled tag is stripped
automatically rather than slipping through. DefaultOptions returns a policy
equivalent to the sanitize-html defaults, allowing common formatting tags plus
href/name/target on <a> and the usual source attributes on <img>.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/sanitizehtml"
```

## Usage

This is the package's own `ExampleSanitize`, so it compiles and its output is
asserted on every `go test ./sanitizehtml/`.

```go
input := `<p>Hello <script>alert(1)</script><b>world</b></p>`
	fmt.Println(sanitizehtml.Sanitize(input, sanitizehtml.DefaultOptions()))
```

```
<p>Hello <b>world</b></p>
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Sanitize(htmlStr string, opts Options) string` | Sanitize returns htmlStr with disallowed tags and attributes removed according to opts. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options controls how Sanitize filters HTML. |

<details>
<summary><code>Options</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func DefaultOptions() Options` | DefaultOptions returns Options mirroring the sanitize-html defaults: a set of common formatting tags is allowed, href/name/target are permitted on… |

</details>

### Constants

`ModeDiscard`, `ModeCompletelyDiscard`, `ModeEscape`, `ModeRecursiveEscape`

### Variables

`DefaultAllowedEmptyAttributes`, `DefaultNonBooleanAttributes`, `DefaultNonTextTags`, `DefaultSchemeAttributes`, `DefaultSchemes`, `DefaultSelfClosing`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/sanitizehtml).

## Measured parity

Compared case-for-case against **`sanitize-html@2.17.6`** by the harness in
[`parity/express/nested/sanitizehtml`](https://github.com/malcolmston/go/tree/main/parity/express/nested/sanitizehtml):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 233 |
| Matching | 223 |
| Mismatching | 0 |
| Declared deviations | 10 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/sanitizehtml/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
