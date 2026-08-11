# uniq — removes duplicate elements from a slice, preserving order

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/uniq.svg)](https://pkg.go.dev/github.com/malcolmston/express/uniq)

Package uniq removes duplicate elements from a slice, preserving order. It is
a standard-library-only Go port of lodash's _.uniq and _.uniqBy
(https://lodash.com/docs/#uniq), two of the most heavily used array helpers in
the JavaScript ecosystem for de-duplicating lists while keeping the first
occurrence of each value. Deduplication is a ubiquitous need — collapsing
tag lists, cleaning user input, building sets of ids — and this package
packages it behind two small generic functions.

Uniq handles the value-equality case. It walks the input once and returns a
new slice containing only the first occurrence of each distinct element, in
the order those elements first appeared. It is generic over any comparable
type, so it works on slices of strings, ints, or any type usable as a map key
without the caller writing an equality function. The result is always a
freshly allocated slice; the input is never modified.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/uniq"
```

## Usage

This is the package's own `ExampleUniqBy`, so it compiles and its output is
asserted on every `go test ./uniq/`.

```go
words := []string{"apple", "banana", "avocado", "cherry", "blueberry"}
	fmt.Println(uniq.UniqBy(words, func(s string) byte { return s[0] }))
```

```
[apple banana cherry]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Uniq[T comparable](s []T) []T` | Uniq returns a new slice containing the first occurrence of each distinct element of s, preserving the order in which those elements first appear. |
| `func UniqBy[T any, K comparable](s []T, key func(T) K) []T` | UniqBy returns a new slice containing the first element of s for each distinct value produced by key. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/uniq).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
