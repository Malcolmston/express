# pluralize — pluralizes and singularizes English words

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/pluralize.svg)](https://pkg.go.dev/github.com/malcolmston/express/pluralize)
[![Parity](https://img.shields.io/badge/parity-100%25-brightgreen)](https://github.com/malcolmston/go/tree/main/parity/express/nested/pluralize)

Package pluralize pluralizes and singularizes English words. It is a faithful
port of the npm package "pluralize", reproducing its rule set, irregular and
uncountable word lists, and case-preserving behaviour using only the Go
standard library (regexp, strconv and strings).

English is used everywhere in software — UI labels, log messages, generated
API and ORM names — and constantly needs the correct grammatical number: "1
file" but "3 files", "1 person" but "2 people". Hand-writing that logic is
error-prone because English pluralization is deeply irregular. This package
centralizes the rules so callers can convert in either direction (Plural and
Singular) or merely ask which form a word is already in (IsPlural and
IsSingular).

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/pluralize"
```

## Usage

This is the package's own `ExampleIsPlural`, so it compiles and its output is
asserted on every `go test ./pluralize/`.

```go
fmt.Println(pluralize.IsPlural("people"))
	fmt.Println(pluralize.IsPlural("person"))
	fmt.Println(pluralize.IsSingular("person"))
	fmt.Println(pluralize.IsPlural("fish"))
```

```
true
false
true
true
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func IsPlural(word string) bool` | IsPlural reports whether the given word is already in plural form. |
| `func IsSingular(word string) bool` | IsSingular reports whether the given word is already in singular form. |
| `func Plural(word string) string` | Plural returns the plural form of the given word. |
| `func Singular(word string) string` | Singular returns the singular form of the given word. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/pluralize).

## Measured parity

Compared case-for-case against **`pluralize@8.0.0`** by the harness in
[`parity/express/nested/pluralize`](https://github.com/malcolmston/go/tree/main/parity/express/nested/pluralize):

| | |
| --- | --- |
| Parity | **100%** |
| Cases | 464 |
| Matching | 450 |
| Mismatching | 0 |
| Declared deviations | 14 |

Regenerate from the aggregator repo with `go test ./parity/express/nested/pluralize/`.
Declared deviations are documented differences, excluded from the denominator
and listed in the harness report.

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
