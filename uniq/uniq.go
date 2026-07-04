// Package uniq provides faithful ports of lodash's `uniq` and `uniqBy`.
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
