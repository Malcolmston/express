# lang — ports the "Lang" category of the npm lodash library

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/lang.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/lang)

Package lang ports the "Lang" category of the npm lodash library, together
with a handful of closely related utility helpers, to idiomatic Go using only
the standard library. It supplies the runtime type predicates (IsNil, IsEmpty,
IsArray, IsMap, IsString, IsNumber, IsBool, IsFunc, IsPointer, IsZero,
IsPlainObject, IsError, IsEqual, Eq), the value coercions (ToString, ToNumber,
ToInteger, ToFinite, ToArray, CastArray, DefaultTo), the relational
comparators (Gt, Gte, Lt, Lte), and the general utility helpers (Times,
Identity, Constant, Noop, Range, UniqueId, Once) that lodash groups under its
Lang and Util documentation sections.

Reach for this package when Go code needs to inspect or normalise values whose
concrete type is not known until run time, for example while decoding
arbitrary JSON into an any, validating loosely typed configuration, or
re-implementing JavaScript-flavoured logic during a port. It is the Go
analogue of scattering _.isEmpty, _.toNumber or _.defaultTo calls throughout a
Node codebase, and it lets you keep the permissive, coercion-heavy semantics
that front-end and Node developers expect rather than fighting Go's static
type system with hand-written type switches at every call site.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/lang"
```

## Usage

This is the package's own `ExampleDefaultTo`, so it compiles and its output is
asserted on every `go test ./lang/`.

```go
fmt.Println(lang.DefaultTo("hello", "fallback"))
	fmt.Println(lang.DefaultTo(nil, "fallback"))
	fmt.Println(lang.DefaultTo(42, 0))
```

```
hello
fallback
42
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func CastArray(value any) []any` | CastArray wraps value in a []any unless it is already a slice or array, in which case its elements are copied into a fresh []any. |
| `func Constant[T any](value T) func() T` | Constant returns a function that always returns value. |
| `func DefaultTo(value, defaultValue any) any` | DefaultTo returns value unless it is nil or (for numbers) NaN, in which case it returns defaultValue. |
| `func Eq(a, b any) bool` | Eq reports whether a and b are equivalent using deep equality, treating two NaN values as equal (unlike Go's == operator). |
| `func Gt(a, b any) bool` | Gt reports whether a is greater than b. |
| `func Gte(a, b any) bool` | Gte reports whether a is greater than or equal to b. |
| `func Identity[T any](value T) T` | Identity returns its argument unchanged. |
| `func IsArray(value any) bool` | IsArray reports whether value is a Go slice or array. |
| `func IsBool(value any) bool` | IsBool reports whether value is a boolean. |
| `func IsEmpty(value any) bool` | IsEmpty reports whether value is considered "empty". |
| `func IsEqual(a, b any) bool` | IsEqual performs a deep comparison between two values and reports whether they are equivalent. |
| `func IsError(value any) bool` | IsError reports whether value implements the error interface. |
| `func IsFinite(f float64) bool` | IsFinite reports whether f is a finite number (neither NaN nor an infinity), mirroring lodash.isFinite for concrete float64 values. |
| `func IsFunc(value any) bool` | IsFunc reports whether value is a function. |
| `func IsInteger(f float64) bool` | IsInteger reports whether f is finite and has no fractional part, mirroring lodash.isInteger. |
| `func IsMap(value any) bool` | IsMap reports whether value is a Go map. |
| `func IsNaN(f float64) bool` | IsNaN reports whether f is an IEEE-754 not-a-number value, mirroring lodash.isNaN for concrete float64 values. |
| `func IsNil(value any) bool` | IsNil reports whether value is nil. |
| `func IsNumber(value any) bool` | IsNumber reports whether value is any built-in integer, unsigned integer or floating point type. |
| `func IsPlainObject(value any) bool` | IsPlainObject reports whether value is a "plain object": a Go map with string keys or a struct (or a pointer to one). |
| `func IsPointer(value any) bool` | IsPointer reports whether value is a pointer. |
| `func IsSafeInteger(f float64) bool` | IsSafeInteger reports whether f is an integer within the safe range [-(2^53-1), 2^53-1], mirroring lodash.isSafeInteger. |
| `func IsString(value any) bool` | IsString reports whether value is a string. |
| `func IsZero(value any) bool` | IsZero reports whether value is the zero value for its type. |
| `func Lt(a, b any) bool` | Lt reports whether a is less than b. |
| `func Lte(a, b any) bool` | Lte reports whether a is less than or equal to b. |
| `func Noop()` | Noop does nothing and returns nothing. |
| `func Once[T any](fn func() T) func() T` | Once returns a function that invokes fn at most once. |
| `func Range(n int) []int` | Range returns a slice of ints from 0 up to, but not including, n. |
| `func Times[T any](n int, fn func(i int) T) []T` | Times invokes fn n times, collecting the results into a slice. |
| `func ToArray(value any) []any` | ToArray converts value into a []any. |
| `func ToFinite(value any) float64` | ToFinite converts value to a finite float64. |
| `func ToInteger(value any) int` | ToInteger converts value to an integer by truncating toward zero. |
| `func ToLength(f float64) int` | ToLength clamps f to a valid array length in [0, 2^32-1] after truncating any fractional part, mirroring lodash.toLength. |
| `func ToNumber(value any) float64` | ToNumber converts value to a float64. |
| `func ToSafeInteger(f float64) float64` | ToSafeInteger converts f to an integer clamped into the safe-integer range, mirroring lodash.toSafeInteger. |
| `func ToString(value any) string` | ToString converts value to its string representation. |
| `func UniqueId(prefix string) string` | UniqueId returns a unique string identifier, optionally prefixed. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/lang).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
