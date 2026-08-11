# flatten — provides faithful ports of lodash's flatten, flattenDeep, and flattenDepth

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/flatten.svg)](https://pkg.go.dev/github.com/malcolmston/express/flatten)

Package flatten provides faithful ports of lodash's flatten, flattenDeep, and
flattenDepth.

The three functions collapse nested slices into a flatter slice. Use Flatten
to merge one level of nesting with full compile-time type safety, use
FlattenDeep to collapse an arbitrarily nested structure completely, and use
FlattenDepth when you need to control exactly how many levels to remove. They
are handy for normalizing grouped results, splicing together lists of lists,
or unwrapping data that arrived more deeply nested than the consumer wants.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/flatten"
```

## Usage

This is the package's own `ExampleFlattenDepth`, so it compiles and its output is
asserted on every `go test ./flatten/`.

```go
nested := []any{1, []any{2, []any{3, []any{4}}}, 5}
	fmt.Println(flatten.FlattenDepth(nested, 1))
	fmt.Println(flatten.FlattenDepth(nested, 2))
```

```
[1 2 [3 [4]] 5]
[1 2 3 [4] 5]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Flatten[T any](s [][]T) []T` | Flatten flattens s by a single level, concatenating all inner slices into one new slice in order. |
| `func FlattenDeep(s any) []any` | FlattenDeep recursively flattens an arbitrarily nested value into a flat []any. |
| `func FlattenDepth(s any, depth int) []any` | FlattenDepth recursively flattens a nested value up to depth levels deep, mirroring lodash's flattenDepth. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/flatten).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
