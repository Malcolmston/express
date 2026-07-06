package totp_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/totp"
)

// ExampleGenerateAt produces a time-based one-time password for an explicit
// instant, which makes the example deterministic rather than depending on the
// current wall clock. The secret, time, and expected code are the RFC 6238
// Appendix B test vector for SHA-1 with eight digits: at Unix time 59 seconds
// the code is "94287082". GenerateAt divides the Unix time by the period to form
// the counter, applies HMAC keyed by the decoded Base32 secret, and truncates
// the result to the requested number of digits. Because both sides derive the
// code from the shared secret and the same clock, no network round trip is
// needed. This is exactly what an authenticator app computes for the same
// inputs.
func ExampleGenerateAt() {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	code, err := totp.GenerateAt(secret, time.Unix(59, 0).UTC(), &totp.Options{Digits: 8})
	if err != nil {
		panic(err)
	}
	fmt.Println(code)
	// Output: 94287082
}

// ExampleVerify checks a user-supplied code against the expected value for a
// window of time steps around a reference instant. To stay deterministic the
// example first computes the code for a fixed time with GenerateAt, then
// verifies it. In real use Verify reads the current clock and a window parameter
// tolerates clock skew: a window of 1 accepts the previous, current, and next
// steps. Each candidate is compared in constant time so a near-miss cannot be
// distinguished from a wrong code by timing. Here the freshly generated code
// verifies successfully while an obviously wrong code does not.
func ExampleVerify() {
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	code, _ := totp.GenerateAt(secret, time.Now(), nil)
	fmt.Println(totp.Verify(secret, code, nil, 1))
	// A 7-character code can never equal a 6-digit code, so this is always false.
	fmt.Println(totp.Verify(secret, "0000000", nil, 1))
	// Output:
	// true
	// false
}
