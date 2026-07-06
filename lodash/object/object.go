// Package object ports the "Object" category of the npm lodash library to Go,
// providing the map- and structure-oriented helpers such as Keys, Values,
// Entries, Pick, Omit, MapKeys, MapValues, Invert, Assign, Merge, Defaults,
// Get, Set, Has, Unset, Update, Clone, CloneDeep, IsEqual and Transform. In
// JavaScript these operate on plain objects; here they operate on Go maps and,
// for the dynamically-typed path and merge helpers, on the any-typed tree of
// map[string]any and []any values that a decoded JSON document produces. The
// package depends only on the Go standard library.
//
// Reach for this package when you are manipulating decoded JSON or other
// loosely-typed nested data and want the ergonomic, batteries-included
// vocabulary that lodash offers on the front end: extracting a subset of keys,
// re-keying or re-mapping values, deeply merging configuration layers, reading
// or writing a value several levels down by a single "a.b.c" path, or comparing
// two trees for structural equality. The statically-typed helpers (Keys,
// Values, Entries, Pick, Omit, MapKeys, MapValues, Invert, FindKey and friends)
// are written with Go generics so they work over any comparable key type and
// any value type without reflection or boxing.
//
// The functions divide into two families by how they are typed. The collection
// helpers are generic over [K comparable, V any] and simply iterate the map to
// build a fresh result; Pick copies only the requested keys that are present,
// Omit copies everything except a drop-set of keys, and the *By variants take a
// predicate or iteratee instead of an explicit key list. The dynamic helpers
// operate on any: Get, Set, Has, Unset and Update accept dot-notation paths
// such as "a.b.c", splitting on "." and walking the tree one segment at a time.
// A numeric segment indexes into a []any slice, so the path "a.0.b" descends
// into the first element of the slice stored under key "a"; Set creates the
// intermediate map or slice containers on demand (choosing a slice when the
// next segment is numeric) and grows slices with nil padding as needed.
//
// Edge-case semantics follow lodash closely. Get returns the supplied
// defaultValue (or nil) whenever any segment along the path is missing, out of
// range, or the resolved value is itself nil. Defaults and DefaultsDeep only
// fill keys that are absent or hold nil, never overwriting an existing
// non-nil value, and DefaultsDeep recurses into nested maps while deep-cloning
// the values it copies in. Merge recursively merges nested maps key by key and
// nested slices index by index, skipping nil source values, whereas Assign is a
// shallow overwrite. CloneDeep recursively copies map and slice containers so
// that later mutation cannot reach the original, and IsEqual compares trees
// structurally, guarding against panics from uncomparable scalar types by
// treating them as unequal. Passing a nil destination to the mutating helpers
// (Assign, Merge, Defaults, DefaultsDeep, Transform) allocates a fresh map.
//
// Parity with Node's lodash is close in behavior but adapted to Go's type
// system and runtime. The most visible difference is ordering: Go map iteration
// is randomized, so Keys, Values, Entries, MapKeys and the For* iterators do
// not preserve insertion order the way lodash does, and functions whose result
// would otherwise be order-dependent take deliberate steps to stay
// deterministic (FindKey examines keys in sorted order). There is no notion of
// JavaScript "undefined" versus a present-but-null value, so nil stands in for
// both; Unset deletes a map key but, matching lodash's array delete, only nils
// out a slice element rather than shifting the slice. Symbol keys, prototype
// chains, getters and lodash's customizer callbacks have no equivalent here.
package object

import (
	"sort"
	"strconv"
	"strings"
)

// Keys returns the keys of m. The order of the returned slice is not
// guaranteed (Go map iteration order is randomized).
func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Values returns the values of m. The order of the returned slice is not
// guaranteed.
func Values[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// Pair is a single key/value entry, as produced by Entries.
type Pair[K comparable, V any] struct {
	// Key is the map key of the entry.
	Key K
	// Value is the map value associated with Key.
	Value V
}

// Entries returns the key/value pairs of m as a slice of Pair. It is the
// generic equivalent of lodash's toPairs.
func Entries[K comparable, V any](m map[K]V) []Pair[K, V] {
	out := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		out = append(out, Pair[K, V]{Key: k, Value: v})
	}
	return out
}

