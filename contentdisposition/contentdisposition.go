// Package contentdisposition creates and parses HTTP Content-Disposition
// headers. It is a port of the npm content-disposition package using only the
// Go standard library, and it exposes the two operations that package offers:
// Format builds a header value from a filename, and Parse decodes a header
// value back into its type and parameters.
//
// The Content-Disposition header tells a client how a response body should be
// handled: "inline" to display it in place, or "attachment" to save it, most
// often under a suggested filename. You use Format on the server side when
// streaming a download so the browser offers a sensible "Save As" name, and you
// use Parse on the client side, or in multipart/form-data handling, to recover
// the disposition type and filename a peer sent.
//
// Format is careful about non-ASCII filenames because a bare filename parameter
// may only carry ASCII. When the name is pure ASCII it emits a single quoted
// filename parameter, escaping embedded quotes and backslashes. When the name
// contains bytes above 0x7f it additionally emits an RFC 5987 filename*
// parameter: the value is prefixed with "UTF-8”" and every byte outside the
// RFC 5987 attr-char set is percent-encoded with lower-case hex. By default a
// legacy filename parameter is emitted alongside it as an ASCII fallback, built
// by replacing each non-ASCII byte with '?'; WithFallback(false) suppresses
// that fallback, and WithType selects the disposition type (default
// "attachment").
//
// Parse splits the header on unquoted semicolons, lower-cases the disposition
// type and each parameter name, and unescapes quoted-string values. A
// parameter whose name ends in "*" is treated as an RFC 5987 ext-value and
// decoded from its charset'lang'percent-encoded form; only the "utf-8" and
// "iso-8859-1" charsets are accepted and any other charset, or a malformed
// percent escape, produces an error. When both a filename and a filename*
// parameter are present the decoded extended value wins and is stored under the
// "filename" key, matching the precedence browsers apply. An empty or
// whitespace-only header is an error, while a header with only a type (for
// example "inline") parses to that type with an empty filename.
//
// Parity with the Node original is close for the common download and upload
// cases: the Format output for ASCII and UTF-8 filenames matches
// content-disposition, and Parse recovers the same type, filename, and
// parameter map. The differences are deliberate simplifications. This port does
// not perform the library's full ISO-8859-1 to UTF-8 transcoding of legacy
// values, it does not reproduce every validation error message verbatim, and
// it decodes rather than aggressively rejecting unusual-but-parseable inputs,
// so it favors round-trip fidelity over byte-for-byte error parity.
package contentdisposition

import (
	"errors"
	"fmt"
	"strings"
)

// ContentDisposition represents a parsed Content-Disposition header.
type ContentDisposition struct {
	// Type is the disposition type, e.g. "attachment" or "inline".
	Type string
	// Filename is the decoded filename, if any (from filename* or filename).
	Filename string
	// Parameters holds all disposition parameters (including "filename").
	Parameters map[string]string
}

// options holds the configuration for Format.
type options struct {
	typ      string
	fallback bool
}

// Option configures how a Content-Disposition header is formatted.
type Option func(*options)

// WithType sets the disposition type (default "attachment").
func WithType(t string) Option {
	return func(o *options) { o.typ = t }
}

// WithFallback controls whether an ASCII fallback filename parameter is
// emitted when the filename contains non-ASCII characters. It is true by
// default; setting it to false suppresses the plain filename parameter.
func WithFallback(fallback bool) Option {
	return func(o *options) { o.fallback = fallback }
}

// isASCII reports whether s consists solely of printable ASCII bytes.
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7f {
			return false
		}
	}
	return true
}

// attrChar reports whether c is allowed unencoded in an RFC 5987 value.
func attrChar(c byte) bool {
	switch {
	case c >= 'A' && c <= 'Z':
		return true
	case c >= 'a' && c <= 'z':
		return true
	case c >= '0' && c <= '9':
		return true
	}
	switch c {
	case '!', '#', '$', '&', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return false
}

// encodeExtValue percent-encodes s per RFC 5987 (ext-value) with UTF-8.
func encodeExtValue(s string) string {
	const hex = "0123456789abcdef"
	var b strings.Builder
	b.WriteString("UTF-8''")
	for i := 0; i < len(s); i++ {
		c := s[i]
		if attrChar(c) {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		}
	}
	return b.String()
}

// quoteString escapes a string for use as a quoted-string parameter value.
func quoteString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			b.WriteByte('\\')
		}
		b.WriteByte(c)
	}
	b.WriteByte('"')
	return b.String()
}

