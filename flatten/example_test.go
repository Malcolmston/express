package flatten_test

import (
	"fmt"

	"github.com/malcolmston/express/flatten"
)

// ExampleFlatten merges a single level of nesting with full type safety. Given a
// [][]int it concatenates every inner slice into one []int, preserving order.
// This form uses generics rather than reflection, so the element type is checked
// at compile time and the result keeps that concrete type. Empty inner slices
// simply contribute nothing. A nil or empty outer slice would yield an empty,
// non-nil slice.
func ExampleFlatten() {
	got := flatten.Flatten([][]int{{1, 2}, {3, 4}})
	fmt.Println(got)
	// Output: [1 2 3 4]
}

// ExampleFlattenDeep collapses an arbitrarily nested value all the way down.
// Because it accepts any and uses reflection, it can descend through mixed
// nesting such as []any wrapping deeper []any values. Every element whose
// dynamic kind is a slice or array is flattened; every other element is kept as
// a leaf. Strings are treated as scalar leaves and are never split into their
// characters. The heterogeneous nesting here therefore collapses to a single
// flat sequence of leaves.
func ExampleFlattenDeep() {
	got := flatten.FlattenDeep([]any{1, []any{2, []any{3, []any{4}}}, 5})
	fmt.Println(got)
	// Output: [1 2 3 4 5]
}

// ExampleFlattenDepth removes a controlled number of nesting levels. With the
// same deeply nested input, a depth of 1 removes only the outermost layer, so
// the inner []any values survive, whereas a depth of 2 removes one more level.
// A depth of 0 or any negative value copies the top-level elements without
// descending at all. A depth larger than the actual nesting flattens
// everything, matching FlattenDeep. The result is a []any because reflection is
// used to walk the heterogeneous structure.
func ExampleFlattenDepth() {
	nested := []any{1, []any{2, []any{3, []any{4}}}, 5}
	fmt.Println(flatten.FlattenDepth(nested, 1))
	fmt.Println(flatten.FlattenDepth(nested, 2))
	// Output:
	// [1 2 [3 [4]] 5]
	// [1 2 3 [4] 5]
}
