// Package cookiesignature signs and verifies cookie values using
// HMAC-SHA256, mirroring the behavior of the npm cookie-signature library.
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
