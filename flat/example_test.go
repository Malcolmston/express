package flat_test

import (
	"fmt"

	"github.com/malcolmston/express/flat"
)

// ExampleFlatten collapses a nested map into a single-level map whose keys are
// joined by the default "." delimiter. The function descends into every value
// that is itself a map[string]any, concatenating each level's key onto the
// accumulated prefix. Scalar values such as the integers here are stored
// unchanged as leaves. The output is printed with fmt, which sorts map keys, so
// the ordering is deterministic. Note how the two-level path a -> b becomes the
// composite key "a.b".
func ExampleFlatten() {
	in := map[string]any{
		"a": map[string]any{"b": 1, "c": 2},
		"d": 3,
	}
	fmt.Println(flat.Flatten(in))
	// Output: map[a.b:1 a.c:2 d:3]
}

// ExampleFlatten_customDelimiter overrides the default separator through
// FlattenOpts. Passing a Delimiter of "/" makes the flattened keys use slashes
// instead of dots, which is convenient when the target key space already uses a
// path-like convention. Every nested level is joined with the chosen delimiter.
// The same delimiter must be supplied to Unflatten to reverse the operation.
// Here the nested a -> b path becomes "a/b".
func ExampleFlatten_customDelimiter() {
	in := map[string]any{"a": map[string]any{"b": 1}}
	fmt.Println(flat.Flatten(in, flat.FlattenOpts{Delimiter: "/"}))
	// Output: map[a/b:1]
}

// ExampleUnflatten reverses Flatten, expanding delimited keys back into a nested
// map structure. Each composite key is split on the delimiter and the nested
// maps are rebuilt segment by segment. Keys that share a prefix are merged into
// the same parent map, so "a.b" and "a.c" both live under "a". The result is
// printed with fmt, which sorts keys recursively for a stable rendering. This
// makes Unflatten the inverse of Flatten for map-only data.
func ExampleUnflatten() {
	in := map[string]any{"a.b": 1, "a.c": 2, "d": 3}
	fmt.Println(flat.Unflatten(in))
	// Output: map[a:map[b:1 c:2] d:3]
}
