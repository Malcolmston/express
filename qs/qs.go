// Package qs parses and serializes URL query strings with support for nested
// objects and arrays via bracket notation. It is a Go port of a subset of the
// npm module "qs".
//
// Supported forms:
//
//	a=1&b=2          -> {"a":"1","b":"2"}
//	a[b]=1&a[c]=2    -> {"a":{"b":"1","c":"2"}}
//	a[]=1&a[]=2      -> {"a":["1","2"]}
//
// Keys and values are URL-decoded. Stringify is the inverse of Parse for the
// documented cases and produces deterministic (key-sorted) output.
package qs

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// Parse parses a URL query string into a nested map. Scalar values are stored
// as strings, nested objects as map[string]any, and repeated bracket ("[]")
// entries as []any. A leading "?" is ignored. Keys and values are URL-decoded.
func Parse(s string) map[string]any {
	root := map[string]any{}

	s = strings.TrimPrefix(s, "?")
	if s == "" {
		return root
	}

	for _, pair := range strings.Split(s, "&") {
		if pair == "" {
			continue
		}

		var rawKey, rawVal string
		if eq := strings.IndexByte(pair, '='); eq >= 0 {
			rawKey = pair[:eq]
			rawVal = pair[eq+1:]
		} else {
			rawKey = pair
		}

		segments := parseSegments(rawKey)
		if len(segments) == 0 || segments[0] == "" {
			continue
		}

		value := decode(rawVal)
		base := segments[0]
		root[base] = merge(root[base], segments[1:], value)
	}

	return root
}

// merge inserts value at the location described by the remaining key segments
// within the existing node, creating maps and slices as needed.
func merge(existing any, segments []string, value any) any {
	if len(segments) == 0 {
		return value
	}

	seg := segments[0]
	rest := segments[1:]

	if seg == "" {
		var arr []any
		if e, ok := existing.([]any); ok {
			arr = e
		}
		return append(arr, merge(nil, rest, value))
	}

	m, ok := existing.(map[string]any)
	if !ok {
		m = map[string]any{}
	}
	m[seg] = merge(m[seg], rest, value)
	return m
}

// parseSegments splits a raw (still-encoded) key into its decoded path
// segments. For example "a[b][]" yields ["a", "b", ""].
func parseSegments(rawKey string) []string {
	idx := strings.IndexByte(rawKey, '[')
	if idx < 0 {
		return []string{decode(rawKey)}
	}

	segments := []string{decode(rawKey[:idx])}
	rest := rawKey[idx:]

	for len(rest) > 0 {
		if rest[0] != '[' {
			// Malformed remainder; treat it as a literal trailing segment.
			segments = append(segments, decode(rest))
			break
		}
		end := strings.IndexByte(rest, ']')
		if end < 0 {
			segments = append(segments, decode(rest))
			break
		}
		segments = append(segments, decode(rest[1:end]))
		rest = rest[end+1:]
	}

	return segments
}

// decode URL-decodes s, returning it unchanged if it is not valid encoding.
func decode(s string) string {
	if v, err := url.QueryUnescape(s); err == nil {
		return v
	}
	return s
}

// Stringify serializes a nested map into a URL query string. Nested maps become
// bracketed keys, and slices become repeated "[]" entries. Output is sorted by
// key for deterministic results. Keys and values are URL-encoded.
func Stringify(m map[string]any) string {
	var pairs []string
	// Top-level keys are emitted in sorted order.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		pairs = append(pairs, encode(url.QueryEscape(k), m[k])...)
	}
	return strings.Join(pairs, "&")
}

// encode builds the "key=value" pairs for a value under the given (already
// escaped) prefix.
func encode(prefix string, val any) []string {
	switch v := val.(type) {
	case map[string]any:
		var pairs []string
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			child := prefix + "[" + url.QueryEscape(k) + "]"
			pairs = append(pairs, encode(child, v[k])...)
		}
		return pairs
	case []any:
		var pairs []string
		child := prefix + "[]"
		for _, elem := range v {
			pairs = append(pairs, encode(child, elem)...)
		}
		return pairs
	default:
		return []string{prefix + "=" + url.QueryEscape(fmt.Sprint(v))}
	}
}
