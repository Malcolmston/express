package uniq_test

import (
	"fmt"

	"github.com/malcolmston/express/uniq"
)

// ExampleUniq removes duplicate values from a slice while preserving the order
// in which each distinct value first appeared. The input contains repeated 1s
// and 2s, and the result keeps only their first occurrences in original order.
// Uniq is generic over any comparable type, so it works on ints, strings, and
// other map-key types without an equality function. The input slice is never
// modified; a new slice is returned. This mirrors lodash's _.uniq.
func ExampleUniq() {
	fmt.Println(uniq.Uniq([]int{1, 2, 2, 3, 1, 3}))
	// Output: [1 2 3]
}

// ExampleUniqBy deduplicates by a derived key rather than by whole-value
// equality. The key function returns each string's first byte, so words sharing
// an initial letter are treated as duplicates and only the first one seen is
// kept. Here "apple" and "avocado" both key on 'a', so "avocado" is dropped
// while "banana" survives. First-appearance order is preserved just as in Uniq.
// This is the idiomatic way to deduplicate structs by a single field.
func ExampleUniqBy() {
	words := []string{"apple", "banana", "avocado", "cherry", "blueberry"}
	fmt.Println(uniq.UniqBy(words, func(s string) byte { return s[0] }))
	// Output: [apple banana cherry]
}
