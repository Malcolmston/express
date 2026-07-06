// Package typeis matches Content-Type header values against a set of type
// candidates. It is a standard-library-only Go port of the npm "type-is"
// package, the content negotiation helper Express exposes as req.is and that
// body-parsing middleware uses to decide whether a request body is JSON, form
// data, or something else. The problem it solves is that a raw Content-Type
// header such as "application/json; charset=utf-8" is awkward to test directly:
// it carries parameters, is case-insensitive, and callers usually want to ask a
// convenience question like "is this JSON?" rather than compare strings.
//
// The high-level entry point is Is, which takes a concrete Content-Type value
// and one or more candidate patterns and returns the matched candidate (in its
// full, normalized form) together with a boolean. Candidates may be written in
// several convenient shapes: a full type such as "application/json"; an
// extension-style shorthand such as "json", "html", "urlencoded", or
// "multipart"; a wildcard such as "*/*", "text/*", or "*/json"; or a structured
// "+suffix" match such as "+json" that matches any subtype ending in that
// suffix. When no candidates are supplied, Is simply reports whether the value
// is a non-empty, parseable type and echoes it back.
//
// Internally each candidate is first expanded into a matchable pattern. A
// recognised shorthand is replaced by its full type from a small lookup table
// (for example "form" and "urlencoded" both become
// "application/x-www-form-urlencoded", and "multipart" becomes "multipart/*");
// a leading "+suffix" becomes "*/*+suffix"; and a bare token containing no
// slash is treated as an extension and falls back to "application/<token>". The
// concrete value being tested is normalized by stripping any ";"-delimited
// parameters and lower-casing what remains, so the charset and boundary
// parameters that real headers carry never affect the match.
//
// Match performs the actual comparison of an expected pattern against a
// concrete type. It supports "*" wildcards in either the type or the subtype
// position and understands the "type/subtype+suffix" structure: an expected
// "+json" matches "application/vnd.api+json" because the suffixes agree, while
// "*/*" matches anything. An empty value on either side never matches, and two
// non-parseable types (a missing or empty type or subtype) are treated as a
// non-match rather than an error.
//
// Parity with the Node original covers the shorthand names, wildcard forms, and
// suffix matching that Express relies on for its own body parsers. The API shape
// is idiomatic Go: Is returns the matched pattern and an ok boolean instead of
// JavaScript's string-or-false union, and the package deliberately omits helpers
// that depend on a live request object (such as reading the body length header),
// leaving those to the surrounding framework.
package typeis

import "strings"

// shorthand maps common extension-style names to full MIME types (or
// wildcards) as recognised by the npm type-is package.
var shorthand = map[string]string{
	"json":       "application/json",
	"html":       "text/html",
	"text":       "text/plain",
	"xml":        "application/xml",
	"urlencoded": "application/x-www-form-urlencoded",
	"multipart":  "multipart/*",
	"form":       "application/x-www-form-urlencoded",
	"js":         "application/javascript",
	"css":        "text/css",
}

// normalizeType strips any parameters (e.g. "; charset=utf-8") from a
// Content-Type value and lower-cases it.
func normalizeType(ct string) string {
	ct = strings.TrimSpace(ct)
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.ToLower(strings.TrimSpace(ct))
}

// expand converts a candidate (which may be a shorthand, a "+suffix", or a
// full/wildcard type) into a matchable pattern.
func expand(candidate string) string {
	c := strings.ToLower(strings.TrimSpace(candidate))
	if c == "" {
		return ""
	}
	if full, ok := shorthand[c]; ok {
		return full
	}
	// A leading "+suffix" becomes "*/*+suffix".
	if strings.HasPrefix(c, "+") {
		return "*/*" + c
	}
	// A bare token with no slash and no plus is treated as an extension
	// shorthand; if unknown, fall back to application/<token>.
	if !strings.ContainsRune(c, '/') {
		return "application/" + c
	}
	return c
}

// splitType splits a normalized "type/subtype" into its parts. The subtype is
// further split into a base and an optional "+suffix".
func splitType(t string) (typ, subtype, suffix string, ok bool) {
	i := strings.IndexByte(t, '/')
	if i < 0 {
		return "", "", "", false
	}
	typ = t[:i]
	sub := t[i+1:]
	if typ == "" || sub == "" {
		return "", "", "", false
	}
	if j := strings.LastIndexByte(sub, '+'); j >= 0 {
		return typ, sub[:j], sub[j+1:], true
	}
	return typ, sub, "", true
}

// Match reports whether the actual (concrete) Content-Type matches the
// expected pattern. The expected pattern may use "*" wildcards for the type
// and/or subtype, and a "+suffix" on the subtype. Both arguments should be
// full types; parameters are ignored.
func Match(expected, actual string) bool {
	exp := normalizeType(expected)
	act := normalizeType(actual)
	if exp == "" || act == "" {
		return false
	}
	if exp == "*/*" {
		return true
	}
	if exp == act {
		return true
	}

	et, es, esuf, eok := splitType(exp)
	at, as, asuf, aok := splitType(act)
	if !eok || !aok {
		return false
	}

	// Match main type.
	if et != "*" && et != at {
		return false
	}

	// Suffix matching: expected "+json" (subtype base empty or "*").
	if esuf != "" {
		if es != "" && es != "*" {
			// Expected has both base and suffix: require exact base and
			// suffix match, or the actual suffix equal.
			if es != as {
				return false
			}
		}
		return esuf == asuf
	}

	// No expected suffix: match subtype with wildcard support.
	if es == "*" {
		return true
	}
	return es == as && esuf == asuf
}

// Is matches a Content-Type value against one or more candidate types. Each
// candidate may be a full type ("application/json"), an extension shorthand
// ("json", "html", "urlencoded", "multipart"), a wildcard ("*/*", "text/*",
// "*/json") or a "+suffix" match ("+json"). It returns the matched candidate,
// normalized to its full type, and true. If nothing matches it returns "" and
// false.
func Is(contentType string, types ...string) (string, bool) {
	act := normalizeType(contentType)
	if act == "" {
		return "", false
	}
	if len(types) == 0 {
		return act, true
	}
	for _, candidate := range types {
		pattern := expand(candidate)
		if pattern == "" {
			continue
		}
		if Match(pattern, act) {
			return pattern, true
		}
	}
	return "", false
}
