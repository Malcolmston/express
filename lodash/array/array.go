// Package array provides idiomatic Go generic ports of lodash's array
// utility functions. It is a standalone package that depends only on the
// standard library and mirrors the "Array" category of the npm "lodash"
// library (the functions exposed individually as lodash.chunk, lodash.compact,
// lodash.difference, lodash.flatten, and so on). The goal is to give Go code
// the same everyday slice helpers a JavaScript project reaches for, without
// pulling in any third-party dependency.
//
// The package is useful whenever a program is doing the small, repetitive slice
// manipulations that the standard library leaves to the caller: splitting a
// slice into fixed-size batches (Chunk), removing empties (Compact),
// deduplicating (Uniq), computing set relationships across slices (Difference,
// Intersection, Union, Xor), slicing from either end (Take, Drop, and their
// Right and While variants), searching (FindIndex, IndexOf), and reshaping
// (Zip, Unzip, ZipObject, FromPairs, Flatten). Because each function is small
// and pure, they compose cleanly and are easy to reason about in tests.
//
// Everything is built on Go generics. Functions that only move elements around
// are parameterized on [T any]; functions that must compare or hash elements
// (Compact, Difference, Intersection, Union, Xor, Uniq, Without, Pull, IndexOf,
// and friends) constrain T to [comparable]; SortedIndex requires
// [T cmp.Ordered] so it can binary-search; and the "By" variants add a second
// type parameter K for the key returned by an iteratee, for example
// UniqBy[T any, K comparable]. This means examples and call sites frequently
// name their type arguments implicitly through inference but must match the
// exact constraints when a value cannot be inferred.
//
// Several conventions are applied uniformly. Input slices are never mutated:
// functions that lodash documents as mutating (Pull, PullAll, Remove, Reverse,
// Fill) instead operate on a copy and return a new slice, and slice-returning
// helpers hand back a fresh backing array rather than an alias into the input.
// Where lodash relies on JavaScript "falsy" values, the Go ports use the zero
// value of the element type, so Compact drops 0, "", and the like. Where lodash
// accepts an "iteratee", the Go ports accept an explicit key or predicate
// function. Functions that may fail to produce an element (Head, Last, Nth)
// return an additional boolean reporting whether a value was found, in place of
// JavaScript's undefined.
//
// Edge cases follow lodash's spirit while staying explicit about bounds. Count
// arguments to Take, Drop, and Slice are clamped into range rather than
// panicking, and negative indexes count from the end in Nth, Fill, and Slice
// (so Nth(s, -1) is the last element). Chunk with a size of zero or less yields
// an empty result, Range treats a zero step as +1 or -1 depending on the
// direction of travel, and set operations preserve first-seen order and
// collapse duplicates the way the corresponding lodash functions do. The main
// intentional divergence from Node is FlattenDeep, which cannot be fully static
// in Go's type system and therefore operates on []any, recursively flattening
// nested []any values while leaving every other value untouched.
package array

import "cmp"

// Compact returns a new slice with all zero-valued elements removed. This is
// the analogue of lodash's _.compact, which drops "falsy" values; in Go the
// zero value of the element type plays that role.
//
//	Compact([]int{0, 1, 0, 2, 3}) => [1 2 3]
func Compact[T comparable](s []T) []T {
	var zero T
	result := make([]T, 0, len(s))
	for _, v := range s {
		if v != zero {
			result = append(result, v)
		}
	}
	return result
}

// Chunk splits s into groups of at most size elements. The final chunk holds
// the remaining elements. If size is less than or equal to zero, an empty
// result is returned.
//
//	Chunk([]string{"a", "b", "c", "d"}, 2) => [[a b] [c d]]
//	Chunk([]string{"a", "b", "c", "d"}, 3) => [[a b c] [d]]
func Chunk[T any](s []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}
	result := make([][]T, 0, (len(s)+size-1)/size)
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		chunk := make([]T, end-i)
		copy(chunk, s[i:end])
		result = append(result, chunk)
	}
	return result
}

// Difference returns the elements of s that are not present in any of the
// other slices. Order and duplicates from s are preserved (each surviving
// element appears as often as it does in s).
//
//	Difference([]int{2, 1}, []int{2, 3}) => [1]
func Difference[T comparable](s []T, others ...[]T) []T {
	exclude := make(map[T]struct{})
	for _, other := range others {
		for _, v := range other {
			exclude[v] = struct{}{}
		}
	}
	result := make([]T, 0, len(s))
	for _, v := range s {
		if _, found := exclude[v]; !found {
			result = append(result, v)
		}
	}
	return result
}

