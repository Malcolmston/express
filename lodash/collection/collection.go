// Package collection is a standalone, dependency-free port of lodash's
// "Collection" category of functions (https://lodash.com/docs -
// _.countBy, _.groupBy, _.keyBy, _.partition, _.map, _.filter, _.reduce,
// _.sortBy, _.orderBy, _.sample, _.shuffle and friends) to Go generics. It
// operates over Go slices ([]T) and depends only on the Go standard library
// (cmp, sort, math/rand, crypto/rand), so it can be dropped into any project
// without pulling in third-party code.
//
// The package exists to give Go programmers the ergonomic, expression-oriented
// data-wrangling vocabulary that JavaScript developers reach for by habit. In
// Node a call such as _.groupBy(users, u => u.role) reads clearly and returns a
// plain object; the naive Go equivalent is a hand-rolled loop plus a map
// declaration at every call site. These helpers collapse that boilerplate into
// a single, type-safe call while remaining allocation-friendly and easy to
// reason about. Reach for them when porting Express/lodash JavaScript to Go, or
// simply when you want map/filter/reduce style pipelines without giving up
// static typing.
//
// Every function that inspects elements does so through an explicit Go closure
// rather than lodash's "iteratee shorthands" (the string, array and object
// shorthands such as _.map(users, 'name')): Go has no first-class property
// paths, so an iteratee is always a func(T) K and a predicate is always a
// func(T) bool. Iteratees that derive a grouping or sort key (CountBy, KeyBy,
// GroupBy, SortBy, MinBy, MaxBy, SumBy, MeanBy) constrain K to a comparable or
// cmp.Ordered type as appropriate; Map and FlatMap are additionally
// parameterized by an independent result type R so the output slice need not
// share the input's element type. Predicate-driven functions (Filter, Reject,
// Find, FindLast, Every, Some, Partition) test truthiness with an ordinary
// bool. ForEach (and its alias Each) is the one iteratee whose bool is a
// control signal: returning false stops iteration early, matching lodash's
// ability to break out of _.forEach by returning false.
//
// Ordering and edge cases follow lodash's observable behavior wherever Go
// allows. All traversals visit elements in slice order, and the functions that
// preserve position (Filter, Reject, Partition, GroupBy within a group, Map)
// keep the input's relative order; SortBy and OrderBy are stable, so equal keys
// retain their original sequence. Empty and nil inputs are handled without
// panicking: an empty or nil slice has length zero, so Filter/Reject/Map return
// empty (non-nil) slices, GroupBy/KeyBy/CountBy return empty (non-nil) maps,
// Every returns true (vacuous truth) while Some returns false, Find/MinBy/MaxBy
// report ok == false with the element type's zero value, and MeanBy returns 0.
// KeyBy resolves duplicate keys last-wins, GroupBy and CountBy accumulate all
// colliding elements, and none of the transforming functions mutate their input
// - SortBy, OrderBy, Shuffle, SampleSize and Sample all operate on a copy.
//
// Parity with Node/lodash is intentional but bounded by the type system. Names,
// argument order and results match the JavaScript originals closely enough that
// a port reads almost line-for-line, and behaviors such as vacuous Every, the
// last-wins semantics of KeyBy and the early-exit of ForEach are reproduced
// deliberately. The differences are those Go forces: functions that JavaScript
// returns a single "not found"/undefined value for instead return a (value,
// bool) pair (Find, FindLast, Sample, MinBy, MaxBy); OrderBy takes explicit key
// functions and a parallel slice of "asc"/"desc" strings (an unknown direction
// is treated as "asc") and compares only the ordered kinds - numbers as
// float64, plus strings and bools; and InvokeMap takes a method value
// (func(T, ...any) R) rather than a string method name, since Go has no dynamic
// dispatch by name. For the randomized helpers (Sample, SampleSize, Shuffle) an
// optional *math/rand.Rand may be supplied so results are deterministic in
// tests; when nil is passed a process-wide generator seeded from crypto/rand is
// used.
package collection

import (
	"cmp"
	cryptorand "crypto/rand"
	"encoding/binary"
	"math/rand"
	"sort"
	"sync"
)

// Number is the set of types over which SumBy and MeanBy operate.
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

// defaultRand is a lazily-initialized, crypto-seeded generator used whenever a
// caller passes a nil *rand.Rand to a randomized function.
var (
	defaultRandOnce sync.Once
	defaultRandMu   sync.Mutex
	defaultRand     *rand.Rand
)

func randOf(r *rand.Rand) *rand.Rand {
	if r != nil {
		return r
	}
	defaultRandOnce.Do(func() {
		var b [8]byte
		var seed int64
		if _, err := cryptorand.Read(b[:]); err == nil {
			seed = int64(binary.LittleEndian.Uint64(b[:]))
		}
		defaultRand = rand.New(rand.NewSource(seed))
	})
	return defaultRand
}

