// Package contenttype parses and formats HTTP Content-Type header values,
// modeled on the npm "content-type" package and RFC 7231.
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
		} else {
			b.WriteByte('"')
			b.WriteString(escapeRE.ReplaceAllString(v, `\$1`))
			b.WriteByte('"')
		}
	}
	return b.String(), nil
}
