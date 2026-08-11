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
//
// The Node constructor takes two further arguments, keygrip(keys, algorithm,
// encoding), which NewWithOptions mirrors: Algorithm selects the HMAC hash
// ("sha1", "sha256", "sha384", "sha512" or "md5") and Encoding how the digest
// is rendered ("base64", the Node default, "base64url" — the same string, since
// Node rewrites '+' and '/' and strips '=' — or "hex"). The one intentional
// divergence from Node is the default hash: New, and NewWithOptions with an
// empty Algorithm, use HMAC-SHA256, whereas the JavaScript library defaults to
// the weaker SHA-1. Pass Algorithm: "sha1" explicitly to interoperate with
// digests produced by a Node service that relies on that default.
package keygrip

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"strings"
)

// Keygrip holds an ordered list of secret keys. The first key is used for
// signing, while all keys are tried during verification to support key
// rotation.
type Keygrip struct {
	keys    []string
	newHash func() hash.Hash
	encode  func([]byte) string
}

// Options configures the digest scheme, mirroring the second and third
// arguments of the Node constructor keygrip(keys, algorithm, encoding).
type Options struct {
	// Algorithm is the HMAC hash: "sha1", "sha256", "sha384", "sha512" or
	// "md5", case-insensitive. Empty means "sha256" (see the package comment:
	// Node defaults to "sha1" instead).
	Algorithm string
	// Encoding is how the digest bytes are rendered: "base64" (the Node
	// default), "base64url" or "hex", case-insensitive. Empty means "base64".
	// Node rewrites '+' to '-', '/' to '_' and strips '=' from its base64
	// output, so "base64" and "base64url" produce the same unpadded base64url
	// string here.
	Encoding string
}

func hasherFor(algorithm string) (func() hash.Hash, error) {
	switch strings.ToLower(strings.TrimSpace(algorithm)) {
	case "", "sha256":
		return sha256.New, nil
	case "sha1":
		return sha1.New, nil
	case "sha384":
		return sha512.New384, nil
	case "sha512":
		return sha512.New, nil
	case "md5":
		return md5.New, nil
	default:
		return nil, fmt.Errorf("keygrip: unsupported algorithm %q", algorithm)
	}
}

func encoderFor(encoding string) (func([]byte) string, error) {
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "", "base64", "base64url":
		return base64.RawURLEncoding.EncodeToString, nil
	case "hex":
		return hex.EncodeToString, nil
	default:
		return nil, fmt.Errorf("keygrip: unsupported encoding %q", encoding)
	}
}

// New creates a Keygrip from the given keys, signing with HMAC-SHA256 and
// unpadded base64url. The first key is used to sign new data; older keys remain
// valid for verification. New panics if keys is empty.
func New(keys []string) *Keygrip {
	k, err := NewWithOptions(keys, Options{})
	if err != nil {
		panic(err)
	}
	return k
}

// NewWithOptions creates a Keygrip with an explicit digest scheme, the
// equivalent of the Node constructor keygrip(keys, algorithm, encoding). It
// returns an error for an unrecognised algorithm or encoding, and panics if
// keys is empty, as New does.
func NewWithOptions(keys []string, opts Options) (*Keygrip, error) {
	if len(keys) == 0 {
		panic("keygrip: keys must not be empty")
	}
	newHash, err := hasherFor(opts.Algorithm)
	if err != nil {
		return nil, err
	}
	encode, err := encoderFor(opts.Encoding)
	if err != nil {
		return nil, err
	}
	cp := make([]string, len(keys))
	copy(cp, keys)
	return &Keygrip{keys: cp, newHash: newHash, encode: encode}, nil
}

// sign computes the digest of data using key under the receiver's scheme.
func (k *Keygrip) sign(data, key string) string {
	mac := hmac.New(k.newHash, []byte(key))
	mac.Write([]byte(data))
	return k.encode(mac.Sum(nil))
}

// Sign returns the digest of data computed with the current (first) key.
func (k *Keygrip) Sign(data string) string {
	return k.sign(data, k.keys[0])
}

// Index returns the index of the first key that produces digest for data,
// using a constant-time comparison, or -1 if no key matches.
func (k *Keygrip) Index(data, digest string) int {
	for i, key := range k.keys {
		if hmac.Equal([]byte(digest), []byte(k.sign(data, key))) {
			return i
		}
	}
	return -1
}

// Verify reports whether digest is a valid signature of data for any key.
func (k *Keygrip) Verify(data, digest string) bool {
	return k.Index(data, digest) >= 0
}