// intn is a small helper that draws a random int in [0, n) using r, guarding
// the shared default generator with a mutex so it is safe for concurrent use.
func intn(r *rand.Rand, n int) int {
	if r != nil {
		return r.Intn(n)
	}
	defaultRandMu.Lock()
	defer defaultRandMu.Unlock()
	return randOf(nil).Intn(n)
}

// CountBy creates a map of keys generated from iteratee to the number of times
// that key occurred across the collection.
func CountBy[T any, K comparable](s []T, iteratee func(T) K) map[K]int {
	out := make(map[K]int)
	for _, v := range s {
		out[iteratee(v)]++
	}
	return out
}

// KeyBy creates a map keyed by the result of iteratee for each element. When
// multiple elements produce the same key the last one wins.
func KeyBy[T any, K comparable](s []T, iteratee func(T) K) map[K]T {
	out := make(map[K]T)
	for _, v := range s {
		out[iteratee(v)] = v
	}
	return out
}

// GroupBy creates a map of keys generated from iteratee to slices of the
// elements responsible for generating each key. Order within a group matches
// the order of the input.
func GroupBy[T any, K comparable](s []T, iteratee func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, v := range s {
		k := iteratee(v)
		out[k] = append(out[k], v)
	}
	return out
}

// Partition splits the collection into two slices: the first holds elements for
// which predicate returns true, the second holds the rest. Relative order is
// preserved in both.
func Partition[T any](s []T, predicate func(T) bool) (truthy []T, falsy []T) {
	for _, v := range s {
		if predicate(v) {
			truthy = append(truthy, v)
		} else {
			falsy = append(falsy, v)
		}
	}
	return truthy, falsy
}

// Every reports whether predicate returns true for all elements. It returns
// true for an empty collection (vacuous truth), matching lodash.
func Every[T any](s []T, predicate func(T) bool) bool {
	for _, v := range s {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// Some reports whether predicate returns true for any element. It returns false
// for an empty collection.
func Some[T any](s []T, predicate func(T) bool) bool {
	for _, v := range s {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Filter returns a new slice of the elements for which predicate returns true.
func Filter[T any](s []T, predicate func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if predicate(v) {
			out = append(out, v)
		}
	}
	return out
}

// Reject is the opposite of Filter: it returns the elements for which predicate
// returns false.
func Reject[T any](s []T, predicate func(T) bool) []T {
	out := make([]T, 0, len(s))
	for _, v := range s {
		if !predicate(v) {
			out = append(out, v)
		}
	}
	return out
}

// Find returns the first element for which predicate returns true. The second
// return value reports whether such an element was found.
func Find[T any](s []T, predicate func(T) bool) (T, bool) {
	for _, v := range s {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindLast returns the last element for which predicate returns true, iterating
// from the end. The second return value reports whether one was found.
func FindLast[T any](s []T, predicate func(T) bool) (T, bool) {
	for i := len(s) - 1; i >= 0; i-- {
		if predicate(s[i]) {
			return s[i], true
		}
	}
	var zero T
	return zero, false
}

// ForEach iterates over the collection invoking iteratee for each element. If
// iteratee returns false iteration stops early, mirroring lodash's ability to
// break out of forEach.
func ForEach[T any](s []T, iteratee func(T) bool) {
	for _, v := range s {
		if !iteratee(v) {
			return
		}
	}
}

// Each is an alias for ForEach.
func Each[T any](s []T, iteratee func(T) bool) {
	ForEach(s, iteratee)
}

// Map creates a new slice of values by running each element through iteratee.
func Map[T, R any](s []T, iteratee func(T) R) []R {
	out := make([]R, len(s))
	for i, v := range s {
		out[i] = iteratee(v)
	}
	return out
}

// FlatMap maps each element through iteratee (which returns a slice) and
// flattens the results one level deep.
func FlatMap[T, R any](s []T, iteratee func(T) []R) []R {
	out := make([]R, 0, len(s))
	for _, v := range s {
		out = append(out, iteratee(v)...)
	}
	return out
}

// Reduce reduces the collection to a single accumulated value, iterating from
// left to right. accumulator provides the initial value.
func Reduce[T, R any](s []T, iteratee func(acc R, cur T) R, accumulator R) R {
	acc := accumulator
	for _, v := range s {
		acc = iteratee(acc, v)
	}
	return acc
}

// ReduceRight is like Reduce except it iterates from right to left.
func ReduceRight[T, R any](s []T, iteratee func(acc R, cur T) R, accumulator R) R {
	acc := accumulator
	for i := len(s) - 1; i >= 0; i-- {
		acc = iteratee(acc, s[i])
	}
	return acc
}

// Includes reports whether value is an element of the collection.
func Includes[T comparable](s []T, value T) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}
	return false
}

// Size returns the number of elements in the collection.
func Size[T any](s []T) int {
	return len(s)
}

// SortBy returns a new slice sorted in ascending order by the key produced by
// iteratee. The sort is stable, so elements with equal keys keep their original
// relative order.
func SortBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) []T {
	out := make([]T, len(s))
	copy(out, s)
	sort.SliceStable(out, func(i, j int) bool {
		return iteratee(out[i]) < iteratee(out[j])
	})
	return out
}

// OrderBy returns a new slice sorted by multiple key functions with a matching
// direction for each ("asc" or "desc"; any other value is treated as "asc").
// Keys return values as any; supported comparable kinds are the signed/unsigned
// integers, floats, strings and bools. Ties on one key fall through to the
// next. The sort is stable.
func OrderBy[T any](s []T, keys []func(T) any, orders []string) []T {
	out := make([]T, len(s))
	copy(out, s)
	sort.SliceStable(out, func(i, j int) bool {
		for idx, key := range keys {
			c := compareAny(key(out[i]), key(out[j]))
			if c == 0 {
				continue
			}
			desc := idx < len(orders) && orders[idx] == "desc"
			if desc {
				return c > 0
			}
			return c < 0
		}
		return false
	})
	return out
}

// compareAny compares two values of supported ordered kinds, returning -1, 0 or
// 1. Numeric values are compared as float64; mismatched or unsupported kinds
// compare as equal.
func compareAny(a, b any) int {
	af, aok := toFloat(a)
	bf, bok := toFloat(b)
	if aok && bok {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	if as, ok := a.(string); ok {
		if bs, ok := b.(string); ok {
			return cmp.Compare(as, bs)
		}
	}
	if ab, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			switch {
			case ab == bb:
				return 0
			case !ab && bb:
				return -1
			default:
				return 1
			}
		}
	}
	return 0
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case uintptr:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	default:
		return 0, false
	}
}

