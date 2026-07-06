// Package chunk provides a faithful port of lodash's `chunk` utility using
// only the Go standard library. In JavaScript, `_.chunk(array, size)` breaks
// an array into a list of smaller arrays each holding at most `size` elements;
// this package exposes the same behavior through the generic Chunk function so
// it works for slices of any element type.
//
// You reach for Chunk whenever a flat slice needs to be processed or displayed
// in fixed-size batches: paginating a result set, rendering a grid with a set
// number of columns per row, sending records to an API that caps how many can
// be submitted per request, or bounding how much work a single goroutine picks
// up. It keeps the batching arithmetic out of the caller so the surrounding
// code can focus on what each group does.
//
// The algorithm mirrors lodash exactly. It first computes the number of groups
// as the ceiling of len(s) divided by size, preallocates the outer slice to
// that capacity, then walks the input in strides of size. Each stride is copied
// into a freshly allocated group slice, so the returned groups never alias the
// backing array of the input; mutating a returned group leaves the caller's
// original slice untouched. Every group holds exactly size elements except
// possibly the last, which holds the remainder when len(s) is not evenly
// divisible by size.
//
// The edge cases follow lodash's semantics. A size of zero or any negative
// value is treated as invalid and yields an empty (non-nil) [][]T rather than
// an error or a panic. An empty or nil input slice likewise yields an empty
// outer slice. When size is greater than or equal to len(s) the result is a
// single group containing a copy of every element. The function never returns
// nil and never mutates its input.
//
// Compared to the Node original, the behavioral parity is intentionally close:
// invalid sizes collapse to an empty result and uneven splits leave a shorter
// final group, exactly as in lodash. The one deliberate difference is that Go's
// static typing replaces lodash's dynamic arrays with a type parameter, so
// Chunk is compile-time type safe and requires no per-element boxing or runtime
// type assertions.
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
