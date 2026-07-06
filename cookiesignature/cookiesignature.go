// Package cookiesignature signs and verifies cookie values using
// HMAC-SHA256, mirroring the behavior of the npm cookie-signature library.
// It is the mechanism Express and cookie-parser use to produce "signed
// cookies": a tamper-evident wrapper around an otherwise plaintext cookie
// value that lets a server detect whether a value it later receives was the
// one it originally issued.
//
// Reach for this package whenever a value is round-tripped through an
// untrusted party (a browser, a proxy, a query string) but must not be
// forgeable. Signing does not hide or encrypt the value; anyone can read it.
// What signing guarantees is integrity and authenticity: only a holder of the
// secret can produce a signature that Unsign will accept, so a client cannot
// alter the value (for example, escalate a session id or a role flag) without
// invalidating the signature. Typical uses are session identifiers, CSRF
// tokens, and remember-me cookies.
//
// The construction is deliberately simple. Sign computes
// HMAC-SHA256(secret, value), encodes the 32-byte digest with standard
// base64, strips the trailing '=' padding, and appends it to the value after
// a '.' separator, yielding "value.signature". Because the value itself is
// stored verbatim in front of the dot, Unsign recovers it by splitting on the
// last '.', re-signing the recovered value with the same secret, and checking
// that the freshly produced string equals the input. This "sign then compare
// whole strings" strategy is exactly how the Node original works, which is
// why values are allowed to contain dots: only the final dot is treated as
// the separator.
//
// Several edge cases are worth calling out. An input with no '.' at all is
// rejected immediately (Unsign returns "", false). An empty value is a valid
// input and round-trips cleanly: Sign("", secret) yields ".signature" and
// Unsign returns "", true, so callers must use the boolean result rather than
// the emptiness of the returned string to decide success. Comparison is
// performed with crypto/hmac.Equal so it runs in constant time and does not
// leak how many leading bytes matched, closing a timing side channel that a
// naive string comparison would open. Sign is deterministic: the same value
// and secret always produce the same output, which is what makes the
// re-sign-and-compare verification possible.
//
// Parity with the npm cookie-signature library is intentional and close.
// Both use HMAC-SHA256, both base64-encode the digest and remove '='
// padding, both join with '.', and both verify by re-signing and comparing in
// constant time. The observable string outputs are byte-for-byte identical
// for the same value and secret, so cookies signed by a Node service can be
// unsigned here and vice versa. The only surface difference is idiomatic: the
// Node unsign returns the value or false, whereas Unsign returns the Go-style
// (value string, ok bool) pair.
package cookiesignature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

// Sign returns value with an HMAC-SHA256 signature appended, in the form
// "value.signature". The signature is the base64url-encoded HMAC of value
// keyed by secret, with trailing '=' padding removed, matching the npm
// cookie-signature library.
func Sign(value, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(value))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	sig = strings.TrimRight(sig, "=")
	return value + "." + sig
}

// Unsign verifies a signed value produced by Sign against secret. It returns
// the original value and true when the signature is valid, or "" and false
// when the signature does not match or the input is malformed (for example,
// when it contains no '.' separator). The comparison is performed in constant
// time.
func Unsign(signed, secret string) (string, bool) {
	idx := strings.LastIndex(signed, ".")
	if idx < 0 {
		return "", false
	}
	value := signed[:idx]
	expected := Sign(value, secret)
	if hmac.Equal([]byte(signed), []byte(expected)) {
		return value, true
	}
	return "", false
}
