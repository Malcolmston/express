package math_test

import (
	"fmt"

	"github.com/malcolmston/express/lodash/math"
)

// ExampleSum totals the values of an integer slice. Because Sum is generic over
// the Number constraint, it operates directly on []int here and returns an int,
// preserving the caller's element type rather than converting through float64.
// The fold starts from the zero value and adds each element in order. An empty
// slice would return 0, the zero value of the accumulator. The computed total is
// printed on a single line.
func ExampleSum() {
	total := math.Sum([]int{4, 2, 8, 6})
	fmt.Println(total)
	// Output:
	// 20
}

// ExampleMean computes the arithmetic mean of a slice of numbers. The mean is
// always returned as a float64 so that averages of integer inputs are not
// silently truncated. Here the four values sum to twenty and divide to exactly
// five. An empty slice would instead return NaN, matching lodash's use of NaN
// for an undefined average. The single float result is printed below.
func ExampleMean() {
	avg := math.Mean([]int{4, 2, 8, 6})
	fmt.Println(avg)
	// Output:
	// 5
}

// ExampleMaxBy finds the element of a collection that maximises a derived value.
// Each person is projected onto their age by the iteratee, and MaxBy returns the
// whole element whose projection is greatest, not merely the maximum age. This
// mirrors lodash's _.maxBy, which is handy for "pick the record with the largest
// field" queries. An empty collection would return the zero value of the element
// type. The example prints the winning person's name.
func ExampleMaxBy() {
	type person struct {
		name string
		age  int
	}
	people := []person{
		{"Ada", 36},
		{"Grace", 45},
		{"Alan", 41},
	}
	oldest := math.MaxBy(people, func(p person) int { return p.age })
	fmt.Println(oldest.name)
	// Output:
	// Grace
}

// ExampleClamp restricts a number to an inclusive range. A value below the lower
// bound is raised to the lower bound, a value above the upper bound is lowered to
// the upper bound, and a value already inside the range passes through unchanged.
// The bounds may be given in either order, so Clamp swaps them internally when
// needed. This matches lodash's _.clamp. Each clamped result is printed on its
// own line.
func ExampleClamp() {
	fmt.Println(math.Clamp(-10, -5, 5))
	fmt.Println(math.Clamp(10, -5, 5))
	fmt.Println(math.Clamp(3, -5, 5))
	// Output:
	// -5
	// 5
	// 3
}

// ExampleRound demonstrates rounding to a chosen number of decimal places. A
// positive precision rounds to the right of the decimal point, while a negative
// precision rounds to the left, affecting tens, hundreds and so on. Halves round
// up toward positive infinity to match JavaScript's Math.round and therefore
// lodash's _.round. NaN and the infinities would pass through unchanged. The two
// rounded values are printed below.
func ExampleRound() {
	fmt.Println(math.Round(4.006, 2))
	fmt.Println(math.Round(4060, -2))
	// Output:
	// 4.01
	// 4100
}

// ExampleInRange reports whether a number falls within a half-open range, that is
// start <= number < end. The lower bound is inclusive and the upper bound is
// exclusive, so a number equal to the start is in range but a number equal to the
// end is not. When the bounds are supplied high-to-low they are swapped first,
// matching lodash's _.inRange. Each boolean result is printed on its own line.
func ExampleInRange() {
	fmt.Println(math.InRange(3, 2, 4))
	fmt.Println(math.InRange(4, 8, 2))
	fmt.Println(math.InRange(4, 2, 4))
	// Output:
	// true
	// true
	// false
}

// ExampleRandom shows deterministic random generation via an injected source.
// Random returns a value in the half-open interval [min, max), computed as
// min + r*(max-min) where r is drawn from the supplied function. By passing a
// function that always returns 0.5, the midpoint of the range is produced, which
// keeps the example reproducible. Real callers would inject a real random source
// instead. The midpoint result is printed below.
func ExampleRandom() {
	mid := math.Random(0, 5, func() float64 { return 0.5 })
	fmt.Println(mid)
	// Output:
	// 2.5
}
