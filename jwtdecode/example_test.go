package jwtdecode_test

import (
	"fmt"

	"github.com/malcolmston/express/jwtdecode"
)

// token is the canonical HS256 example JWT whose payload carries the claims
// {"sub":"1234567890","name":"John Doe","iat":1516239022}. It is used unchanged
// across the examples below so their output is deterministic.
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
	"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

// ExampleDecode reads the payload of a JWT without verifying its signature. The
// token is split on its dots, the second (payload) segment is base64url-decoded,
// and the resulting JSON is unmarshalled into a map of claims. String claims such
// as "sub" and "name" come back as Go strings, while the numeric "iat" claim
// unmarshals into a float64 because that is how encoding/json represents JSON
// numbers, so it is printed with a %.0f verb to show it as the integer timestamp
// it represents. Because no signature check is performed, this decoded result
// must never be trusted for authorization.
func ExampleDecode() {
	claims, err := jwtdecode.Decode(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(claims["sub"])
	fmt.Println(claims["name"])
	fmt.Printf("%.0f\n", claims["iat"])
	// Output:
	// 1234567890
	// John Doe
	// 1516239022
}

// ExampleDecodeHeader decodes the first segment of the same token, the JWT
// header, again without any cryptographic verification. The header describes how
// the token was signed rather than who it is about, so it typically contains the
// signing algorithm "alg" and the token type "typ". Reading it is useful for
// deciding which key or verifier to hand a token to before validation. As with
// Decode, the values are returned in a plain map and the header contents are not
// trusted or checked here.
func ExampleDecodeHeader() {
	header, err := jwtdecode.DecodeHeader(token)
	if err != nil {
		panic(err)
	}
	fmt.Println(header["alg"])
	fmt.Println(header["typ"])
	// Output:
	// HS256
	// JWT
}

// ExampleDecode_invalid shows the error path. A string that does not split into
// exactly three dot-separated segments, or whose selected segment is not valid
// base64url or valid JSON, is rejected as malformed. Here a token with only two
// segments is passed, so Decode returns the sentinel ErrInvalidToken and a nil
// map. Callers should always check the returned error before reading claims,
// since jwtdecode validates structure even though it never validates the
// signature.
func ExampleDecode_invalid() {
	_, err := jwtdecode.Decode("not.a-valid-jwt")
	fmt.Println(err == jwtdecode.ErrInvalidToken)
	// Output: true
}
