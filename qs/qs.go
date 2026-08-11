// Package qs parses and serializes URL query strings with support for nested
// objects and arrays via bracket notation. It is a Go port of the npm module
// "qs", the query-string library Express uses when its extended body/query
// parser is enabled.
//
// Where the standard library's net/url and this repository's querystringlite
// treat a query string as a flat mapping of keys to values, qs understands the
// bracket syntax that lets a single query string describe arbitrarily nested
// structures. This is useful whenever an HTML form, an Express client, or an
// AJAX request encodes an object or array into a URL, and the server needs to
// reconstruct that shape rather than a flat list of strings.
//
// The bracket notation works by reading each key as a path. A plain key such as
// "a" addresses a top-level scalar; "a[b]" nests a value under key "b" of the
// object at "a"; "a[b][c]" nests one level deeper; and an empty bracket pair
// "a[]" appends to an array. Parse walks these path segments and materializes
// them as Go values: scalars are stored as strings, nested objects as
// map[string]any, and repeated "[]" entries as []any, so the three canonical
// forms below round-trip through the natural Go types.
//
// Supported forms:
//
//	a=1&b=2          -> {"a":"1","b":"2"}
//	a[b]=1&a[c]=2    -> {"a":{"b":"1","c":"2"}}
//	a[]=1&a[]=2      -> {"a":["1","2"]}
//
// Keys and values are URL-decoded on parse and URL-encoded on stringify. Empty
// pairs are skipped and a key with no "=" is treated as having an empty-string
// value. A leading "?" is only stripped when ParseOptions.IgnoreQueryPrefix is
// set, exactly as upstream: by default "?a=b" has the key "?a".
// Percent-encoded brackets (%5B/%5D) are normalized to literal brackets across
// the whole query string before any key is split, so "a%5Bb%5D=c" nests the same
// as "a[b]=c". Following upstream's default "combine" duplicate handling, a
// repeated key gathers its values into an array ("a=b&a=c" yields
// {"a":["b","c"]}) rather than the later value overwriting the earlier one; set
// ParseOptions.Duplicates to change that. Stringify is the inverse of Parse for
// the documented cases: nested maps become bracketed keys, slices become
// repeated "[]" entries, and both the top-level keys and every nested object's
// keys are emitted in sorted order so output is deterministic and safe to
// compare in tests.
//
// # Bounded, like the original
//
// Adversarial input is bounded the way upstream qs bounds it. ParseOptions.Depth
// caps how many bracket levels are followed (five by default): once the cap is
// reached the entire remainder of the key becomes a single literal segment, so
// "a[b][c][d][e][f][g]=h" nests five levels below "a" and stores the rest under
// the literal key "[g]" rather than allocating one map per bracket pair.
// ParseOptions.ParameterLimit caps how many key/value pairs are read at all
// (1000 by default). Both limits exist because a query string is
// attacker-controlled: without them a short request body can force a large
// number of allocations.
//
// Numeric bracket segments build slices rather than maps, matching upstream:
// "a[0]=b&a[1]=c" yields {"a":["b","c"]}, index gaps are compacted away, and an
// index that reaches ParseOptions.ArrayLimit (20 by default, and tested with a
// strict "<" as upstream does) falls back to a map key so that "a[9999]=b"
// cannot be used to force a huge allocation. The same limit caps how many values
// a repeated key may gather into one slice; past it the slice becomes an
// index-keyed map. The top level is always a map, so "0=foo" stays {"0":"foo"}.
//
// # Prototype-pollution guard
//
// Upstream refuses, by default, to build any key path whose parent or bracket
// content is an own property of JavaScript's Object.prototype — "__proto__",
// "constructor", "hasOwnProperty", "toString" and the rest — and drops the whole
// pair rather than just the offending level. A Go map has no prototype chain, so
// nothing of that class is exploitable here; the guard is ported anyway so that
// the *shape* of the parsed data matches, because an application moved over from
// Express would otherwise see keys upstream would never have produced.
// ParseOptions.AllowPrototypes and ParseOptions.PlainObjects switch it off, as
// upstream's options of the same names do.
//
// # Not implemented
//
// This is a large subset of the npm qs surface, not all of it. The following
// upstream options have no equivalent here and are documented in
// express/API-DEVIATIONS.md: charset, charsetSentinel, interpretNumericEntities,
// decodeDotInKeys/encodeDotInKeys, decoder/encoder, filter, sort (this port
// always sorts), serializeDate, allowSparse, parseArrays, strictDepth,
// throwOnLimitExceeded, commaRoundTrip, a regexp delimiter, and the deprecated
// "indices" flag. Neither ParseWith nor StringifyWith has an error return, so an
// option value upstream would reject with a TypeError is treated here as the
// documented default.
package qs

