package hotp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"math"
)

// Generate computes an RFC 4226 HOTP value for the given secret and counter,
// returning a zero-padded decimal string of the requested number of digits.
func Generate(secret []byte, counter uint64, digits int) string {
	if digits <= 0 {
		digits = 6
	}

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	mac := hmac.New(sha1.New, secret)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	// Dynamic truncation (RFC 4226 section 5.3).
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	mod := uint32(math.Pow10(digits))
	otp := binCode % mod

	return fmt.Sprintf("%0*d", digits, otp)
}

// Verify reports whether code matches the HOTP value for the secret and counter.
func Verify(secret []byte, counter uint64, code string, digits int) bool {
	expected := Generate(secret, counter, digits)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1
}