// DifferenceBy is like Difference except elements are compared by the value
// returned from applying iteratee to each element.
//
//	DifferenceBy([]float64{2.1, 1.2}, math.Floor, []float64{2.3, 3.4}) => [1.2]
func DifferenceBy[T any, K comparable](s []T, iteratee func(T) K, others ...[]T) []T {
	exclude := make(map[K]struct{})
	for _, other := range others {
		for _, v := range other {
			exclude[iteratee(v)] = struct{}{}
		}
	}
	result := make([]T, 0, len(s))
	for _, v := range s {
		if _, found := exclude[iteratee(v)]; !found {
			result = append(result, v)
		}
	}
	return result
}

// Intersection returns the unique elements that are present in every one of
// the provided slices, in the order they first appear in the first slice.
//
//	Intersection([]int{2, 1}, []int{2, 3}) => [2]
func Intersection[T comparable](arrays ...[]T) []T {
	if len(arrays) == 0 {
		return []T{}
	}
	result := make([]T, 0, len(arrays[0]))
	seen := make(map[T]struct{})
	for _, v := range arrays[0] {
		if _, done := seen[v]; done {
			continue
		}
		inAll := true
		for _, other := range arrays[1:] {
			if !contains(other, v) {
				inAll = false
				break
			}
		}
		if inAll {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// IntersectionBy is like Intersection except elements are compared by the
// value returned from applying iteratee to each element.
//
//	IntersectionBy(math.Floor, []float64{2.1, 1.2}, []float64{2.3, 3.4}) => [2.1]
func IntersectionBy[T any, K comparable](iteratee func(T) K, arrays ...[]T) []T {
	if len(arrays) == 0 {
		return []T{}
	}
	// Precompute key sets for the other slices.
	otherKeys := make([]map[K]struct{}, len(arrays)-1)
	for i, other := range arrays[1:] {
		set := make(map[K]struct{}, len(other))
		for _, v := range other {
			set[iteratee(v)] = struct{}{}
		}
		otherKeys[i] = set
	}
	result := make([]T, 0, len(arrays[0]))
	seen := make(map[K]struct{})
	for _, v := range arrays[0] {
		k := iteratee(v)
		if _, done := seen[k]; done {
			continue
		}
		inAll := true
		for _, set := range otherKeys {
			if _, found := set[k]; !found {
				inAll = false
				break
			}
		}
		if inAll {
			seen[k] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Union returns the unique elements of all provided slices, in first-seen
// order.
//
//	Union([]int{2}, []int{1, 2}) => [2 1]
func Union[T comparable](arrays ...[]T) []T {
	result := make([]T, 0)
	seen := make(map[T]struct{})
	for _, arr := range arrays {
		for _, v := range arr {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// UnionBy is like Union except uniqueness is determined by the value returned
// from applying iteratee to each element.
//
//	UnionBy(math.Floor, []float64{2.1}, []float64{1.2, 2.3}) => [2.1 1.2]
func UnionBy[T any, K comparable](iteratee func(T) K, arrays ...[]T) []T {
	result := make([]T, 0)
	seen := make(map[K]struct{})
	for _, arr := range arrays {
		for _, v := range arr {
			k := iteratee(v)
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Without returns a new slice with all occurrences of the given values
// removed. Order of the remaining elements is preserved.
//
//	Without([]int{2, 1, 2, 3}, 1, 2) => [3]
func Without[T comparable](s []T, values ...T) []T {
	return Difference(s, values)
}

// Xor returns the symmetric difference of the provided slices: the unique
// elements that appear in exactly one of the slices.
//
//	Xor([]int{2, 1}, []int{2, 3}) => [1 3]
func Xor[T comparable](arrays ...[]T) []T {
	counts := make(map[T]int)
	// Count in how many distinct slices each value appears.
	for _, arr := range arrays {
		local := make(map[T]struct{})
		for _, v := range arr {
			local[v] = struct{}{}
		}
		for v := range local {
			counts[v]++
		}
	}
	result := make([]T, 0)
	seen := make(map[T]struct{})
	for _, arr := range arrays {
		for _, v := range arr {
			if counts[v] != 1 {
				continue
			}
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// Uniq returns a new slice with duplicate elements removed, keeping the first
// occurrence of each.
//
//	Uniq([]int{2, 1, 2}) => [2 1]
func Uniq[T comparable](s []T) []T {
	result := make([]T, 0, len(s))
	seen := make(map[T]struct{})
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// UniqBy is like Uniq except uniqueness is determined by the value returned
// from applying iteratee to each element.
//
//	UniqBy([]float64{2.1, 1.2, 2.3}, math.Floor) => [2.1 1.2]
func UniqBy[T any, K comparable](s []T, iteratee func(T) K) []T {
	result := make([]T, 0, len(s))
	seen := make(map[K]struct{})
	for _, v := range s {
		k := iteratee(v)
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Drop returns a new slice with the first n elements dropped. n is clamped to
// the range [0, len(s)].
//
//	Drop([]int{1, 2, 3}, 2) => [3]
func Drop[T any](s []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return clone(s[n:])
}

// DropRight returns a new slice with the last n elements dropped. n is clamped
// to the range [0, len(s)].
//
//	DropRight([]int{1, 2, 3}, 2) => [1]
func DropRight[T any](s []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return clone(s[:len(s)-n])
}

// DropWhile returns a new slice excluding the leading elements for which pred
// returns true.
//
//	DropWhile([]int{1, 2, 3, 4}, func(v int) bool { return v < 3 }) => [3 4]
func DropWhile[T any](s []T, pred func(T) bool) []T {
	i := 0
	for i < len(s) && pred(s[i]) {
		i++
	}
	return clone(s[i:])
}

// Take returns a new slice with the first n elements. n is clamped to the
// range [0, len(s)].
//
//	Take([]int{1, 2, 3}, 2) => [1 2]
func Take[T any](s []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return clone(s[:n])
}

// TakeRight returns a new slice with the last n elements. n is clamped to the
// range [0, len(s)].
//
//	TakeRight([]int{1, 2, 3}, 2) => [2 3]
func TakeRight[T any](s []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(s) {
		n = len(s)
	}
	return clone(s[len(s)-n:])
}

// TakeWhile returns a new slice of the leading elements for which pred returns
// true.
//
//	TakeWhile([]int{1, 2, 3, 4}, func(v int) bool { return v < 3 }) => [1 2]
func TakeWhile[T any](s []T, pred func(T) bool) []T {
	i := 0
	for i < len(s) && pred(s[i]) {
		i++
	}
	return clone(s[:i])
}

// Head returns the first element of s. ok is false when s is empty, in which
// case the zero value is returned.
//
//	Head([]int{1, 2, 3}) => 1, true
func Head[T any](s []T) (value T, ok bool) {
	if len(s) == 0 {
		return value, false
	}
	return s[0], true
}

// Tail returns a new slice with all but the first element of s.
//
//	Tail([]int{1, 2, 3}) => [2 3]
func Tail[T any](s []T) []T {
	if len(s) == 0 {
		return []T{}
	}
	return clone(s[1:])
}

// Initial returns a new slice with all but the last element of s.
//
//	Initial([]int{1, 2, 3}) => [1 2]
func Initial[T any](s []T) []T {
	if len(s) == 0 {
		return []T{}
	}
	return clone(s[:len(s)-1])
}

// Last returns the last element of s. ok is false when s is empty, in which
// case the zero value is returned.
//
//	Last([]int{1, 2, 3}) => 3, true
func Last[T any](s []T) (value T, ok bool) {
	if len(s) == 0 {
		return value, false
	}
	return s[len(s)-1], true
}

// Nth returns the element at index n. A negative n counts from the end. ok is
// false when the resolved index is out of range.
//
//	Nth([]int{1, 2, 3}, 1)  => 2, true
//	Nth([]int{1, 2, 3}, -1) => 3, true
func Nth[T any](s []T, n int) (value T, ok bool) {
	if n < 0 {
		n += len(s)
	}
	if n < 0 || n >= len(s) {
		return value, false
	}
	return s[n], true
}

// FindIndex returns the index of the first element for which pred returns
// true, or -1 if none match.
//
//	FindIndex([]int{1, 2, 3}, func(v int) bool { return v == 2 }) => 1
func FindIndex[T any](s []T, pred func(T) bool) int {
	for i, v := range s {
		if pred(v) {
			return i
		}
	}
	return -1
}

// FindLastIndex returns the index of the last element for which pred returns
// true, or -1 if none match.
//
//	FindLastIndex([]int{1, 2, 1}, func(v int) bool { return v == 1 }) => 2
func FindLastIndex[T any](s []T, pred func(T) bool) int {
	for i := len(s) - 1; i >= 0; i-- {
		if pred(s[i]) {
			return i
		}
	}
	return -1
}

// IndexOf returns the index of the first occurrence of value in s, or -1 if it
// is not present.
//
//	IndexOf([]int{1, 2, 1, 2}, 2) => 1
func IndexOf[T comparable](s []T, value T) int {
	for i, v := range s {
		if v == value {
			return i
		}
	}
	return -1
}

// LastIndexOf returns the index of the last occurrence of value in s, or -1 if
// it is not present.
//
//	LastIndexOf([]int{1, 2, 1, 2}, 2) => 3
func LastIndexOf[T comparable](s []T, value T) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == value {
			return i
		}
	}
	return -1
}

// Fill returns a new slice with the elements in the half-open range
// [start, end) replaced by value. Negative indexes count from the end and
// both bounds are clamped to the range [0, len(s)].
//
//	Fill([]int{1, 2, 3, 4}, 0, 1, 3) => [1 0 0 4]
func Fill[T any](s []T, value T, start, end int) []T {
	result := clone(s)
	n := len(result)
	if start < 0 {
		start += n
	}
	if end < 0 {
		end += n
	}
	if start < 0 {
		start = 0
	}
	if end > n {
		end = n
	}
	for i := start; i < end; i++ {
		result[i] = value
	}
	return result
}

// Flatten flattens s a single level deep.
//
//	Flatten([][]int{{1, 2}, {3, 4}}) => [1 2 3 4]
func Flatten[T any](s [][]T) []T {
	result := make([]T, 0)
	for _, inner := range s {
		result = append(result, inner...)
	}
	return result
}

// FlattenDeep recursively flattens s. Because Go slices are not arbitrarily
// nestable in a single static type, the input is a slice of any; nested values
// of type []any are flattened to any depth while all other values are kept.
//
//	FlattenDeep([]any{1, []any{2, []any{3, 4}}}) => [1 2 3 4]
func FlattenDeep(s []any) []any {
	result := make([]any, 0, len(s))
	for _, v := range s {
		if inner, ok := v.([]any); ok {
			result = append(result, FlattenDeep(inner)...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

// Pair is a key/value pair used by FromPairs and produced by other utilities
// that model lodash's two-element array pairs.
type Pair[K comparable, V any] struct {
	// Key is the pair's key, used as the map key by FromPairs.
	Key K
	// Value is the pair's value, stored under Key by FromPairs.
	Value V
}

// FromPairs builds a map from a slice of key/value pairs. Later pairs override
// earlier ones that share a key.
//
//	FromPairs([]Pair[string, int]{{"a", 1}, {"b", 2}}) => map[a:1 b:2]
func FromPairs[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	result := make(map[K]V, len(pairs))
	for _, p := range pairs {
		result[p.Key] = p.Value
	}
	return result
}

// Zip groups the elements of the provided slices by index. The result has as
// many groups as the longest input slice; missing elements are filled with the
// zero value of T.
//
//	Zip([]int{1, 2}, []int{3, 4}) => [[1 3] [2 4]]
func Zip[T any](arrays ...[]T) [][]T {
	maxLen := 0
	for _, arr := range arrays {
		if len(arr) > maxLen {
			maxLen = len(arr)
		}
	}
	result := make([][]T, maxLen)
	for i := 0; i < maxLen; i++ {
		group := make([]T, len(arrays))
		for j, arr := range arrays {
			if i < len(arr) {
				group[j] = arr[i]
			}
		}
		result[i] = group
	}
	return result
}

// Unzip is the inverse of Zip: it regroups grouped elements by their position
// within each group. Missing elements are filled with the zero value of T.
//
//	Unzip([][]int{{1, 3}, {2, 4}}) => [[1 2] [3 4]]
func Unzip[T any](groups [][]T) [][]T {
	return Zip(groups...)
}

// ZipObject builds a map by pairing each key with the value at the same index.
// Keys without a corresponding value receive the zero value of V.
//
//	ZipObject([]string{"a", "b"}, []int{1, 2}) => map[a:1 b:2]
func ZipObject[K comparable, V any](keys []K, values []V) map[K]V {
	result := make(map[K]V, len(keys))
	for i, k := range keys {
		var v V
		if i < len(values) {
			v = values[i]
		}
		result[k] = v
	}
	return result
}

// Reverse returns a new slice with the elements of s in reverse order. Unlike
// lodash's _.reverse, the input is not mutated.
//
//	Reverse([]int{1, 2, 3}) => [3 2 1]
func Reverse[T any](s []T) []T {
	result := make([]T, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// Concat returns a new slice that is the concatenation of all provided slices.
//
//	Concat([]int{1}, []int{2, 3}, []int{4}) => [1 2 3 4]
func Concat[T any](slices ...[]T) []T {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	result := make([]T, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// Pull returns a new slice with all occurrences of the given values removed.
// Unlike lodash's _.pull, the input is not mutated.
//
//	Pull([]int{1, 2, 3, 1, 2}, 2, 3) => [1 1]
func Pull[T comparable](s []T, values ...T) []T {
	return Without(s, values...)
}

// PullAll is like Pull except it accepts the values to remove as a single
// slice. The input is not mutated.
//
//	PullAll([]int{1, 2, 3, 1, 2}, []int{2, 3}) => [1 1]
func PullAll[T comparable](s []T, values []T) []T {
	return Without(s, values...)
}

// Remove returns a new slice containing the elements of s for which pred
// returns false. Whereas lodash's _.remove mutates its argument and returns
// the removed elements, this port keeps the input intact and returns the
// retained elements.
//
//	Remove([]int{1, 2, 3, 4}, func(v int) bool { return v%2 == 0 }) => [1 3]
func Remove[T any](s []T, pred func(T) bool) []T {
	result := make([]T, 0, len(s))
	for _, v := range s {
		if !pred(v) {
			result = append(result, v)
		}
	}
	return result
}

// SortedIndex returns the lowest index at which value can be inserted into the
// sorted slice s to keep it sorted.
//
//	SortedIndex([]int{30, 50}, 40) => 1
func SortedIndex[T cmp.Ordered](s []T, value T) int {
	lo, hi := 0, len(s)
	for lo < hi {
		mid := (lo + hi) / 2
		if s[mid] < value {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// Slice returns a copy of the portion of s in the half-open range
// [start, end). Negative indexes count from the end and both bounds are
// clamped to the range [0, len(s)].
//
//	Slice([]int{1, 2, 3, 4}, 1, 3) => [2 3]
func Slice[T any](s []T, start, end int) []T {
	n := len(s)
	if start < 0 {
		start += n
		if start < 0 {
			start = 0
		}
	}
	if start > n {
		start = n
	}
	if end < 0 {
		end += n
	}
	if end > n {
		end = n
	}
	if end < start {
		end = start
	}
	return clone(s[start:end])
}

// Range returns a slice of integers from start (inclusive) up to end
// (exclusive), stepping by step. If step is zero it defaults to 1 (or -1 when
// end is less than start). An empty slice is returned when no values fit.
//
//	Range(0, 4, 1)  => [0 1 2 3]
//	Range(0, 6, 2)  => [0 2 4]
//	Range(4, 0, -1) => [4 3 2 1]
func Range(start, end, step int) []int {
	if step == 0 {
		if end < start {
			step = -1
		} else {
			step = 1
		}
	}
	result := make([]int, 0)
	if step > 0 {
		for i := start; i < end; i += step {
			result = append(result, i)
		}
	} else {
		for i := start; i > end; i += step {
			result = append(result, i)
		}
	}
	return result
}

// RangeRight is like Range except the resulting values are in descending
// (reversed) order.
//
//	RangeRight(0, 4, 1) => [3 2 1 0]
func RangeRight(start, end, step int) []int {
	return Reverse(Range(start, end, step))
}

// clone returns a fresh copy of s so callers never receive an alias into the
// input backing array.
func clone[T any](s []T) []T {
	result := make([]T, len(s))
	copy(result, s)
	return result
}

// contains reports whether value is present in s.
func contains[T comparable](s []T, value T) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}
	return false
}