import (
	"fmt"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Default limits applied by Parse, matching the npm qs defaults. Each may be
// overridden through ParseOptions.
const (
	// DefaultDepth is the maximum number of bracket levels Parse follows before
	// treating the remainder of a key as one literal segment.
	DefaultDepth = 5
	// DefaultParameterLimit is the maximum number of key/value pairs Parse reads.
	DefaultParameterLimit = 1000
	// DefaultArrayLimit is the exclusive upper bound on a numeric bracket index
	// that still produces a slice rather than a map key.
	DefaultArrayLimit = 20
	// DefaultDelimiter separates key/value pairs.
	DefaultDelimiter = "&"
)

// Duplicate-key strategies accepted by ParseOptions.Duplicates, matching the npm
// qs `duplicates` option.
const (
	// DuplicatesCombine gathers repeated values into a slice. It is the upstream
	// default and the default of ParseWith.
	DuplicatesCombine = "combine"
	// DuplicatesFirst keeps the first value seen for a key.
	DuplicatesFirst = "first"
	// DuplicatesLast keeps the last value seen for a key.
	DuplicatesLast = "last"
)

// ParseOptions controls ParseWith. The zero value is valid and reproduces the
// npm qs defaults, which is also what Parse uses.
//
// Because Go cannot distinguish an unset int field from an explicit zero, the
// numeric limits use 0 to mean "the documented default" and a negative value to
// mean "no limit" (or, for ArrayLimit, "never build slices from numeric
// indices"). This differs from the JavaScript original only in spelling:
// upstream expresses "no limit" as Infinity.
type ParseOptions struct {
	// Depth caps the number of bracket levels followed below the first key
	// segment. 0 selects DefaultDepth and a negative value removes the cap.
	// Upstream's depth: 0, which disables bracket parsing entirely so that
	// "a[b]=c" yields the literal key "a[b]", is not reachable through this
	// field; parse such input with net/url instead.
	Depth int
	// ParameterLimit caps how many key/value pairs are read. 0 selects
	// DefaultParameterLimit; a negative value removes the cap.
	ParameterLimit int
	// ArrayLimit is the exclusive upper bound on a numeric bracket index that
	// still produces a slice, and the largest number of values one key may
	// gather into a slice. 0 selects DefaultArrayLimit; a negative value
	// disables slices entirely so every numeric segment becomes a map key.
	ArrayLimit int
	// Delimiter separates key/value pairs. An empty string selects
	// DefaultDelimiter ("&"). Setting it to ";" parses ";"-separated pairs.
	Delimiter string
	// AllowDots additionally treats dotted paths as nesting, so "a.b=c" parses
	// the same as "a[b]=c". It is off by default, matching upstream.
	AllowDots bool
	// IgnoreQueryPrefix strips a single leading "?" before parsing. It is off by
	// default, matching upstream, so "?a=b" otherwise yields the key "?a".
	IgnoreQueryPrefix bool
	// PlainObjects disables the prototype-pollution guard, as upstream's option
	// of the same name does. Upstream's version also switches the returned
	// objects to a null prototype, which a Go map already has no analogue for.
	PlainObjects bool
	// AllowPrototypes disables the prototype-pollution guard, so that keys such
	// as "constructor" or "toString" become ordinary map entries instead of
	// being dropped.
	AllowPrototypes bool
	// StrictNullHandling distinguishes a key with no "=" from a key with an
	// empty value: with it set, "a" parses to a nil value and "a=" to "".
	StrictNullHandling bool
	// Comma splits a value containing "," into a slice, so "a=b,c" parses to
	// {"a":["b","c"]}.
	Comma bool
	// AllowEmptyArrays makes "a[]=" parse to an empty slice rather than a
	// one-element slice holding "".
	AllowEmptyArrays bool
	// Duplicates selects what happens when a key is seen more than once: one of
	// DuplicatesCombine (the default), DuplicatesFirst or DuplicatesLast. An
	// empty string, or any unrecognized value, selects DuplicatesCombine.
	Duplicates string
}

// resolved returns opts with every zero-valued field replaced by its default and
// every "no limit" turned into a concrete bound, so the parser can compare
// against plain ints.
func (o ParseOptions) resolved() ParseOptions {
	switch {
	case o.Depth == 0:
		o.Depth = DefaultDepth
	case o.Depth < 0:
		o.Depth = math.MaxInt
	}
	switch {
	case o.ParameterLimit == 0:
		o.ParameterLimit = DefaultParameterLimit
	case o.ParameterLimit < 0:
		o.ParameterLimit = math.MaxInt
	}
	// ArrayLimit keeps its sign: a negative bound is meaningful, because every
	// comparison against it is upstream's own "index < arrayLimit".
	if o.ArrayLimit == 0 {
		o.ArrayLimit = DefaultArrayLimit
	}
	if o.Delimiter == "" {
		o.Delimiter = DefaultDelimiter
	}
	switch o.Duplicates {
	case DuplicatesFirst, DuplicatesLast:
	default:
		o.Duplicates = DuplicatesCombine
	}
	return o
}

// Parse parses a URL query string into a nested map using the npm qs defaults:
// depth 5, parameter limit 1000, array limit 20, and "&" as the delimiter.
// Scalar values are stored as strings, nested objects as map[string]any, and
// bracket ("[]") or numeric-index entries as []any. Keys and values are
// URL-decoded. Parse never panics and always returns a non-nil map.
func Parse(s string) map[string]any {
	return ParseWith(s, ParseOptions{})
}

// ParseWith parses a URL query string into a nested map using opts. See
// ParseOptions for the meaning of each field and for how its zero value maps
// onto the upstream defaults.
func ParseWith(s string, opts ParseOptions) map[string]any {
	o := opts.resolved()
	if o.IgnoreQueryPrefix {
		s = strings.TrimPrefix(s, "?")
	}
	if s == "" {
		return map[string]any{}
	}

	order, values := parseValues(s, o)

	var root any = map[string]any{}
	for _, key := range order {
		node, ok := parseKeys(key, values[key], o)
		if !ok {
			continue
		}
		root = mergeValues(root, node, o)
	}

	// Drop the holes sparse numeric indices left behind, as upstream's
	// utils.compact does.
	if m, ok := asMap(compactValue(root)); ok {
		return m
	}
	return map[string]any{}
}

// ------------------------------------------------------------- phase 1: pairs

// parseValues splits the query string into decoded key/value pairs, applying the
// parameter limit and the duplicate strategy. It returns the keys in the order
// they were first seen, because the merge that follows is order-sensitive, and
// the values indexed by key.
func parseValues(s string, o ParseOptions) ([]string, map[string]any) {
	// Match upstream qs: percent-encoded brackets are normalized to literal
	// brackets over the whole string, before it is even split, so encoded and
	// unencoded bracket notation parse identically.
	s = normalizeBrackets(s)

	parts := strings.Split(s, o.Delimiter)
	if len(parts) > o.ParameterLimit {
		// Upstream truncates with String.prototype.split's limit argument rather
		// than reporting an error, so surplus pairs are silently dropped.
		parts = parts[:o.ParameterLimit]
	}

	order := make([]string, 0, len(parts))
	values := make(map[string]any, len(parts))

	for _, part := range parts {
		// Locate the value separator the way upstream qs does: prefer the "]="
		// that terminates a bracket group so keys may contain '=' (e.g.
		// "a[<=>]==23" splits into key "a[<=>]" and value "=23"), falling back to
		// the first '=' otherwise.
		pos := strings.IndexByte(part, '=')
		if b := strings.Index(part, "]="); b >= 0 {
			pos = b + 1
		}

		var key string
		var val any
		if pos < 0 {
			key = decode(part)
			if o.StrictNullHandling {
				val = nil
			} else {
				val = ""
			}
		} else {
			key = decode(part[:pos])
			val = parseArrayValue(part[pos+1:], o)
		}

		// An explicit "[]=" push whose value comma-split into a slice nests one
		// level deeper, so the whole slice becomes a single element.
		if strings.Contains(part, "[]=") {
			if arr, isArr := val.([]any); isArr {
				val = []any{arr}
			}
		}
		if arr, isArr := val.([]any); isArr && o.Comma && len(arr) > o.ArrayLimit {
			val = combineValues([]any{}, arr, o)
		}

		prev, existing := values[key]
		switch {
		case existing && (o.Duplicates == DuplicatesCombine || strings.Contains(part, "[]=")):
			values[key] = combineValues(prev, val, o)
		case !existing:
			order = append(order, key)
			values[key] = val
		case o.Duplicates == DuplicatesLast:
			values[key] = val
		}
	}

	return order, values
}

// parseArrayValue decodes a raw value, splitting it on "," when Comma is set.
func parseArrayValue(raw string, o ParseOptions) any {
	if o.Comma && strings.Contains(raw, ",") {
		parts := strings.Split(raw, ",")
		out := make([]any, len(parts))
		for i, p := range parts {
			out[i] = decode(p)
		}
		return out
	}
	return decode(raw)
}

// -------------------------------------------------------------- phase 2: keys

// prototypeProps are the own property names of JavaScript's Object.prototype,
// obtained from Object.getOwnPropertyNames(Object.prototype). Upstream drops any
// key path that would address one of them, and the port matches so that the
// parsed shape agrees; see the package comment for why a Go map is not at risk.
var prototypeProps = map[string]bool{
	"__defineGetter__":     true,
	"__defineSetter__":     true,
	"__lookupGetter__":     true,
	"__lookupSetter__":     true,
	"__proto__":            true,
	"constructor":          true,
	"hasOwnProperty":       true,
	"isPrototypeOf":        true,
	"propertyIsEnumerable": true,
	"toLocaleString":       true,
	"toString":             true,
	"valueOf":              true,
}

// parseKeys turns one decoded key into the nested value it describes. The second
// result is false when the key must be dropped entirely: an empty key, or one the
// prototype guard rejects.
func parseKeys(key string, val any, o ParseOptions) (any, bool) {
	if key == "" {
		return nil, false
	}
	segments, ok := splitKeySegments(key, o)
	if !ok {
		return nil, false
	}
	return buildObject(segments, val, o), true
}

// splitKeySegments splits a key such as "a[b][c[]]" into ["a", "[b]", "[c[]]"],
// keeping the brackets on every child segment the way upstream's own splitter
// does. Bracket groups are balanced rather than split at the first ']', an
// unterminated group is kept verbatim as one literal segment, and at most
// o.Depth groups are consumed — past the cap the whole unconsumed remainder
// becomes a single literal segment, which is what keeps a hostile key from
// allocating one map per bracket pair.
//
// It returns false when the prototype guard rejects the key.
func splitKeySegments(key string, o ParseOptions) ([]string, bool) {
	if o.AllowDots {
		key = dotsToBrackets(key)
	}
	guard := !o.PlainObjects && !o.AllowPrototypes

	var segments []string
	first := strings.IndexByte(key, '[')
	parent := key
	if first >= 0 {
		parent = key[:first]
	}
	if parent != "" {
		if guard && prototypeProps[parent] {
			return nil, false
		}
		segments = append(segments, parent)
	}

	open := first
	collected := 0
	for open >= 0 && collected < o.Depth {
		// Balance nested '[' and ']' inside this group with a level counter.
		level := 1
		end := -1
		for i := open + 1; i < len(key) && end < 0; i++ {
			switch key[i] {
			case '[':
				level++
			case ']':
				level--
				if level == 0 {
					end = i
				}
			}
		}
		if end < 0 {
			// Unterminated group: wrap the raw remainder in one bracket pair so
			// it stays a single literal segment. No missing ']' is inferred.
			segments = append(segments, "["+key[open:]+"]")
			return segments, true
		}
		if guard && prototypeProps[key[open+1:end]] {
			return nil, false
		}
		segments = append(segments, key[open:end+1])
		collected++

		if next := strings.IndexByte(key[end+1:], '['); next >= 0 {
			open = end + 1 + next
		} else {
			open = -1
		}
	}
	if open >= 0 {
		// Depth cap reached with brackets left over.
		segments = append(segments, "["+key[open:]+"]")
	}
	return segments, true
}

// buildObject materializes a segment chain from the leaf upwards, deciding at
// each level whether the segment addresses a slice element or a map key.
func buildObject(chain []string, val any, o ParseOptions) any {
	leaf := val
	for i := len(chain) - 1; i >= 0; i-- {
		root := chain[i]

		if root == "[]" {
			switch {
			case isOverflowMap(leaf):
				// Already an index-keyed object because a limit was exceeded;
				// keep it rather than nesting it inside a slice.
			case o.AllowEmptyArrays && (leaf == "" || (o.StrictNullHandling && leaf == nil)):
				leaf = []any{}
			default:
				leaf = combineValues([]any{}, leaf, o)
			}
			continue
		}

		clean := root
		if len(root) >= 2 && root[0] == '[' && root[len(root)-1] == ']' {
			clean = root[1 : len(root)-1]
		}
		idx, isIndex := arrayIndex(clean)
		// Only a bracketed segment can be an index: upstream requires
		// root !== decodedRoot, so the top-level "0" of "0=foo" is a map key.
		isIndex = isIndex && root != clean

		switch {
		case isIndex && idx < o.ArrayLimit:
			slots := make([]any, idx+1)
			for j := range slots {
				slots[j] = undefined
			}
			slots[idx] = leaf
			leaf = slots
		case isIndex:
			// Past the array limit the index becomes an ordinary map key, so a
			// short input cannot force a large allocation.
			leaf = map[string]any{strconv.Itoa(idx): leaf}
		case clean != "__proto__":
			leaf = map[string]any{clean: leaf}
		default:
			// Upstream silently refuses to write the literal __proto__ property
			// even when the guard has been switched off.
			leaf = map[string]any{}
		}
	}
	return leaf
}

// dotsToBrackets rewrites the dotted portions of a key into bracket notation,
// mirroring upstream's allowDots transform /\.([^.[]+)/g -> "[$1]".
func dotsToBrackets(key string) string {
	var b strings.Builder
	b.Grow(len(key) + 8)
	i := 0
	for i < len(key) {
		if key[i] != '.' {
			b.WriteByte(key[i])
			i++
			continue
		}
		j := i + 1
		for j < len(key) && key[j] != '.' && key[j] != '[' {
			j++
		}
		if j == i+1 {
			// A '.' with nothing after it is literal.
			b.WriteByte('.')
			i++
			continue
		}
		b.WriteByte('[')
		b.WriteString(key[i+1 : j])
		b.WriteByte(']')
		i = j
	}
	return b.String()
}

// ------------------------------------------------------------------- merging

// undefinedValue marks a hole in a slice, standing in for JavaScript's
// undefined. Holes appear when a sparse numeric index skips slots, and
// compactValue removes them at the end of a parse, exactly as upstream's
// utils.compact does.
type undefinedValue struct{}

var undefined any = undefinedValue{}

// asMap unwraps any of the map shapes a parsed node can take.
func asMap(v any) (map[string]any, bool) {
	m, ok := v.(map[string]any)
	return m, ok && m != nil
}

func isObjectLike(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return true
	}
	return false
}

