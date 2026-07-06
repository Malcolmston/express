// Package vary manipulates the HTTP Vary response header. It is a
// standard-library-only Go port of the npm "vary" package, the small helper
// Express and its middleware use to record which request headers a response
// varies on. The Vary header tells caches that a stored response may only be
// reused for a later request when the listed request headers match, so getting
// it right is essential for correct content negotiation: middleware that
// branches on Accept-Encoding, Accept-Language, or Origin must append the
// corresponding field to Vary or a shared cache may serve the wrong variant.
//
// The core operation is appending a field name to an existing Vary value
// without introducing duplicates. Because HTTP header field names are
// case-insensitive, deduplication here is also case-insensitive: appending
// "accept-encoding" to a header that already contains "Accept-Encoding" leaves
// the header unchanged rather than listing the field twice. Existing fields are
// preserved in their original order and casing, and newly appended fields are
// added in the order given, so the result is stable and predictable.
//
// The "*" wildcard is handled with its special HTTP meaning. A Vary value of
// "*" means the response varies on factors beyond the request headers and is
// effectively uncacheable per-variant. This package honors that: if the
// existing header already contains "*", or if any field being appended is "*",
// the result collapses to exactly "*" and the individual field names are
// discarded, because listing specific fields alongside "*" would be redundant.
//
// The package exposes three related entry points. Vary mutates an http.Header in
// place, reading its current "Vary" value, appending the given fields, and
// writing the result back; it silently ignores invalid field names so it can be
// called without error handling in a middleware chain. Append is the pure
// string form: it takes an existing header value and the fields to add and
// returns the new value together with an error when a field name is not a valid
// HTTP token. Field builds a Vary value from scratch out of a slice of names and
// is simply Append against an empty starting header.
//
// Field-name validation follows RFC 7230: a valid name is a non-empty token
// composed only of the token characters (letters, digits, and a fixed set of
// punctuation such as "!#$%&'*+-.^_`|~"). Names containing spaces, commas, or
// other separators are rejected by Append and Field with an error and skipped by
// Vary. Parity with the Node original covers the deduplication, wildcard
// collapsing, and token validation; the difference is idiomatic Go shape, with
// an http.Header-mutating Vary alongside the pure Append and Field string
// helpers instead of JavaScript's single overloaded function.
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
