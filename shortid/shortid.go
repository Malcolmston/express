// Package shortid is a standard-library-only short, URL-friendly unique id
// generator inspired by the npm "shortid" library. It produces compact,
// non-sequential-looking strings suitable for use in URLs, filenames, and other
// contexts where a full UUID would be unnecessarily long. Every character comes
// from a 64-character URL-safe alphabet, so ids never require percent-encoding.
//
// An id is built from two concatenated components. The first is a time
// component: the current time in Unix milliseconds is written out in the
// alphabet's base (64), least-significant digit first, which yields a handful of
// characters that advance as the clock advances. The second is a random
// component of six characters, each drawn independently from crypto/rand via
// math/big, which disambiguates ids created within the same millisecond. The
// two parts together are typically 7 to 14 characters long depending on the
// magnitude of the current timestamp.
//
// The default alphabet is
// "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-", the same
// URL-safe set of 64 symbols used by the npm library. SetAlphabet lets you
// substitute your own ordering or character set; it requires exactly 64 unique
// characters and returns an error otherwise. The alphabet is guarded by a mutex
// and read under lock by both Generate and IsValid, so it is safe to change and
// use concurrently, though changing it mid-flight means previously issued ids
// may no longer validate against the new set.
//
// Note on ordering and uniqueness: although the leading time component means
// ids created in later milliseconds generally compare greater than earlier
// ones, shortid does not guarantee strict lexicographic sortability the way
// ULID or xid do, because the time digits are emitted least-significant first
// and the trailing random characters dominate string comparison. Uniqueness
// rests on the six random characters (64^6 possibilities) combined with the
// millisecond timestamp; this makes practical collisions extremely unlikely but,
// unlike a counter-based scheme, not theoretically impossible within a single
// millisecond.
//
// IsValid reports whether a string is non-empty and composed solely of
// characters from the current alphabet; it is a cheap membership check, not a
// cryptographic verification of provenance. Compared with the npm "shortid"
// package (now itself deprecated in favour of nanoid), this port keeps the
// time-plus-random construction and URL-safe alphabet but deliberately drops the
// worker/seed cluster configuration, exposing instead a simpler Generate,
// IsValid, and SetAlphabet surface with idiomatic Go error returns.
package shortid

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"
)

const defaultAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

var (
	mu       sync.Mutex
	alphabet = defaultAlphabet
)

// SetAlphabet replaces the alphabet used by Generate and IsValid. The alphabet
// must consist of exactly 64 unique characters.
func SetAlphabet(a string) error {
	runes := []rune(a)
	if len(runes) != 64 {
		return errors.New("shortid: alphabet must be exactly 64 characters")
	}
	seen := make(map[rune]bool, 64)
	for _, r := range runes {
		if seen[r] {
			return errors.New("shortid: alphabet characters must be unique")
		}
		seen[r] = true
	}
	mu.Lock()
	alphabet = a
	mu.Unlock()
	return nil
}

// Generate returns a new short id combining a time component and a random
// component. Ids are between 7 and 14 characters long.
func Generate() (string, error) {
	mu.Lock()
	a := alphabet
	mu.Unlock()

	runes := []rune(a)
	base := int64(len(runes))

	// Time component (millisecond precision) encoded in the alphabet base.
	t := time.Now().UnixMilli()
	out := make([]rune, 0, 16)
	for t > 0 {
		out = append(out, runes[t%base])
		t /= base
	}

	// Random component for uniqueness within the same millisecond.
	n := big.NewInt(base)
	for i := 0; i < 6; i++ {
		idx, err := rand.Int(rand.Reader, n)
		if err != nil {
			return "", err
		}
		out = append(out, runes[idx.Int64()])
	}

	return string(out), nil
}

// IsValid reports whether s is non-empty and contains only characters from the
// current alphabet.
func IsValid(s string) bool {
	if s == "" {
		return false
	}
	mu.Lock()
	a := alphabet
	mu.Unlock()

	set := make(map[rune]bool, len(a))
	for _, r := range a {
		set[r] = true
	}
	for _, r := range s {
		if !set[r] {
			return false
		}
	}
	return true
}
