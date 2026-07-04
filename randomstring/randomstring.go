// Package randomstring is a standard-library port of the npm "randomstring"
// library. It generates random strings from named character-set presets or a
// custom character set using crypto/rand for unbiased sampling.
package randomstring

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	lowerAlpha = "abcdefghijklmnopqrstuvwxyz"
	upperAlpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits     = "0123456789"
)

// presets maps preset names to their concrete character sets.
var presets = map[string]string{
	"alphanumeric": lowerAlpha + upperAlpha + digits,
	"alphabetic":   lowerAlpha + upperAlpha,
	"numeric":      digits,
	"hex":          "0123456789abcdef",
	"binary":       "01",
	"octal":        "01234567",
}

// GenerateFrom returns a random string of the requested length drawn uniformly
// from the literal character set chars using crypto/rand.
func GenerateFrom(length int, chars string) (string, error) {
	if length < 0 {
		return "", errors.New("randomstring: length must be non-negative")
	}
	runes := []rune(chars)
	if len(runes) == 0 {
		return "", errors.New("randomstring: charset must not be empty")
	}
	n := big.NewInt(int64(len(runes)))
	out := make([]rune, length)
	for i := range out {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		out[i] = runes[idx.Int64()]
	}
	return string(out), nil
}

// Generate returns a random string of the requested length using a preset
// character set. An empty charset defaults to "alphanumeric".
func Generate(length int, charset string) (string, error) {
	if charset == "" {
		charset = "alphanumeric"
	}
	chars, ok := presets[charset]
	if !ok {
		return "", errors.New("randomstring: unknown charset: " + charset)
	}
	return GenerateFrom(length, chars)
}

// New returns a 32-character alphanumeric random string.
func New() (string, error) {
	return Generate(32, "alphanumeric")
}
