// Package uidsafe generates cryptographically secure, URL-safe unique
// identifiers. It is a standard-library-only Go port of the npm "uid-safe"
// library, the token generator used by express-session and cookie-session to
// mint session ids and by other middleware that needs an unguessable, opaque
// string. The name captures the two guarantees it provides: the identifier is
// safe to embed unescaped in URLs, cookies, and headers, and it is drawn from a
// cryptographically secure source so it cannot be predicted by an attacker.
//
// The generation strategy is deliberately simple. Bytes reads n bytes from
// crypto/rand, the operating system's CSPRNG, and encodes them with base64
// URL-safe encoding (the alphabet that uses '-' and '_' in place of '+' and
// '/') without any '=' padding. Because base64 packs three input bytes into
// four output characters, the resulting string is longer than n and its length
// is not equal to n; for example 18 random bytes produce a 24-character token.
// Callers who care about the exact string length should choose n accordingly
// rather than assuming a one-to-one mapping.
//
// Randomness quality is the whole point. Every byte comes from crypto/rand, so
// the tokens are suitable for security-sensitive roles such as session
// identifiers, CSRF tokens, password-reset links, and API keys, where a
// predictable value would be a vulnerability. The URL-safe, unpadded encoding
// means the token never needs percent-encoding and never ends in stray "="
// characters that some routers or caches mishandle, so the same string can be
// used verbatim as a path segment, query value, or cookie value.
//
// Two entry points cover the common cases. Bytes returns the token and an error,
// which is non-nil only in the rare event that the system random source fails.
// MustBytes is the convenience variant for program initialization: it calls
// Bytes and panics if the random source fails, which is appropriate when a
// missing entropy source is unrecoverable and the alternative would be running
// without secure tokens. Neither function retries or blocks; on a healthy system
// crypto/rand does not fail.
//
// Parity with the Node original is at the level of behavior rather than call
// shape. The npm uid-safe exposes an async and a ".sync" form that take a byte
// count and return a URL-safe base64 string; this port collapses that to Bytes
// (error-returning) and MustBytes (panicking), using Go's crypto/rand instead of
// Node's crypto.randomBytes and Go's encoding/base64 RawURLEncoding to match the
// unpadded URL-safe output. The encoding, alphabet, and padding behavior are the
// same, so a token of a given byte count has the same length and character set
// as its JavaScript counterpart.
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
