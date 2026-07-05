// Package timingsafe compares byte slices and strings in constant time,
// mirroring Node's crypto.timingSafeEqual to avoid timing side channels.
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
