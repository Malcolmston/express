// Package groupby provides a faithful port of lodash's `groupBy`.
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
