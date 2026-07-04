package base32

import (
	"encoding/base32"
	"strings"
)

// Encode encodes data using the uppercase RFC 4648 alphabet with padding.
func Encode(data []byte) string {
	return base32.StdEncoding.EncodeToString(data)
}

// EncodeNoPadding encodes data without the "=" padding characters.
func EncodeNoPadding(data []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data)
}

// Decode decodes a base32 string, case-insensitively and tolerating missing padding.
func Decode(s string) ([]byte, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimRight(s, "=")
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
}
