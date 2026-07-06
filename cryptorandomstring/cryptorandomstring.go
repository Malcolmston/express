// Package cryptorandomstring is a standard-library port of the npm
// "crypto-random-string" library. It generates cryptographically strong random
// strings via crypto/rand with unbiased sampling, and offers the same set of
// named character-set presets as the original module.
//
// Use this package anywhere you need a random token that must be hard to
// guess: session identifiers, API keys, password-reset tokens, one-time
// codes, filenames for temporary uploads, and similar. Because it draws from
// crypto/rand rather than math/rand, the output is suitable for security
// contexts where predictability would be a vulnerability. When you only need
// a random hex string, the Hex convenience wrapper covers the common case;
// for anything else, Generate takes an Options struct that selects a preset
// character set or supplies a completely custom one.
//
// The generation algorithm produces one output rune at a time. For each
// position it calls crypto/rand.Int with an upper bound equal to the size of
// the character set, then indexes into the set with the result. Using
// crypto/rand.Int (which performs rejection sampling internally) rather than
// reducing a random byte modulo the set size is what keeps the distribution
// unbiased: every character is equally likely regardless of whether the set
// length divides evenly into a power of two. The character set is treated as
// a slice of runes, so multi-byte Unicode characters in a custom Characters
// string each count as a single output element.
//
// The available presets mirror the Node library: "hex" (the default,
// 0-9a-f), "base64", "url-safe", "numeric", "distinguishable" (a reduced set
// that omits visually ambiguous glyphs such as 0/O and 1/l), "ascii-printable",
// and "alphanumeric". Options.Characters, when non-empty, overrides Type and
// is used verbatim as the literal set. Regarding semantics and edge cases:
// a negative Length is an error; a Length of zero yields an empty string with
// no error; an unknown Type is an error; and an explicitly empty character set
// (an empty Characters together with no usable preset) is an error. Errors
// from the underlying crypto/rand reader are propagated to the caller rather
// than swallowed.
//
// Parity with the npm original is close in behavior but idiomatic in shape.
// The preset names and their character contents match, the unbiased sampling
// matches, and the default type is "hex" in both. The differences are those
// you would expect from a Go port: the API is a single Generate function
// taking an Options value plus a Hex helper, instead of a JavaScript function
// with an options object; errors are returned explicitly instead of thrown;
// and output is a Go string of runes. Length is measured in characters
// (runes), matching the Node library's character-count semantics rather than
// raw byte count.
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
