package sha256hex_test

import (
	"fmt"

	"github.com/malcolmston/express/sha256hex"
)

// ExampleSHA256String hashes a string with SHA-256 and returns the digest as a
// 64-character lowercase hexadecimal string. The value shown is the well-known
// SHA-256 digest of "abc" and matches what Node's
// crypto.createHash("sha256").update("abc").digest("hex") produces. The String
// variant simply converts its argument to bytes and delegates to SHA256, so the
// two forms are equivalent for UTF-8 text. The lowercase hex output is safe to
// embed in URLs, filenames, and cache keys without escaping.
func ExampleSHA256String() {
	fmt.Println(sha256hex.SHA256String("abc"))
	// Output: ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad
}

// ExampleMD5String hashes a string with MD5 and returns a 32-character lowercase
// hex digest. The value is the classic MD5 of "abc" and matches Node's
// crypto md5 hex output byte for byte. MD5 is provided for legacy interop and
// checksums only; it is not collision-resistant and should not be used for
// security-sensitive work. Prefer SHA-256 or HMAC-SHA256 for anything that must
// resist tampering. The digest is deterministic for a given input.
func ExampleMD5String() {
	fmt.Println(sha256hex.MD5String("abc"))
	// Output: 900150983cd24fb0d6963f7d28e17f72
}

// ExampleHMACSHA256String computes a keyed HMAC-SHA256 and returns it as
// lowercase hex. Unlike a bare hash, an HMAC mixes in a secret key so the
// result cannot be recomputed by anyone who does not know the key, which is why
// it is used for signing tokens and verifying webhook payloads. The value here
// is the standard HMAC-SHA256 test vector for the key "key" over the pangram
// message. The output matches Node's crypto.createHmac equivalent. Digests
// should be compared with a constant-time function when verifying signatures.
func ExampleHMACSHA256String() {
	fmt.Println(sha256hex.HMACSHA256String("key", "The quick brown fox jumps over the lazy dog"))
	// Output: f7bc83f430538424b13298e6aa6fb143ef4d59a14946175997479dbc2d1a3cd8
}
