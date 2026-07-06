package array_test

import (
	"fmt"

	"github.com/malcolmston/express/lodash/array"
)

// ExampleChunk splits a slice into consecutive groups of at most the requested
// size, mirroring lodash's _.chunk. The elements are copied into fresh backing
// slices so the result never aliases the input. When the length is not an exact
// multiple of the size, the final group holds the remaining elements, which is
// why five elements chunked by two yield two full pairs and a trailing single.
// A size of zero or less would instead yield an empty result. This is handy for
// batching work or laying out grid rows.
func ExampleChunk() {
	fmt.Println(array.Chunk([]int{1, 2, 3, 4, 5}, 2))
	// Output: [[1 2] [3 4] [5]]
}

// ExampleUniq removes duplicate elements while preserving the first occurrence
// of each, the analogue of lodash's _.uniq. Comparison uses Go's == operator, so
// the element type is constrained to comparable. The relative order of the
// surviving elements matches the order in which they first appeared in the input.
// Here the second 2 and the trailing 1 are dropped because those values were
// already seen. The input slice itself is left unmodified.
func ExampleUniq() {
	fmt.Println(array.Uniq([]int{1, 2, 2, 3, 1}))
	// Output: [1 2 3]
}

// ExampleDifference returns the elements of the first slice that are absent from
// every other slice, like lodash's _.difference. Membership of the excluded
// values is collected into a set, then the first slice is scanned in order,
// keeping only the values not present in that set. Order and any duplicates from
// the first slice are preserved for the surviving elements. Here 2 and 4 are
// removed because they appear in the second slice, leaving 1 and 3. The result
// is a newly allocated slice.
func ExampleDifference() {
	fmt.Println(array.Difference([]int{1, 2, 3, 4}, []int{2, 4}))
	// Output: [1 3]
}

// ExampleZip regroups several slices by index, pairing the first elements
// together, the second elements together, and so on, matching lodash's _.zip.
// The number of groups equals the length of the longest input, and any missing
// positions are filled with the zero value of the element type. Because both
// inputs here have equal length, every group is fully populated. The result is a
// slice of groups suitable for iterating in parallel over the original slices.
func ExampleZip() {
	fmt.Println(array.Zip([]int{1, 2}, []int{3, 4}))
	// Output: [[1 3] [2 4]]
}
