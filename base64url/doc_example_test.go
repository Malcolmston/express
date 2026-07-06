package base64url_test

import (
	"fmt"

	"github.com/malcolmston/express/base64url"
)

// ExampleEncodeString encodes a string using the RFC 4648 url-safe alphabet
// without padding. Unlike standard base64 it uses "-" and "_" in place of "+"
// and "/" and emits no trailing "=", so the result can be dropped into a URL,
// an HTTP header, or a JWT segment without further escaping. Here a string full
// of angle brackets encodes to a url-safe token. This is the encoding used for
// JWT headers and payloads.
func ExampleEncodeString() {
	fmt.Println(base64url.EncodeString("<<hello>>"))
	// Output: PDxoZWxsbz4-
}

// ExampleDecodeString decodes a url-safe base64 string back into text. It trims
// any trailing "=" first, so it accepts both the unpadded output of this package
// and padded strings copied from a standard-base64 source. Input outside the
// url-safe alphabet or of invalid length returns a non-nil error together with
// an empty string. Here the token round-trips back to the original value.
func ExampleDecodeString() {
	s, err := base64url.DecodeString("PDxoZWxsbz4-")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(s)
	// Output: <<hello>>
}

// ExampleToBase64 converts a base64url string into standard base64. It replaces
// "-" and "_" with "+" and "/" and re-appends as many "=" characters as are
// needed to make the length a multiple of four. This is a pure character
// rewrite and does not validate the input. Here a url-safe encoding of "Hi!" is
// turned into its padded standard-base64 equivalent.
func ExampleToBase64() {
	urlSafe := base64url.EncodeString("Hi!")
	fmt.Println(base64url.ToBase64(urlSafe))
	// Output: SGkh
}
