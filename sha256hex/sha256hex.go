// Package sha256hex provides small, allocation-light helpers that hash bytes or
// strings and return the digest as a lowercase hexadecimal string. It is a
// stdlib-only port of the "sha256 hex" idiom exposed by JavaScript hashing
// libraries such as js-sha256 (and its sibling js-sha1 / js-md5 packages), whose
// default call form hashes an input and returns its digest already encoded as a
// lowercase hex string. The same output is what Node's built-in
// crypto.createHash("sha256").update(data).digest("hex") produces, so this
// package is a drop-in equivalent for code that was relying on either API.
//
// The primary function is SHA256: it feeds the input through crypto/sha256 and
// hex-encodes the resulting 32-byte digest with encoding/hex, yielding a
// 64-character lowercase string. The package also mirrors the companion
// algorithms those JS libraries ship, offering SHA-1 (40 hex characters) and MD5
// (32 hex characters), plus keyed HMAC-SHA256. Every function has a "String"
// convenience variant that accepts a string argument instead of a []byte; the
// String form simply converts its argument with []byte(s) and delegates, so the
// two forms are exact equivalents for UTF-8 text.
//
// The algorithm is the standard one from the Go standard library: the byte-array
// digest helpers (sha256.Sum256, sha1.Sum, md5.Sum) compute the fixed-size
// digest in one call, and hex.EncodeToString turns each digest byte into two
// lowercase hex characters ('0'-'9', 'a'-'f'). HMACSHA256 constructs an
// hmac.Hash with sha256.New and the supplied key, writes the data, and hex-
// encodes mac.Sum(nil). Because the output alphabet is always lowercase hex, the
// results are safe to embed in URLs, filenames, cache keys, and ETags without
// further escaping.
//
// Semantics and edge cases are deliberately simple and total: none of these
// functions returns an error and none can panic on ordinary input. Hashing nil
// or an empty input is well defined and returns the algorithm's digest of the
// empty message (for example SHA256(nil) is the well-known
// "e3b0c442..." digest). An empty HMAC key is accepted and produces a valid, if
// cryptographically weak, MAC. The input bytes are never mutated. Callers who
// need constant-time comparison of two digests should compare the raw bytes with
// crypto/hmac.Equal rather than comparing the hex strings, though string
// comparison is adequate for non-adversarial equality checks.
//
// Parity with the Node/JS originals is exact for the digest value: for identical
// input bytes these helpers return byte-for-byte the same lowercase hex string
// as js-sha256, js-sha1, js-md5, and crypto.digest("hex"). The differences are
// only in surface API shape, which follows Go conventions: results are returned
// as plain strings rather than through a chainable hasher object, there is no
// incremental update/streaming interface (hash the full message at once), and
// output is always lowercase hex only (the JS ability to request an array,
// ArrayBuffer, or base64 encoding is intentionally omitted, since callers who
// need those can reach for the crypto/* and encoding/* packages directly).
// SHA-1 and MD5 are provided for compatibility and legacy interop only; prefer
// SHA-256 or HMAC-SHA256 for any security-sensitive use.
package sha256hex

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// SHA256 returns the lowercase hex SHA-256 digest of data.
func SHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// SHA256String returns the lowercase hex SHA-256 digest of s.
func SHA256String(s string) string {
	return SHA256([]byte(s))
}

// SHA1 returns the lowercase hex SHA-1 digest of data.
func SHA1(data []byte) string {
	sum := sha1.Sum(data)
	return hex.EncodeToString(sum[:])
}

// SHA1String returns the lowercase hex SHA-1 digest of s.
func SHA1String(s string) string {
	return SHA1([]byte(s))
}

// MD5 returns the lowercase hex MD5 digest of data.
func MD5(data []byte) string {
	sum := md5.Sum(data)
	return hex.EncodeToString(sum[:])
}

// MD5String returns the lowercase hex MD5 digest of s.
func MD5String(s string) string {
	return MD5([]byte(s))
}

// HMACSHA256 returns the lowercase hex HMAC-SHA256 of data using key.
func HMACSHA256(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACSHA256String returns the lowercase hex HMAC-SHA256 of data using key.
func HMACSHA256String(key, data string) string {
	return HMACSHA256([]byte(key), []byte(data))
}
