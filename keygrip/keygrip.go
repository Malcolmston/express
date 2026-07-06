// Package keygrip provides signing and verification of data using a rotating
// list of secret keys, a stdlib-only Go port of the npm "keygrip" library used
// by frameworks such as Express and Koa to protect signed cookies and other
// tamper-evident tokens. Rather than encrypting data, keygrip attaches a
// keyed message-authentication code (a digest) that a later request can present
// alongside the data so the server can confirm the data was produced by a
// holder of one of its secret keys and has not been altered in transit.
//
// The motivating problem is secret rotation. A long-lived service wants to
// replace its signing secret periodically without invalidating every cookie or
// token that is still in the wild. Keygrip solves this by holding an ordered
// list of keys: the newest key sits at the front and signs all new data, while
// the older keys are retained only so that previously issued digests still
// verify. When it is time to rotate, the operator prepends a fresh key and
// eventually drops the oldest one from the tail, and outstanding tokens signed
// under the retiring key keep working until they expire.
//
// A digest is computed by sign, which runs HMAC-SHA256 over the data using a
// key as the MAC secret and encodes the result with unpadded base64url
// (base64.RawURLEncoding). The URL-safe, padding-free encoding means a digest
// is safe to place directly in a cookie value or URL: it never contains '=',
// '+', or '/'. Sign always uses the first (current) key, so every freshly
// produced digest reflects the newest secret.
//
// Verification is where rotation shows through. Index walks the key list from
// front to back, i.e. newest to oldest, recomputing the digest under each key
// and comparing it to the supplied digest with crypto/hmac.Equal, a
// constant-time comparison that does not leak how many leading bytes matched.
// It returns the index of the first key that reproduces the digest, or -1 when
// none do; that index tells the caller how far down the rotation the matching
// key sits, so an application can, for example, re-sign a token whenever it
// verifies under a non-zero index to migrate it onto the current key. Verify is
// a thin boolean wrapper that reports whether Index found any match.
//
// The construction and semantics track the Node original closely. New defends
// against caller-side mutation by copying the supplied slice, so later changes
// to the caller's array do not affect signing, and it panics when given an
// empty key list because a keygrip with no keys can neither sign nor verify.
// The chief intentional divergence from Node is that this port fixes the digest
// algorithm to HMAC-SHA256 with base64url output, whereas the JavaScript
// library lets the caller configure the hash and encoding; callers needing a
// different scheme should adapt sign accordingly.
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
