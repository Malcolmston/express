# math — ports the "Math" and "Number" category functions of the npm lodash library to Go

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/math.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/math)

Package math ports the "Math" and "Number" category functions of the npm
lodash library to Go, implemented with generics and depending only on the
standard library. It provides the aggregation helpers (Sum, SumBy, Mean,
MeanBy, Max, Min, MaxBy, MinBy), the arithmetic wrappers (Add, Subtract,
Multiply, Divide), the rounding helpers (Round, Ceil, Floor), the range and
bounds helpers (Clamp, InRange), and the random helpers (Random, RandomInt)
that lodash documents under _.sum, _.mean, _.max, _.clamp, _.round and
friends.

Use this package whenever Go code needs the small, familiar numeric utilities
that a JavaScript or Node developer would reach for out of lodash: totalling
or averaging a slice, finding the extreme element of a collection by a derived
key, clamping a value into a range, or rounding to a given number of decimal
places. Because the API is generic over a Number constraint rather than fixed
to float64, the aggregation and arithmetic helpers preserve the caller's own
integer or floating-point type (and named types derived from them) instead of
forcing everything through float64 and back.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/math"
```

## Usage

This is the package's own `ExampleMaxBy`, so it compiles and its output is
asserted on every `go test ./math/`.

```go
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
```

```
Grace
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Add[T Number](augend, addend T) T` | Add returns augend + addend. |
| `func Ceil(number float64, precision int) float64` | Ceil rounds number up to precision decimal places. |
| `func Clamp[T Number](number, lower, upper T) T` | Clamp restricts number to be within the inclusive lower and upper bounds. |
| `func Divide[T Number](dividend, divisor T) T` | Divide returns dividend / divisor. |
| `func Floor(number float64, precision int) float64` | Floor rounds number down to precision decimal places. |
| `func InRange[T Number](number, start, end T) bool` | InRange reports whether number is in the half-open range bounded by start and end, that is start <= number < end. |
| `func Max[T Number](nums []T) T` | Max returns the largest value in nums. |
| `func MaxBy[T any, N Number](coll []T, iteratee func(T) N) T` | MaxBy returns the element of coll that yields the largest value when passed through iteratee. |
| `func Mean[T Number](nums []T) float64` | Mean computes the arithmetic mean of the values in nums. |
| `func MeanBy[T any, N Number](coll []T, iteratee func(T) N) float64` | MeanBy computes the arithmetic mean of applying iteratee to each element of coll. |
| `func Min[T Number](nums []T) T` | Min returns the smallest value in nums. |
| `func MinBy[T any, N Number](coll []T, iteratee func(T) N) T` | MinBy returns the element of coll that yields the smallest value when passed through iteratee. |
| `func Multiply[T Number](multiplier, multiplicand T) T` | Multiply returns multiplier * multiplicand. |
| `func Random(min, max float64, rnd func() float64) float64` | Random returns a floating-point number between min and max (inclusive of min, exclusive of max). |
| `func RandomInt(min, max int, rnd func() float64) int` | RandomInt returns an integer between min and max, inclusive of both bounds. |
| `func Round(number float64, precision int) float64` | Round rounds number to precision decimal places. |
| `func Subtract[T Number](minuend, subtrahend T) T` | Subtract returns minuend - subtrahend. |
| `func Sum[T Number](nums []T) T` | Sum computes the sum of the values in nums. |
| `func SumBy[T any, N Number](coll []T, iteratee func(T) N) N` | SumBy computes the sum of applying iteratee to each element of coll. |

### Types

| Type | What it is |
| --- | --- |
| `Number` | Number is the set of built-in numeric types (and named types derived from them) that the helpers in this package operate on. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/math).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
