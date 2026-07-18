// Package base62 is a standard-library-only Go implementation of Base62
// encoding, the alphanumeric (0-9, A-Z, a-z) binary-to-text scheme used by URL
// shorteners and short-id generators across the npm ecosystem (base62,
// base-x). Because the alphabet contains only characters that are safe in URLs
// and identifiers, Base62 is a common choice for compact, human-friendly tokens.
//
// Encode and Decode convert between arbitrary byte slices and Base62 strings,
// preserving leading zero bytes as leading '0' characters. EncodeInt and
// DecodeInt convert unsigned 64-bit integers, the form most short-id schemes
// use. Decode and DecodeInt return ErrInvalidCharacter for input outside the
// alphabet. The implementation uses math/big for the base conversion, is
// deterministic, and depends only on the standard library.
package base62

import (
	"errors"
	"math/big"
)

// Alphabet is the standard Base62 alphabet: digits, then uppercase, then
// lowercase letters.
const Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// ErrInvalidCharacter is returned when decoding input containing a character
// outside the Base62 alphabet.
var ErrInvalidCharacter = errors.New("base62: invalid character")

var bigRadix = big.NewInt(62)

// Encode returns the Base62 encoding of input. Leading zero bytes are encoded
// as leading '0' characters, and empty input encodes to the empty string.
func Encode(input []byte) string {
	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}
	x := new(big.Int).SetBytes(input)
	mod := new(big.Int)
	buf := make([]byte, 0, len(input)*137/100+1)
	for x.Sign() > 0 {
		x.DivMod(x, bigRadix, mod)
		buf = append(buf, Alphabet[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		buf = append(buf, Alphabet[0])
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// Decode returns the bytes represented by the Base62-encoded string s. Leading
// '0' characters decode to leading zero bytes. It returns ErrInvalidCharacter
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
	for zeros < len(s) && s[zeros] == Alphabet[0] {
		zeros++
	}
	out := make([]byte, zeros+len(decoded))
	copy(out[zeros:], decoded)
	return out, nil
}

// EncodeInt returns the Base62 encoding of an unsigned integer. Zero encodes to
// "0".
func EncodeInt(n uint64) string {
	if n == 0 {
		return string(Alphabet[0])
	}
	var buf []byte
	for n > 0 {
		buf = append(buf, Alphabet[n%62])
		n /= 62
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

// DecodeInt returns the unsigned integer represented by the Base62 string s. It
// returns ErrInvalidCharacter for input outside the alphabet.
func DecodeInt(s string) (uint64, error) {
	var n uint64
	for i := 0; i < len(s); i++ {
		idx := indexOfChar(s[i])
		if idx < 0 {
			return 0, ErrInvalidCharacter
		}
		n = n*62 + uint64(idx)
	}
	return n, nil
}

func indexOfChar(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'Z':
		return int(c-'A') + 10
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 36
	default:
		return -1
	}
}
