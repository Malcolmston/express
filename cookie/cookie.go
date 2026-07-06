// Package cookie parses and serializes HTTP cookie headers, a port of the npm
// "cookie" package that Express and cookie-parser use to read the Cookie request
// header and to build Set-Cookie response headers. It exposes Parse to turn a
// raw "Cookie" header into a map of names to values and Serialize to render a
// single name/value pair plus its attributes into a Set-Cookie value, using only
// the Go standard library.
//
// You reach for this package on both sides of the request/response cycle. On the
// way in, Parse gives you the individual cookies a client sent so a handler can
// look up a session id or a preference flag without string-splitting the header
// itself. On the way out, Serialize produces a well-formed Set-Cookie value with
// Path, Domain, Expires, Max-Age, Secure, HttpOnly, and SameSite attributes, so
// you can issue login cookies, clear a cookie by expiring it, or scope a cookie
// to a path without hand-assembling the syntax and risking a malformed header.
//
// Values are transported using JavaScript's encodeURIComponent conventions.
// Serialize percent-encodes every byte of the value that is not an
// encodeURIComponent-safe character (letters, digits, and the set -_.!~*'()),
// which keeps semicolons, commas, spaces, and other separators from corrupting
// the header. Parse reverses this: it URL-decodes each value, tolerating values
// that were never encoded, and additionally strips one layer of surrounding
// double quotes so quoted-string cookie values round-trip. When the same cookie
// name appears more than once in a header the first occurrence wins, matching the
// npm package, and pairs with an empty name or no '=' are skipped.
//
// Serialize validates its inputs rather than emitting a header that a client
// would reject. The name must be an RFC 6265 token and the encoded value must be
// a valid cookie-octet sequence, otherwise an error is returned; Path and Domain
// must be valid field-content. Max-Age follows net/http conventions rather than
// dotenv-style literalism: a positive value sets Max-Age to that many seconds, a
// negative value emits "Max-Age=0" to delete the cookie immediately, and zero
// omits the attribute entirely. A zero Expires time omits Expires, and an empty
// SameSite omits SameSite while "Lax", "Strict", and "None" (case-insensitive)
// are accepted and anything else is an error. Passing a nil *Options serializes
// just the name and value with no attributes.
//
// Parity with the Node original is close for everyday cookie handling: the
// encodeURIComponent-based value codec, the token and cookie-octet validation,
// the first-wins duplicate rule, and the attribute names and formats all match.
// The deliberate differences are idiomatic Go ones. Attributes are supplied
// through a typed Options struct instead of a plain JavaScript object, the value
// codec is fixed to encode/decode rather than accepting caller-supplied
// encode/decode callbacks, Serialize returns an explicit error instead of
// throwing a TypeError, and Max-Age is interpreted with net/http's sign
// conventions so it composes naturally with the rest of the standard library.
package cookie

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Options holds the attributes used when serializing a cookie.
type Options struct {
	// Path sets the Path attribute.
	Path string
	// Domain sets the Domain attribute.
	Domain string
	// Expires sets the Expires attribute. A zero time omits the attribute.
	Expires time.Time
	// MaxAge sets the Max-Age attribute. Following net/http conventions:
	// a positive value sets Max-Age to that number of seconds, a negative
	// value sets Max-Age=0 (delete now), and zero omits the attribute.
	MaxAge int
	// Secure sets the Secure attribute.
	Secure bool
	// HttpOnly sets the HttpOnly attribute.
	HttpOnly bool
	// SameSite sets the SameSite attribute; recognized values are "Lax",
	// "Strict", and "None" (case-insensitive). An empty value omits it.
	SameSite string
}

// Parse parses a Cookie header (for example "a=1; b=2") into a map of names to
// values. Values are URL-decoded. When a name appears more than once, the
// first occurrence wins.
func Parse(header string) map[string]string {
	m := make(map[string]string)
	if header == "" {
		return m
	}

	for _, pair := range strings.Split(header, ";") {
		eq := strings.IndexByte(pair, '=')
		if eq < 0 {
			continue
		}
		name := strings.TrimSpace(pair[:eq])
		if name == "" {
			continue
		}
		if _, exists := m[name]; exists {
			continue
		}
		val := strings.TrimSpace(pair[eq+1:])
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		}
		m[name] = decode(val)
	}
	return m
}

