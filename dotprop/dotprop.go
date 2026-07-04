// Package dotprop provides get/set/has/delete access to nested
// map[string]any structures using dotted path strings, mirroring the
// behavior of the npm "dot-prop" library.
//
// Paths are dot-separated (e.g. "a.b.c"). A literal dot inside a key can be
// escaped with a backslash ("a\\.b" refers to the single key "a.b"). Numeric
// path segments index into []any slices when the current value is a slice,
// otherwise they are treated as plain string map keys.
package dotprop

import (
	"strconv"
	"strings"
)

// parsePath splits a dotted path into its individual segments, honoring
// backslash escaping of the dot separator (and of the backslash itself).
func parsePath(path string) []string {
	var segments []string
	var b strings.Builder
	escaped := false
	for i := 0; i < len(path); i++ {
		c := path[i]
		if escaped {
			b.WriteByte(c)
			escaped = false
			continue
		}
		switch c {
		case '\\':
			escaped = true
		case '.':
			segments = append(segments, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	if escaped {
		// Trailing backslash is kept literally.
		b.WriteByte('\\')
	}
	segments = append(segments, b.String())
	return segments
}

// Get retrieves the value at the given dotted path. It returns the value and
// true when the path resolves, or nil and false otherwise. Numeric segments
// index into []any slices; all other segments are map keys.
func Get(obj map[string]any, path string) (any, bool) {
	if obj == nil || path == "" {
		return nil, false
	}
	segments := parsePath(path)
	var current any = obj
	for _, seg := range segments {
		switch node := current.(type) {
		case map[string]any:
			v, ok := node[seg]
			if !ok {
				return nil, false
			}
			current = v
		case []any:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(node) {
				return nil, false
			}
			current = node[idx]
		default:
			return nil, false
		}
	}
	return current, true
}

// Has reports whether the given dotted path resolves to a value in obj.
func Has(obj map[string]any, path string) bool {
	_, ok := Get(obj, path)
	return ok
}

// Set assigns value at the given dotted path, creating intermediate
// map[string]any nodes as needed. It returns obj to allow chaining. If obj is
// nil or path is empty, obj is returned unchanged. Existing non-map
// intermediate values along the path are overwritten with new maps.
func Set(obj map[string]any, path string, value any) map[string]any {
	if obj == nil || path == "" {
		return obj
	}
	segments := parsePath(path)
	current := obj
	for i := 0; i < len(segments)-1; i++ {
		seg := segments[i]
		next, ok := current[seg].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[seg] = next
		}
		current = next
	}
	current[segments[len(segments)-1]] = value
	return obj
}

// Delete removes the value at the given dotted path. It returns true if a
// value was present and removed, false otherwise.
func Delete(obj map[string]any, path string) bool {
	if obj == nil || path == "" {
		return false
	}
	segments := parsePath(path)
	current := obj
	for i := 0; i < len(segments)-1; i++ {
		next, ok := current[segments[i]].(map[string]any)
		if !ok {
			return false
		}
		current = next
	}
	last := segments[len(segments)-1]
	if _, ok := current[last]; !ok {
		return false
	}
	delete(current, last)
	return true
}
