package object_test

import (
	"fmt"

	"github.com/malcolmston/express/lodash/object"
)

// ExampleGet demonstrates reading a deeply nested value by a single
// dot-notation path. The root is the kind of any-typed tree that decoding a
// JSON document produces, mixing map[string]any and []any containers. A numeric
// path segment such as "0" indexes into a slice, so "a.b.0.c" walks a map, a
// slice, and then another map in one call. When any segment along the path is
// missing or out of range, Get returns the supplied default value instead of
// panicking. This makes Get a safe way to probe optional configuration values.
func ExampleGet() {
	root := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{"c": 42},
			},
		},
	}
	fmt.Println(object.Get(root, "a.b.0.c"))
	fmt.Println(object.Get(root, "a.b.5.c", "missing"))
	// Output:
	// 42
	// missing
}

// ExamplePick builds a new map containing only a chosen subset of keys. It is
// the map-shaped analogue of selecting columns from a record, and it never
// mutates the source map. Keys that are requested but not present in the source
// are simply skipped rather than added with a zero value. Because Go's fmt
// package prints map keys in sorted order, the output here is deterministic.
// Pick is generic, so it works over any comparable key type and any value type.
func ExamplePick() {
	m := map[string]int{"name": 1, "age": 2, "email": 3}
	picked := object.Pick(m, "name", "email")
	fmt.Println(picked)
	// Output:
	// map[email:3 name:1]
}

// ExampleOmit is the complement of Pick: it copies every entry except the ones
// whose keys are listed. Like Pick it returns a fresh map and leaves the
// original untouched. This is handy for stripping sensitive fields such as a
// password before logging or serializing a record. Listing a key that is not
// present has no effect. The printed output is deterministic because fmt sorts
// map keys.
func ExampleOmit() {
	m := map[string]int{"name": 1, "age": 2, "password": 3}
	safe := object.Omit(m, "password")
	fmt.Println(safe)
	// Output:
	// map[age:2 name:1]
}

// ExampleSet writes a value at a dot-notation path, creating the intermediate
// containers as it goes. Passing a nil root lets Set allocate the top-level
// container for you, choosing a map or a slice based on whether the first
// segment is numeric. Here the path "a.b.c" is entirely non-numeric, so nested
// maps are created at each level. Set returns the (possibly newly allocated)
// root so the caller can capture it. Reading the value back with Get confirms
// the write landed where expected.
func ExampleSet() {
	root := object.Set(nil, "a.b.c", 10)
	fmt.Println(object.Get(root, "a.b.c"))
	// Output:
	// 10
}

// ExampleMerge recursively combines source maps into a destination. Nested
// maps are merged key by key rather than replaced wholesale, so keys unique to
// either side survive. This mirrors how lodash deep-merges plain objects and is
// useful for layering configuration defaults with overrides. Merge mutates and
// returns the destination map. The nested result prints deterministically
// because fmt sorts keys at every level.
func ExampleMerge() {
	dst := map[string]any{"a": map[string]any{"x": 1}}
	src := map[string]any{"a": map[string]any{"y": 2}}
	merged := object.Merge(dst, src)
	fmt.Println(merged)
	// Output:
	// map[a:map[x:1 y:2]]
}

// ExampleIsEqual compares two values for deep structural equality. It descends
// into nested maps and slices, comparing them key by key and element by
// element, and falls back to == for scalar leaves. Two independently
// constructed trees with the same shape and contents are reported as equal.
// This avoids the pitfalls of comparing maps or slices directly, which is not
// allowed with ==. The comparison is order-independent for maps and
// order-sensitive for slices.
func ExampleIsEqual() {
	a := map[string]any{"x": []any{1, 2}}
	b := map[string]any{"x": []any{1, 2}}
	fmt.Println(object.IsEqual(a, b))
	// Output:
	// true
}

// ExampleInvert swaps keys and values, producing a new map whose keys are the
// original values. Both the key and value types must be comparable so the
// result is a valid map. When the source contains duplicate values, later
// entries overwrite earlier ones, so the inverse may be smaller than the
// original. Here the values are distinct, so the mapping is one-to-one. The
// output prints in sorted key order.
func ExampleInvert() {
	m := map[string]string{"a": "1", "b": "2"}
	fmt.Println(object.Invert(m))
	// Output:
	// map[1:a 2:b]
}

// ExampleMapValues transforms every value in a map while preserving its keys.
// The iteratee receives both the value and its key and returns the replacement
// value, which may be of a different type. This is the map-shaped counterpart
// of a slice transform and does not mutate the source. Here each integer value
// is doubled. The result prints deterministically in sorted key order.
func ExampleMapValues() {
	m := map[string]int{"a": 1, "b": 2}
	doubled := object.MapValues(m, func(v int, k string) int { return v * 2 })
	fmt.Println(doubled)
	// Output:
	// map[a:2 b:4]
}

// ExampleFindKey returns the key of the first entry satisfying a predicate.
// Because Go map iteration order is randomized, FindKey examines keys in sorted
// order to keep its result deterministic when several entries would match. It
// returns the found key together with true, or the zero key and false when
// nothing matches. The predicate receives both the value and the key. Here it
// locates the key whose value equals 2.
func ExampleFindKey() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	key, ok := object.FindKey(m, func(v int, k string) bool { return v == 2 })
	fmt.Println(key, ok)
	// Output:
	// b true
}
