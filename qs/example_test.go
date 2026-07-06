package qs_test

import (
	"fmt"

	"github.com/malcolmston/express/qs"
)

// ExampleParse reconstructs a nested object from bracket notation. The key
// "a[b]" nests the value under key "b" of the object stored at "a", and "a[c]"
// adds a sibling, so the two pairs together build a single map. Parse stores
// scalar leaves as strings and nested objects as map[string]any, decoding each
// key and value along the way. The result is printed with fmt, which renders
// maps with their keys in sorted order for a stable rendering. This is the shape
// Express reconstructs when its extended query parser is enabled.
func ExampleParse() {
	fmt.Println(qs.Parse("a[b]=1&a[c]=2"))
	// Output: map[a:map[b:1 c:2]]
}

// ExampleParse_array shows how repeated empty brackets build a slice. Each
// "a[]" entry appends to the array stored at "a" in the order the pairs appear,
// so "a[]=1&a[]=2" yields a two-element slice rather than a map. Array elements
// are stored as []any and printed by fmt in index order. This mirrors the way
// HTML forms and Express clients encode a list into a query string. The bracket
// notation is what lets a flat URL describe a structured value.
func ExampleParse_array() {
	fmt.Println(qs.Parse("a[]=1&a[]=2"))
	// Output: map[a:[1 2]]
}

// ExampleStringify serializes a nested map back into bracket notation, the
// inverse of Parse. Nested maps become bracketed keys such as "a[b]", and both
// the top-level keys and every nested object's keys are emitted in sorted order
// so the output is deterministic. Keys and values are URL-encoded, though the
// bracket characters themselves are written literally. The result here round-
// trips back to the same structure through Parse. Deterministic ordering makes
// Stringify output safe to compare directly in tests.
func ExampleStringify() {
	m := map[string]any{"a": map[string]any{"b": "1", "c": "2"}}
	fmt.Println(qs.Stringify(m))
	// Output: a[b]=1&a[c]=2
}
