// Package base64url encodes and decodes byte data using the RFC 4648 url-safe
// base64 alphabet without padding, as used in JWTs and other web tokens. It
// ports the npm base64url package's surface on top of the standard library's
// encoding/base64, exposing the same encode, decode, and standard-base64
// interconversion helpers while delegating the underlying transform to Go.
//
// You want this package whenever base64 data has to travel through a URL, an
// HTTP header, a filename, or a JSON web token. Ordinary base64 uses "+" and
// "/", which are reserved or unsafe in those positions, and its "=" padding is
// often stripped by intermediaries; base64url swaps those two characters for "-"
// and "_" and omits padding so the text can be dropped into a query string, a
// path segment, or a JWT segment without any further percent-encoding. It is the
// encoding you see in JWT headers and payloads, in OAuth tokens, and in many
// web-facing identifiers.
//
// Encoding uses the RFC 4648 "URL and Filename Safe" alphabet in its raw
// (unpadded) form: input bytes are grouped into 6-bit units, each unit maps to
// one of A-Z, a-z, 0-9, "-" or "_", and no trailing "=" characters are emitted.
// Encode operates on a byte slice and EncodeString is a convenience wrapper that
// encodes a string's UTF-8 bytes. Because the alphabet is a straightforward
// character substitution of standard base64, FromBase64 can convert an existing
// standard-base64 string to base64url purely by trimming padding and replacing
// "+"/"/" with "-"/"_", and ToBase64 reverses that by replacing "-"/"_" with
// "+"/"/" and re-appending as many "=" characters as are needed to make the
// length a multiple of four.
//
// Decoding is tolerant of missing or extra padding. Decode and DecodeString
// trim any trailing "=" before decoding, so they accept both the canonical
// unpadded output of this package and padded strings copied from a
// standard-base64 source. An empty input round-trips to empty output. Input that
// contains characters outside the url-safe alphabet, or that has an invalid
// length, produces a non-nil error from encoding/base64 rather than partial or
// silently truncated output; DecodeString additionally returns the empty string
// alongside that error. Note that FromBase64 and ToBase64 are pure string
// rewrites and do not themselves validate that their argument is well-formed
// base64.
//
// Compared with the Node original the behavior is intentionally matched:
// Encode/EncodeString correspond to base64url() and its Buffer form,
// Decode/DecodeString to base64url.decode, and FromBase64/ToBase64 to the
// fromBase64/toBase64 helpers. The differences are the idiomatic Go ones.
// Functions take and return []byte and string rather than Buffers, decoding
// reports failure through an error return instead of throwing, and DecodeString
// decodes to a Go string (interpreting the bytes as UTF-8) just as the npm
// package returns a JavaScript string by default.
package base64url

import (
	"encoding/base64"
	"strings"
)

// Encode encodes data as RFC 4648 url-safe base64 without padding.
func Encode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// EncodeString encodes a string as url-safe base64 without padding.
func EncodeString(s string) string {
	return Encode([]byte(s))
}

// Decode decodes a url-safe base64 string, tolerating missing padding.
func Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

// DecodeString decodes a url-safe base64 string and returns it as a string.
func DecodeString(s string) (string, error) {
	b, err := Decode(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// FromBase64 converts a standard base64 string to a base64url string.
func FromBase64(s string) string {
	s = strings.TrimRight(s, "=")
	s = strings.ReplaceAll(s, "+", "-")
	s = strings.ReplaceAll(s, "/", "_")
	return s
}

// ToBase64 converts a base64url string to a standard base64 string with padding.
func ToBase64(s string) string {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return s
}
