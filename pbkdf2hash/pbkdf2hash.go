// Package pbkdf2hash implements PBKDF2-HMAC-SHA256 key derivation together with
// a self-describing, textual password-hash encoding. It occupies the same niche
// as the npm pbkdf2-password module and the broader family of PHC/@node-rs-style
// password hashers: it turns a plaintext password into a storable string and can
// later verify a candidate password against that string without keeping the
// original around. Everything here is built on the Go standard library
// (crypto/hmac, crypto/sha256, crypto/rand and crypto/subtle), so it has no
// third-party dependencies.
//
// You reach for this package whenever you need to persist passwords safely.
// Storing plaintext (or a plain SHA-256) is unsafe because a leaked database
// immediately exposes every credential. PBKDF2 defends against that by making
// each guess deliberately expensive: an attacker must repeat many thousands of
// HMAC evaluations per candidate password, which throttles offline brute-force
// and dictionary attacks. The per-hash random salt additionally defeats rainbow
// tables and ensures two users with the same password produce different hashes.
//
// The mechanism is standard PBKDF2 (RFC 2898) with HMAC-SHA256 as the underlying
// pseudo-random function. DeriveKey is the raw primitive: given a password, a
// salt, an iteration count and a desired key length it produces that many bytes
// of derived key material by computing HMAC blocks and XOR-folding each block
// across the requested number of iterations. Higher iteration counts increase
// the cost of every guess (and every legitimate login), so the value is a
// deliberate security/latency trade-off.
//
// Hash and Verify build a Django-compatible encoded string on top of DeriveKey.
// Hash generates a fresh 16-byte cryptographically random salt, derives a 32-byte
// key and returns the string "pbkdf2_sha256$<iterations>$<hexsalt>$<hexkey>". All
// parameters needed to check a password later — the algorithm tag, iteration
// count, salt and expected key — are embedded in that single self-describing
// field, so no side-channel storage is required. If iterations is zero or
// negative Hash substitutes a default of 100000. Verify reverses the process: it
// parses the four dollar-separated fields, re-derives a key of the same length
// from the candidate password and salt, and compares it against the stored key
// using subtle.ConstantTimeCompare so that comparison time does not leak how many
// leading bytes matched.
//
// Edge cases and Node parity: an empty password is accepted and hashed like any
// other byte string. A malformed encoded string — wrong field count, wrong
// algorithm tag, or non-hex salt/key — makes Verify return false rather than
// panicking or erroring. Verify is fully deterministic given the encoded hash, so
// repeated verification of the same password/hash pair always agrees. Compared to
// the Node ecosystem this port fixes the digest to SHA-256 and the encoding to
// the Django "pbkdf2_sha256" layout rather than exposing pluggable digests; the
// numeric key-derivation output of DeriveKey matches PBKDF2 as implemented
// everywhere, so a hash produced here verifies against any conforming PBKDF2
// implementation that is told the same parameters.
package pbkdf2hash

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// DeriveKey derives a key of keyLen bytes from password and salt using
// PBKDF2 with HMAC-SHA256 over the given number of iterations.
func DeriveKey(password, salt []byte, iterations, keyLen int) []byte {
	prf := hmac.New(sha256.New, password)
	hLen := prf.Size()
	numBlocks := (keyLen + hLen - 1) / hLen

	dk := make([]byte, 0, numBlocks*hLen)
	var block [4]byte
	u := make([]byte, hLen)

	for i := 1; i <= numBlocks; i++ {
		block[0] = byte(i >> 24)
		block[1] = byte(i >> 16)
		block[2] = byte(i >> 8)
		block[3] = byte(i)

		prf.Reset()
		prf.Write(salt)
		prf.Write(block[:])
		u = prf.Sum(u[:0])

		t := make([]byte, hLen)
		copy(t, u)

		for n := 2; n <= iterations; n++ {
			prf.Reset()
			prf.Write(u)
			u = prf.Sum(u[:0])
			for j := range t {
				t[j] ^= u[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

// Hash derives a 32-byte key from password with a random 16-byte salt and
// returns an encoded string: pbkdf2_sha256$<iterations>$<hexsalt>$<hexkey>.
func Hash(password string, iterations int) (string, error) {
	if iterations <= 0 {
		iterations = 100000
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := DeriveKey([]byte(password), salt, iterations, 32)
	return fmt.Sprintf("pbkdf2_sha256$%d$%s$%s",
		iterations, hex.EncodeToString(salt), hex.EncodeToString(key)), nil
}

// Verify parses an encoded hash, re-derives the key from password, and
// compares in constant time.
func Verify(password, encoded string) bool {
	iterations, salt, key, err := parse(encoded)
	if err != nil {
		return false
	}
	derived := DeriveKey([]byte(password), salt, iterations, len(key))
	return subtle.ConstantTimeCompare(derived, key) == 1
}

func parse(encoded string) (iterations int, salt, key []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2_sha256" {
		return 0, nil, nil, errors.New("invalid encoded hash format")
	}
	iterations, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, nil, nil, err
	}
	salt, err = hex.DecodeString(parts[2])
	if err != nil {
		return 0, nil, nil, err
	}
	key, err = hex.DecodeString(parts[3])
	if err != nil {
		return 0, nil, nil, err
	}
	return iterations, salt, key, nil
}
