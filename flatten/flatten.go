// Package flatten provides faithful ports of lodash's flatten, flattenDeep,
// and flattenDepth.
//
// The three functions collapse nested slices into a flatter slice. Use Flatten
// to merge one level of nesting with full compile-time type safety, use
// FlattenDeep to collapse an arbitrarily nested structure completely, and use
// FlattenDepth when you need to control exactly how many levels to remove. They
// are handy for normalizing grouped results, splicing together lists of lists,
// or unwrapping data that arrived more deeply nested than the consumer wants.
//
// Flatten is generic over the element type and works on a [][]T, concatenating
// every inner slice into a single []T in order; it is a pure, allocation-sized
// pass with no reflection. FlattenDeep and FlattenDepth instead accept any and
// use reflection so they can descend through mixed and heterogeneously typed
// nesting, including []any, []int, arrays, and interface-wrapped slices. Any
// element whose dynamic kind is a slice or array is descended into; every other
// element is appended as a leaf. Interface values are unwrapped to their
// concrete value before this test, so an []any holding a typed slice is still
// flattened.
//
// FlattenDeep descends without limit. FlattenDepth takes an explicit depth:
// depth 1 removes a single level, each larger depth removes one more, and a
// depth of 0 or any negative value copies the top-level elements without
// descending at all. A depth larger than the actual nesting simply flattens
// everything, matching FlattenDeep.
//
// Several edge cases are handled deliberately. Strings are scalar leaves and are
// never split into runes or bytes, so []any{"hi"} stays a single "hi". A nil,
// empty, or non-slice input yields an empty but non-nil slice rather than nil,
// which keeps reflect.DeepEqual comparisons and JSON output predictable.
// Invalid or nil interface elements are appended as a literal nil. Compared to
// lodash, these ports drop the iteratee/predicate variants (flatMap and
// friends) and, because Go slices are homogeneous at the type level, expose the
// deep and depth forms as any-based functions returning []any while keeping the
// single-level Flatten fully typed.
package flatten

import "reflect"

// Flatten flattens s by a single level, concatenating all inner slices into
// one new slice in order.
//
// A nil or empty input yields an empty (non-nil) slice.
func Flatten[T any](s [][]T) []T {
	total := 0
	for _, inner := range s {
		total += len(inner)
	}
	result := make([]T, 0, total)
	for _, inner := range s {
		result = append(result, inner...)
	}
	return result
}

// FlattenDeep recursively flattens an arbitrarily nested value into a flat
// []any. Any element that is itself a slice or array (of any type, including
// []any) is descended into to unlimited depth; all other elements are
// appended as leaves.
//
// Strings are treated as scalar leaves and are never flattened into their
// bytes or runes. A nil or non-slice input yields an empty (non-nil) slice.
func FlattenDeep(s any) []any {
	result := make([]any, 0)
	v := deref(reflect.ValueOf(s))
	if !isFlattenable(v) {
		return result
	}
	return baseFlatten(result, v, -1)
}

// FlattenDepth recursively flattens a nested value up to depth levels deep,
// mirroring lodash's flattenDepth.
//
// FlattenDepth(x, 1) is equivalent to a single-level flatten. A depth of 0
// (or negative) copies the top-level elements without descending into any
// nested slices. Larger depths descend further. Elements that are not slices
// are always appended as leaves.
func FlattenDepth(s any, depth int) []any {
	if depth < 0 {
		depth = 0
	}
	result := make([]any, 0)
	v := deref(reflect.ValueOf(s))
	if !isFlattenable(v) {
		return result
	}
	return baseFlatten(result, v, depth)
}

// deref unwraps interface values to their concrete underlying value.
func deref(v reflect.Value) reflect.Value {
	for v.IsValid() && v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// isFlattenable reports whether v is a slice or array that should be descended
// into.
func isFlattenable(v reflect.Value) bool {
	if !v.IsValid() {
		return false
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

// baseFlatten iterates the elements of the container v, appending each element
// to acc. A slice/array element is recursively flattened while depth allows it;
// depth < 0 means unlimited depth, and depth == 0 stops further descent.
func baseFlatten(acc []any, v reflect.Value, depth int) []any {
	for i := 0; i < v.Len(); i++ {
		el := deref(v.Index(i))
		if depth != 0 && isFlattenable(el) {
			next := -1
			if depth > 0 {
				next = depth - 1
			}
			acc = baseFlatten(acc, el, next)
			continue
		}
		if !el.IsValid() {
			acc = append(acc, nil)
			continue
		}
		acc = append(acc, el.Interface())
	}
	return acc
}