// asciiFallback replaces non-ASCII bytes with '?' to build a legacy filename.
func asciiFallback(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7f {
			b.WriteByte('?')
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

// Format builds a Content-Disposition header value for the given filename.
// By default the type is "attachment". When the filename contains non-ASCII
// characters, both a plain ASCII fallback filename and an RFC 5987
// filename* parameter are emitted.
func Format(filename string, opts ...Option) string {
	cfg := options{typ: "attachment", fallback: true}
	for _, opt := range opts {
		opt(&cfg)
	}

	var b strings.Builder
	b.WriteString(cfg.typ)

	if filename == "" {
		return b.String()
	}

	if isASCII(filename) {
		b.WriteString("; filename=")
		b.WriteString(quoteString(filename))
		return b.String()
	}

	if cfg.fallback {
		b.WriteString("; filename=")
		b.WriteString(quoteString(asciiFallback(filename)))
	}
	b.WriteString("; filename*=")
	b.WriteString(encodeExtValue(filename))
	return b.String()
}

// hexVal returns the numeric value of a hex digit and whether it was valid.
func hexVal(c byte) (byte, bool) {
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

// decodeExtValue decodes an RFC 5987 ext-value (charset'lang'percent-encoded).
func decodeExtValue(s string) (string, error) {
	idx := strings.Index(s, "'")
	if idx < 0 {
		return "", errors.New("invalid extended field value")
	}
	charset := strings.ToLower(s[:idx])
	rest := s[idx+1:]
	idx2 := strings.Index(rest, "'")
	if idx2 < 0 {
		return "", errors.New("invalid extended field value")
	}
	encoded := rest[idx2+1:]

	var raw []byte
	for i := 0; i < len(encoded); i++ {
		c := encoded[i]
		if c == '%' {
			if i+2 >= len(encoded) {
				return "", errors.New("invalid percent-encoding")
			}
			hi, ok1 := hexVal(encoded[i+1])
			lo, ok2 := hexVal(encoded[i+2])
			if !ok1 || !ok2 {
				return "", errors.New("invalid percent-encoding")
			}
			raw = append(raw, hi<<4|lo)
			i += 2
		} else {
			raw = append(raw, c)
		}
	}

	switch charset {
	case "utf-8", "iso-8859-1":
		// UTF-8 bytes are used directly; ISO-8859-1 bytes map 1:1 to runes,
		// but for our purposes we treat the raw bytes as UTF-8 which is the
		// common case for filenames.
		if charset == "iso-8859-1" {
			var b strings.Builder
			for _, by := range raw {
				b.WriteRune(rune(by))
			}
			return b.String(), nil
		}
		return string(raw), nil
	default:
		return "", fmt.Errorf("unsupported charset %q", charset)
	}
}

// unquote removes surrounding quotes and unescapes a quoted-string value.
func unquote(s string) string {
	if len(s) < 2 || s[0] != '"' || s[len(s)-1] != '"' {
		return s
	}
	inner := s[1 : len(s)-1]
	var b strings.Builder
	for i := 0; i < len(inner); i++ {
		if inner[i] == '\\' && i+1 < len(inner) {
			i++
			b.WriteByte(inner[i])
			continue
		}
		b.WriteByte(inner[i])
	}
	return b.String()
}

// splitParams splits a header value on unquoted semicolons.
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

// Parse parses a Content-Disposition header value. If no disposition type is
// present the default "attachment" is used. A filename* parameter takes
// precedence over a plain filename parameter.
func Parse(s string) (ContentDisposition, error) {
	if strings.TrimSpace(s) == "" {
		return ContentDisposition{}, errors.New("empty content-disposition header")
	}

	parts := splitParams(s)
	cd := ContentDisposition{
		Type:       strings.ToLower(strings.TrimSpace(parts[0])),
		Parameters: map[string]string{},
	}
	if cd.Type == "" {
		cd.Type = "attachment"
	}

	var plainFilename string
	var extFilename string
	haveExt := false

	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		eq := strings.IndexByte(p, '=')
		if eq < 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(p[:eq]))
		val := strings.TrimSpace(p[eq+1:])

		if strings.HasSuffix(key, "*") {
			decoded, err := decodeExtValue(val)
			if err != nil {
				return ContentDisposition{}, err
			}
			base := strings.TrimSuffix(key, "*")
			cd.Parameters[base] = decoded
			if base == "filename" {
				extFilename = decoded
				haveExt = true
			}
			continue
		}

		val = unquote(val)
		cd.Parameters[key] = val
		if key == "filename" {
			plainFilename = val
		}
	}

	if haveExt {
		cd.Filename = extFilename
		cd.Parameters["filename"] = extFilename
	} else {
		cd.Filename = plainFilename
	}

	return cd, nil
}
