// Package base58 is a standard-library-only Go implementation of Base58 and
// Base58Check encoding, the compact, ambiguity-free binary-to-text encoding
// popularised by Bitcoin and used across the npm ecosystem (bs58, base58check,
// bs58check) for keys, addresses and short identifiers. Base58 uses the digits
// and letters excluding the visually confusable 0, O, I and l, so encoded
// strings survive hand transcription.
//
// Encode and Decode implement the raw Bitcoin-alphabet encoding, preserving
// leading zero bytes as leading '1' characters. CheckEncode and CheckDecode
// implement Base58Check: they prepend a one-byte version, append a four-byte
// checksum (the first four bytes of the double SHA-256 of the version-and-
// payload), and verify that checksum on decode, returning ErrInvalidChecksum
// when it does not match. Decode returns ErrInvalidCharacter for any character
// outside the alphabet.
//
// The implementation uses math/big for the base conversion and crypto/sha256
// for the checksum, is deterministic, and depends only on the standard library.
package base58

import (
	"crypto/sha256"
	"errors"
	"math/big"
)

// alphabet is the Bitcoin Base58 alphabet.
const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

// ErrInvalidCharacter is returned by Decode when the input contains a character
// that is not part of the Base58 alphabet.
var ErrInvalidCharacter = errors.New("base58: invalid character")

// ErrInvalidChecksum is returned by CheckDecode when the trailing checksum does
// not match the payload.
var ErrInvalidChecksum = errors.New("base58: invalid checksum")

// ErrTooShort is returned by CheckDecode when the decoded data is too short to
// contain a version byte and a four-byte checksum.
var ErrTooShort = errors.New("base58: decoded data too short for checksum")

var bigRadix = big.NewInt(58)

// Encode returns the Base58 encoding of input. Leading zero bytes are encoded
// as leading '1' characters.
func Encode(input []byte) string {
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}
	x := new(big.Int).SetBytes(input)
	mod := new(big.Int)
	buf := make([]byte, 0, len(input)*138/100+1)
	for x.Sign() > 0 {
		x.DivMod(x, bigRadix, mod)
		buf = append(buf, alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		buf = append(buf, alphabet[0])
	}
	// reverse in place
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// Decode returns the bytes represented by the Base58-encoded string s. Leading
// '1' characters decode to leading zero bytes. It returns ErrInvalidCharacter
// for any character outside the alphabet.
func Decode(s string) ([]byte, error) {
	x := new(big.Int)
	for i := 0; i < len(s); i++ {
		idx := indexOfChar(s[i])
		if idx < 0 {
			return nil, ErrInvalidCharacter
		}
		x.Mul(x, bigRadix)
		x.Add(x, big.NewInt(int64(idx)))
	}
	decoded := x.Bytes()
	zeros := 0
	for zeros < len(s) && s[zeros] == alphabet[0] {
		zeros++
	}
	out := make([]byte, zeros+len(decoded))
	copy(out[zeros:], decoded)
	return out, nil
}

// CheckEncode returns the Base58Check encoding of payload with the given version
// byte: it encodes version||payload||checksum, where checksum is the first four
// bytes of the double SHA-256 of version||payload.
func CheckEncode(payload []byte, version byte) string {
	body := make([]byte, 0, 1+len(payload)+4)
	body = append(body, version)
	body = append(body, payload...)
	sum := doubleSHA256(body)
	body = append(body, sum[:4]...)
	return Encode(body)
}

// CheckDecode decodes a Base58Check string, verifies its four-byte checksum and
// returns the version byte and the payload. It returns ErrTooShort when the data
// cannot hold a checksum, ErrInvalidChecksum when verification fails, or the
// error from Decode for malformed input.
func CheckDecode(s string) (version byte, payload []byte, err error) {
	decoded, err := Decode(s)
	if err != nil {
		return 0, nil, err
	}
	if len(decoded) < 5 {
		return 0, nil, ErrTooShort
	}
	body := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	sum := doubleSHA256(body)
	for i := 0; i < 4; i++ {
		if checksum[i] != sum[i] {
			return 0, nil, ErrInvalidChecksum
		}
	}
	return body[0], body[1:], nil
}

func doubleSHA256(b []byte) []byte {
	first := sha256.Sum256(b)
	second := sha256.Sum256(first[:])
	return second[:]
}

func indexOfChar(c byte) int {
	for i := 0; i < len(alphabet); i++ {
		if alphabet[i] == c {
			return i
		}
	}
	return -1
}
