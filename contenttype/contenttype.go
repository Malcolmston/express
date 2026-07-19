// Package contenttype parses and formats HTTP Content-Type header values,
// modeled on the npm "content-type" package and RFC 7231, using only the Go
// standard library. It provides Parse to decode a header value into a
// ContentType struct and Format to serialize a ContentType back into a header
// value.
//
// The Content-Type header names the media type of a message body and carries
// optional parameters such as the character set. You use Parse on the receiving
// side to learn the media type and, for example, the charset of a request or
// response body, and Format on the sending side to build a well-formed header
// from a media type and a parameter map. Both operations are pure functions, so
// they are convenient to test and to reuse outside of a live HTTP handler.
//
// Parse splits the value at the first ';' into a media type and a parameter
// list. The media type is trimmed and lower-cased and must match the RFC 7230
// token "/" token grammar or an error is returned. The remainder is scanned one
// "; name=value" parameter at a time with a regular expression built from the
// token character class; each parameter name is lower-cased, and a value given
// as a quoted string is unquoted and has its backslash escapes removed. A
// value that is a bare token is taken verbatim. Any input that does not match
// the parameter grammar, including a trailing "name" with no "=value", produces
// an error.
//
// Format is the inverse and validates as it goes. The Type field must be a
// valid media type or Format returns an error. Parameters are emitted in sorted
// name order so the output is deterministic, which makes it stable to compare
// or snapshot; each name must be a valid token. A parameter value that is
// itself a valid token is written unquoted, while any other value is wrapped in
// double quotes with backslashes and quotes escaped. Note that parameter names
// are lower-cased on Parse but Format writes the names it is given as-is, and
// that parameter values preserve their original case in both directions.
//
// The result is faithful to the npm content-type package for typical headers:
// the token and quoted-string grammar, the lower-casing of the type and
// parameter names, the case preservation of values, and the quoting rules on
// output all match. The main intentional differences are that this port sorts
// parameters by name for deterministic output rather than preserving insertion
// order, and it uses Go-idiomatic error values instead of thrown TypeError
// objects, so error messages are prefixed with "contenttype:" rather than
// matching the Node text verbatim.
package contenttype

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// ContentType represents a parsed media type and its parameters.
type ContentType struct {
	// Type is the media type, e.g. "text/html". It is always lower-cased.
	Type string
	// Parameters holds the media type parameters keyed by their lower-cased
	// names, e.g. {"charset": "utf-8"}.
	Parameters map[string]string
}

// token is the RFC 7230 token character class.
const token = "[!#$%&'*+.^_`|~0-9A-Za-z-]"

var (
	typeRE  = regexp.MustCompile(`^` + token + `+/` + token + `+$`)
	tokenRE = regexp.MustCompile(`^` + token + `+$`)
	// textRE matches values that may be written as a quoted-string: HTAB, SP,
	// printable ASCII, and obs-text (%x80-FF), per RFC 9110 sec 5.6.4. Values
	// outside this class (e.g. NUL or vertical tab) cannot be represented and
	// are rejected by Format, matching upstream qstring().
	textRE = regexp.MustCompile(`^[\x{09}\x{20}-\x{7e}\x{80}-\x{ff}]*$`)
	// paramRE matches a single "; name=value" parameter at the start of the
	// remaining input. The value is either a token or a quoted string.
	paramRE    = regexp.MustCompile(`^;[ \t]*(` + token + `+)[ \t]*=[ \t]*("(?:\\.|[^"\\])*"|` + token + `+)[ \t]*`)
	unescapeRE = regexp.MustCompile(`\\(.)`)
	escapeRE   = regexp.MustCompile(`([\\"])`)
)

// Parse parses a Content-Type header value such as "text/html; charset=utf-8".
// The media type and parameter names are lower-cased; quoted parameter values
// are unquoted and unescaped. It returns an error if the value is malformed.
func Parse(s string) (ContentType, error) {
	idx := strings.IndexByte(s, ';')
	typePart := s
	rest := ""
	if idx >= 0 {
		typePart = s[:idx]
		rest = s[idx:]
	}
	typePart = strings.TrimSpace(typePart)
	lower := strings.ToLower(typePart)
	if !typeRE.MatchString(lower) {
		return ContentType{}, fmt.Errorf("contenttype: invalid media type %q", typePart)
	}

	ct := ContentType{Type: lower, Parameters: map[string]string{}}

	pos := 0
	for pos < len(rest) {
		m := paramRE.FindStringSubmatch(rest[pos:])
		if m == nil {
			return ContentType{}, fmt.Errorf("contenttype: invalid parameter format")
		}
		key := strings.ToLower(m[1])
		value := m[2]
		if len(value) > 0 && value[0] == '"' {
			value = unescapeRE.ReplaceAllString(value[1:len(value)-1], "$1")
		}
		ct.Parameters[key] = value
		pos += len(m[0])
	}
	return ct, nil
}

// Format serializes a ContentType back into a header value such as
// "text/html; charset=utf-8". Parameter values that are not valid tokens are
// quoted and escaped. Parameters are emitted in sorted name order. It returns
// an error if the type or any parameter name is invalid.
func Format(ct ContentType) (string, error) {
	if !typeRE.MatchString(ct.Type) {
		return "", fmt.Errorf("contenttype: invalid media type %q", ct.Type)
	}

	var b strings.Builder
	b.WriteString(ct.Type)

	keys := make([]string, 0, len(ct.Parameters))
	for k := range ct.Parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if !tokenRE.MatchString(k) {
			return "", fmt.Errorf("contenttype: invalid parameter name %q", k)
		}
		b.WriteString("; ")
		b.WriteString(k)
		b.WriteByte('=')
		v := ct.Parameters[k]
		if tokenRE.MatchString(v) {
			b.WriteString(v)
		} else if textRE.MatchString(v) {
			b.WriteByte('"')
			b.WriteString(escapeRE.ReplaceAllString(v, `\$1`))
			b.WriteByte('"')
		} else {
			return "", fmt.Errorf("contenttype: invalid parameter value %q", v)
		}
	}
	return b.String(), nil
}