// falsy reports whether v is one of the values JavaScript would treat as false.
// Upstream's merge returns its target untouched for a falsy source, so an empty
// value cannot overwrite a structure already built at the same path.
func falsy(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return t == ""
	case bool:
		return !t
	case undefinedValue:
		return true
	}
	return false
}

// isOverflowMap reports whether m looks like the index-keyed object upstream
// produces when an array limit is exceeded: a non-empty map whose every key is a
// canonical non-negative integer. Upstream tracks that fact in a side channel
// keyed by object identity, which Go cannot do for a map; the shape is an
// accurate stand-in because such objects are only ever built from a slice.
func isOverflowMap(v any) bool {
	m, ok := asMap(v)
	if !ok || len(m) == 0 {
		return false
	}
	for k := range m {
		if _, isIdx := arrayIndex(k); !isIdx {
			return false
		}
	}
	return true
}

// maxIndex returns the largest canonical numeric key of m, or -1 if it has none.
func maxIndex(m map[string]any) int {
	max := -1
	for k := range m {
		if n, ok := arrayIndex(k); ok && n > max {
			max = n
		}
	}
	return max
}

// arrayToObject converts a slice into the index-keyed map upstream falls back to
// past an array limit, skipping holes.
func arrayToObject(arr []any) map[string]any {
	m := make(map[string]any, len(arr))
	for i, v := range arr {
		if v == undefined {
			continue
		}
		m[strconv.Itoa(i)] = v
	}
	return m
}

