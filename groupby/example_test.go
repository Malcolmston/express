package groupby_test

import (
	"fmt"

	"github.com/malcolmston/express/groupby"
)

// ExampleGroupBy partitions a slice of integers into buckets by parity. The key
// function returns "even" or "odd" for each element, and GroupBy collects the
// elements that produced each key into a slice. Within every group the elements
// keep their original input order, so the odd bucket reads 1, 3, 5. Because a
// returned map has no defined iteration order, this example prints selected keys
// explicitly rather than ranging over the map. The result map is keyed by the
// string values the key function returned.
func ExampleGroupBy() {
	in := []int{1, 2, 3, 4, 5, 6}
	groups := groupby.GroupBy(in, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	fmt.Println("odd:", groups["odd"])
	fmt.Println("even:", groups["even"])
	// Output:
	// odd: [1 3 5]
	// even: [2 4 6]
}

// ExampleGroupBy_length groups strings by their length using an integer key.
// The key function may return any comparable type, and here it returns len(s),
// so the map is keyed by int. Strings of the same length are gathered together
// in input order, which is why "one", "two", and "six" share the group for
// length three. This shows that keys are kept in their native type rather than
// being stringified. Specific length keys are printed for deterministic output.
func ExampleGroupBy_length() {
	in := []string{"one", "two", "three", "four", "six"}
	groups := groupby.GroupBy(in, func(s string) int { return len(s) })
	fmt.Println("3:", groups[3])
	fmt.Println("4:", groups[4])
	fmt.Println("5:", groups[5])
	// Output:
	// 3: [one two six]
	// 4: [four]
	// 5: [three]
}

// ExampleGroupBy_empty demonstrates the behavior on empty input. Given a nil
// slice, GroupBy returns an empty but non-nil map, so callers can safely index
// or range over the result without a nil check. The length of that map is zero
// because no elements were grouped. This matches the guarantee that the returned
// map is never nil. The example prints the length to confirm it is empty.
func ExampleGroupBy_empty() {
	groups := groupby.GroupBy([]int(nil), func(n int) int { return n })
	fmt.Println(len(groups))
	// Output: 0
}
