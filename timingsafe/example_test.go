package timingsafe_test

import (
	"fmt"

	"github.com/malcolmston/express/timingsafe"
)

// ExampleEqual compares two byte slices in constant time. It returns true when
// the contents are identical and false otherwise, but crucially its running time
// depends only on the input length, not on where the first differing byte is, so
// it cannot leak a secret through timing measurement. Inputs of different lengths
// return false immediately. This is the comparison to use when checking an HMAC,
// token, or API key supplied by an untrusted caller. Using == or bytes.Equal for
// such a check would be an exploitable vulnerability.
func ExampleEqual() {
	fmt.Println(timingsafe.Equal([]byte("secret"), []byte("secret")))
	fmt.Println(timingsafe.Equal([]byte("secret"), []byte("sekret")))
	fmt.Println(timingsafe.Equal([]byte("secret"), []byte("secretx")))
	// Output:
	// true
	// false
	// false
}

// ExampleEqualString is the string convenience wrapper, giving the same
// constant-time guarantee for string inputs by converting them to bytes and
// delegating to Equal. It is the natural choice for comparing a user-supplied
// token against an expected value held by the server. Equal contents yield true
// and any difference yields false. As with Equal, a length mismatch simply
// returns false rather than throwing, unlike Node's crypto.timingSafeEqual. The
// length of a secret is not generally considered sensitive, so only the content
// comparison is protected.
func ExampleEqualString() {
	fmt.Println(timingsafe.EqualString("token-abc", "token-abc"))
	fmt.Println(timingsafe.EqualString("token-abc", "token-xyz"))
	// Output:
	// true
	// false
}
