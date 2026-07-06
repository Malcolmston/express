package lang_test

import (
	"fmt"

	"github.com/malcolmston/express/lodash/lang"
)

// ExampleIsEmpty demonstrates how the IsEmpty predicate classifies a range of
// dynamically typed values. Nil, empty strings and empty collections are all
// reported as empty, exactly as lodash's _.isEmpty would report them. Because Go
// has no single "object" type, a non-collection scalar such as an integer is
// also treated as empty, mirroring lodash's treatment of primitives. A populated
// slice, by contrast, is not empty. The example prints one boolean per input so
// the output is fully deterministic.
func ExampleIsEmpty() {
	fmt.Println(lang.IsEmpty(nil))
	fmt.Println(lang.IsEmpty(""))
	fmt.Println(lang.IsEmpty([]int{}))
	fmt.Println(lang.IsEmpty(42))
	fmt.Println(lang.IsEmpty([]int{1, 2, 3}))
	// Output:
	// true
	// true
	// true
	// true
	// false
}

// ExampleIsNil shows that IsNil recognises both an untyped nil interface and a
// typed nil pointer. A plain nil literal is nil, and so is a *int variable that
// has never been assigned. A pointer to an actual value is not nil, and neither
// is a non-pointer value such as an integer. This is stricter and more precise
// than a bare == nil comparison, which cannot see through a typed nil stored in
// an interface. Each result is printed on its own line for determinism.
func ExampleIsNil() {
	var p *int
	value := 7
	fmt.Println(lang.IsNil(nil))
	fmt.Println(lang.IsNil(p))
	fmt.Println(lang.IsNil(&value))
	fmt.Println(lang.IsNil(value))
	// Output:
	// true
	// true
	// false
	// false
}

// ExampleToNumber demonstrates lodash-style numeric coercion. A numeric string
// is parsed to its float value, surrounding whitespace is trimmed, and a boolean
// is mapped to 0 or 1. An empty string coerces to 0, matching lodash, while a
// value that cannot be parsed at all would instead yield NaN. This is the
// permissive behaviour that Node developers expect from _.toNumber. The results
// are printed as a fixed sequence of lines for a deterministic output.
func ExampleToNumber() {
	fmt.Println(lang.ToNumber("  3.5 "))
	fmt.Println(lang.ToNumber(true))
	fmt.Println(lang.ToNumber(false))
	fmt.Println(lang.ToNumber(""))
	// Output:
	// 3.5
	// 1
	// 0
	// 0
}

// ExampleDefaultTo shows how DefaultTo supplies a fallback for missing or
// invalid values. A concrete, present value is returned unchanged. A nil value is
// replaced by the provided default. A floating-point NaN is likewise treated as
// missing and replaced, because NaN represents a failed numeric computation just
// as it does in JavaScript. This mirrors lodash's _.defaultTo, which falls back
// on null, undefined and NaN. The example prints the resolved value for each
// case.
func ExampleDefaultTo() {
	fmt.Println(lang.DefaultTo("hello", "fallback"))
	fmt.Println(lang.DefaultTo(nil, "fallback"))
	fmt.Println(lang.DefaultTo(42, 0))
	// Output:
	// hello
	// fallback
	// 42
}

// ExampleEq contrasts Eq with Go's built-in equality. Eq performs a deep,
// value-based comparison, so two slices with the same contents are equal even
// though they are distinct backing arrays. Crucially, Eq also treats two NaN
// values as equal, which the == operator never does, matching lodash's SameValueZero
// semantics. Ordinary distinct values compare as unequal. Each comparison prints
// a single boolean for a deterministic output.
func ExampleEq() {
	nan := lang.ToNumber("not a number")
	fmt.Println(lang.Eq([]int{1, 2}, []int{1, 2}))
	fmt.Println(lang.Eq(nan, nan))
	fmt.Println(lang.Eq(1, 2))
	// Output:
	// true
	// true
	// false
}

// ExampleTimes demonstrates the generic Times helper, which invokes a function n
// times and collects the results. Here it builds a slice of the first four
// squares by squaring each index. Times is generic, so the element type is
// inferred from the function's return type, in this case int. A non-positive
// count would produce an empty slice. The resulting slice is printed directly.
func ExampleTimes() {
	squares := lang.Times(4, func(i int) int { return i * i })
	fmt.Println(squares)
	// Output:
	// [0 1 4 9]
}

// ExampleUniqueId shows that UniqueId returns distinct, monotonically increasing
// identifiers, optionally prefixed with a caller-supplied string. Because the
// counter is process-global and advanced atomically, successive calls never
// collide, which makes the helper suitable for generating DOM-style ids or
// temporary keys. The exact numeric suffix depends on how many times the counter
// has advanced, so this example only asserts that two successive ids differ. It
// prints that comparison as a boolean.
func ExampleUniqueId() {
	a := lang.UniqueId("item_")
	b := lang.UniqueId("item_")
	fmt.Println(a == b)
	// Output:
	// false
}
