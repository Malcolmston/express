# titlecase — converts strings to Title Case, matching the behavior of the npm "title-case" package

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/titlecase.svg)](https://pkg.go.dev/github.com/malcolmston/express/titlecase)

Package titlecase converts strings to Title Case, matching the behavior of the
npm "title-case" package from the change-case family
(blakeembrey/change-case). It is a faithful port of that library's algorithm
(packages/title-case/src/index.ts), preserving the original characters of the
input and only upper-casing the first letter of each word, with English "small
word" handling, sentence detection, and special-casing of URLs, acronyms, and
manually cased words such as "iPhone".

Unlike a naive title-caser, this package never lower-cases letters: acronyms
like "NASA" and manually cased words like "camelCase" are preserved as-is,
punctuation and whitespace (including newlines and repeated spaces) are kept
intact, and small function words such as "a", "of", and "the" are left in
lower case except at the start or end of a sentence.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/titlecase"
```

## Usage

This is the package's own `ExampleTitleCase_preserved`, so it compiles and its output is
asserted on every `go test ./titlecase/`.

```go
fmt.Println(titlecase.TitleCase("we keep NASA capitalized"))
	fmt.Println(titlecase.TitleCase("pass camelCase through"))
```

```
We Keep NASA Capitalized
Pass camelCase Through
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func TitleCase(s string, opts ...Options) string` | TitleCase converts s to Title Case following the upstream change-case algorithm. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures TitleCase. |

### Variables

`DefaultSmallWords`

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/titlecase).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
