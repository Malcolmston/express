package object

// This file extends the object package with generic, map-oriented helpers that
// complement the base set: At looks up several keys at once, AssignWith merges
// with a customizer, and EveryEntry/SomeEntry/FindEntry/ReduceEntries iterate a
// map's entries. Because Go map iteration order is unspecified, the predicate
// helpers make no ordering guarantee; FindEntry returns an arbitrary matching
// entry. All functions leave their inputs unmodified (AssignWith writes into a
// fresh map) and depend only on the standard library.

// At returns the values of m at the given keys, in the order the keys are
// supplied. A key that is not present yields the zero value of V. It mirrors
// lodash.at for a map.
func At[K comparable, V any](m map[K]V, keys ...K) []V {
	out := make([]V, len(keys))
	for i, k := range keys {
		out[i] = m[k]
	}
	return out
}

// AssignWith merges src into a copy of dst, resolving each key with customizer,
// which receives the destination and source values (the destination value is
// the zero value of V when the key is only in src) and returns the value to
// store. It mirrors lodash.assignWith and never mutates its arguments.
func AssignWith[K comparable, V any](dst, src map[K]V, customizer func(dstVal, srcVal V, key K) V) map[K]V {
	out := make(map[K]V, len(dst))
	for k, v := range dst {
		out[k] = v
	}
	for k, sv := range src {
		out[k] = customizer(out[k], sv, k)
	}
	return out
}

// EveryEntry reports whether predicate returns true for every entry of m. It
// returns true for an empty map.
func EveryEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) bool {
	for k, v := range m {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// SomeEntry reports whether predicate returns true for at least one entry of m.
// It returns false for an empty map.
func SomeEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) bool {
	for k, v := range m {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// FindEntry returns an arbitrary entry of m for which predicate is true, along
// with true; if no entry matches it returns zero values and false.
func FindEntry[K comparable, V any](m map[K]V, predicate func(key K, value V) bool) (K, V, bool) {
	for k, v := range m {
		if predicate(k, v) {
			return k, v, true
		}
	}
	var zk K
	var zv V
	return zk, zv, false
}

// ReduceEntries folds the entries of m into an accumulator using iteratee.
// Because map iteration order is unspecified, use it only with an associative,
// commutative reduction (such as summing values). It mirrors lodash.reduce over
// an object.
func ReduceEntries[K comparable, V any, R any](m map[K]V, iteratee func(acc R, key K, value V) R, accumulator R) R {
	for k, v := range m {
		accumulator = iteratee(accumulator, k, v)
	}
	return accumulator
}