// concatValues is JavaScript's [].concat(a, b): a slice contributes its elements,
// anything else contributes itself.
func concatValues(a, b any) []any {
	out := []any{}
	for _, v := range []any{a, b} {
		if arr, ok := v.([]any); ok {
			out = append(out, arr...)
		} else {
			out = append(out, v)
		}
	}
	return out
}

// combineValues appends b to a, falling back to an index-keyed map once the
// result would exceed the array limit. That fallback is upstream's, and it is
// what stops a repeated key from growing an unbounded slice.
func combineValues(a, b any, o ParseOptions) any {
	if m, ok := asMap(a); ok && isOverflowMap(a) {
		m[strconv.Itoa(maxIndex(m)+1)] = b
		return m
	}
	result := concatValues(a, b)
	if len(result) > o.ArrayLimit {
		return arrayToObject(result)
	}
	return result
}

// mergeValues is a port of upstream's utils.merge, with strictMerge left at its
// default: an object and a scalar at the same path are kept side by side rather
// than the scalar becoming a key of the object.
func mergeValues(target, source any, o ParseOptions) any {
	if falsy(source) {
		return target
	}

	// A scalar source.
	if !isObjectLike(source) {
		if arr, ok := target.([]any); ok {
			if len(arr) >= o.ArrayLimit {
				return arrayToObject(append(arr, source))
			}
			return append(arr, source)
		}
		if m, ok := asMap(target); ok {
			if isOverflowMap(target) {
				m[strconv.Itoa(maxIndex(m)+1)] = source
				return m
			}
			return []any{target, source}
		}
		return []any{target, source}
	}

	// A structural source merged into a missing or scalar target.
	if !isObjectLike(target) || falsy(target) {
		if sm, ok := asMap(source); ok && isOverflowMap(source) {
			shifted := map[string]any{"0": target}
			for k, v := range sm {
				n, _ := arrayIndex(k)
				shifted[strconv.Itoa(n+1)] = v
			}
			return shifted
		}
		combined := concatValues(target, source)
		if len(combined) > o.ArrayLimit {
			return arrayToObject(combined)
		}
		return combined
	}

	ta, targetIsArray := target.([]any)
	sa, sourceIsArray := source.([]any)

	if targetIsArray && sourceIsArray {
		for i, item := range sa {
			if item == undefined {
				continue
			}
			if i < len(ta) && ta[i] != undefined {
				if isObjectLike(ta[i]) && isObjectLike(item) {
					ta[i] = mergeValues(ta[i], item, o)
				} else {
					ta = append(ta, item)
				}
				continue
			}
			for len(ta) <= i {
				ta = append(ta, undefined)
			}
			ta[i] = item
		}
		if len(ta) > o.ArrayLimit {
			return arrayToObject(ta)
		}
		return ta
	}

	// Everything else merges key by key into a map. A slice target facing a map
	// source becomes an index-keyed map first, as upstream's arrayToObject does.
	acc, _ := asMap(target)
	if targetIsArray {
		acc = arrayToObject(ta)
	}
	for _, k := range sourceKeys(source) {
		v := sourceValue(source, k)
		if _, exists := acc[k]; exists {
			acc[k] = mergeValues(acc[k], v, o)
		} else {
			acc[k] = v
		}
	}
	return acc
}

