// Package nanoid generates compact, URL-safe, cryptographically random
// identifiers, a Go port of the npm "nanoid" package. nanoid is a popular
// modern alternative to UUIDs: the identifiers are shorter, need no hyphenation
// or special encoding to travel safely in URLs, and are produced from a
// cryptographically secure random source. This implementation uses only the Go
// standard library, drawing its randomness from crypto/rand.
//
// You use this package whenever you need unique, unguessable, opaque tokens —
// database primary keys, short links, session or resource identifiers, file
// names — and want something more compact and URL-friendly than a canonical
// UUID string. Because the bytes come from a CSPRNG the ids are suitable for
// contexts where predictability would be a security problem, and because the
// default alphabet is URL- and filename-safe they can be dropped into a path
// or query string without escaping.
//
// Internally an id is built by sampling random bytes and mapping them onto the
// alphabet with an unbiased mask-and-reject scheme. A bit mask the size of the
// smallest power of two that covers the alphabet is applied to each random
// byte; values that fall outside the alphabet range are discarded and more
// randomness is drawn, which keeps every character equally likely rather than
// skewing toward the start of the alphabet the way a naive modulo would. Bytes
// are requested from crypto/rand in batches sized to make rejections cheap.
//
// The default alphabet is the 64-character URL-safe set of A–Z, a–z, 0–9, "_"
// and "-" exposed as DefaultAlphabet — the same characters, in the same
// compression-friendly order, as the npm original's urlAlphabet — and the
// default length is DefaultSize
// (21), which gives a collision probability comparable to a v4 UUID. New
// returns an id with those defaults; NewSize keeps the default alphabet but
// lets you choose the length; and Custom lets you supply both a custom
// alphabet and a custom size, for example to restrict ids to lowercase hex or
// to trade length for a larger keyspace.
//
// The inputs are validated rather than silently coerced: a negative size, or
// an alphabet whose length is outside 1..256, produces an error. A size of
// zero yields the empty id (matching the original's nanoid(0) === "").
// Note that a shorter id or a smaller alphabet increases the chance of
// collisions, so those are trade-offs the caller makes deliberately. Compared
// with the Node original the generation algorithm and default alphabet and
// size are the same, and the output is the same kind of string; the main
// differences are that errors (including any failure to read from the system
// RNG) are returned as Go error values instead of thrown, and that each
// generated value is inherently random, so callers should test properties such
// as length and alphabet membership rather than exact output.
package nanoid

import (
	"crypto/rand"
	"errors"
	"math/bits"
)

// DefaultAlphabet is the URL-safe alphabet used by nanoid. It is the same
// 64-character A-Za-z0-9_- set as the npm original's urlAlphabet, in the same
// order (chosen to compress well under gzip and brotli).
const DefaultAlphabet = "useandom-26T198340PX75pxJACKVERYMINDBUSHWOLF_GQZbfghjklqvwyzrict"

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
	// Match the npm original: a negative size is an error ("Wrong ID size"),
	// while a zero size yields the empty id without inspecting the alphabet
	// (nanoid(0) === '', customAlphabet('')(0) === '', etc.).
	if size < 0 {
		return "", errors.New("nanoid: size must not be negative")
	}
	if size == 0 {
		return "", nil
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
