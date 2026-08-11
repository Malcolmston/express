# function — provides idiomatic, generic Go ports of lodash's "Function" category

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/lodash/function.svg)](https://pkg.go.dev/github.com/malcolmston/express/lodash/function)

Package function provides idiomatic, generic Go ports of lodash's "Function"
category (https://lodash.com/docs — after, before, negate, memoize, flip,
wrap, over, overEvery, overSome, flow, flowRight, curry and partial). These
are the higher-order helpers that build new functions out of existing ones,
letting Go code adopt the compositional style JavaScript projects rely on
without pulling in a third-party dependency.

Because Go is statically typed and has no arguments object, each helper is
expressed with explicit generic type parameters and concrete arities. Curry
and partial come in numbered variants (Curry2, Curry3, Partial1, Partial2)
rather than a single variadic function, and Flip swaps the two arguments of a
binary function. Memoize caches results keyed by a comparable argument, and
MemoizeWith lets the caller derive the cache key from an arbitrary argument.
After and Before gate how many times an underlying function actually runs,
Negate inverts a predicate, Over runs several functions over one input and
collects the results, and OverEvery/OverSome combine predicates with logical
AND/OR. Flow and FlowRight compose same-typed unary functions left-to-right
and right-to-left.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/lodash/function"
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func After[T any](n int, fn func() T) func() T` | After returns a function that invokes fn only once it has itself been called at least n times; earlier calls return the zero value of T. It mirrors… |
| `func Before[T any](n int, fn func() T) func() T` | Before returns a function that invokes fn on each of its first n-1 calls and thereafter returns the result of the last invocation, mirroring… |
| `func Curry2[A, B, R any](fn func(A, B) R) func(A) func(B) R` | Curry2 curries a two-argument function into a chain of single-argument functions: Curry2(fn)(a)(b) equals fn(a, b). |
| `func Curry3[A, B, C, R any](fn func(A, B, C) R) func(A) func(B) func(C) R` | Curry3 curries a three-argument function into a chain of single-argument functions: Curry3(fn)(a)(b)(c) equals fn(a, b, c). |
| `func Flip[A, B, R any](fn func(A, B) R) func(B, A) R` | Flip returns a binary function with its two arguments swapped. |
| `func Flow[T any](fns ...func(T) T) func(T) T` | Flow returns a function that pipes its argument through each of fns from left to right, so Flow(f, g)(x) is g(f(x)). |
| `func FlowRight[T any](fns ...func(T) T) func(T) T` | FlowRight returns a function that pipes its argument through each of fns from right to left, so FlowRight(f, g)(x) is f(g(x)). |
| `func Memoize[K comparable, V any](fn func(K) V) func(K) V` | Memoize returns a memoized version of fn that caches results keyed by the (comparable) argument, so fn runs at most once per distinct input. |
| `func MemoizeWith[T any, K comparable, V any](fn func(T) V, key func(T) K) func(T) V` | MemoizeWith returns a memoized version of fn using key to derive a comparable cache key from the argument, mirroring lodash.memoize's resolver option. |
| `func Negate[T any](pred func(T) bool) func(T) bool` | Negate returns a predicate that yields the logical negation of pred. |
| `func Over[T, R any](fns ...func(T) R) func(T) []R` | Over returns a function that invokes each of fns with its argument and returns the slice of results in order. |
| `func OverEvery[T any](preds ...func(T) bool) func(T) bool` | OverEvery returns a predicate that reports whether the argument satisfies every predicate in preds (logical AND, short-circuiting). |
| `func OverSome[T any](preds ...func(T) bool) func(T) bool` | OverSome returns a predicate that reports whether the argument satisfies at least one predicate in preds (logical OR, short-circuiting). |
| `func Partial1[A, B, R any](fn func(A, B) R, a A) func(B) R` | Partial1 partially applies the first argument of a binary function, returning a unary function of the remaining argument. |
| `func Partial2[A, B, C, R any](fn func(A, B, C) R, a A, b B) func(C) R` | Partial2 partially applies the first two arguments of a ternary function, returning a unary function of the remaining argument. |
| `func Wrap[T, R any](value T, wrapper func(T) R) func() R` | Wrap returns a niladic function that calls wrapper with value, mirroring lodash.wrap for the common single-value case. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/lodash/function).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
