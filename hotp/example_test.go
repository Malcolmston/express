package hotp_test

import (
	"fmt"

	"github.com/malcolmston/express/hotp"
)

// ExampleGenerate derives an RFC 4226 one-time password from a shared secret
// and a counter. The secret is used exactly as provided, so a Base32-encoded
// secret from an authenticator app must be decoded to raw bytes first. The
// counter is a value both the client and server track and advance on each use.
// Requesting 6 digits produces the canonical HOTP length, and the result is
// zero-padded so it always has exactly that many characters. This example uses
// the RFC 4226 Appendix D test secret, whose counter 0 code is "755224".
func ExampleGenerate() {
	secret := []byte("12345678901234567890")
	code := hotp.Generate(secret, 0, 6)
	fmt.Println(code)
	// Output: 755224
}

// ExampleVerify checks a user-supplied code against the expected value for a
// given secret and counter. Verify recomputes the code with Generate and
// compares the two in constant time, so a mismatch cannot be distinguished
// from a match by timing. It returns true only when the candidate matches the
// code for that exact secret, counter and digit length. Here the correct code
// for counter 0 verifies successfully while an obviously wrong code does not.
func ExampleVerify() {
	secret := []byte("12345678901234567890")
	fmt.Println(hotp.Verify(secret, 0, "755224", 6))
	fmt.Println(hotp.Verify(secret, 0, "000000", 6))
	// Output:
	// true
	// false
}