// ToPairs is an alias for Entries, matching lodash's naming.
func ToPairs[K comparable, V any](m map[K]V) []Pair[K, V] {
	return Entries(m)
}

// FromEntries builds a map from a slice of Pair. Later pairs overwrite earlier
// ones that share a key.
func FromEntries[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	out := make(map[K]V, len(pairs))
	for _, p := range pairs {
		out[p.Key] = p.Value
	}
	return out
}

// Pick returns a new map composed of the given keys picked from m. Keys that
// are not present in m are skipped.
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	out := make(map[K]V)
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out
}

// PickBy returns a new map composed of the entries of m for which predicate
// returns true.
func PickBy[K comparable, V any](m map[K]V, predicate func(value V, key K) bool) map[K]V {
	out := make(map[K]V)
	for k, v := range m {
		if predicate(v, k) {
			out[k] = v
		}
	}
	return out
}

// Omit returns a new map with the given keys removed from m.
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	drop := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		drop[k] = struct{}{}
	}
	out := make(map[K]V)
	for k, v := range m {
		if _, skip := drop[k]; !skip {
			out[k] = v
		}
	}
	return out
}

// OmitBy returns a new map with the entries of m for which predicate returns
// true removed.
func OmitBy[K comparable, V any](m map[K]V, predicate func(value V, key K) bool) map[K]V {
	out := make(map[K]V)
	for k, v := range m {
		if !predicate(v, k) {
			out[k] = v
		}
	}
	return out
}

// MapKeys returns a new map with the same values as m but with keys produced
// by running each entry through iteratee. If iteratee maps two keys to the
// same result, the last one written wins.
func MapKeys[K comparable, V any, R comparable](m map[K]V, iteratee func(value V, key K) R) map[R]V {
	out := make(map[R]V, len(m))
	for k, v := range m {
		out[iteratee(v, k)] = v
	}
	return out
}

// MapValues returns a new map with the same keys as m but with values produced
// by running each entry through iteratee.
func MapValues[K comparable, V any, R any](m map[K]V, iteratee func(value V, key K) R) map[K]R {
	out := make(map[K]R, len(m))
	for k, v := range m {
		out[k] = iteratee(v, k)
	}
	return out
}

// Invert returns a new map composed of the inverted keys and values of m. If m
// contains duplicate values, subsequent values overwrite prior ones.
func Invert[K comparable, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// InvertBy returns a new map composed of keys generated from the results of
// running each value of m through iteratee. Each key maps to the slice of
// original keys responsible for generating it.
func InvertBy[K comparable, V any, R comparable](m map[K]V, iteratee func(value V) R) map[R][]K {
	out := make(map[R][]K)
	for k, v := range m {
		r := iteratee(v)
		out[r] = append(out[r], k)
	}
	return out
}

// Defaults assigns, for each source in sources, the values of keys that are
// missing (or hold the zero value equivalent of "undefined") in dst. Existing,
// non-nil keys in dst are preserved. It mutates and returns dst.
//
// Faithful to lodash, only the first value encountered for a key is used, and
// keys already present with a non-nil value are left untouched.
func Defaults(dst map[string]any, sources ...map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}
	for _, src := range sources {
		for k, v := range src {
			if cur, ok := dst[k]; !ok || cur == nil {
				dst[k] = v
			}
		}
	}
	return dst
}

// DefaultsDeep is like Defaults except it recurses into nested map[string]any
// values, filling in missing keys at every depth. It mutates and returns dst.
func DefaultsDeep(dst map[string]any, sources ...map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}
	for _, src := range sources {
		defaultsDeepInto(dst, src)
	}
	return dst
}

func defaultsDeepInto(dst, src map[string]any) {
	for k, sv := range src {
		dv, ok := dst[k]
		if !ok || dv == nil {
			dst[k] = CloneDeep(sv)
			continue
		}
		dm, dIsMap := dv.(map[string]any)
		sm, sIsMap := sv.(map[string]any)
		if dIsMap && sIsMap {
			defaultsDeepInto(dm, sm)
		}
	}
}

// Assign copies the own enumerable keys of each source into dst, with later
// sources overwriting earlier ones. It mutates and returns dst. This is a
// shallow copy (lodash's assign / Object.assign).
func Assign(dst map[string]any, sources ...map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}
	for _, src := range sources {
		for k, v := range src {
			dst[k] = v
		}
	}
	return dst
}