// sourceKeys lists the keys of a structural value in a deterministic order:
// index order for a slice, and sorted order for a map, because a Go map has no
// insertion order to follow. Nodes built by buildObject carry a single key, so
// the choice is only visible for the index-keyed maps an exceeded limit
// produces, where sorted order and index order agree numerically.
func sourceKeys(source any) []string {
	if arr, ok := source.([]any); ok {
		keys := make([]string, 0, len(arr))
		for i, v := range arr {
			if v == undefined {
				continue
			}
			keys = append(keys, strconv.Itoa(i))
		}
		return keys
	}
	m, _ := asMap(source)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		ni, iok := arrayIndex(keys[i])
		nj, jok := arrayIndex(keys[j])
		if iok && jok {
			return ni < nj
		}
		return keys[i] < keys[j]
	})
	return keys
}

func sourceValue(source any, key string) any {
	if arr, ok := source.([]any); ok {
		if n, isIdx := arrayIndex(key); isIdx && n < len(arr) {
			return arr[n]
		}
		return nil
	}
	m, _ := asMap(source)
	return m[key]
}

// compactValue removes the holes sparse numeric indices left in slices, and the
// undefined-valued map entries JSON would drop, matching utils.compact.
func compactValue(v any) any {
	switch t := v.(type) {
	case []any:
		out := make([]any, 0, len(t))
		for _, e := range t {
			if e == undefined {
				continue
			}
			out = append(out, compactValue(e))
		}
		return out
	case map[string]any:
		for k, e := range t {
			if e == undefined {
				delete(t, k)
				continue
			}
			t[k] = compactValue(e)
		}
		return t
	default:
		return v
	}
}

