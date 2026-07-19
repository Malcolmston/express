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
// and one or more candidate patterns and returns the matched candidate together
// with a boolean. Candidates may be written in several convenient shapes: a full
// type such as "application/json"; an extension-style shorthand such as "json",
// "html", or "png"; the specials "urlencoded" and "multipart"; a wildcard such
// as "*/*", "text/*", or "*/json"; or a structured "+suffix" match such as
// "+json" that matches any subtype ending in that suffix. When no candidates are
// supplied, Is reports whether the value is a non-empty, valid type and echoes
// it back.
//
// Following the Node original, Is's returned string mirrors the matched
// candidate: when the matching candidate contains a "*" wildcard or begins with
// "+", Is returns the concrete (normalized) value that was tested; otherwise it
// returns the candidate exactly as it was supplied (for example "json",
// "urlencoded", or "multipart"). If nothing matches it returns "" and false.
//
// Internally each candidate is first expanded by Normalize into a matchable
// pattern: the specials "urlencoded" and "multipart" expand to their full
// forms; a leading "+suffix" becomes "*/*+suffix"; a bare extension token (no
// slash) is resolved through a small MIME table, and an unrecognised extension
// is treated as no match (matching the upstream mime.lookup returning false);
// everything containing a slash is passed through unchanged. The concrete value
// being tested is normalized by NormalizeType, which strips any ";"-delimited
// parameters, lower-cases what remains, and validates it as a real
// "type/subtype" (so a bogus value such as "bogus" or "text/html**" never
// matches anything).
//
// Match performs the actual comparison of an expected pattern against a
// concrete type. It supports "*" wildcards in either the type or the subtype
// position and understands the "*+suffix" structure: an expected "*/*+xml"
// matches "text/html+xml" because the suffixes agree, while "*/*" matches any
// valid type/subtype. An empty (false) expected value never matches, and a
// value that is not exactly "type/subtype" is treated as a non-match.
//
// The API shape is idiomatic Go: Is and Normalize return an ok boolean instead
// of JavaScript's string-or-false union, and the package deliberately omits the
// helpers that depend on a live request object (typeis.hasBody and the
// request-object form of typeis), leaving those to the surrounding framework.
package typeis

import (
	"regexp"
	"strings"
)

// mimeExt maps extension-style names to full MIME types, standing in for the
// npm mime-types database that the original type-is consults via mime.lookup.
// An extension that is not present here resolves to no match, exactly as an
// unknown extension yields false from mime.lookup upstream.
var mimeExt = map[string]string{
	"json": "application/json",
	"html": "text/html",
	"htm":  "text/html",
	"xml":  "application/xml",
	"text": "text/plain",
	"txt":  "text/plain",
	"js":   "application/javascript",
	"css":  "text/css",
	"png":  "image/png",
	"jpeg": "image/jpeg",
	"jpg":  "image/jpeg",
	"gif":  "image/gif",
}

// mediaTypeRe validates a normalized, lower-cased "type/subtype" value. The
// character classes mirror media-typer's grammar: the type name allows no "."
// or "+", the subtype name additionally allows "." and "+" (for suffixes), and
// neither allows "*". This is what makes values like "bogus" (no slash) or
// "text/html**" (invalid character) fail to normalize.
var mediaTypeRe = regexp.MustCompile(`^[a-z0-9][a-z0-9!#$&^_-]*/[a-z0-9][a-z0-9!#$&^_.+-]*$`)

// lookupExtension resolves a bare extension token (with or without a leading
// dot) to a full MIME type, returning "" when the extension is unknown.
func lookupExtension(t string) string {
	e := strings.ToLower(t)
	e = strings.TrimPrefix(e, ".")
	if m, ok := mimeExt[e]; ok {
		return m
	}
	return ""
}

// NormalizeType strips any parameters (e.g. "; charset=utf-8") from a
// Content-Type value, lower-cases it, and validates it as a "type/subtype"
// media type. It returns "" for an empty or invalid value.
func NormalizeType(value string) string {
	if value == "" {
		return ""
	}
	ct := value
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	ct = strings.ToLower(strings.TrimSpace(ct))
	if !mediaTypeRe.MatchString(ct) {
		return ""
	}
	return ct
}