// Merge recursively merges the sources into dst. Nested map[string]any values
// are merged key by key; nested []any values are merged index by index (as
// lodash does for array-like objects). Non-mergeable values from a source
// overwrite the destination unless the source value is nil, in which case the
// destination is preserved. It mutates and returns dst.
func Merge(dst map[string]any, sources ...map[string]any) map[string]any {
	if dst == nil {
		dst = make(map[string]any)
	}
	for _, src := range sources {
		for k, sv := range src {
			dst[k] = mergeValue(dst[k], sv)
		}
	}
	return dst
}

func mergeValue(dv, sv any) any {
	if sv == nil {
		if dv != nil {
			return dv
		}
		return sv
	}
	switch s := sv.(type) {
	case map[string]any:
		dm, ok := dv.(map[string]any)
		if !ok || dm == nil {
			dm = make(map[string]any)
		}
		for k, v := range s {
			dm[k] = mergeValue(dm[k], v)
		}
		return dm
	case []any:
		var da []any
		if existing, ok := dv.([]any); ok {
			da = existing
		}
		for i, v := range s {
			if i < len(da) {
				da[i] = mergeValue(da[i], v)
			} else {
				da = append(da, CloneDeep(v))
			}
		}
		return da
	default:
		return sv
	}
}

// splitPath splits a dot-notation path into its segments.
func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}

// Get retrieves the value at the given dot-notation path within root. If any
// segment along the path is missing, the provided defaultValue is returned (or
// nil if none is supplied). Numeric segments index into []any slices.
func Get(root any, path string, defaultValue ...any) any {
	def := any(nil)
	if len(defaultValue) > 0 {
		def = defaultValue[0]
	}
	cur := root
	for _, seg := range splitPath(path) {
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[seg]
			if !ok {
				return def
			}
			cur = v
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(c) {
				return def
			}
			cur = c[idx]
		default:
			return def
		}
	}
	if cur == nil {
		return def
	}
	return cur
}

// Has reports whether root contains a value at the given dot-notation path.
func Has(root any, path string) bool {
	cur := root
	segs := splitPath(path)
	if len(segs) == 0 {
		return false
	}
	for _, seg := range segs {
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[seg]
			if !ok {
				return false
			}
			cur = v
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(c) {
				return false
			}
			cur = c[idx]
		default:
			return false
		}
	}
	return true
}

// Set sets the value at the given dot-notation path within root, creating
// intermediate map[string]any containers as needed. Numeric segments index
// into (and, when necessary, grow) []any slices. It returns the (possibly
// replaced) root so callers can capture a newly created container.
func Set(root any, path string, value any) any {
	segs := splitPath(path)
	if len(segs) == 0 {
		return root
	}
	if root == nil {
		root = newContainer(segs[0])
	}
	setRec(root, segs, value)
	return root
}

// newContainer creates the appropriate container for the given first segment:
// a []any when the segment is numeric, otherwise a map[string]any.
func newContainer(seg string) any {
	if _, err := strconv.Atoi(seg); err == nil {
		return []any{}
	}
	return map[string]any{}
}

// setRec walks segs within cur, assigning value at the final segment. It
// returns the (possibly reallocated) container so parents can rewire slices
// that had to grow.
func setRec(cur any, segs []string, value any) any {
	seg := segs[0]
	last := len(segs) == 1

	switch c := cur.(type) {
	case map[string]any:
		if last {
			c[seg] = value
			return c
		}
		child, ok := c[seg]
		if !ok || child == nil || !isContainer(child) {
			child = newContainer(segs[1])
		}
		c[seg] = setRec(child, segs[1:], value)
		return c
	case []any:
		idx, err := strconv.Atoi(seg)
		if err != nil {
			return c
		}
		for idx >= len(c) {
			c = append(c, nil)
		}
		if last {
			c[idx] = value
			return c
		}
		child := c[idx]
		if child == nil || !isContainer(child) {
			child = newContainer(segs[1])
		}
		c[idx] = setRec(child, segs[1:], value)
		return c
	default:
		return cur
	}
}

func isContainer(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}