// arrayIndex reports whether s is the canonical decimal form of a non-negative
// integer, matching upstream's String(parseInt(s, 10)) === s test. "01" and "+1"
// are therefore ordinary object keys, not indices.
func arrayIndex(s string) (int, bool) {
	if s == "" || (len(s) > 1 && s[0] == '0') {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// normalizeBrackets replaces percent-encoded square brackets (%5B / %5D, in any
// letter case) with their literal forms, matching upstream qs, which performs
// this substitution over the whole query string before splitting keys so that
// encoded and unencoded bracket notation parse identically.
func normalizeBrackets(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	s = strings.ReplaceAll(s, "%5B", "[")
	s = strings.ReplaceAll(s, "%5b", "[")
	s = strings.ReplaceAll(s, "%5D", "]")
	s = strings.ReplaceAll(s, "%5d", "]")
	return s
}

// decode URL-decodes s the way upstream's utils.decode does: '+' becomes a space
// first, and if what remains is not valid percent-encoded UTF-8 the plus-replaced
// text is returned as-is rather than reported as an error. decodeURIComponent
// rejects malformed UTF-8 too, so a truncated sequence such as "%E2%9C" survives
// verbatim instead of decoding to replacement runes.
func decode(s string) string {
	withoutPlus := strings.ReplaceAll(s, "+", " ")
	v, err := url.QueryUnescape(withoutPlus)
	if err != nil || !utf8.ValidString(v) {
		return withoutPlus
	}
	return v
}

// Array serialization styles accepted by StringifyOptions.ArrayFormat, matching
// the npm qs `arrayFormat` option.
const (
	// ArrayFormatIndices writes "a[0]=b&a[1]=c". It is the upstream default and
	// the default of StringifyWith.
	ArrayFormatIndices = "indices"
	// ArrayFormatBrackets writes "a[]=b&a[]=c". It is what Stringify emits.
	ArrayFormatBrackets = "brackets"
	// ArrayFormatRepeat writes "a=b&a=c".
	ArrayFormatRepeat = "repeat"
	// ArrayFormatComma writes "a=b,c".
	ArrayFormatComma = "comma"
)

// Percent-encoding styles accepted by StringifyOptions.Format.
const (
	// FormatRFC3986 percent-encodes a space as "%20". It is the upstream default
	// and the default of StringifyWith.
	FormatRFC3986 = "RFC3986"
	// FormatRFC1738 encodes a space as "+" and leaves parentheses unescaped, the
	// application/x-www-form-urlencoded style. It is what Stringify emits.
	FormatRFC1738 = "RFC1738"
)

// StringifyOptions controls StringifyWith. Its zero value reproduces the npm qs
// defaults: arrayFormat "indices", RFC3986 encoding, "&" as the delimiter, and
// no leading "?".
type StringifyOptions struct {
	// ArrayFormat selects how slices are keyed. An empty string selects
	// ArrayFormatIndices.
	ArrayFormat string
	// Format selects the percent-encoding style. An empty string, or any
	// unrecognized value, selects FormatRFC3986.
	Format string
	// Delimiter separates key/value pairs. An empty string selects "&".
	Delimiter string
	// AddQueryPrefix prepends "?" to a non-empty result.
	AddQueryPrefix bool
	// AllowDots writes nested keys as "a.b=c" instead of "a[b]=c".
	AllowDots bool
	// SkipEncoding emits keys and values verbatim, matching upstream's
	// encode: false. Only use it for output that is not going into a URL.
	SkipEncoding bool
	// EncodeValuesOnly encodes values but leaves keys verbatim, so brackets stay
	// readable: "a[b]=c%20d".
	EncodeValuesOnly bool
	// SkipNulls omits a nil value entirely instead of writing "key=".
	SkipNulls bool
	// StrictNullHandling writes a nil value as a bare key with no "=", the
	// inverse of ParseOptions.StrictNullHandling.
	StrictNullHandling bool
	// AllowEmptyArrays writes an empty slice as "key[]" instead of omitting it.
	// Upstream emits that marker without percent-encoding its brackets.
	AllowEmptyArrays bool
}

// Stringify serializes a nested map into a URL query string. Nested maps become
// bracketed keys, and slices become repeated "[]" entries. Output is sorted by
// key for deterministic results. Keys and values are URL-encoded.
//
// Stringify deliberately keeps this port's historical output shape, which
// differs from the npm original in two ways: arrays use ArrayFormatBrackets
// rather than upstream's ArrayFormatIndices, and encoding is
// application/x-www-form-urlencoded (space becomes "+", brackets stay literal)
// rather than RFC3986. Use StringifyWith, whose zero-value options reproduce the
// upstream defaults exactly, when byte-identical qs output matters.
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
		return []string{prefix + "=" + url.QueryEscape(scalarString(v))}
	}
}