// Sample returns a random element from the collection. The second return value
// is false when the collection is empty. Pass a seeded *rand.Rand for
// deterministic behavior, or nil to use a crypto-seeded default.
func Sample[T any](s []T, r *rand.Rand) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	return s[intn(r, len(s))], true
}

// SampleSize returns n elements sampled without replacement from the
// collection, in random order. If n is greater than or equal to the length the
// whole (shuffled) collection is returned; n <= 0 yields an empty slice.
func SampleSize[T any](s []T, n int, r *rand.Rand) []T {
	if n <= 0 || len(s) == 0 {
		return []T{}
	}
	if n > len(s) {
		n = len(s)
	}
	// Partial Fisher-Yates on a copy.
	pool := make([]T, len(s))
	copy(pool, s)
	for i := 0; i < n; i++ {
		j := i + intn(r, len(pool)-i)
		pool[i], pool[j] = pool[j], pool[i]
	}
	out := make([]T, n)
	copy(out, pool[:n])
	return out
}

// Shuffle returns a new slice with the elements of the collection in random
// order, using the Fisher-Yates algorithm.
func Shuffle[T any](s []T, r *rand.Rand) []T {
	out := make([]T, len(s))
	copy(out, s)
	for i := len(out) - 1; i > 0; i-- {
		j := intn(r, i+1)
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// MinBy returns the element that yields the smallest value when passed through
// iteratee. The second return value is false for an empty collection.
func MinBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best := s[0]
	bestKey := iteratee(best)
	for _, v := range s[1:] {
		k := iteratee(v)
		if k < bestKey {
			best, bestKey = v, k
		}
	}
	return best, true
}

// MaxBy returns the element that yields the largest value when passed through
// iteratee. The second return value is false for an empty collection.
func MaxBy[T any, K cmp.Ordered](s []T, iteratee func(T) K) (T, bool) {
	if len(s) == 0 {
		var zero T
		return zero, false
	}
	best := s[0]
	bestKey := iteratee(best)
	for _, v := range s[1:] {
		k := iteratee(v)
		if k > bestKey {
			best, bestKey = v, k
		}
	}
	return best, true
}

// SumBy sums the values produced by running each element through iteratee.
func SumBy[T any, N Number](s []T, iteratee func(T) N) N {
	var sum N
	for _, v := range s {
		sum += iteratee(v)
	}
	return sum
}

// MeanBy returns the mean (as float64) of the values produced by iteratee. It
// returns 0 for an empty collection.
func MeanBy[T any, N Number](s []T, iteratee func(T) N) float64 {
	if len(s) == 0 {
		return 0
	}
	var sum float64
	for _, v := range s {
		sum += float64(iteratee(v))
	}
	return sum / float64(len(s))
}

// InvokeMap invokes method for each element of the collection, passing args to
// each call, and returns a slice of the results. Because Go has no dynamic
// method dispatch by name, method is supplied as a function; this is the
// idiomatic Go equivalent of lodash's _.invokeMap(collection, methodName).
func InvokeMap[T, R any](s []T, method func(T, ...any) R, args ...any) []R {
	out := make([]R, len(s))
	for i, v := range s {
		out[i] = method(v, args...)
	}
	return out
}
