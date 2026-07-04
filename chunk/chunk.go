// Package chunk provides a faithful port of lodash's `chunk` utility.
package chunk

// Chunk splits s into groups of at most size elements. The final group holds
// the remaining elements when len(s) is not evenly divisible by size.
//
// If size is less than or equal to zero, Chunk returns an empty slice.
// The returned chunks reference freshly allocated slices; the input is not
// mutated.
func Chunk[T any](s []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}
	n := (len(s) + size - 1) / size
	result := make([][]T, 0, n)
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		group := make([]T, end-i)
		copy(group, s[i:end])
		result = append(result, group)
	}
	return result
}
