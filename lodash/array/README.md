# array — provides idiomatic Go generic ports of lodash's array utility functions

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/array.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/array)

Package array provides idiomatic Go generic ports of lodash's array utility
functions. It is a standalone package that depends only on the standard
library and mirrors the "Array" category of the npm "lodash" library (the
functions exposed individually as lodash.chunk, lodash.compact,
lodash.difference, lodash.flatten, and so on). The goal is to give Go code the
same everyday slice helpers a JavaScript project reaches for, without pulling
in any third-party dependency.

The package is useful whenever a program is doing the small, repetitive slice
manipulations that the standard library leaves to the caller: splitting a
slice into fixed-size batches (Chunk), removing empties (Compact),
deduplicating (Uniq), computing set relationships across slices (Difference,
Intersection, Union, Xor), slicing from either end (Take, Drop, and their
Right and While variants), searching (FindIndex, IndexOf), and reshaping (Zip,
Unzip, ZipObject, FromPairs, Flatten). Because each function is small and
pure, they compose cleanly and are easy to reason about in tests.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/array"
```

## Usage

This is the package's own `ExampleDifference`, so it compiles and its output is
asserted on every `go test ./array/`.

```go
fmt.Println(array.Difference([]int{1, 2, 3, 4}, []int{2, 4}))
```

```
[1 3]
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Chunk[T any](s []T, size int) [][]T` | Chunk splits s into groups of at most size elements. |
| `func Compact[T comparable](s []T) []T` | Compact returns a new slice with all zero-valued elements removed. |
| `func Concat[T any](slices ...[]T) []T` | Concat returns a new slice that is the concatenation of all provided slices. |
| `func Difference[T comparable](s []T, others ...[]T) []T` | Difference returns the elements of s that are not present in any of the other slices. |
| `func DifferenceBy[T any, K comparable](s []T, iteratee func(T) K, others ...[]T) []T` | DifferenceBy is like Difference except elements are compared by the value returned from applying iteratee to each element. |
| `func DifferenceWith[T any](s []T, values []T, eq func(a, b T) bool) []T` | DifferenceWith returns the elements of s for which no element of values is equal according to eq, preserving the order and first-occurrence of s. |
| `func Drop[T any](s []T, n int) []T` | Drop returns a new slice with the first n elements dropped. |
| `func DropRight[T any](s []T, n int) []T` | DropRight returns a new slice with the last n elements dropped. |
| `func DropRightWhile[T any](s []T, pred func(T) bool) []T` | DropRightWhile returns s without the trailing elements for which pred is true. |
| `func DropWhile[T any](s []T, pred func(T) bool) []T` | DropWhile returns a new slice excluding the leading elements for which pred returns true. |
| `func Fill[T any](s []T, value T, start, end int) []T` | Fill returns a new slice with the elements in the half-open range [start, end) replaced by value. |
| `func FindIndex[T any](s []T, pred func(T) bool) int` | FindIndex returns the index of the first element for which pred returns true, or -1 if none match. |
| `func FindLastIndex[T any](s []T, pred func(T) bool) int` | FindLastIndex returns the index of the last element for which pred returns true, or -1 if none match. |
| `func Flatten[T any](s [][]T) []T` | Flatten flattens s a single level deep. |
| `func FlattenDeep(s []any) []any` | FlattenDeep recursively flattens s. |
| `func FromPairs[K comparable, V any](pairs []Pair[K, V]) map[K]V` | FromPairs builds a map from a slice of key/value pairs. |
| `func Head[T any](s []T) (value T, ok bool)` | Head returns the first element of s. |
| `func IndexOf[T comparable](s []T, value T) int` | IndexOf returns the index of the first occurrence of value in s, or -1 if it is not present. |
| `func Initial[T any](s []T) []T` | Initial returns a new slice with all but the last element of s. |
| `func Intersection[T comparable](arrays ...[]T) []T` | Intersection returns the unique elements that are present in every one of the provided slices, in the order they first appear in the first slice. |
| `func IntersectionBy[T any, K comparable](iteratee func(T) K, arrays ...[]T) []T` | IntersectionBy is like Intersection except elements are compared by the value returned from applying iteratee to each element. |
| `func IntersectionWith[T any](a, b []T, eq func(x, y T) bool) []T` | IntersectionWith returns the elements present in both a and b according to eq, in the order they appear in a and without duplicates. |
| `func Last[T any](s []T) (value T, ok bool)` | Last returns the last element of s. |
| `func LastIndexOf[T comparable](s []T, value T) int` | LastIndexOf returns the index of the last occurrence of value in s, or -1 if it is not present. |
| `func Nth[T any](s []T, n int) (value T, ok bool)` | Nth returns the element at index n. |
| `func Pull[T comparable](s []T, values ...T) []T` | Pull returns a new slice with all occurrences of the given values removed. |
| `func PullAll[T comparable](s []T, values []T) []T` | PullAll is like Pull except it accepts the values to remove as a single slice. |
| `func PullAt[T any](slice []T, indexes ...int) []T` | PullAt returns a copy of slice with the elements at the given indexes removed. |
| `func Range(start, end, step int) []int` | Range returns a slice of integers from start (inclusive) up to end (exclusive), stepping by step. |
| `func RangeRight(start, end, step int) []int` | RangeRight is like Range except the resulting values are in descending (reversed) order. |
| `func Remove[T any](s []T, pred func(T) bool) []T` | Remove returns a new slice containing the elements of s for which pred returns false. |
| `func Reverse[T any](s []T) []T` | Reverse returns a new slice with the elements of s in reverse order. |
| `func Slice[T any](s []T, start, end int) []T` | Slice returns a copy of the portion of s in the half-open range [start, end). |
| `func SortedIndex[T cmp.Ordered](s []T, value T) int` | SortedIndex returns the lowest index at which value can be inserted into the sorted slice s to keep it sorted. |
| `func SortedIndexBy[T any, K cmp.Ordered](s []T, value T, iteratee func(T) K) int` | SortedIndexBy returns the lowest index at which value should be inserted into s (assumed sorted by the key returned by iteratee) to keep it sorted. |
| `func SortedIndexOf[T cmp.Ordered](s []T, value T) int` | SortedIndexOf returns the index of the first occurrence of value in the sorted slice s using binary search, or -1 if value is not present. |
| `func SortedLastIndex[T cmp.Ordered](s []T, value T) int` | SortedLastIndex returns the highest index at which value should be inserted into the sorted slice s to keep it sorted (after any equal elements). |
| `func SortedLastIndexOf[T cmp.Ordered](s []T, value T) int` | SortedLastIndexOf returns the index of the last occurrence of value in the sorted slice s using binary search, or -1 if value is not present. |
| `func SortedUniq[T comparable](s []T) []T` | SortedUniq returns a new slice with consecutive duplicate elements removed. |
| `func SortedUniqBy[T any, K comparable](s []T, iteratee func(T) K) []T` | SortedUniqBy returns a new slice with consecutive elements sharing a key (as returned by iteratee) collapsed to their first occurrence. |
| `func Tail[T any](s []T) []T` | Tail returns a new slice with all but the first element of s. |
| `func Take[T any](s []T, n int) []T` | Take returns a new slice with the first n elements. |
| `func TakeRight[T any](s []T, n int) []T` | TakeRight returns a new slice with the last n elements. |
| `func TakeRightWhile[T any](s []T, pred func(T) bool) []T` | TakeRightWhile returns the trailing elements of s for which pred is true, stopping at the last element from the end for which pred is false. |
| `func TakeWhile[T any](s []T, pred func(T) bool) []T` | TakeWhile returns a new slice of the leading elements for which pred returns true. |
| `func Union[T comparable](arrays ...[]T) []T` | Union returns the unique elements of all provided slices, in first-seen order. |
| `func UnionBy[T any, K comparable](iteratee func(T) K, arrays ...[]T) []T` | UnionBy is like Union except uniqueness is determined by the value returned from applying iteratee to each element. |
| `func UnionWith[T any](eq func(x, y T) bool, slices ...[]T) []T` | UnionWith returns the unique elements of all slices, treating two elements as equal when eq reports so, in first-occurrence order. |
| `func Uniq[T comparable](s []T) []T` | Uniq returns a new slice with duplicate elements removed, keeping the first occurrence of each. |
| `func UniqBy[T any, K comparable](s []T, iteratee func(T) K) []T` | UniqBy is like Uniq except uniqueness is determined by the value returned from applying iteratee to each element. |
| `func Unzip[T any](groups [][]T) [][]T` | Unzip is the inverse of Zip: it regroups grouped elements by their position within each group. |
| `func UnzipWith[T, R any](fn func(group []T) R, grouped [][]T) []R` | UnzipWith is the inverse of ZipWith: it applies fn to each column of grouped and returns the slice of results. |
| `func Without[T comparable](s []T, values ...T) []T` | Without returns a new slice with all occurrences of the given values removed. |
| `func Xor[T comparable](arrays ...[]T) []T` | Xor returns the symmetric difference of the provided slices: the unique elements that appear in exactly one of the slices. |
| `func XorBy[T any, K comparable](iteratee func(T) K, slices ...[]T) []T` | XorBy returns the symmetric difference of the slices, comparing elements by the key returned by iteratee: values whose key appears in exactly one… |
| `func XorWith[T any](eq func(x, y T) bool, slices ...[]T) []T` | XorWith returns the symmetric difference of the slices using eq for equality. |
| `func Zip[T any](arrays ...[]T) [][]T` | Zip groups the elements of the provided slices by index. |
| `func ZipObject[K comparable, V any](keys []K, values []V) map[K]V` | ZipObject builds a map by pairing each key with the value at the same index. |
| `func ZipWith[T, R any](fn func(group []T) R, slices ...[]T) []R` | ZipWith groups the i-th elements of each slice and applies fn to each group, returning the slice of results. |

### Types

| Type | What it is |
| --- | --- |
| `Pair` | Pair is a key/value pair used by FromPairs and produced by other utilities that model lodash's two-element array pairs. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/array).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
