// Package shortid is a standard-library short, URL-friendly unique id
// generator inspired by the npm "shortid" library. Ids combine a time
// component with a crypto/rand random component over a 64-character
// URL-safe alphabet.
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