// StringifyWith serializes a nested map into a URL query string using opts.
// With the zero-value options it reproduces the npm qs defaults byte for byte:
// nested keys are bracketed and percent-encoded ("a%5Bb%5D=c"), slices use
// numeric indices ("a%5B0%5D=b"), and a space encodes as "%20".
//
// Keys are emitted in sorted order at every level so that output is
// deterministic and safe to compare in tests; upstream instead follows
// JavaScript object insertion order, which Go maps do not have.
func StringifyWith(m map[string]any, opts StringifyOptions) string {
	if opts.ArrayFormat == "" {
		opts.ArrayFormat = ArrayFormatIndices
	}
	if opts.Format != FormatRFC1738 {
		opts.Format = FormatRFC3986
	}
	if opts.Delimiter == "" {
		opts.Delimiter = DefaultDelimiter
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var pairs []string
	for _, k := range keys {
		if opts.SkipNulls && m[k] == nil {
			continue
		}
		pairs = append(pairs, encodeWith(k, m[k], opts)...)
	}
	out := strings.Join(pairs, opts.Delimiter)
	if opts.AddQueryPrefix && out != "" {
		return "?" + out
	}
	return out
}

// encodeWith builds the "key=value" pairs for val under prefix, which is the
// *unencoded* key path. Upstream builds the whole prefix first and encodes it
// once at the leaf, and doing the same here is what makes EncodeValuesOnly and
// AllowEmptyArrays — both of which need the raw path — come out identical.
func encodeWith(prefix string, val any, opts StringifyOptions) []string {
	switch v := val.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var pairs []string
		for _, k := range keys {
			if opts.SkipNulls && v[k] == nil {
				continue
			}
			var child string
			if opts.AllowDots {
				child = prefix + "." + k
			} else {
				child = prefix + "[" + k + "]"
			}
			pairs = append(pairs, encodeWith(child, v[k], opts)...)
		}
		return pairs

	case []any:
		if len(v) == 0 {
			if opts.AllowEmptyArrays {
				// Upstream emits this marker from the raw prefix, so its
				// brackets are not percent-encoded even when encoding is on.
				return []string{prefix + "[]"}
			}
			return nil
		}
		if opts.ArrayFormat == ArrayFormatComma {
			parts := make([]string, len(v))
			for i, e := range v {
				parts[i] = scalarString(e)
			}
			return []string{encodeKey(prefix, opts) + "=" + encodeValue(strings.Join(parts, ","), opts)}
		}
		var pairs []string
		for i, elem := range v {
			var child string
			switch opts.ArrayFormat {
			case ArrayFormatBrackets:
				child = prefix + "[]"
			case ArrayFormatRepeat:
				child = prefix
			default: // ArrayFormatIndices
				child = prefix + "[" + strconv.Itoa(i) + "]"
			}
			pairs = append(pairs, encodeWith(child, elem, opts)...)
		}
		return pairs

	case nil:
		if opts.StrictNullHandling {
			// A bare key, with no "=", so that a parse with the matching option
			// reads it back as nil rather than "".
			return []string{encodeKey(prefix, opts)}
		}
		return []string{encodeKey(prefix, opts) + "="}

	default:
		return []string{encodeKey(prefix, opts) + "=" + encodeValue(scalarString(v), opts)}
	}
}

