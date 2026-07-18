package array

import "cmp"

// This file extends the array package with further lodash "Array" helpers that
// complement the base set: comparator-based set operations (DifferenceWith,
// IntersectionWith, UnionWith, XorWith), key-based Xor (XorBy), the zip-with and
// unzip-with reshapers, index removal (PullAt), the sorted-slice search and
// dedup family (SortedIndexBy, SortedIndexOf, SortedLastIndex,
// SortedLastIndexOf, SortedUniq, SortedUniqBy) and the right-hand while slicers
// (TakeRightWhile, DropRightWhile). Like the rest of the package, inputs are
// never mutated and comparisons are expressed with explicit equality, key or
// ordering functions rather than lodash's implicit iteratees.

// DifferenceWith returns the elements of s for which no element of values is
// equal according to eq, preserving the order and first-occurrence of s.
func DifferenceWith[T any](s []T, values []T, eq func(a, b T) bool) []T {
	out := make([]T, 0, len(s))
	for _, x := range s {
		found := false
		for _, v := range values {
			if eq(x, v) {
				found = true
				break
			}
		}
		if !found {
			out = append(out, x)
		}
	}
	return out
}

// IntersectionWith returns the elements present in both a and b according to eq,
// in the order they appear in a and without duplicates.
func IntersectionWith[T any](a, b []T, eq func(x, y T) bool) []T {
	out := make([]T, 0)
	for _, x := range a {
		inB := false
		for _, y := range b {
			if eq(x, y) {
				inB = true
				break
			}
		}
		if !inB {
			continue
		}
		dup := false
		for _, o := range out {
			if eq(x, o) {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, x)
		}
	}
	return out
}

// UnionWith returns the unique elements of all slices, treating two elements as
// equal when eq reports so, in first-occurrence order.
func UnionWith[T any](eq func(x, y T) bool, slices ...[]T) []T {
	out := make([]T, 0)
	for _, s := range slices {
		for _, x := range s {
			dup := false
			for _, o := range out {
				if eq(x, o) {
					dup = true
					break
				}
			}
			if !dup {
				out = append(out, x)
			}
		}
	}
	return out
}

// XorBy returns the symmetric difference of the slices, comparing elements by
// the key returned by iteratee: values whose key appears in exactly one slice
// are kept, in first-occurrence order.
func XorBy[T any, K comparable](iteratee func(T) K, slices ...[]T) []T {
	counts := make(map[K]int)
	seen := make(map[K]bool)
	var order []T
	for _, s := range slices {
		local := make(map[K]bool)
		for _, x := range s {
			k := iteratee(x)
			if !seen[k] {
				seen[k] = true
				order = append(order, x)
			}
			if !local[k] {
				local[k] = true
				counts[k]++
			}
		}
	}
	out := make([]T, 0)
	for _, x := range order {
		if counts[iteratee(x)] == 1 {
			out = append(out, x)
		}
	}
	return out
}

// XorWith returns the symmetric difference of the slices using eq for equality.
// Values equal to an element of exactly one slice are kept, in first-occurrence
// order.
func XorWith[T any](eq func(x, y T) bool, slices ...[]T) []T {
	type entry struct {
		val   T
		count int
	}
	var entries []*entry
	find := func(v T) *entry {
		for _, e := range entries {
			if eq(e.val, v) {
				return e
			}
		}
		return nil
	}
	for _, s := range slices {
		var localSeen []*entry
		for _, x := range s {
			e := find(x)
			if e == nil {
				e = &entry{val: x}
				entries = append(entries, e)
			}
			already := false
			for _, le := range localSeen {
				if le == e {
					already = true
					break
				}
			}
			if !already {
				e.count++
				localSeen = append(localSeen, e)
			}
		}
	}
	out := make([]T, 0)
	for _, e := range entries {
		if e.count == 1 {
			out = append(out, e.val)
		}
	}
	return out
}

// ZipWith groups the i-th elements of each slice and applies fn to each group,
// returning the slice of results. The number of results equals the length of
// the longest input slice; missing elements are the zero value of T.
func ZipWith[T, R any](fn func(group []T) R, slices ...[]T) []R {
	maxLen := 0
	for _, s := range slices {
		if len(s) > maxLen {
			maxLen = len(s)
		}
	}
	out := make([]R, maxLen)
	for i := 0; i < maxLen; i++ {
		group := make([]T, len(slices))
		for j, s := range slices {
			if i < len(s) {
				group[j] = s[i]
			}
		}
		out[i] = fn(group)
	}
	return out
}

