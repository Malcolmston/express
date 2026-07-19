// Package dotprop provides get/set/has/delete access to nested map[string]any
// structures using dotted path strings, mirroring the behavior of the npm
// "dot-prop" library, using only the Go standard library. It exposes four
// functions — Get, Has, Set, and Delete — that address a value deep inside a
// tree of maps and slices by a single string path such as "a.b.c" rather than by
// a chain of hand-written map lookups and type assertions.
//
// This is the utility you want when working with dynamically shaped data:
// configuration loaded from JSON or YAML, a decoded request body, or any
// map[string]any whose structure is known by path rather than by a Go type. It
// lets a caller read config.Get(cfg, "server.tls.enabled") or set a deeply nested
// default without first checking that every intermediate map exists, which keeps
// option-merging and template-context code short and free of nil-map panics.
//
// A path is split on unescaped dots into segments. A literal dot inside a key can
// be escaped with a backslash, so the Go string "a\\.b" (the path a\.b) refers to
// the single key "a.b" rather than to a nested a then b; a backslash itself is
// escaped the same way, and a trailing backslash is kept literally. At each step
// the resolver looks at the current node: if it is a map[string]any the segment
// is used as a string key, and if it is a []any the segment is parsed as a
// base-ten index into the slice. Numeric segments therefore index slices only
// when the current value actually is a slice; against a map the same digits are
// an ordinary string key.
//
// The four operations have well-defined behavior on the awkward inputs. Get
// returns (value, true) when the whole path resolves and (nil, false) otherwise —
// including a missing key, an out-of-range or negative slice index, a
// non-numeric index into a slice, or an attempt to descend into a scalar leaf —
// and Has is simply Get with the value discarded. Set creates intermediate
// map[string]any nodes as needed and overwrites any non-map value that sits in
// the way of the path, then returns obj so calls can be chained; it does not
// create slice nodes. Delete removes the addressed key and reports whether
// something was actually removed. A nil obj or an empty path is a no-op: Get and
// Has report absence, Set returns obj unchanged, and Delete returns false.
//
// Parity with the Node original covers the core get/set/has/delete surface and
// the dot-with-backslash-escape path grammar. The deliberate differences are
// idiomatic and reflect Go's type system: values are map[string]any and []any
// instead of arbitrary JavaScript objects, Get returns the Go-style (value, ok)
// pair rather than an optional-or-default, array indices only apply to real
// []any slices, and this port does not implement dot-prop's default-value
// argument or its escaping of bracket-style paths — paths are dot-and-backslash
// only.
package dotprop

import (
	"strconv"
	"strings"
)

// disallowedKeys mirrors dot-prop's prototype-pollution guard: a path whose
// segments include any of these keys is treated as if it does not resolve, so
// Get/Has report absence, and Set/Delete are no-ops. Upstream collapses such a
// path to empty, yielding the default value on read and no mutation on write.
var disallowedKeys = map[string]struct{}{
	"__proto__":   {},
	"prototype":   {},
	"constructor": {},
}

// hasDisallowedSegment reports whether any segment is a guarded key.
func hasDisallowedSegment(segments []string) bool {
	for _, s := range segments {
		if _, bad := disallowedKeys[s]; bad {
			return true
		}
	}
	return false
}

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
	if hasDisallowedSegment(segments) {
		return nil, false
	}
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
	if hasDisallowedSegment(segments) {
		return obj
	}
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
	if hasDisallowedSegment(segments) {
		return false
	}
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
