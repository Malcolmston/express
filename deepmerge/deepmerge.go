// Package deepmerge deeply merges maps, modeled on the npm "deepmerge" package.
// It recursively combines a target map with a source map, producing a brand new
// map in which nested objects are merged key by key rather than being replaced
// wholesale. The public API is Merge for the common case, MergeWith for callers
// that need to customize how slices combine, and MergeAll for folding an
// arbitrary number of maps together left to right.
//
// Reach for deepmerge whenever you need to layer configuration or state: default
// settings overlaid by user settings, a base document patched by an override, or
// several partial fragments assembled into one. Because later values win only at
// the leaves where they actually appear, a source map that sets a single nested
// field does not clobber sibling fields that only exist in the target. This is
// the behavior that a shallow copy or a plain map assignment cannot give you.
//
// Internally the merge walks both maps in parallel. Keys present only in the
// target are cloned into the result; keys present only in the source are cloned
// in as well; keys present in both are reconciled by mergeValues. When both
// values are map[string]any they are merged recursively. When both values are
// []any they are combined by the ArrayMerge strategy (concatenation by default).
// For every other combination, including a type mismatch such as a slice in the
// target and a map in the source, the source value simply replaces the target
// value. Scalars, functions, and any non-map/non-slice values are always treated
// as opaque leaves.
//
// Merging never mutates its inputs. Every value that ends up in the result is
// deep-cloned first, so nested maps and slices in the returned map are fully
// independent of the originals: mutating the result cannot reach back and change
// target or source, and vice versa. Nil inputs are accepted and treated as empty
// maps, so Merge(nil, m) and Merge(m, nil) both yield a clone of the non-nil
// argument, and MergeAll with no arguments returns an empty, non-nil map. Because
// Go maps have no defined iteration order the result is order-independent, which
// matters only for the default array behavior where target elements precede
// source elements deterministically regardless of map traversal.
//
// Compared with the Node original, this port keeps the same default semantics of
// recursive object merge plus array concatenation and the same guarantee of no
// input mutation. The main difference is idiomatic: it operates on Go's
// map[string]any and []any rather than arbitrary JavaScript objects, it has no
// notion of a customMerge-per-key callback beyond the single ArrayMerge hook, and
// it does not special-case class instances or non-plain objects the way the
// JavaScript isMergeableObject check does. Only map[string]any is treated as a
// mergeable object; everything else is a leaf.
package deepmerge

// Options configures MergeWith.
type Options struct {
	// ArrayMerge, when non-nil, combines a target slice and a source slice into
	// the merged result. Both arguments are clones owned by the merge, so the
	// function may reuse or return them freely. When nil, slices are
	// concatenated (target then source).
	ArrayMerge func(target, source []any) []any
}

// Merge returns a new map that is the deep merge of target and source. Neither
// input is mutated. Slices are concatenated (target then source).
func Merge(target, source map[string]any) map[string]any {
	return MergeWith(target, source, Options{})
}

// MergeWith returns a new map that is the deep merge of target and source using
// opts. Neither input is mutated.
func MergeWith(target, source map[string]any, opts Options) map[string]any {
	am := opts.ArrayMerge
	if am == nil {
		am = concatArrays
	}
	return mergeMaps(target, source, am)
}

// MergeAll deeply merges any number of maps left to right and returns a new
// map. Later maps take precedence over earlier ones, slices are concatenated,
// and no input is mutated.
func MergeAll(maps ...map[string]any) map[string]any {
	result := map[string]any{}
	for _, m := range maps {
		result = Merge(result, m)
	}
	return result
}

// concatArrays is the default ArrayMerge: it appends source onto target.
func concatArrays(target, source []any) []any {
	out := make([]any, 0, len(target)+len(source))
	out = append(out, target...)
	out = append(out, source...)
	return out
}

// mergeMaps merges source into a clone of target using the array-merge strategy.
func mergeMaps(target, source map[string]any, am func([]any, []any) []any) map[string]any {
	result := make(map[string]any, len(target)+len(source))
	for k, v := range target {
		result[k] = cloneValue(v)
	}
	for k, sv := range source {
		if tv, ok := result[k]; ok {
			result[k] = mergeValues(tv, sv, am)
		} else {
			result[k] = cloneValue(sv)
		}
	}
	return result
}

// mergeValues merges a single source value into an existing (already cloned)
// target value. Maps merge recursively, slices use am, and any other type has
// the source value replace the target.
func mergeValues(t, s any, am func([]any, []any) []any) any {
	tm, tIsMap := t.(map[string]any)
	sm, sIsMap := s.(map[string]any)
	if tIsMap && sIsMap {
		return mergeMaps(tm, sm, am)
	}

	ts, tIsSlice := t.([]any)
	ss, sIsSlice := s.([]any)
	if tIsSlice && sIsSlice {
		return am(cloneSlice(ts), cloneSlice(ss))
	}

	return cloneValue(s)
}

// cloneValue deep-clones maps and slices; other values are returned as-is.
func cloneValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		return cloneMap(x)
	case []any:
		return cloneSlice(x)
	default:
		return v
	}
}

// cloneMap returns a deep copy of m.
func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = cloneValue(v)
	}
	return out
}

// cloneSlice returns a deep copy of s.
func cloneSlice(s []any) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = cloneValue(v)
	}
	return out
}
