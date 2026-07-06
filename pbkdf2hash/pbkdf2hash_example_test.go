package pbkdf2hash_test

import (
	"encoding/hex"
	"fmt"

	"github.com/malcolmston/express/pbkdf2hash"
)

// ExampleDeriveKey shows the raw PBKDF2-HMAC-SHA256 primitive. Because DeriveKey
// takes an explicit salt rather than generating a random one, its output is fully
// deterministic and can be asserted byte-for-byte. Here we derive a 32-byte key
// from the password "password" and the salt "salt" using a single iteration,
// which matches the well-known PBKDF2 test vector for these inputs. The result is
// printed as hex so the exact key material is visible. This is the building block
// that Hash and Verify layer their encoding on top of.
func ExampleDeriveKey() {
	key := pbkdf2hash.DeriveKey([]byte("password"), []byte("salt"), 1, 32)
	fmt.Println(hex.EncodeToString(key))
	// Output: 120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b
}

// ExampleHash demonstrates the encode-then-verify round trip. Hash embeds a fresh
// random 16-byte salt in every call, so the encoded string itself is different
// each time and must not be asserted directly. Instead we hash a password and
// immediately verify both the correct password and an incorrect one against the
// resulting string. Verify re-derives the key from the candidate password and the
// salt stored inside the encoded hash, comparing in constant time. The two
// boolean results are deterministic, so they make a stable example output.
func ExampleHash() {
	encoded, err := pbkdf2hash.Hash("s3cret", 1000)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(pbkdf2hash.Verify("s3cret", encoded))
	fmt.Println(pbkdf2hash.Verify("wrong", encoded))
	// Output:
	// true
	// false
}

// ExampleVerify checks a password against a pre-computed encoded hash string. The
// encoded value uses the Django-style "pbkdf2_sha256$<iterations>$<hexsalt>$<hexkey>"
// layout, carrying every parameter needed for verification in one self-describing
// field. Because the salt and iteration count are fixed inside the string, Verify
// is completely deterministic for a given password. A malformed encoding — wrong
// field count, wrong algorithm tag, or non-hex data — makes Verify return false
// rather than erroring. Here the first call supplies the right password and the
// second supplies a garbage encoding.
func ExampleVerify() {
	// Encoded hash for the password "hunter2" (1 iteration, fixed salt).
	salt := []byte("staticsalt1234!!")
	key := pbkdf2hash.DeriveKey([]byte("hunter2"), salt, 1, 32)
	encoded := fmt.Sprintf("pbkdf2_sha256$1$%s$%s",
		hex.EncodeToString(salt), hex.EncodeToString(key))

	fmt.Println(pbkdf2hash.Verify("hunter2", encoded))
	fmt.Println(pbkdf2hash.Verify("hunter2", "not-a-valid-hash"))
	// Output:
	// true
	// false
}
