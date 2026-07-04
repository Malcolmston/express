// Package mediatyper parses and formats HTTP media type strings of the form
// type/subtype+suffix; params. It is a port of the npm media-typer package
// using only the Go standard library.
package mediatyper

import (
	"errors"
	"sort"
	"strings"
)

// MediaType represents a parsed media type.
type MediaType struct {
	// Type is the top-level type, e.g. "application".
	Type string
	// Subtype is the subtype without any suffix, e.g. "vnd.api".
	Subtype string
	// Suffix is the optional structured suffix without the "+", e.g. "json".
	Suffix string
	// Parameters holds media type parameters keyed by lower-case name.
	Parameters map[string]string
}

// isToken reports whether s is a valid RFC 7230 token (non-empty, only tchar).
func isToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isTChar(s[i]) {
			return false
		}
	}
	return true
}

// isTChar reports whether c is a valid token character.
func isTChar(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// quoteValue wraps a parameter value in a quoted-string if it is not a valid
// token.
func quoteValue(v string) string {
	if isToken(v) {
		return v
	}
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(v); i++ {
		if v[i] == '"' || v[i] == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(v[i])
	}
	b.WriteByte('"')
	return b.String()
}

// Format reconstructs a media type string from a MediaType value. It returns
// an error if the type or subtype are not valid tokens.
func Format(m MediaType) (string, error) {
	if !isToken(m.Type) {
		return "", errors.New("invalid type")
	}
	if !isToken(m.Subtype) {
		return "", errors.New("invalid subtype")
	}
	if m.Suffix != "" && !isToken(m.Suffix) {
		return "", errors.New("invalid suffix")
	}

	var b strings.Builder
	b.WriteString(strings.ToLower(m.Type))
	b.WriteByte('/')
	b.WriteString(strings.ToLower(m.Subtype))
	if m.Suffix != "" {
		b.WriteByte('+')
		b.WriteString(strings.ToLower(m.Suffix))
	}

	// Emit parameters in a stable (sorted) order.
	keys := make([]string, 0, len(m.Parameters))
	for k := range m.Parameters {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !isToken(k) {
			return "", errors.New("invalid parameter name")
		}
		b.WriteString("; ")
		b.WriteString(strings.ToLower(k))
		b.WriteByte('=')
		b.WriteString(quoteValue(m.Parameters[k]))
	}
	return b.String(), nil
}

// unquote unescapes a quoted-string value (without surrounding quotes handling
// beyond stripping matched quotes).
func unquote(s string) (string, bool) {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return "", false
	}
	inner := s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			i++
		}
		b.WriteByte(inner[i])
	}
	return b.String(), true
}

// splitParams splits on unquoted semicolons.
func splitParams(s string) []string {
	var parts []string
	var cur strings.Builder
	inQuotes := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			inQuotes = !inQuotes
			cur.WriteByte(c)
		case c == '\\' && inQuotes && i+1 < len(s):
			cur.WriteByte(c)
			i++
			cur.WriteByte(s[i])
		case c == ';' && !inQuotes:
			parts = append(parts, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	parts = append(parts, cur.String())
	return parts
}

// Parse parses a media type string. It validates token syntax for the type,
// subtype, optional "+suffix" and parameter names, returning an error on
// invalid input.
func Parse(s string) (MediaType, error) {
	if strings.TrimSpace(s) == "" {
		return MediaType{}, errors.New("invalid media type")
	}

	parts := splitParams(s)
	head := strings.TrimSpace(parts[0])

	slash := strings.IndexByte(head, '/')
	if slash < 0 {
		return MediaType{}, errors.New("invalid media type: missing subtype")
	}
	typ := strings.ToLower(strings.TrimSpace(head[:slash]))
	sub := strings.ToLower(strings.TrimSpace(head[slash+1:]))
	if !isToken(typ) {
		return MediaType{}, errors.New("invalid type")
	}

	m := MediaType{Type: typ, Parameters: map[string]string{}}

	if plus := strings.LastIndexByte(sub, '+'); plus >= 0 {
		base := sub[:plus]
		suffix := sub[plus+1:]
		if !isToken(base) || !isToken(suffix) {
			return MediaType{}, errors.New("invalid subtype")
		}
		m.Subtype = base
		m.Suffix = suffix
	} else {
		if !isToken(sub) {
			return MediaType{}, errors.New("invalid subtype")
		}
		m.Subtype = sub
	}

	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		if eq < 0 {
			return MediaType{}, errors.New("invalid parameter: missing value")
		}
		key := strings.ToLower(strings.TrimSpace(p[:eq]))
		val := strings.TrimSpace(p[eq+1:])
		if !isToken(key) {
			return MediaType{}, errors.New("invalid parameter name")
		}
		if uq, ok := unquote(val); ok {
			val = uq
		} else if !isToken(val) {
			return MediaType{}, errors.New("invalid parameter value")
		}
		m.Parameters[key] = val
	}

	return m, nil
}
