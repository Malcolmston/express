// Package timingsafe compares byte slices and strings in constant time to
// avoid leaking secrets through timing side channels. It is a stdlib-only Go
// port of Node's crypto.timingSafeEqual
// (https://nodejs.org/api/crypto.html#cryptotimingsafeequala-b), which is also
// republished on npm as the "timing-safe-equal" package. The problem it solves
// is that the obvious way to compare two byte strings — a byte-by-byte loop
// that stops at the first difference — takes longer the more leading bytes
// match, so an attacker who can measure comparison time can recover a secret
// one byte at a time.
//
// The canonical place this matters is verifying secrets supplied by an
// untrusted caller against a value the server holds: HMAC signatures, session
// tokens, password-reset tokens, API keys, CSRF tokens, and one-time codes. For
// any such check the comparison must take the same amount of time whether the
// candidate is completely wrong or wrong only in its final byte. Using an
// ordinary == or bytes.Equal for these comparisons is a real, exploitable
// vulnerability; this package exists so that the safe comparison is a single
// obvious call.
//
// Equal reports whether two byte slices are equal in constant time. Internally
// it delegates to crypto/subtle.ConstantTimeCompare, the Go standard library's
// timing-safe primitive, which reads every byte of both inputs and accumulates
// differences without an early return, so its running time depends only on the
// input length and not on the position or number of differing bytes. Equal
// returns false for unequal, true for equal, and never panics.
//
// Length handling deserves note. crypto/subtle.ConstantTimeCompare requires its
// two arguments to have equal length and returns 0 otherwise; this package first
// checks len(a) != len(b) and returns false immediately for a mismatch. That
// length check is itself fast and non-constant-time with respect to length, but
// this matches Node's crypto.timingSafeEqual, which likewise rejects
// differently sized inputs (in Node's case by throwing). The length of a secret
// is not generally considered sensitive, so this is the standard and accepted
// behavior; only the content comparison is protected.
//
// EqualString is a convenience wrapper that compares two strings by converting
// them to byte slices and calling Equal, giving the same constant-time guarantee
// for string inputs. The one behavioral parity note versus Node is that
// crypto.timingSafeEqual throws a TypeError when the two inputs differ in length
// or are not buffers, whereas the Go functions here simply return false for a
// length mismatch; callers that want Node's throwing semantics can check lengths
// themselves before calling.
package timingsafe

import "crypto/subtle"

// Equal reports whether a and b are equal using a constant-time comparison.
func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

// EqualString reports whether a and b are equal using a constant-time comparison.
func EqualString(a, b string) bool {
	return Equal([]byte(a), []byte(b))
}
