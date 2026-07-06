package deepmerge_test

import (
	"fmt"

	"github.com/malcolmston/express/deepmerge"
)

// ExampleMerge deeply combines two maps into a brand new map. Keys present in
// only one input are carried over, and keys present in both are reconciled at
// the leaves: the nested map under "nested" is merged key by key rather than
// replaced wholesale, so both "x" from the target and "y" from the source
// survive. Neither input is mutated. This is the behavior you want when layering
// default settings under user overrides.
func ExampleMerge() {
	target := map[string]any{"a": 1, "nested": map[string]any{"x": 1}}
	source := map[string]any{"b": 2, "nested": map[string]any{"y": 2}}

	merged := deepmerge.Merge(target, source)
	fmt.Println(merged)
	// Output: map[a:1 b:2 nested:map[x:1 y:2]]
}

// ExampleMergeAll folds any number of maps together left to right, with later
// maps taking precedence over earlier ones. It is equivalent to reducing the
// maps with Merge, so nested objects still merge key by key rather than
// clobbering siblings. Here three fragments assemble into a single map. Calling
// it with no arguments returns an empty, non-nil map.
func ExampleMergeAll() {
	merged := deepmerge.MergeAll(
		map[string]any{"a": 1},
		map[string]any{"b": 2},
		map[string]any{"a": 3},
	)
	fmt.Println(merged)
	// Output: map[a:3 b:2]
}
