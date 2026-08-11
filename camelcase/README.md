# camelcase — converts dash/dot/underscore/space separated (and mixed-case) strings into camelCase

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/camelcase.svg)](https://pkg.go.dev/github.com/malcolmston/express/camelcase)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/camelcase)

Package camelcase converts dash/dot/underscore/space separated (and
mixed-case) strings into camelCase or PascalCase, modeled on the npm
"camelcase" package (sindresorhus/camelcase v9) that Express-adjacent tooling
uses to normalize option keys, identifiers, and header-derived names. It
exposes CamelCase for the default lower-first form, PascalCase for the
upper-first form, and CamelCaseWith for callers that want to choose the style
through an Options value, all built on only the Go standard library.

Reach for this package whenever human-written or wire-format names have to be
turned into program-friendly identifiers: converting "user-id" or "user_id"
from a config file into "userId", mapping "Content-Type" style header names to
struct-field-like tokens, or normalizing the several separator conventions a
single input source might mix together. Keeping the word-splitting rules in
one place means "foo-bar", "foo_bar", "foo bar", and "fooBar" all collapse to
the same canonical output regardless of how the caller happened to write them.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/camelcase"
```

## Usage

This is the package's own `ExampleCamelCase`, so it compiles and its output is
asserted on every `go test ./camelcase/`.

```go
fmt.Println(camelcase.CamelCase("foo-bar-baz"))
```

```
fooBarBaz
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CamelCase(s string) string` | CamelCase converts s to camelCase. |
| `func CamelCaseWith(s string, opts Options) string` | CamelCaseWith converts s to camelCase or PascalCase according to opts. |
| `func PascalCase(s string) string` | PascalCase converts s to PascalCase (also called UpperCamelCase). |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures how CamelCaseWith transforms a string. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/camelcase).

## Measured parity

Compared case-for-case against **`camelcase@9.0.0`** by the harness in
[`parity/express/nested/camelcase`](https://github.com/malcolmston/go/tree/main/parity/express/nested/camelcase):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 260 |
| Matching | 260 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/camelcase/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
