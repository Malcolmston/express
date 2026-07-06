// Package hotp implements HMAC-based one-time passwords as defined in RFC 4226,
// used for counter-based two-factor authentication codes. It is a
// standard-library port of the npm "hotp" / "otplib" style HOTP primitive,
// exposing the two operations applications actually need: Generate, which
// derives a one-time code from a shared secret and a counter, and Verify,
// which checks a user-supplied code against the expected value.
//
// HOTP was standardized by the IETF in RFC 4226 (2005) as the foundation of the
// Initiative for Open Authentication (OATH). It exists so that a server and a
// hardware token or mobile app, sharing only a secret key, can independently
// arrive at the same short numeric code without any network round trip. HOTP is
// the counter-based cousin of TOTP: where TOTP feeds the current time window
// into the algorithm, HOTP feeds a monotonically increasing counter that both
// sides advance on each successful use. This port targets the same behavior as
// the JavaScript reference implementations and validates against the RFC 4226
// Appendix D test vectors.
//
// The algorithm works by computing an HMAC-SHA1 of the 8-byte big-endian
// counter using the secret as the key, then applying the "dynamic truncation"
// described in RFC 4226 section 5.3. The low four bits of the final HMAC byte
// select an offset into the 20-byte digest; four bytes are read from that
// offset, the top bit is masked off to avoid sign ambiguity, and the resulting
// 31-bit integer is reduced modulo 10^digits. The remainder is formatted as a
// zero-padded decimal string, so a 6-digit code such as "007236" keeps its
// leading zeros. HMAC-SHA1 is mandated by the standard here even though SHA-1
// is deprecated for general hashing; HOTP's security rests on the HMAC
// construction and the secrecy of the key, not on SHA-1's collision
// resistance.
//
// Generate takes the raw secret bytes, the counter and the desired number of
// digits, and returns the code. The digits argument follows the common default:
// any value less than or equal to zero is treated as 6, the canonical HOTP
// length, while callers may request other lengths (typically 6 to 8). The
// secret is used exactly as provided; if your secret is stored in Base32 (as it
// usually is in authenticator apps), decode it before calling Generate. Verify
// recomputes the expected code with the same parameters and compares it against
// the candidate using crypto/subtle.ConstantTimeCompare, so a mismatch cannot
// be distinguished from a match by timing. A wrong code, or a code computed for
// a different counter or digit length, causes Verify to report false.
//
// Parity with the Node ecosystem is deliberate but narrow. This package
// implements only the standard HMAC-SHA1 HOTP path and offers no counter
// resynchronization window, no configurable hash algorithm, and no Base32
// handling or otpauth:// URI generation; those concerns are left to the caller.
// Because Verify checks a single counter value, servers that must tolerate a
// token drifting ahead should call Verify across a small look-ahead range of
// counters and, on success, advance their stored counter past the matched
// value to prevent code reuse.
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