// UnzipWith is the inverse of ZipWith: it applies fn to each column of grouped
// and returns the slice of results.
func UnzipWith[T, R any](fn func(group []T) R, grouped [][]T) []R {
	maxLen := 0
	for _, g := range grouped {
		if len(g) > maxLen {
			maxLen = len(g)
		}
	}
	out := make([]R, maxLen)
	for i := 0; i < maxLen; i++ {
		col := make([]T, len(grouped))
		for j, g := range grouped {
			if i < len(g) {
				col[j] = g[i]
			}
		}
		out[i] = fn(col)
	}
	return out
}

// PullAt returns a copy of slice with the elements at the given indexes removed.
// Out-of-range and duplicate indexes are ignored. Unlike lodash's mutating
// pullAt, the input is left unchanged.
func PullAt[T any](slice []T, indexes ...int) []T {
	remove := make(map[int]bool, len(indexes))
	for _, i := range indexes {
		if i >= 0 && i < len(slice) {
			remove[i] = true
		}
	}
	out := make([]T, 0, len(slice))
	for i, v := range slice {
		if !remove[i] {
			out = append(out, v)
		}
	}
	return out
}

// SortedIndexOf returns the index of the first occurrence of value in the sorted
// slice s using binary search, or -1 if value is not present.
func SortedIndexOf[T cmp.Ordered](s []T, value T) int {
	i := SortedIndex(s, value)
	if i < len(s) && s[i] == value {
		return i
	}
	return -1
}

// SortedIndexBy returns the lowest index at which value should be inserted into
// s (assumed sorted by the key returned by iteratee) to keep it sorted.
func SortedIndexBy[T any, K cmp.Ordered](s []T, value T, iteratee func(T) K) int {
	target := iteratee(value)
	lo, hi := 0, len(s)
	for lo < hi {
		mid := (lo + hi) / 2
		if iteratee(s[mid]) < target {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// SortedLastIndex returns the highest index at which value should be inserted
// into the sorted slice s to keep it sorted (after any equal elements).
func SortedLastIndex[T cmp.Ordered](s []T, value T) int {
	lo, hi := 0, len(s)
	for lo < hi {
		mid := (lo + hi) / 2
		if s[mid] <= value {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// SortedLastIndexOf returns the index of the last occurrence of value in the
// sorted slice s using binary search, or -1 if value is not present.
func SortedLastIndexOf[T cmp.Ordered](s []T, value T) int {
	i := SortedLastIndex(s, value) - 1
	if i >= 0 && s[i] == value {
		return i
	}
	return -1
}

// SortedUniq returns a new slice with consecutive duplicate elements removed. It
// is optimised for already-sorted input, where it removes all duplicates.
func SortedUniq[T comparable](s []T) []T {
	out := make([]T, 0, len(s))
	for i, v := range s {
		if i == 0 || v != s[i-1] {
			out = append(out, v)
		}
	}
	return out
}

// SortedUniqBy returns a new slice with consecutive elements sharing a key
// (as returned by iteratee) collapsed to their first occurrence.
func SortedUniqBy[T any, K comparable](s []T, iteratee func(T) K) []T {
	out := make([]T, 0, len(s))
	var prev K
	for i, v := range s {
		k := iteratee(v)
		if i == 0 || k != prev {
			out = append(out, v)
		}
		prev = k
	}
	return out
}

// TakeRightWhile returns the trailing elements of s for which pred is true,
// stopping at the last element from the end for which pred is false.
func TakeRightWhile[T any](s []T, pred func(T) bool) []T {
	i := len(s)
	for i > 0 && pred(s[i-1]) {
		i--
	}
	out := make([]T, len(s)-i)
	copy(out, s[i:])
	return out
}

// DropRightWhile returns s without the trailing elements for which pred is true.
func DropRightWhile[T any](s []T, pred func(T) bool) []T {
	i := len(s)
	for i > 0 && pred(s[i-1]) {
		i--
	}
	out := make([]T, i)
	copy(out, s[:i])
	return out
}
