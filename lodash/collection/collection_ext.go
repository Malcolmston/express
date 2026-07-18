package collection

// This file extends the collection package with index-aware variants of the
// core iteration helpers (MapWithIndex, FilterWithIndex, ReduceWithIndex,
// FlatMapWithIndex, ForEachRight), a counting helper (CountWhere) and the
// lodash sequence utilities Tap and Thru. lodash's JavaScript iteratees receive
// the element index as a second argument; because the base helpers in this
// package take single-argument iteratees for brevity, these variants expose the
// index explicitly for the cases that need it. All functions are deterministic
// and depend only on the standard library.

// MapWithIndex returns a new slice holding iteratee(element, index) for each
// element of s, the index-aware form of Map.
func MapWithIndex[T, R any](s []T, iteratee func(value T, index int) R) []R {
	out := make([]R, len(s))
	for i, v := range s {
		out[i] = iteratee(v, i)
	}
	return out
}

// FilterWithIndex returns the elements of s for which predicate(element, index)
// is true, the index-aware form of Filter.
func FilterWithIndex[T any](s []T, predicate func(value T, index int) bool) []T {
	out := make([]T, 0, len(s))
	for i, v := range s {
		if predicate(v, i) {
			out = append(out, v)
		}
	}
	return out
}

// ReduceWithIndex folds s left-to-right into an accumulator, passing each
// element's index to the iteratee.
func ReduceWithIndex[T, R any](s []T, iteratee func(acc R, cur T, index int) R, accumulator R) R {
	for i, v := range s {
		accumulator = iteratee(accumulator, v, i)
	}
	return accumulator
}

// FlatMapWithIndex maps each element through iteratee(element, index), which
// returns a slice, and concatenates the results, the index-aware form of
// FlatMap.
func FlatMapWithIndex[T, R any](s []T, iteratee func(value T, index int) []R) []R {
	out := make([]R, 0, len(s))
	for i, v := range s {
		out = append(out, iteratee(v, i)...)
	}
	return out
}

// ForEachRight invokes iteratee for each element of s from right to left,
// stopping early if iteratee returns false. It mirrors lodash.forEachRight.
func ForEachRight[T any](s []T, iteratee func(T) bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if !iteratee(s[i]) {
			return
		}
	}
}

// CountWhere returns the number of elements of s that satisfy predicate.
func CountWhere[T any](s []T, predicate func(T) bool) int {
	n := 0
	for _, v := range s {
		if predicate(v) {
			n++
		}
	}
	return n
}

// Tap invokes interceptor with value for its side effects and returns value
// unchanged, allowing a value to be inspected or logged inside a pipeline. It
// mirrors lodash.tap.
func Tap[T any](value T, interceptor func(T)) T {
	interceptor(value)
	return value
}

// Thru passes value through interceptor and returns the interceptor's result,
// letting a pipeline transform a value inline. It mirrors lodash.thru.
func Thru[T, R any](value T, interceptor func(T) R) R {
	return interceptor(value)
}
