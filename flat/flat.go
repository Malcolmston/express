// Package flat flattens nested maps into single-depth maps with delimited keys
// and unflattens them back, mirroring the npm "flat" library.
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
	out := make(map[string]any)
	for k, v := range m {
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
