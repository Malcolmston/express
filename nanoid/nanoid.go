// Package nanoid generates compact, URL-safe, cryptographically random
// identifiers, a Go port of the npm "nanoid" package.
package nanoid

import (
	"crypto/rand"
	"errors"
	"math/bits"
)

// DefaultAlphabet is the URL-safe alphabet used by nanoid.
const DefaultAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-"

// DefaultSize is the default id length.
const DefaultSize = 21

// New returns a nanoid using the default alphabet and size.
func New() (string, error) {
	return Custom(DefaultAlphabet, DefaultSize)
}

// NewSize returns a nanoid using the default alphabet and the given size.
func NewSize(size int) (string, error) {
	return Custom(DefaultAlphabet, size)
}

// Custom returns a nanoid using the given alphabet and size, sampling bytes
// from crypto/rand with an unbiased mask/reject method.
func Custom(alphabet string, size int) (string, error) {
	if size <= 0 {
		return "", errors.New("nanoid: size must be positive")
	}
	if len(alphabet) < 1 || len(alphabet) > 256 {
		return "", errors.New("nanoid: alphabet length must be in 1..256")
	}

	// mask is the smallest 2^n-1 >= len(alphabet)-1.
	mask := 1
	if len(alphabet) > 1 {
		mask = (2 << (bits.Len(uint(len(alphabet)-1)) - 1)) - 1
	}
	// step: how many random bytes to request per batch.
	step := (size * mask * 8) / (len(alphabet) * 5)
	if step < 1 {
		step = size
	}

	id := make([]byte, 0, size)
	buf := make([]byte, step)
	for {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for i := 0; i < step; i++ {
			idx := int(buf[i]) & mask
			if idx < len(alphabet) {
				id = append(id, alphabet[idx])
				if len(id) == size {
					return string(id), nil
				}
			}
		}
	}
}
