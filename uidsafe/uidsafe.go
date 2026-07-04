// Package uidsafe generates cryptographically secure, URL-safe unique
// identifiers, mirroring the behavior of the npm uid-safe library.
package uidsafe

import (
	"crypto/rand"
	"encoding/base64"
)

// Bytes returns a URL-safe base64 string generated from n cryptographically
// random bytes. The result uses base64 URL-safe encoding without padding, so
// its length differs from n. It returns an error only if the system random
// source fails.
func Bytes(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// MustBytes is like Bytes but panics if the random source fails. It is
// convenient for initialization where an error is not recoverable.
func MustBytes(n int) string {
	s, err := Bytes(n)
	if err != nil {
		panic(err)
	}
	return s
}
