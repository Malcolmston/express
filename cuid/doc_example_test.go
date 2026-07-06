package cuid_test

import (
	"fmt"

	"github.com/malcolmston/express/cuid"
)

// ExampleNew generates a collision-resistant unique identifier of the default
// length. Because a cuid mixes a timestamp, a counter, and random entropy, its
// exact value differs on every call and cannot be shown literally. This example
// instead prints the length and the result of a syntactic validity check, both
// of which are deterministic. New always returns a 24-character id that starts
// with a letter and is URL-safe.
func ExampleNew() {
	id, err := cuid.New()
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(len(id), cuid.IsCuid(id))
	// Output: 24 true
}

// ExampleNewLength generates a cuid of a caller-chosen length. The requested
// length is clamped to the range 2..32 rather than rejected, so out-of-range
// values are never returned. Here a length of 10 produces a ten-character id.
// As with New, the value is random, so the example checks the length and
// validity instead of the literal string.
func ExampleNewLength() {
	id, err := cuid.NewLength(10)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(len(id), cuid.IsCuid(id))
	// Output: 10 true
}

// ExampleIsCuid performs a purely syntactic check on a candidate string. It
// accepts strings of length 2..32 that start with a lowercase letter and
// otherwise contain only lowercase letters and digits. It cannot verify that a
// string was actually produced by this package. Here a well-formed id passes
// while a value starting with a digit is rejected.
func ExampleIsCuid() {
	fmt.Println(cuid.IsCuid("abc123"))
	fmt.Println(cuid.IsCuid("1abc"))
	// Output:
	// true
	// false
}
