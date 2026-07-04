// Package typeis matches Content-Type header values against a set of type
// candidates. It is a port of the npm type-is package using only the Go
// standard library.
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
