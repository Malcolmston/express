// Package function provides idiomatic, generic Go ports of lodash's "Function"
// category (https://lodash.com/docs — after, before, negate, memoize, flip,
// wrap, over, overEvery, overSome, flow, flowRight, curry and partial). These
// are the higher-order helpers that build new functions out of existing ones,
// letting Go code adopt the compositional style JavaScript projects rely on
// without pulling in a third-party dependency.
//
// Because Go is statically typed and has no arguments object, each helper is
// expressed with explicit generic type parameters and concrete arities. Curry
// and partial come in numbered variants (Curry2, Curry3, Partial1, Partial2)
// rather than a single variadic function, and Flip swaps the two arguments of a
// binary function. Memoize caches results keyed by a comparable argument, and
// MemoizeWith lets the caller derive the cache key from an arbitrary argument.
// After and Before gate how many times an underlying function actually runs,
// Negate inverts a predicate, Over runs several functions over one input and
// collects the results, and OverEvery/OverSome combine predicates with logical
// AND/OR. Flow and FlowRight compose same-typed unary functions left-to-right
// and right-to-left.
//
// The returned closures are not safe for concurrent use unless noted; After,
// Before and Memoize keep mutable call state and should be confined to a single
// goroutine or guarded by the caller. Everything depends only on the standard
// library and is deterministic.
package function

// After returns a function that invokes fn only once it has itself been called
// at least n times; earlier calls return the zero value of T. It mirrors
// lodash.after, which is handy for running an action after n asynchronous steps.
func After[T any](n int, fn func() T) func() T {
	count := 0
	return func() T {
		count++
		if count >= n {
			return fn()
		}
		var zero T
		return zero
	}
}

// Before returns a function that invokes fn on each of its first n-1 calls and
// thereafter returns the result of the last invocation, mirroring lodash.before.
func Before[T any](n int, fn func() T) func() T {
	count := 0
	var result T
	return func() T {
		count++
		if count < n {
			result = fn()
		}
		return result
	}
}

// Negate returns a predicate that yields the logical negation of pred.
func Negate[T any](pred func(T) bool) func(T) bool {
	return func(v T) bool { return !pred(v) }
}

// Memoize returns a memoized version of fn that caches results keyed by the
// (comparable) argument, so fn runs at most once per distinct input.
func Memoize[K comparable, V any](fn func(K) V) func(K) V {
	cache := make(map[K]V)
	return func(k K) V {
		if v, ok := cache[k]; ok {
			return v
		}
		v := fn(k)
		cache[k] = v
		return v
	}
}

// MemoizeWith returns a memoized version of fn using key to derive a comparable
// cache key from the argument, mirroring lodash.memoize's resolver option.
func MemoizeWith[T any, K comparable, V any](fn func(T) V, key func(T) K) func(T) V {
	cache := make(map[K]V)
	return func(t T) V {
		k := key(t)
		if v, ok := cache[k]; ok {
			return v
		}
		v := fn(t)
		cache[k] = v
		return v
	}
}

// Flip returns a binary function with its two arguments swapped.
func Flip[A, B, R any](fn func(A, B) R) func(B, A) R {
	return func(b B, a A) R { return fn(a, b) }
}

// Wrap returns a niladic function that calls wrapper with value, mirroring
// lodash.wrap for the common single-value case.
func Wrap[T, R any](value T, wrapper func(T) R) func() R {
	return func() R { return wrapper(value) }
}

// Over returns a function that invokes each of fns with its argument and returns
// the slice of results in order.
func Over[T, R any](fns ...func(T) R) func(T) []R {
	return func(v T) []R {
		out := make([]R, len(fns))
		for i, fn := range fns {
			out[i] = fn(v)
		}
		return out
	}
}

// OverEvery returns a predicate that reports whether the argument satisfies
// every predicate in preds (logical AND, short-circuiting). With no predicates
// it returns true.
func OverEvery[T any](preds ...func(T) bool) func(T) bool {
	return func(v T) bool {
		for _, p := range preds {
			if !p(v) {
				return false
			}
		}
		return true
	}
}

// OverSome returns a predicate that reports whether the argument satisfies at
// least one predicate in preds (logical OR, short-circuiting). With no
// predicates it returns false.
func OverSome[T any](preds ...func(T) bool) func(T) bool {
	return func(v T) bool {
		for _, p := range preds {
			if p(v) {
				return true
			}
		}
		return false
	}
}

// Flow returns a function that pipes its argument through each of fns from left
// to right, so Flow(f, g)(x) is g(f(x)). With no functions it is the identity.
func Flow[T any](fns ...func(T) T) func(T) T {
	return func(v T) T {
		for _, fn := range fns {
			v = fn(v)
		}
		return v
	}
}

// FlowRight returns a function that pipes its argument through each of fns from
// right to left, so FlowRight(f, g)(x) is f(g(x)). It mirrors lodash.flowRight
// (a.k.a. compose).
func FlowRight[T any](fns ...func(T) T) func(T) T {
	return func(v T) T {
		for i := len(fns) - 1; i >= 0; i-- {
			v = fns[i](v)
		}
		return v
	}
}

// Curry2 curries a two-argument function into a chain of single-argument
// functions: Curry2(fn)(a)(b) equals fn(a, b).
func Curry2[A, B, R any](fn func(A, B) R) func(A) func(B) R {
	return func(a A) func(B) R {
		return func(b B) R { return fn(a, b) }
	}
}

// Curry3 curries a three-argument function into a chain of single-argument
// functions: Curry3(fn)(a)(b)(c) equals fn(a, b, c).
func Curry3[A, B, C, R any](fn func(A, B, C) R) func(A) func(B) func(C) R {
	return func(a A) func(B) func(C) R {
		return func(b B) func(C) R {
			return func(c C) R { return fn(a, b, c) }
		}
	}
}

// Partial1 partially applies the first argument of a binary function, returning
// a unary function of the remaining argument.
func Partial1[A, B, R any](fn func(A, B) R, a A) func(B) R {
	return func(b B) R { return fn(a, b) }
}

// Partial2 partially applies the first two arguments of a ternary function,
// returning a unary function of the remaining argument.
func Partial2[A, B, C, R any](fn func(A, B, C) R, a A, b B) func(C) R {
	return func(c C) R { return fn(a, b, c) }
}
