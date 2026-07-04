// Package sha256hex provides hex-encoded hashing and HMAC helpers.
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
