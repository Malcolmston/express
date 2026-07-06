package jsonwebtoken_test

import (
	"fmt"

	"github.com/malcolmston/express/jsonwebtoken"
)

// ExampleSign signs a set of claims and then verifies the resulting token to
// read them back. A nil *SignOptions selects the defaults, signing an HS256
// token whose only automatically added claim is iat (issued-at). Because the
// token embeds that issued-at timestamp, the raw token string is not stable, so
// this example round-trips through Verify and prints individual claim values
// instead of the token bytes. Verify recomputes the HMAC signature with the
// shared secret and compares it in constant time before returning the claims.
// String claims come back as Go strings, exactly as they were signed.
func ExampleSign() {
	secret := []byte("my-fixed-secret")
	token, err := jsonwebtoken.Sign(jsonwebtoken.Claims{
		"name": "John Doe",
		"role": "admin",
	}, secret, nil)
	if err != nil {
		fmt.Println("sign error:", err)
		return
	}

	claims, err := jsonwebtoken.Verify(token, secret)
	if err != nil {
		fmt.Println("verify error:", err)
		return
	}
	fmt.Println(claims["name"])
	fmt.Println(claims["role"])
	// Output:
	// John Doe
	// admin
}

// ExampleDecode reads a token's payload WITHOUT verifying its signature. Decode
// splits the token, base64url-decodes the payload segment, and unmarshals it as
// JSON, ignoring the signature entirely. It is convenient for inspecting a token
// you have not validated, but its result must never be trusted for authorization.
// Here the token is signed only so the example is self-contained; a real caller
// might receive it from elsewhere. The subject claim is printed to show that the
// payload is readable without the secret.
func ExampleDecode() {
	secret := []byte("my-fixed-secret")
	token, _ := jsonwebtoken.Sign(jsonwebtoken.Claims{
		"sub": "user-42",
	}, secret, nil)

	claims, err := jsonwebtoken.Decode(token)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}
	fmt.Println(claims["sub"])
	// Output: user-42
}

// ExampleVerify demonstrates that verification fails when the secret is wrong.
// Verify recomputes the expected signature from the header and payload using the
// supplied secret and compares it against the token's signature with a constant-
// time comparison. A token signed with one secret cannot be verified with a
// different secret, so Verify returns ErrInvalidSignature. The returned error is
// a sentinel value suitable for errors.Is checks. This is the core guarantee
// that makes a signed JWT trustworthy.
func ExampleVerify() {
	token, _ := jsonwebtoken.Sign(jsonwebtoken.Claims{
		"name": "Alice",
	}, []byte("correct-secret"), nil)

	_, err := jsonwebtoken.Verify(token, []byte("wrong-secret"))
	fmt.Println(err)
	// Output: jsonwebtoken: invalid signature
}
