// Package deepmerge deeply merges maps, modeled on the npm "deepmerge" package.
//
// Merging never mutates its inputs: nested maps and slices are cloned so the
// returned map is fully independent of the target and source. Nested maps are
// merged recursively. By default, slices ([]any) are concatenated (target
// elements followed by source elements), matching deepmerge's default array
// behavior; callers can override this via MergeWith and Options.ArrayMerge.
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
