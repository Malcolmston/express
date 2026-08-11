# chunk — provides a faithful port of lodash's `chunk` utility using only the Go standard library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/chunk.svg)](https://pkg.go.dev/github.com/malcolmston/express/chunk)

Package chunk provides a faithful port of lodash's `chunk` utility using only
the Go standard library. In JavaScript, `_.chunk(array, size)` breaks an array
into a list of smaller arrays each holding at most `size` elements; this
package exposes the same behavior through the generic Chunk function so it
works for slices of any element type.

You reach for Chunk whenever a flat slice needs to be processed or displayed
in fixed-size batches: paginating a result set, rendering a grid with a set
number of columns per row, sending records to an API that caps how many can be
submitted per request, or bounding how much work a single goroutine picks up.
It keeps the batching arithmetic out of the caller so the surrounding code can
focus on what each group does.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/chunk"
```

## Usage

This is the package's own `ExampleChunk_strings`, so it compiles and its output is
asserted on every `go test ./chunk/`.

```go
groups := chunk.Chunk([]string{"a", "b", "c", "d"}, 2)
	fmt.Println(groups)
```

```
[[a b] [c d]]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Chunk[T any](s []T, size int) [][]T` | Chunk splits s into groups of at most size elements. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/chunk).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
