// Package base64url encodes and decodes byte data using the RFC 4648 url-safe
// base64 alphabet without padding, as used in JWTs and other web tokens.
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
