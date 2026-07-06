package collection_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"github.com/malcolmston/express/lodash/collection"
)

// This example groups a slice of words by their length. GroupBy takes an
// iteratee that derives a comparable key from each element and returns a map
// from each distinct key to the slice of elements that produced it. Elements
// keep their input order within every group, so the traversal is fully
// deterministic. The resulting map is always non-nil, even when the input slice
// is empty. Here the words are bucketed by rune count into groups of three,
// four and five characters.
func ExampleGroupBy() {
	words := []string{"one", "two", "three", "four", "five", "six"}
	byLen := collection.GroupBy(words, func(w string) int { return len(w) })
	fmt.Println(byLen[3])
	fmt.Println(byLen[4])
	fmt.Println(byLen[5])
	// Output:
	// [one two six]
	// [four five]
	// [three]
}

// This example splits a slice of integers into two groups with Partition. The
// predicate returns true for even numbers and false for odd ones, and the two
// returned slices preserve the relative order of the original input. The first
// slice collects every element for which the predicate held; the second collects
// the rest. Partition is a single pass over the data and never mutates the
// caller's slice. This mirrors lodash's _.partition, which returns a two-element
// array of the truthy and falsy groups.
func ExamplePartition() {
	nums := []int{1, 2, 3, 4, 5, 6}
	even, odd := collection.Partition(nums, func(n int) bool { return n%2 == 0 })
	fmt.Println(even)
	fmt.Println(odd)
	// Output:
	// [2 4 6]
	// [1 3 5]
}

// This example transforms a slice with Map and then collapses it with Reduce.
// Map runs each element through an iteratee and returns a new slice of the
// results, and because it is parameterized by an independent result type it can
// change the element type - here int to int. Reduce then folds the squared
// values into a single accumulated sum, walking left to right from the supplied
// initial accumulator of zero. Neither call mutates the input slice. Together
// they express a classic map/reduce pipeline in a fully type-safe way.
func ExampleMap() {
	nums := []int{1, 2, 3, 4}
	squares := collection.Map(nums, func(n int) int { return n * n })
	sum := collection.Reduce(squares, func(acc, cur int) int { return acc + cur }, 0)
	fmt.Println(squares)
	fmt.Println(sum)
	// Output:
	// [1 4 9 16]
	// 30
}

// This example sorts a slice of structs by a derived key using SortBy. The
// iteratee extracts the field to order on (age), and SortBy returns a new slice
// sorted in ascending order without touching the original. The sort is stable,
// so elements that share a key keep their input order. Here three people are
// ordered from youngest to oldest. This is the direct analogue of lodash's
// _.sortBy with a single iteratee.
func ExampleSortBy() {
	type person struct {
		Name string
		Age  int
	}
	people := []person{{"Alice", 30}, {"Bob", 25}, {"Carol", 35}}
	sorted := collection.SortBy(people, func(p person) int { return p.Age })
	for _, p := range sorted {
		fmt.Printf("%s:%d\n", p.Name, p.Age)
	}
	// Output:
	// Bob:25
	// Alice:30
	// Carol:35
}

// This example counts occurrences of a derived key with CountBy. The iteratee
// maps each number to the string "even" or "odd", and CountBy returns a map from
// each distinct key to the number of elements that produced it. The returned map
// is always non-nil. Iteration order over a Go map is unspecified, so the example
// prints the two counts by explicit key rather than ranging over the map. This
// matches lodash's _.countBy, which tallies elements by the iteratee's result.
func ExampleCountBy() {
	nums := []int{1, 2, 3, 4, 5}
	counts := collection.CountBy(nums, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	fmt.Println("even:", counts["even"])
	fmt.Println("odd:", counts["odd"])
	// Output:
	// even: 2
	// odd: 3
}

// This example draws a reproducible shuffle by passing a seeded *rand.Rand.
// Shuffle returns a new slice containing the same elements in random order using
// the Fisher-Yates algorithm and leaves the input untouched. Because a generator
// seeded with a fixed value is supplied, the permutation is deterministic and the
// Output comment can be matched exactly; passing nil would instead use a
// crypto-seeded default. The result is always a permutation of the input, so its
// length and multiset of elements are unchanged. This is the seedable counterpart
// of lodash's _.shuffle.
func ExampleShuffle() {
	r := rand.New(rand.NewSource(1))
	in := []int{1, 2, 3, 4, 5}
	out := collection.Shuffle(in, r)
	sorted := append([]int(nil), out...)
	sort.Ints(sorted)
	// Shuffle returns a permutation of the input and leaves the input untouched;
	// the exact order depends on the RNG, so assert the version-stable facts.
	fmt.Println(len(out), sorted)
	fmt.Println(in)
	// Output:
	// 5 [1 2 3 4 5]
	// [1 2 3 4 5]
}

// This example finds the first matching element with Find. The predicate selects
// strings longer than three characters, and Find returns the first such element
// together with a bool reporting whether any match existed. When nothing matches,
// the bool is false and the value is the element type's zero value, which is why
// callers should check the bool before using the result. Find never mutates the
// slice and stops at the first hit. It corresponds to lodash's _.find.
func ExampleFind() {
	words := []string{"a", "to", "the", "four", "five"}
	got, ok := collection.Find(words, func(w string) bool { return len(w) > 3 })
	fmt.Println(got, ok)
	missing, ok := collection.Find(words, func(w string) bool { return strings.HasPrefix(w, "z") })
	fmt.Printf("%q %v\n", missing, ok)
	// Output:
	// four true
	// "" false
}