// normalize expands a candidate into a matchable pattern, mirroring the npm
// type-is normalize(): the specials "urlencoded" and "multipart" expand to full
// forms, a leading "+suffix" becomes "*/*+suffix", a bare extension is resolved
// through the MIME table (unknown => ""), and anything with a slash passes
// through unchanged.
func normalize(t string) string {
	switch t {
	case "urlencoded":
		return "application/x-www-form-urlencoded"
	case "multipart":
		return "multipart/*"
	}
	if len(t) > 0 && t[0] == '+' {
		return "*/*" + t
	}
	if !strings.Contains(t, "/") {
		return lookupExtension(t)
	}
	return t
}

// Normalize expands a single candidate type into its full, matchable form. It
// returns the expanded pattern and true, or "" and false when the candidate is
// an unknown extension. It is the exported form of the npm typeis.normalize.
func Normalize(t string) (string, bool) {
	n := normalize(t)
	if n == "" {
		return "", false
	}
	return n, true
}

// mimeMatch reports whether the expected pattern matches the actual concrete
// type. It is a faithful port of the npm type-is mimeMatch: "*" wildcards are
// supported in the type and subtype positions, and a subtype beginning with
// "*+" matches any subtype sharing the same suffix.
func mimeMatch(expected, actual string) bool {
	// invalid type
	if expected == "" {
		return false
	}

	actualParts := strings.Split(actual, "/")
	expectedParts := strings.Split(expected, "/")

	// invalid format
	if len(actualParts) != 2 || len(expectedParts) != 2 {
		return false
	}

	// validate type
	if expectedParts[0] != "*" && expectedParts[0] != actualParts[0] {
		return false
	}

	es := expectedParts[1]
	as := actualParts[1]

	// validate suffix wildcard: subtype starts with "*+"
	if strings.HasPrefix(es, "*+") {
		// expectedParts[1].length <= actualParts[1].length + 1 &&
		//   expectedParts[1].slice(1) === actualParts[1].slice(1 - expectedParts[1].length)
		if len(es) > len(as)+1 {
			return false
		}
		esSuffix := es[1:]
		n := len(es) - 1
		if n > len(as) {
			return false
		}
		return esSuffix == as[len(as)-n:]
	}

	// validate subtype
	if es != "*" && es != as {
		return false
	}

	return true
}

// Match reports whether the expected mime type matches the actual mime type,
// with "*" wildcard and "*+suffix" support. It is the exported form of the npm
// typeis.match: both arguments are expected to be concrete "type/subtype"
// strings (the expected side may use wildcards); no parameter stripping is
// performed. An empty expected value never matches.
func Match(expected, actual string) bool {
	return mimeMatch(expected, actual)
}

// Is matches a Content-Type value against one or more candidate types. Each
// candidate may be a full type ("application/json"), an extension shorthand
// ("json", "png"), a special ("urlencoded", "multipart"), a wildcard ("*/*",
// "text/*", "*/json") or a "+suffix" match ("+json"). It returns the matched
// candidate and true, or "" and false when nothing matches. When no candidates
// are supplied it returns the normalized value and true, or "" and false if the
// value is empty or invalid.
//
// The returned string follows the upstream convention: for a matching candidate
// that contains "*" or begins with "+", the concrete (normalized) value is
// returned; otherwise the candidate is returned exactly as supplied.
func Is(contentType string, types ...string) (string, bool) {
	val := NormalizeType(contentType)
	if val == "" {
		return "", false
	}
	if len(types) == 0 {
		return val, true
	}
	for _, t := range types {
		pattern := normalize(t)
		if pattern == "" {
			continue
		}
		if mimeMatch(pattern, val) {
			if len(t) > 0 && (t[0] == '+' || strings.ContainsRune(t, '*')) {
				return val, true
			}
			return t, true
		}
	}
	return "", false
}
