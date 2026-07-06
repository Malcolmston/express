package encodeurl_test

import (
	"fmt"

	"github.com/malcolmston/express/encodeurl"
)

// ExampleEncode percent-encodes the unsafe characters in a URL while leaving
// valid structure intact. Characters such as the space become "%20", but
// reserved URL punctuation like '/', '?', '&', and '=' is preserved so the
// URL's shape survives. Existing valid escapes such as "%20" are passed through
// unchanged rather than being doubled into "%2520". This is what Express uses to
// safely reflect a request path into a header.
func ExampleEncode() {
	fmt.Println(encodeurl.Encode("/foo/bar?msg=hello world&x=1%202"))
	// Output: /foo/bar?msg=hello%20world&x=1%202
}

// ExampleEncode_strayPercent shows the special handling of the '%' character. A
// '%' that begins a valid escape is preserved, but a '%' that is not the start
// of a valid two-hex-digit escape is itself encoded to "%25". Here the "%zz" is
// not a valid escape, so its percent sign is encoded while the valid "%20" is
// left alone. This prevents both corruption and accidental double-encoding.
func ExampleEncode_strayPercent() {
	fmt.Println(encodeurl.Encode("a%zzb%20c"))
	// Output: a%25zzb%20c
}
