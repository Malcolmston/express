// Package cryptorandomstring is a standard-library port of the npm
// "crypto-random-string" library. It generates cryptographically strong random
// strings via crypto/rand with unbiased sampling.
package cryptorandomstring

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	hexChars             = "0123456789abcdef"
	numericChars         = "0123456789"
	distinguishableChars = "CDEHKMPRTUWXY012458964"
	alphanumericChars    = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	urlSafeChars         = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	base64Chars          = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	asciiPrintableChars  = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"
)

// Options controls the generated string.
type Options struct {
	// Length is the number of characters to generate.
	Length int
	// Type selects a preset character set. One of "hex" (default), "base64",
	// "url-safe", "numeric", "distinguishable", "ascii-printable",
	// "alphanumeric".
	Type string
	// Characters, if non-empty, overrides Type and is used as the literal
	// character set.
	Characters string
}

func charsForType(t string) (string, error) {
	switch t {
	case "", "hex":
		return hexChars, nil
	case "base64":
		return base64Chars, nil
	case "url-safe":
		return urlSafeChars, nil
	case "numeric":
		return numericChars, nil
	case "distinguishable":
		return distinguishableChars, nil
	case "ascii-printable":
		return asciiPrintableChars, nil
	case "alphanumeric":
		return alphanumericChars, nil
	default:
		return "", errors.New("cryptorandomstring: unknown type: " + t)
	}
}

// Generate returns a secure random string according to opts.
func Generate(opts Options) (string, error) {
	if opts.Length < 0 {
		return "", errors.New("cryptorandomstring: length must be non-negative")
	}
	var chars string
	if opts.Characters != "" {
		chars = opts.Characters
	} else {
		c, err := charsForType(opts.Type)
		if err != nil {
			return "", err
		}
		chars = c
	}
	runes := []rune(chars)
	if len(runes) == 0 {
		return "", errors.New("cryptorandomstring: characters must not be empty")
	}
	n := big.NewInt(int64(len(runes)))
	out := make([]rune, opts.Length)
	for i := range out {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		out[i] = runes[idx.Int64()]
	}
	return string(out), nil
}

// Hex returns a secure random hexadecimal string of the given length.
func Hex(length int) (string, error) {
	return Generate(Options{Length: length, Type: "hex"})
}
