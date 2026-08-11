# striptags — removes HTML tags from a string while keeping the text content

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/striptags.svg)](https://pkg.go.dev/github.com/malcolmston/express/striptags)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/striptags)

Package striptags removes HTML tags from a string while keeping the text
content, mirroring the behavior of the npm "striptags" library. It is a small
utility for turning a fragment of markup into plain text, for example when
generating a preview, building a search index, or sanitizing a value that must
not contain any HTML at all.

By default every tag is stripped: the angle-bracket markup is removed and the
text that sits between tags is preserved. When one or more allowed tag names
are supplied to StripTags as variadic arguments, those tags (and their
matching closing tags) are kept verbatim while all other tags are removed.
Allowed tag names may be written either bare ("p") or wrapped in angle
brackets ("<p>"); normalizeAllowed accepts either form, and comparison is
case-insensitive.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/striptags"
```

## Usage

This is the package's own `ExampleStripTags_allowed`, so it compiles and its output is
asserted on every `go test ./striptags/`.

```go
fmt.Println(striptags.StripTags("<p>Hi <b>there</b></p>", "b"))
```

```
Hi <b>there</b>
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func StripTags(html string, allowed ...string) string` | StripTags removes HTML tags from html. |
| `func StripTagsWith(html string, allowed []string, tagReplacement string) string` | StripTagsWith removes HTML tags from html, writing tagReplacement in place of every removed tag. |

### Types

| Type | What it is |
| --- | --- |
| `Streamer` | Streamer strips tags from a sequence of chunks, carrying the parser state from one chunk to the next so that a tag split across a chunk boundary is… |

<details>
<summary><code>Streamer</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func NewStreamer(allowed []string, tagReplacement string) *Streamer` | NewStreamer returns a Streamer that keeps the tags named in allowed (in any of the forms StripTags accepts) and writes replacement in place of every… |
| `func (s *Streamer) Write(html string) string` | Write feeds one chunk through the stripper and returns the output produced by that chunk. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/striptags).

## Measured parity

Compared case-for-case against **`striptags@3.2.0`** by the harness in
[`parity/express/nested/striptags`](https://github.com/malcolmston/go/tree/main/parity/express/nested/striptags):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 116 |
| Matching | 109 |
| Mismatching | 0 |
| Declared deviations | 7 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/striptags/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
