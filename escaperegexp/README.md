# escaperegexp — escapes regular-expression metacharacters in a string so the string can be embedded

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/escaperegexp.svg)](https://pkg.go.dev/github.com/malcolmston/express/escaperegexp)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/escaperegexp)

Package escaperegexp escapes regular-expression metacharacters in a string so
the string can be embedded as a literal inside a larger pattern. It is a
faithful, standard-library-only Go port of the npm package
escape-string-regexp, exposing a single function, EscapeRegExp, that takes an
arbitrary string and returns a version in which every character that would
otherwise carry special meaning to a regex engine has been neutralized with a
backslash.

The problem it solves appears whenever user- or data-derived text has to be
spliced into a pattern: a search term typed into a box, a filename, a
delimiter read from configuration. Concatenating such a value directly into a
regular expression is both a correctness bug (a "." in the input would match
any character, "(" would open an unbalanced group and fail to compile) and a
denial-of-service risk (a crafted value can build a catastrophically
backtracking pattern). Escaping the value first guarantees it is matched byte
for byte, exactly as written.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/escaperegexp"
```

## Usage

This is the package's own `ExampleEscapeRegExp_compile`, so it compiles and its output is
asserted on every `go test ./escaperegexp/`.

```go
re := regexp.MustCompile(escaperegexp.EscapeRegExp("1+1=2?"))
	fmt.Println(re.MatchString("1+1=2?"))
```

```
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func EscapeRegExp(s string) string` | EscapeRegExp escapes the characters \|\{}()[]^$+*?. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/escaperegexp).

## Measured parity

Compared case-for-case against **`escape-string-regexp@5.0.0`** by the harness in
[`parity/express/nested/escaperegexp`](https://github.com/malcolmston/go/tree/main/parity/express/nested/escaperegexp):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 61 |
| Matching | 61 |
| Mismatching | 0 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/escaperegexp/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
