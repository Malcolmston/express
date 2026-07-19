// Package flat flattens nested maps into single-depth maps with delimited keys
// and unflattens them back, mirroring the npm "flat" library.
//
// Flattening turns a tree such as {"a": {"b": 1}} into the single-level map
// {"a.b": 1}, and unflattening reverses the process. This is convenient when a
// nested structure must be projected onto a flat key space: environment
// variables, form fields, dot-notation config, database columns, or any store
// that only accepts string keys and scalar values.
//
// Flatten walks the input recursively. For every value that is itself a
// map[string]any it descends, joining each level's key to the accumulated
// prefix with the delimiter (default "."); any other value, including a slice
// or a struct, is stored unchanged as a leaf. Unflatten performs the inverse:
// it splits each composite key on the delimiter and rebuilds the nested maps
// segment by segment. Before rebuilding, Unflatten re-flattens any entry whose
// value is itself a non-empty nested map, merging it in under its own key; this
// matches the original library's handling of "messy" inputs and keeps siblings
// like {"a.b": {...}, "a": {...}} from clobbering one another. The delimiter is
// configurable through FlattenOpts and UnflattenOpts, but both sides must agree
// on it for a round trip to succeed.
//
// One deliberate edge case is that an empty nested map is preserved as a leaf
// rather than vanishing, so {"a": {}} flattens to {"a": {}} with the empty map
// intact; this keeps Flatten followed by Unflatten a faithful round trip for
// map-only data. An empty top-level map flattens to an empty map. There is no
// configurable depth limit: Flatten always descends to the bottom of the map
// tree, and Unflatten always rebuilds the full nesting implied by the keys.
//
// Because Go maps use string keys and carry no notion of arrays, this port
// operates purely on map[string]any and does not attempt the array handling,
// key transformation, safe/overwrite, or maxDepth options of the JavaScript
// original. Values that are not map[string]any, including slices, are treated
// as opaque leaves. Neither function returns an error; malformed keys simply
// produce whatever nesting the delimiter split implies.
package flat

import "strings"

// DefaultDelimiter is the key separator used when none is configured.
const DefaultDelimiter = "."

// FlattenOpts configures Flatten.
type FlattenOpts struct {
	// Delimiter separates nested key segments. Defaults to ".".
	Delimiter string
}

// UnflattenOpts configures Unflatten.
type UnflattenOpts struct {
	// Delimiter separates nested key segments. Defaults to ".".
	Delimiter string
}

// Flatten recursively flattens a nested map into a single-level map whose keys
// are joined by the configured delimiter (default "."). Non-map leaf values,
// including empty maps, are preserved as-is.
func Flatten(m map[string]any, opts ...FlattenOpts) map[string]any {
	delim := DefaultDelimiter
	if len(opts) > 0 && opts[0].Delimiter != "" {
		delim = opts[0].Delimiter
	}
	out := make(map[string]any)
	flatten("", m, delim, out)
	return out
}

func flatten(prefix string, m map[string]any, delim string, out map[string]any) {
	if len(m) == 0 && prefix != "" {
		// Preserve empty maps as leaves.
		out[prefix] = m
		return
	}
	for k, v := range m {
		key := k
		if prefix != "" {
			key = prefix + delim + k
		}
		if child, ok := v.(map[string]any); ok {
			flatten(key, child, delim, out)
			continue
		}
		out[key] = v
	}
}

// Unflatten expands a single-level map with delimited keys back into a nested
// map structure. It inverts Flatten for map[string]any inputs.
func Unflatten(m map[string]any, opts ...UnflattenOpts) map[string]any {
	delim := DefaultDelimiter
	if len(opts) > 0 && opts[0].Delimiter != "" {
		delim = opts[0].Delimiter
	}
	// Mirror upstream's pre-processing pass: any entry whose value is itself a
	// non-empty nested map is re-flattened and merged in under its own key
	// before the tree is rebuilt. This makes "messy" inputs (delimited keys
	// pointing at further nested maps) round-trip correctly and stops siblings
	// such as {"a.b": {...}, "a": {...}} from clobbering one another regardless
	// of map iteration order. Empty maps are left untouched as leaves.
	target := make(map[string]any, len(m))
	for k, v := range m {
		if child, ok := v.(map[string]any); ok && len(child) > 0 {
			flatten(k, child, delim, target)
			continue
		}
		target[k] = v
	}

	out := make(map[string]any)
	for k, v := range target {
		parts := strings.Split(k, delim)
		cur := out
		for i, p := range parts {
			if i == len(parts)-1 {
				cur[p] = v
				break
			}
			next, ok := cur[p].(map[string]any)
			if !ok {
				next = make(map[string]any)
				cur[p] = next
			}
			cur = next
		}
	}
	return out
}
