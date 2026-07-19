// Package randomstring is a standard-library port of the npm "randomstring"
// library. It generates random strings from named character-set presets or a
// custom character set using crypto/rand for unbiased sampling.
//
// Random strings are a common building block for tokens, temporary passwords,
// session identifiers, file names, and test fixtures. This package packages
// that need behind a small API: pick a length and either a named charset or a
// literal set of characters, and receive a string of that length drawn from
// those characters. Because the source of randomness is crypto/rand, the output
// is suitable for security-sensitive identifiers, unlike helpers built on
// math/rand.
//
// Sampling is uniform and unbiased. For each output position the code draws a
// value in [0, len(charset)) with crypto/rand's rand.Int, which uses rejection
// sampling internally, so no character is favored even when the charset length
// does not evenly divide the size of the random space. The charset is treated
// as a slice of runes, so multi-byte Unicode characters count as a single
// symbol and are emitted whole.
//
// Three entry points cover the common cases. GenerateFrom takes an explicit
// character set and is the lowest-level primitive. Generate selects one of the
// named presets: "alphanumeric" (the default when the name is empty),
// "alphabetic", "numeric", "hex", "binary", and "octal". New is a convenience
// wrapper returning a 32-character alphanumeric string, the library's typical
// default token.
//
// Error semantics are explicit. A negative length is rejected, an empty charset
// is rejected, and an unknown preset name returns an error rather than falling
// back silently. A length of zero is valid and yields the empty string. The
// main parity difference from the npm original is shape rather than behavior:
// where the JavaScript version exposes a single options object with fields like
// length, charset, capitalization, and readable, this port offers a small set
// of explicit functions and leaves higher-level formatting to the caller.
package randomstring

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
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

// Options mirrors the options object accepted by the npm original's
// randomstring.generate(options). It groups the length, character-set
// selection, capitalization, and readability controls that the JavaScript
// version exposes on a single object.
//
// Charset holds one or more entries; each entry is resolved with the same rule
// the upstream Charset.getCharacters uses: a recognized preset name
// ("alphanumeric", "alphabetic", "numeric", "hex", "binary", "octal") expands
// to that preset's characters, and any other value is treated as a literal set
// of characters. Multiple entries are concatenated, matching upstream's support
// for an array of charsets (for example []string{"alphabetic", "!"}). An empty
// Charset defaults to "alphanumeric".
//
// Capitalization is "uppercase", "lowercase", or "" (leave as-is), applied to
// the assembled character set before sampling, exactly as upstream does.
// Readable, when true, removes the visually ambiguous characters 0, O, I, and l
// from the set. After these transforms duplicate characters are removed so that
// sampling stays uniform, matching upstream's removeDuplicates step.
type Options struct {
	Length         int
	Charset        []string
	Capitalization string
	Readable       bool
}

// getCharacters resolves a single charset entry, mirroring upstream's
// Charset.getCharacters: a known preset name expands to its characters, and any
// other value is returned as a literal character set.
func getCharacters(entry string) string {
	if chars, ok := presets[entry]; ok {
		return chars
	}
	return entry
}

// removeUnreadable drops the visually ambiguous characters 0, O, I, and l,
// mirroring upstream's Charset.removeUnreadable (/[0OIl]/g).
func removeUnreadable(chars string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '0', 'O', 'I', 'l':
			return -1
		default:
			return r
		}
	}, chars)
}

// removeDuplicates removes repeated characters while preserving first-seen
// order, mirroring upstream's Charset.removeDuplicates.
func removeDuplicates(chars string) string {
	seen := make(map[rune]bool)
	var b strings.Builder
	for _, r := range chars {
		if !seen[r] {
			seen[r] = true
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GenerateWith returns a random string built from opts, reproducing the
// character-set pipeline of the npm original's generate(options): the charset
// entries are resolved and concatenated, capitalization is applied, unreadable
// characters are optionally removed, and duplicates are stripped before uniform
// sampling with crypto/rand.
func GenerateWith(opts Options) (string, error) {
	var chars string
	if len(opts.Charset) == 0 {
		chars = presets["alphanumeric"]
	} else {
		for _, entry := range opts.Charset {
			chars += getCharacters(entry)
		}
	}

	switch opts.Capitalization {
	case "uppercase":
		chars = strings.ToUpper(chars)
	case "lowercase":
		chars = strings.ToLower(chars)
	}

	if opts.Readable {
		chars = removeUnreadable(chars)
	}

	chars = removeDuplicates(chars)

	return GenerateFrom(opts.Length, chars)
}
