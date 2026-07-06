// Package uniq removes duplicate elements from a slice, preserving order. It is
// a standard-library-only Go port of lodash's _.uniq and _.uniqBy
// (https://lodash.com/docs/#uniq), two of the most heavily used array helpers in
// the JavaScript ecosystem for de-duplicating lists while keeping the first
// occurrence of each value. Deduplication is a ubiquitous need — collapsing tag
// lists, cleaning user input, building sets of ids — and this package packages
// it behind two small generic functions.
//
// Uniq handles the value-equality case. It walks the input once and returns a
// new slice containing only the first occurrence of each distinct element, in
// the order those elements first appeared. It is generic over any comparable
// type, so it works on slices of strings, ints, or any type usable as a map key
// without the caller writing an equality function. The result is always a
// freshly allocated slice; the input is never modified.
//
// UniqBy is the accessor-based variant. Instead of comparing elements directly,
// it applies a caller-supplied key function to each element and treats two
// elements as duplicates when their keys are equal. This is the idiomatic way to
// deduplicate structs by one field, or numbers by a derived bucket such as
// math.Floor, mirroring the iteratee argument of lodash's _.uniqBy. Like Uniq it
// keeps the first element seen for each distinct key and preserves first-
// appearance order.
//
// Both functions are implemented with a single pass and a set (a Go map) of
// values or keys already seen, so they run in linear time and use memory
// proportional to the number of distinct elements. Order preservation is a
// deliberate guarantee, not an accident of the map: the output slice is built in
// iteration order, and only the membership test consults the map. This makes the
// output deterministic for a given input, which the lodash originals also
// promise.
//
// A note on the empty and nil cases: both functions always return a non-nil
// slice, returning an empty (length-zero) slice rather than nil when the input
// is nil or empty, which keeps callers from having to special-case nil. The main
// parity difference from lodash is idiomatic typing — Go generics replace
// JavaScript's dynamic arrays, and UniqBy takes a typed func(T) K instead of a
// property-name string or function — while the ordering and first-wins semantics
// match.
package uniq

// Uniq returns a new slice containing the first occurrence of each distinct
// element of s, preserving the order in which those elements first appear.
//
// A nil or empty input yields an empty (non-nil) slice.
func Uniq[T comparable](s []T) []T {
	seen := make(map[T]struct{}, len(s))
	result := make([]T, 0, len(s))
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// UniqBy returns a new slice containing the first element of s for each
// distinct value produced by key. Order of first appearance is preserved.
//
// This is the accessor-based variant of Uniq: two elements are considered
// duplicates when key returns equal values for them.
func UniqBy[T any, K comparable](s []T, key func(T) K) []T {
	seen := make(map[K]struct{}, len(s))
	result := make([]T, 0, len(s))
	for _, v := range s {
		k := key(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		result = append(result, v)
	}
	return result
}
