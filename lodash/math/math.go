// Package math provides generic ports of the math and number utility functions
// found in the JavaScript library lodash. Every function is written against the
// Number constraint and depends only on the standard library.
package math

import "math"

// Number is the set of built-in numeric types (and named types derived from
// them) that the helpers in this package operate on.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sum computes the sum of the values in nums.
//
//	Sum([]int{4, 2, 8, 6}) => 20
func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// SumBy computes the sum of applying iteratee to each element of coll.
//
//	SumBy([]obj, func(o obj) int { return o.N }) => total of every o.N
func SumBy[T any, N Number](coll []T, iteratee func(T) N) N {
	var total N
	for _, v := range coll {
		total += iteratee(v)
	}
	return total
}

// Mean computes the arithmetic mean of the values in nums. An empty slice
// yields NaN, matching lodash's use of NaN for an empty mean.
//
//	Mean([]int{4, 2, 8, 6}) => 5
func Mean[T Number](nums []T) float64 {
	if len(nums) == 0 {
		return math.NaN()
	}
	return float64(Sum(nums)) / float64(len(nums))
}

// MeanBy computes the arithmetic mean of applying iteratee to each element of
// coll. An empty collection yields NaN.
func MeanBy[T any, N Number](coll []T, iteratee func(T) N) float64 {
	if len(coll) == 0 {
		return math.NaN()
	}
	return float64(SumBy(coll, iteratee)) / float64(len(coll))
}

// Max returns the largest value in nums. An empty slice returns the zero value
// of T (lodash returns undefined in this case).
//
//	Max([]int{4, 2, 8, 6}) => 8
func Max[T Number](nums []T) T {
	var out T
	for i, n := range nums {
		if i == 0 || n > out {
			out = n
		}
	}
	return out
}

// Min returns the smallest value in nums. An empty slice returns the zero value
// of T (lodash returns undefined in this case).
//
//	Min([]int{4, 2, 8, 6}) => 2
func Min[T Number](nums []T) T {
	var out T
	for i, n := range nums {
		if i == 0 || n < out {
			out = n
		}
	}
	return out
}

// MaxBy returns the element of coll that yields the largest value when passed
// through iteratee. An empty collection returns the zero value of T.
//
//	MaxBy(objs, func(o obj) int { return o.N }) => object with greatest N
func MaxBy[T any, N Number](coll []T, iteratee func(T) N) T {
	var out T
	var best N
	for i, v := range coll {
		c := iteratee(v)
		if i == 0 || c > best {
			best = c
			out = v
		}
	}
	return out
}

// MinBy returns the element of coll that yields the smallest value when passed
// through iteratee. An empty collection returns the zero value of T.
//
//	MinBy(objs, func(o obj) int { return o.N }) => object with least N
func MinBy[T any, N Number](coll []T, iteratee func(T) N) T {
	var out T
	var best N
	for i, v := range coll {
		c := iteratee(v)
		if i == 0 || c < best {
			best = c
			out = v
		}
	}
	return out
}

// Clamp restricts number to be within the inclusive lower and upper bounds. The
// bounds may be supplied in either order.
//
//	Clamp(-10, -5, 5) => -5
//	Clamp(10, -5, 5)  => 5
func Clamp[T Number](number, lower, upper T) T {
	if lower > upper {
		lower, upper = upper, lower
	}
	if number < lower {
		return lower
	}
	if number > upper {
		return upper
	}
	return number
}

// InRange reports whether number is in the half-open range bounded by start and
// end, that is start <= number < end. The bounds may be supplied in either
// order, matching lodash which swaps them when start is greater than end.
//
//	InRange(3, 2, 4) => true
//	InRange(4, 8, 2) => true
//	InRange(2, 2, 4) => true
func InRange[T Number](number, start, end T) bool {
	lo, hi := start, end
	if lo > hi {
		lo, hi = hi, lo
	}
	return number >= lo && number < hi
}

// roundTo applies op (one of math.Floor, math.Ceil or half-up rounding) to n
// after scaling by 10**precision, then rescales. It mirrors lodash's precise
// createRound helper closely enough for typical precisions.
func roundTo(n float64, precision int, op func(float64) float64) float64 {
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return n
	}
	factor := math.Pow(10, float64(precision))
	return op(n*factor) / factor
}

// halfUp rounds x to the nearest integer, rounding halves toward positive
// infinity to match JavaScript's Math.round (and therefore lodash's _.round).
func halfUp(x float64) float64 {
	return math.Floor(x + 0.5)
}

// Round rounds number to precision decimal places. A negative precision rounds
// to the left of the decimal point.
//
//	Round(4.006, 2) => 4.01
//	Round(4060, -2) => 4100
func Round(number float64, precision int) float64 {
	return roundTo(number, precision, halfUp)
}

// Ceil rounds number up to precision decimal places.
//
//	Ceil(6.004, 2) => 6.01
//	Ceil(6040, -2) => 6100
func Ceil(number float64, precision int) float64 {
	return roundTo(number, precision, math.Ceil)
}

// Floor rounds number down to precision decimal places.
//
//	Floor(0.046, 2) => 0.04
//	Floor(4060, -2) => 4000
func Floor(number float64, precision int) float64 {
	return roundTo(number, precision, math.Floor)
}

// Add returns augend + addend.
//
//	Add(6, 4) => 10
func Add[T Number](augend, addend T) T {
	return augend + addend
}

// Subtract returns minuend - subtrahend.
//
//	Subtract(6, 4) => 2
func Subtract[T Number](minuend, subtrahend T) T {
	return minuend - subtrahend
}

// Multiply returns multiplier * multiplicand.
//
//	Multiply(6, 4) => 24
func Multiply[T Number](multiplier, multiplicand T) T {
	return multiplier * multiplicand
}

// Divide returns dividend / divisor. As with any Go division, dividing an
// integer type by zero panics, while a float type yields +Inf, -Inf or NaN.
//
//	Divide(6, 4) => 1.5 (for floats)
func Divide[T Number](dividend, divisor T) T {
	return dividend / divisor
}

// Random returns a floating-point number between min and max (inclusive of min,
// exclusive of max). rnd must return a value in the half-open interval [0, 1);
// injecting it keeps the result deterministic and testable. The bounds may be
// supplied in either order.
//
//	Random(0, 5, func() float64 { return 0.5 }) => 2.5
func Random(min, max float64, rnd func() float64) float64 {
	if min > max {
		min, max = max, min
	}
	return min + rnd()*(max-min)
}

// RandomInt returns an integer between min and max, inclusive of both bounds.
// rnd must return a value in the half-open interval [0, 1). The bounds may be
// supplied in either order.
//
//	RandomInt(0, 5, func() float64 { return 0.5 }) => 2
func RandomInt(min, max int, rnd func() float64) int {
	if min > max {
		min, max = max, min
	}
	span := max - min + 1
	offset := int(math.Floor(rnd() * float64(span)))
	if offset >= span {
		offset = span - 1
	}
	return min + offset
}