// scalarString renders a leaf the way JavaScript's String() would, so that a
// whole float prints as "1" rather than "1e+00".
func scalarString(v any) string {
	if f, ok := v.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	return fmt.Sprint(v)
}

// encodeKey percent-encodes a key path unless encoding is off entirely or
// restricted to values.
func encodeKey(prefix string, opts StringifyOptions) string {
	if opts.SkipEncoding || opts.EncodeValuesOnly {
		return prefix
	}
	return encodeComponent(prefix, opts)
}

func encodeValue(s string, opts StringifyOptions) string {
	if opts.SkipEncoding {
		return s
	}
	return encodeComponent(s, opts)
}

// encodeComponent percent-encodes s in the style selected by opts.Format,
// mirroring upstream's utils.encode: only the unreserved set
// (A-Z a-z 0-9 - . _ ~) survives, hex digits are upper case, and RFC1738
// additionally writes a space as "+" and leaves parentheses alone.
func encodeComponent(s string, opts StringifyOptions) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '-' || c == '.' || c == '_' || c == '~' ||
			(c >= '0' && c <= '9') || (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z'):
			b.WriteByte(c)
		case opts.Format == FormatRFC1738 && (c == ' '):
			b.WriteByte('+')
		case opts.Format == FormatRFC1738 && (c == '(' || c == ')'):
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		}
	}
	return b.String()
}