// Serialize builds a Set-Cookie header value for the given name and value,
// applying the attributes in opts (which may be nil). The value is URL-encoded.
// It returns an error if the name or value contains invalid characters, or if
// an attribute value is invalid.
func Serialize(name, value string, opts *Options) (string, error) {
	if !isCookieName(name) {
		return "", errors.New("cookie: name is invalid")
	}

	encoded := encode(value)
	if !isCookieValue(encoded) {
		return "", errors.New("cookie: value is invalid")
	}

	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('=')
	b.WriteString(encoded)

	if opts == nil {
		return b.String(), nil
	}

	if opts.Path != "" {
		if !isFieldContent(opts.Path) {
			return "", errors.New("cookie: path is invalid")
		}
		b.WriteString("; Path=")
		b.WriteString(opts.Path)
	}

	if opts.Domain != "" {
		if !isFieldContent(opts.Domain) {
			return "", errors.New("cookie: domain is invalid")
		}
		b.WriteString("; Domain=")
		b.WriteString(opts.Domain)
	}

	if !opts.Expires.IsZero() {
		b.WriteString("; Expires=")
		b.WriteString(opts.Expires.UTC().Format(http.TimeFormat))
	}

	if opts.MaxAge != 0 {
		ma := opts.MaxAge
		if ma < 0 {
			ma = 0
		}
		b.WriteString("; Max-Age=")
		b.WriteString(strconv.Itoa(ma))
	}

	if opts.HttpOnly {
		b.WriteString("; HttpOnly")
	}

	if opts.Secure {
		b.WriteString("; Secure")
	}

	switch strings.ToLower(opts.SameSite) {
	case "":
		// omit
	case "lax":
		b.WriteString("; SameSite=Lax")
	case "strict":
		b.WriteString("; SameSite=Strict")
	case "none":
		b.WriteString("; SameSite=None")
	default:
		return "", errors.New("cookie: sameSite is invalid")
	}

	return b.String(), nil
}

// decode URL-decodes a cookie value, leaving it unchanged if it is not encoded
// or cannot be decoded.
func decode(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	if dec, err := url.PathUnescape(s); err == nil {
		return dec
	}
	return s
}

// encode percent-encodes a cookie value using encodeURIComponent semantics.
func encode(s string) string {
	const hexDigits = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isURIComponentSafe(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hexDigits[c>>4])
			b.WriteByte(hexDigits[c&0x0F])
		}
	}
	return b.String()
}

// isURIComponentSafe reports whether b is left unencoded by
// JavaScript's encodeURIComponent.
func isURIComponentSafe(b byte) bool {
	if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') {
		return true
	}
	switch b {
	case '-', '_', '.', '!', '~', '*', '\'', '(', ')':
		return true
	}
	return false
}

// isCookieName reports whether name is a valid cookie name (an RFC 6265 token).
func isCookieName(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		if !isTokenChar(name[i]) {
			return false
		}
	}
	return true
}

func isTokenChar(c byte) bool {
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
		return true
	}
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// isCookieValue reports whether s is a valid cookie-octet sequence per RFC 6265.
func isCookieValue(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x21 ||
			(c >= 0x23 && c <= 0x2B) ||
			(c >= 0x2D && c <= 0x3A) ||
			(c >= 0x3C && c <= 0x5B) ||
			(c >= 0x5D && c <= 0x7E) {
			continue
		}
		return false
	}
	return true
}

// isFieldContent reports whether s is valid for an attribute value such as
// Path or Domain (RFC 7230 field-content, roughly printable characters).
func isFieldContent(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0x09 || (c >= 0x20 && c <= 0x7E) || c >= 0x80 {
			continue
		}
		return false
	}
	return true
}
