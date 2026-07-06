package cookiesignature_test

import (
	"fmt"

	"github.com/malcolmston/express/cookiesignature"
)

// ExampleSign appends an HMAC-SHA256 signature to a value, producing the
// "value.signature" form. The signature is the base64-encoded digest with
// trailing '=' padding removed, keyed by the secret. Signing is deterministic:
// the same value and secret always produce the same output. It does not hide the
// value; it makes tampering with it detectable.
func ExampleSign() {
	fmt.Println(cookiesignature.Sign("hello", "secret"))
	// Output: hello.iKqz7ejTrflNJquQ07r9SiCDBww7zOnAFO4EpEOEfAs
}

// ExampleUnsign verifies a signed value and recovers the original. It splits on
// the last '.', re-signs the recovered value with the same secret, and compares
// in constant time. It returns the value and true on success, or "" and false
// when the signature does not match or the input is malformed. Callers must
// check the boolean rather than the emptiness of the returned string.
func ExampleUnsign() {
	signed := cookiesignature.Sign("hello", "secret")
	value, ok := cookiesignature.Unsign(signed, "secret")
	fmt.Println(value, ok)

	_, ok = cookiesignature.Unsign(signed, "wrong-secret")
	fmt.Println(ok)
	// Output:
	// hello true
	// false
}
