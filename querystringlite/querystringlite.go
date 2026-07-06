// Package querystringlite is a faithful port of Node.js's built-in querystring
// module for flat (non-nested) query strings. It mirrors the querystring.parse,
// querystring.stringify, querystring.escape, and querystring.unescape functions
// that Node exposes for reading and writing application/x-www-form-urlencoded
// data.
//
// Use this package when you want Node-compatible query-string handling without
// the nested bracket semantics of the qs package. It treats a query string as a
// flat multimap: every key maps to the ordered list of values it was given, so
// repeated keys such as "a=1&a=3" are preserved rather than collapsed. This is
// the right tool for straightforward form submissions and API query parameters
// where "a[b]" is just a literal key and not a request to build a nested object.
//
// Parsing splits the input on "&" and then on the first "=" in each pair,
// treats "+" as a space, and percent-decodes both keys and values. A pair with
// no "=" maps its key to a single empty-string value, and empty pairs produced
// by leading, trailing, or doubled "&" are skipped. Unlike some parsers, a
// leading "?" is not stripped, matching Node's querystring.parse, which leaves
// the "?" as part of the first key.
//
// Stringifying is the inverse: keys are emitted in sorted order for
// deterministic output, multi-valued keys expand to repeated "key=value" pairs,
// and a key with an empty value list is omitted entirely. Both keys and values
// are percent-encoded via Escape, which follows encodeURIComponent's unreserved
// set and encodes spaces as "%20" rather than "+", again matching Node.
// Unescape is deliberately lenient: a malformed "%XX" sequence is left
// untouched instead of raising an error, mirroring Node's fallback behavior.
//
// For the common case of one value per key, the ParseSingle and StringifySingle
// convenience wrappers work with map[string]string instead of
// map[string][]string. The main parity difference from Node is surface shape
// rather than behavior: Go's static typing means values are exposed as
// []string (or string) maps instead of the dynamic objects Node returns, and
// this port omits Node's configurable options such as custom separators,
// maxKeys limits, and pluggable encode/decode hooks.
package querystringlite

import (
	"sort"
	"strings"
)

// isUnreserved reports whether a byte is left untouched by encoding, matching
// the unreserved set used by JavaScript's encodeURIComponent.
func isUnreserved(b byte) bool {
	switch {
	case b >= 'A' && b <= 'Z':
		return true
	case b >= 'a' && b <= 'z':
		return true
	case b >= '0' && b <= '9':
		return true
	}
	switch b {
	case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
		return true
	}
	return false
}

const upperhex = "0123456789ABCDEF"

// Escape percent-encodes s the way Node's querystring.escape does. Spaces are
// encoded as "%20".
func Escape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
			continue
		}
		b.WriteByte('%')
		b.WriteByte(upperhex[c>>4])
		b.WriteByte(upperhex[c&0x0f])
	}
	return b.String()
}

// unhex returns the value of a hexadecimal digit and whether it was valid.
func unhex(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// Unescape decodes s the way Node's querystring.unescape does: "+" becomes a
// space and valid "%XX" sequences are percent-decoded. Malformed percent
// sequences are left as-is (lenient, like Node's fallback behavior).
func Unescape(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '+':
			b.WriteByte(' ')
		case c == '%' && i+2 < len(s):
			hi, ok1 := unhex(s[i+1])
			lo, ok2 := unhex(s[i+2])
			if ok1 && ok2 {
				b.WriteByte(hi<<4 | lo)
				i += 2
			} else {
				b.WriteByte('%')
			}
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// Parse parses a URL query string into a map of keys to their values. A key
// that appears multiple times (as in "a=1&a=3") maps to all of its values in
// order. A segment without "=" maps the key to a single empty-string value.
func Parse(s string) map[string][]string {
	result := make(map[string][]string)
	if s == "" {
		return result
	}
	// A leading "?" is not stripped by Node's querystring.parse, so we don't
	// strip it either.
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
			rawVal = ""
		}
		key := Unescape(rawKey)
		val := Unescape(rawVal)
		result[key] = append(result[key], val)
	}
	return result
}

// Stringify serializes a map of keys to values into a query string. Keys are
// emitted in sorted order for determinism; multi-valued keys are emitted as
// repeated "key=value" pairs. A key with no values is omitted.
func Stringify(v map[string][]string) string {
	if len(v) == 0 {
		return ""
	}
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	first := true
	for _, k := range keys {
		ek := Escape(k)
		vals := v[k]
		if len(vals) == 0 {
			continue
		}
		for _, val := range vals {
			if !first {
				b.WriteByte('&')
			}
			first = false
			b.WriteString(ek)
			b.WriteByte('=')
			b.WriteString(Escape(val))
		}
	}
	return b.String()
}

// ParseSingle is a convenience wrapper around Parse that keeps only the first
// value seen for each key.
func ParseSingle(s string) map[string]string {
	parsed := Parse(s)
	out := make(map[string]string, len(parsed))
	for k, vals := range parsed {
		if len(vals) > 0 {
			out[k] = vals[0]
		} else {
			out[k] = ""
		}
	}
	return out
}

// StringifySingle is a convenience wrapper around Stringify for maps with a
// single value per key.
func StringifySingle(v map[string]string) string {
	multi := make(map[string][]string, len(v))
	for k, val := range v {
		multi[k] = []string{val}
	}
	return Stringify(multi)
}
