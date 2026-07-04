// Package vary manipulates the HTTP Vary response header, a port of the npm
// "vary" package. It appends field names to a Vary header with case-insensitive
// deduplication and correct handling of the "*" wildcard.
package vary

import (
	"errors"
	"net/http"
	"strings"
)

// errInvalidField is returned when a field name is not a valid HTTP token.
var errInvalidField = errors.New("field argument contains an invalid header name")

// Vary appends the given field(s) to the Vary header of h, mutating it in
// place. Invalid field names are ignored to keep the mutation side-effect free
// of error handling; use Append for error reporting.
func Vary(h http.Header, field ...string) {
	if h == nil {
		return
	}
	val, err := Append(h.Get("Vary"), field...)
	if err != nil {
		return
	}
	if val != "" {
		h.Set("Vary", val)
	}
}

// Append appends field(s) to the given Vary header value, returning the new
// value. Deduplication is case-insensitive. If the existing header contains "*"
// or any appended field is "*", the result is "*". An error is returned if any
// field is not a valid header field name.
func Append(header string, field ...string) (string, error) {
	for _, f := range field {
		if !validFieldName(f) {
			return "", errInvalidField
		}
	}

	// Existing fields.
	var existing []string
	for _, f := range strings.Split(header, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			existing = append(existing, f)
		}
	}

	// If existing already has "*", the header stays "*".
	for _, f := range existing {
		if f == "*" {
			return "*", nil
		}
	}

	for _, f := range field {
		if f == "*" {
			return "*", nil
		}
		if !containsFold(existing, f) {
			existing = append(existing, f)
		}
	}

	return strings.Join(existing, ", "), nil
}

// Field builds a Vary header value from a list of field names, deduplicating
// case-insensitively and collapsing to "*" when appropriate.
func Field(fields []string) (string, error) {
	return Append("", fields...)
}

func containsFold(list []string, s string) bool {
	for _, item := range list {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// validFieldName reports whether s is a valid HTTP token (RFC 7230).
func validFieldName(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isTokenChar(s[i]) {
			return false
		}
	}
	return true
}

func isTokenChar(c byte) bool {
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
