package keygrip_test

import (
	"fmt"

	"github.com/malcolmston/express/keygrip"
)

// ExampleKeygrip demonstrates the basic sign-then-verify round trip. A Keygrip
// is created from a list of secret keys; Sign computes an HMAC-SHA256 digest of
// the data using the first (current) key and encodes it as unpadded base64url so
// it is safe to place directly in a cookie or URL. Verify recomputes the digest
// and compares it in constant time, reporting whether the data and digest belong
// together. Index goes a step further and reports which key in the rotation
// produced the digest, which is 0 here because the current key signed it. The
// digest itself is not printed because its exact bytes are an implementation
// detail of the keys chosen.
func ExampleKeygrip() {
	kg := keygrip.New([]string{"current-secret", "previous-secret"})

	digest := kg.Sign("user=42")
	fmt.Println(kg.Verify("user=42", digest))
	fmt.Println(kg.Index("user=42", digest))
	// Output:
	// true
	// 0
}

// ExampleKeygrip_rotation shows why Keygrip keeps a list of keys rather than a
// single secret. A digest is first produced by a keygrip whose only key is the
// now-retired "previous-secret", simulating a token that was signed before a
// rotation. A second keygrip is then configured with a freshly prepended
// "current-secret" ahead of the old one. Verify still accepts the old digest
// because every key is tried during verification, and Index returns 1 to reveal
// that the match came from the second (older) key. An application can use that
// non-zero index as a signal to re-sign the token under the current key.
func ExampleKeygrip_rotation() {
	oldDigest := keygrip.New([]string{"previous-secret"}).Sign("user=42")

	rotated := keygrip.New([]string{"current-secret", "previous-secret"})
	fmt.Println(rotated.Verify("user=42", oldDigest))
	fmt.Println(rotated.Index("user=42", oldDigest))
	// Output:
	// true
	// 1
}
