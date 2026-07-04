// Package keygrip provides signing and verification of data using a rotating
// list of secret keys, mirroring the behavior of the npm keygrip library.
package keygrip

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
)

// Keygrip holds an ordered list of secret keys. The first key is used for
// signing, while all keys are tried during verification to support key
// rotation.
type Keygrip struct {
	keys []string
}

// New creates a Keygrip from the given keys. The first key is used to sign new
// data; older keys remain valid for verification. New panics if keys is empty.
func New(keys []string) *Keygrip {
	if len(keys) == 0 {
		panic("keygrip: keys must not be empty")
	}
	cp := make([]string, len(keys))
	copy(cp, keys)
	return &Keygrip{keys: cp}
}

// sign computes the base64url (unpadded) HMAC-SHA256 digest of data using key.
func sign(data, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Sign returns the digest of data computed with the current (first) key.
func (k *Keygrip) Sign(data string) string {
	return sign(data, k.keys[0])
}

// Index returns the index of the first key that produces digest for data,
// using a constant-time comparison, or -1 if no key matches.
func (k *Keygrip) Index(data, digest string) int {
	for i, key := range k.keys {
		if hmac.Equal([]byte(digest), []byte(sign(data, key))) {
			return i
		}
	}
	return -1
}

// Verify reports whether digest is a valid signature of data for any key.
func (k *Keygrip) Verify(data, digest string) bool {
	return k.Index(data, digest) >= 0
}
