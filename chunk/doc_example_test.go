package chunk_test

import (
	"fmt"

	"github.com/malcolmston/express/chunk"
)

// ExampleChunk splits a flat slice into groups of at most the given size,
// mirroring lodash's chunk. Every group holds exactly size elements except
// possibly the last, which holds the remainder when the length is not evenly
// divisible. Here five elements split into two pairs and a trailing single. The
// returned groups are freshly allocated copies, so mutating them never touches
// the input.
func ExampleChunk() {
	groups := chunk.Chunk([]int{1, 2, 3, 4, 5}, 2)
	fmt.Println(groups)
	// Output: [[1 2] [3 4] [5]]
}

// ExampleChunk_strings shows that Chunk is generic and works for a slice of any
// element type, here strings. It is useful for batching work such as paginating
// results or capping how many records are sent per request. A size greater than
// or equal to the length yields a single group containing every element. This
// call batches four names two at a time.
func ExampleChunk_strings() {
	groups := chunk.Chunk([]string{"a", "b", "c", "d"}, 2)
	fmt.Println(groups)
	// Output: [[a b] [c d]]
}

// ExampleChunk_invalidSize demonstrates the edge-case semantics. A size of zero
// or any negative value is treated as invalid and yields an empty, non-nil
// result rather than an error or a panic. The input slice is never mutated and
// the function never returns nil. This matches lodash, which collapses invalid
// sizes to an empty array.
func ExampleChunk_invalidSize() {
	groups := chunk.Chunk([]int{1, 2, 3}, 0)
	fmt.Println(len(groups))
	// Output: 0
}
