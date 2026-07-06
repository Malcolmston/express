// Package groupby provides a faithful port of lodash's `groupBy`. It partitions
// a slice into groups according to a key derived from each element, returning a
// map from key to the slice of elements that produced it.
//
// Use this package whenever you need to bucket a collection by some computed
// property: grouping records by category, numbers by parity, strings by length,
// or structs by a field. It is the Go analogue of the collection helper that
// lodash exposes as _.groupBy and that many JavaScript and Express codebases
// rely on for shaping data before rendering or aggregation.
//
// The implementation is a single generic function. GroupBy walks the input
// slice once, calls the supplied key function on each element, and appends the
// element to the slice stored under that key. Because iteration proceeds in
// input order and elements are appended, the order of elements within each
// group reflects their original order in the input. The overall cost is linear
// in the length of the slice.
//
// The key function may return any comparable type (its result is constrained by
// Go's comparable constraint), so keys can be strings, integers, booleans, or
// any other comparable value including structs of comparable fields. Elements
// themselves may be of any type. A nil or empty input slice yields an empty but
// non-nil map, so callers can index or range over the result without a nil
// check; the returned map is never nil.
//
// Compared to the lodash original, the semantics of grouping and of preserving
// per-group order are the same, but the Go version is type-safe by construction
// rather than coercing keys to strings. lodash converts every computed key to a
// string property name on a plain object, whereas GroupBy keeps the key's
// native type as the map key, so numeric and boolean keys remain distinct and
// are not stringified.
package groupby

// GroupBy partitions s into a map keyed by the value returned by key. Each
// map value is a slice of the elements that produced that key, in the order
// they appear in s.
//
// A nil or empty input yields an empty (non-nil) map.
func GroupBy[T any, K comparable](s []T, key func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range s {
		k := key(v)
		result[k] = append(result[k], v)
	}
	return result
}