// Update is like Set except the new value is produced by running the current
// value at path (or nil if absent) through updater. It returns the (possibly
// replaced) root.
func Update(root any, path string, updater func(current any) any) any {
	cur := Get(root, path)
	return Set(root, path, updater(cur))
}

// Unset removes the value at the given dot-notation path from root. It reports
// whether a value was removed. Numeric segments index into []any slices; for
// slices the element is set to nil (lodash-style delete) rather than removed.
func Unset(root any, path string) bool {
	segs := splitPath(path)
	if len(segs) == 0 {
		return false
	}
	cur := root
	for i := 0; i < len(segs)-1; i++ {
		seg := segs[i]
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[seg]
			if !ok {
				return false
			}
			cur = v
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(c) {
				return false
			}
			cur = c[idx]
		default:
			return false
		}
	}
	leaf := segs[len(segs)-1]
	switch c := cur.(type) {
	case map[string]any:
		if _, ok := c[leaf]; !ok {
			return false
		}
		delete(c, leaf)
		return true
	case []any:
		idx, err := strconv.Atoi(leaf)
		if err != nil || idx < 0 || idx >= len(c) {
			return false
		}
		c[idx] = nil
		return true
	default:
		return false
	}
}

// Clone returns a shallow copy of value. For maps and slices a new top-level
// container is allocated but nested values are shared with the original. Scalar
// values are returned unchanged.
func Clone(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, e := range v {
			out[k] = e
		}
		return out
	case []any:
		out := make([]any, len(v))
		copy(out, v)
		return out
	default:
		return value
	}
}

// CloneDeep returns a deep copy of value, recursively copying nested
// map[string]any and []any containers so that mutating the result never
// affects the original. Scalar values are returned unchanged.
func CloneDeep(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, e := range v {
			out[k] = CloneDeep(e)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, e := range v {
			out[i] = CloneDeep(e)
		}
		return out
	default:
		return value
	}
}

// IsEqual reports whether a and b are deeply, structurally equal. It compares
// map[string]any values key by key, []any values element by element, and falls
// back to == for scalars. Two nil values are equal.
func IsEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for k, va := range av {
			vb, present := bv[k]
			if !present || !IsEqual(va, vb) {
				return false
			}
		}
		return true
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !IsEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	default:
		return equalScalar(a, b)
	}
}

// equalScalar compares two non-container values. It guards against comparing
// uncomparable types (which would panic with ==) by returning false.
func equalScalar(a, b any) (eq bool) {
	defer func() {
		if recover() != nil {
			eq = false
		}
	}()
	return a == b
}

// Transform is a variant of Reduce for objects. It runs each entry of m through
// iteratee, threading an accumulator that iteratee mutates in place. If
// iteratee returns false, iteration stops early. The accumulator is returned.
func Transform(m map[string]any, iteratee func(acc map[string]any, value any, key string) bool, accumulator map[string]any) map[string]any {
	if accumulator == nil {
		accumulator = make(map[string]any)
	}
	for k, v := range m {
		if !iteratee(accumulator, v, k) {
			break
		}
	}
	return accumulator
}

// ForIn iterates over the entries of m, invoking iteratee for each. If iteratee
// returns false, iteration stops early. (For plain maps ForIn and ForOwn behave
// identically; both are provided to match lodash.)
func ForIn(m map[string]any, iteratee func(value any, key string) bool) {
	for k, v := range m {
		if !iteratee(v, k) {
			break
		}
	}
}

// ForOwn iterates over the own enumerable entries of m, invoking iteratee for
// each. If iteratee returns false, iteration stops early.
func ForOwn(m map[string]any, iteratee func(value any, key string) bool) {
	for k, v := range m {
		if !iteratee(v, k) {
			break
		}
	}
}

// FindKey returns the key of the first entry of m for which predicate returns
// true, along with true. If no entry matches, it returns the zero key and
// false.
//
// Because Go map iteration order is randomized, callers that need a
// deterministic result over multiple matches should not rely on which matching
// key is returned. To keep FindKey predictable in the common case, keys are
// examined in sorted order.
func FindKey[V any](m map[string]V, predicate func(value V, key string) bool) (string, bool) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if predicate(m[k], k) {
			return k, true
		}
	}
	return "", false
}
