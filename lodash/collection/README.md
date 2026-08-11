# collection — standalone, dependency-free port of lodash's "Collection" category of functions

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/collection.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/collection)

Package collection is a standalone, dependency-free port of lodash's
"Collection" category of functions (https://lodash.com/docs - _.countBy,
_.groupBy, _.keyBy, _.partition, _.map, _.filter, _.reduce, _.sortBy,
_.orderBy, _.sample, _.shuffle and friends) to Go generics. It operates over
Go slices ([]T) and depends only on the Go standard library (cmp, sort,
math/rand, crypto/rand), so it can be dropped into any project without pulling
in third-party code.

The package exists to give Go programmers the ergonomic, expression-oriented
data-wrangling vocabulary that JavaScript developers reach for by habit. In
Node a call such as _.groupBy(users, u => u.role) reads clearly and returns a
plain object; the naive Go equivalent is a hand-rolled loop plus a map
declaration at every call site. These helpers collapse that boilerplate into a
single, type-safe call while remaining allocation-friendly and easy to reason
about. Reach for them when porting Express/lodash JavaScript to Go, or simply
when you want map/filter/reduce style pipelines without giving up static
typing.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/collection"
```

## Usage

This is the package's own `ExampleCountBy`, so it compiles and its output is
asserted on every `go test ./collection/`.

```go
nums := []int{1, 2, 3, 4, 5}
	counts := collection.CountBy(nums, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	fmt.Println("even:", counts["even"])
	fmt.Println("odd:", counts["odd"])
```

```
even: 2
odd: 3
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CountBy[T any, K comparable](s []T, iteratee func(T) K) map[K]int` | CountBy creates a map of keys generated from iteratee to the number of times that key occurred across the collection. |
| `func CountWhere[T any](s []T, predicate func(T) bool) int` | CountWhere returns the number of elements of s that satisfy predicate. |
| `func Each[T any](s []T, iteratee func(T) bool)` | Each is an alias for ForEach. |
| `func Every[T any](s []T, predicate func(T) bool) bool` | Every reports whether predicate returns true for all elements. |
| `func Filter[T any](s []T, predicate func(T) bool) []T` | Filter returns a new slice of the elements for which predicate returns true. |
| `func FilterWithIndex[T any](s []T, predicate func(value T, index int) bool) []T` | FilterWithIndex returns the elements of s for which predicate(element, index) is true, the index-aware form of Filter. |
| `func Find[T any](s []T, predicate func(T) bool) (T, bool)` | Find returns the first element for which predicate returns true. |
| `func FindLast[T any](s []T, predicate func(T) bool) (T, bool)` | FindLast returns the last element for which predicate returns true, iterating from the end. |
| `func FlatMap[T, R any](s []T, iteratee func(T) []R) []R` | FlatMap maps each element through iteratee (which returns a slice) and flattens the results one level deep. |
| `func FlatMapWithIndex[T, R any](s []T, iteratee func(value T, index int) []R) []R` | FlatMapWithIndex maps each element through iteratee(element, index), which returns a slice, and concatenates the results, the index-aware form of… |
| `func ForEach[T any](s []T, iteratee func(T) bool)` | ForEach iterates over the collection invoking iteratee for each element. |
| `func ForEachRight[T any](s []T, iteratee func(T) bool)` | ForEachRight invokes iteratee for each element of s from right to left, stopping early if iteratee returns false. |
| `func GroupBy[T any, K comparable](s []T, iteratee func(T) K) map[K][]T` | GroupBy creates a map of keys generated from iteratee to slices of the elements responsible for generating each key. |
| `func Includes[T comparable](s []T, value T) bool` | Includes reports whether value is an element of the collection. |
| `func InvokeMap[T, R any](s []T, method func(T, ...any) R, args ...any) []R` | InvokeMap invokes method for each element of the collection, passing args to each call, and returns a slice of the results. |
| `func KeyBy[T any, K comparable](s []T, iteratee func(T) K) map[K]T` | KeyBy creates a map keyed by the result of iteratee for each element. |
| `func Map[T, R any](s []T, iteratee func(T) R) []R` | Map creates a new slice of values by running each element through iteratee. |
| `func MapWithIndex[T, R any](s []T, iteratee func(value T, index int) R) []R` | MapWithIndex returns a new slice holding iteratee(element, index) for each element of s, the index-aware form of Map. |
| `func MaxBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) (T, bool)` | MaxBy returns the element that yields the largest value when passed through iteratee. |
| `func MeanBy[T any, N Number](s []T, iteratee func(T) N) float64` | MeanBy returns the mean (as float64) of the values produced by iteratee. |
| `func MinBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) (T, bool)` | MinBy returns the element that yields the smallest value when passed through iteratee. |
| `func OrderBy[T any](s []T, keys []func(T) any, orders []string) []T` | OrderBy returns a new slice sorted by multiple key functions with a matching direction for each ("asc" or "desc"; any other value is treated as… |
| `func Partition[T any](s []T, predicate func(T) bool) (truthy []T, falsy []T)` | Partition splits the collection into two slices: the first holds elements for which predicate returns true, the second holds the rest. |
| `func Reduce[T, R any](s []T, iteratee func(acc R, cur T) R, accumulator R) R` | Reduce reduces the collection to a single accumulated value, iterating from left to right. |
| `func ReduceRight[T, R any](s []T, iteratee func(acc R, cur T) R, accumulator R) R` | ReduceRight is like Reduce except it iterates from right to left. |
| `func ReduceWithIndex[T, R any](s []T, iteratee func(acc R, cur T, index int) R, accumulator R) R` | ReduceWithIndex folds s left-to-right into an accumulator, passing each element's index to the iteratee. |
| `func Reject[T any](s []T, predicate func(T) bool) []T` | Reject is the opposite of Filter: it returns the elements for which predicate returns false. |
| `func Sample[T any](s []T, r *rand.Rand) (T, bool)` | Sample returns a random element from the collection. |
| `func SampleSize[T any](s []T, n int, r *rand.Rand) []T` | SampleSize returns n elements sampled without replacement from the collection, in random order. |
| `func Shuffle[T any](s []T, r *rand.Rand) []T` | Shuffle returns a new slice with the elements of the collection in random order, using the Fisher-Yates algorithm. |
| `func Size[T any](s []T) int` | Size returns the number of elements in the collection. |
| `func Some[T any](s []T, predicate func(T) bool) bool` | Some reports whether predicate returns true for any element. |
| `func SortBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) []T` | SortBy returns a new slice sorted in ascending order by the key produced by iteratee. |
| `func SumBy[T any, N Number](s []T, iteratee func(T) N) N` | SumBy sums the values produced by running each element through iteratee. |
| `func Tap[T any](value T, interceptor func(T)) T` | Tap invokes interceptor with value for its side effects and returns value unchanged, allowing a value to be inspected or logged inside a pipeline. |
| `func Thru[T, R any](value T, interceptor func(T) R) R` | Thru passes value through interceptor and returns the interceptor's result, letting a pipeline transform a value inline. |

### Types

| Type | What it is |
| --- | --- |
| `Number` | Number is the set of types over which SumBy and MeanBy operate. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/collection).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
