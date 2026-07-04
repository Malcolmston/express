package uuid

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
)

// Predefined namespace UUIDs from RFC 4122.
const (
	NamespaceDNS  = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	NamespaceURL  = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	NamespaceOID  = "6ba7b812-9dad-11d1-80b4-00c04fd430c8"
	NamespaceX500 = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
)

// V4 returns a randomly generated UUID version 4 string.
func V4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Set version to 4.
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant to RFC 4122 (10xx).
	b[8] = (b[8] & 0x3f) | 0x80
	return Format(b), nil
}

// V5 returns a name-based UUID version 5 (SHA-1) string.
func V5(namespace string, name string) (string, error) {
	ns, err := Parse(namespace)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write(ns[:])
	h.Write([]byte(name))
	sum := h.Sum(nil)
	var b [16]byte
	copy(b[:], sum[:16])
	// Set version to 5.
	b[6] = (b[6] & 0x0f) | 0x50
	// Set variant to RFC 4122 (10xx).
	b[8] = (b[8] & 0x3f) | 0x80
	return Format(b), nil
}

// Parse parses a UUID string into a 16-byte array.
func Parse(s string) ([16]byte, error) {
	var b [16]byte
	if len(s) != 36 {
		return b, errors.New("uuid: invalid length")
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return b, errors.New("uuid: invalid format")
	}
	stripped := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	decoded, err := hex.DecodeString(stripped)
	if err != nil {
		return b, errors.New("uuid: invalid hex")
	}
	copy(b[:], decoded)
	return b, nil
}

// Format renders a 16-byte array as a canonical lowercase UUID string.
func Format(b [16]byte) string {
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// Validate reports whether s is a well-formed UUID string.
func Validate(s string) bool {
	_, err := Parse(s)
	return err == nil
}
