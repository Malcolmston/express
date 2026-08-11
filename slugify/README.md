# slugify — converts strings into URL-safe slugs

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/slugify.svg)](https://pkg.go.dev/github.com/malcolmston/express/slugify)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/slugify)

Package slugify converts strings into URL-safe slugs, mirroring the behavior
of the npm "slugify" library (simov/slugify). A slug is the lower-friction,
path-safe form of a title used in URLs, filenames, and anchor identifiers, for
example turning "Crème brûlée" into "creme-brulee". The goal is a
deterministic, readable ASCII string that survives copy-pasting and routing
unchanged.

The transformation runs in a few stages, matching the reference
implementation. Each input rune is looked up in charMap, a transliteration
table that maps accented Latin letters, Greek, Cyrillic, Turkish, Armenian,
Arabic, Georgian, Vietnamese, currency, and symbol characters to ASCII
equivalents (for example 'é' to "e", 'ß' to "ss", '&' to "and", and '©' to
"(c)"). A rune with no mapping is passed through as-is. If a mapped result
equals the separator it is turned into a space so separator runs collapse.
After mapping, each appended fragment has any character that is not in the
allowed set removed (the default keeps word characters, whitespace, and the
punctuation $*_+~.()'"!-:@). Then, optionally, strict filtering drops every
non-alphanumeric, non-space character; whitespace is trimmed; whitespace runs
are replaced by the separator; and the result is optionally lowercased.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/slugify"
```

## Usage

This is the package's own `ExampleSlugify_transliterate`, so it compiles and its output is
asserted on every `go test ./slugify/`.

```go
fmt.Println(slugify.Slugify("Crème Brûlée", slugify.Options{Lower: true}))
```

```
creme-brulee
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Slugify(s string, opts ...Options) string` | Slugify converts s into a slug. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures Slugify. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/slugify).

## Measured parity

Compared case-for-case against **`slugify@1.6.9`** by the harness in
[`parity/express/nested/slugify`](https://github.com/malcolmston/go/tree/main/parity/express/nested/slugify):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 75 |
| Matching | 70 |
| Mismatching | 0 |
| Declared deviations | 5 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/slugify/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
