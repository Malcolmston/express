# groupby — provides a faithful port of lodash's `groupBy`

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/groupby.svg)](https://pkg.go.dev/github.com/malcolmston/express/groupby)

Package groupby provides a faithful port of lodash's `groupBy`. It partitions
a slice into groups according to a key derived from each element, returning a
map from key to the slice of elements that produced it.

Use this package whenever you need to bucket a collection by some computed
property: grouping records by category, numbers by parity, strings by length,
or structs by a field. It is the Go analogue of the collection helper that
lodash exposes as _.groupBy and that many JavaScript and Express codebases
rely on for shaping data before rendering or aggregation.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/groupby"
```

## Usage

This is the package's own `ExampleGroupBy`, so it compiles and its output is
asserted on every `go test ./groupby/`.

```go
in := []int{1, 2, 3, 4, 5, 6}
	groups := groupby.GroupBy(in, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	fmt.Println("odd:", groups["odd"])
	fmt.Println("even:", groups["even"])
```

```
odd: [1 3 5]
even: [2 4 6]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func GroupBy[T any, K comparable](s []T, key func(T) K) map[K][]T` | GroupBy partitions s into a map keyed by the value returned by key. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/groupby).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
