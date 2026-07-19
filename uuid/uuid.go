// Package uuid generates and parses RFC 4122 universally unique identifiers.
// It is a small, standard-library-only Go port of the npm "uuid" package,
// exposing the two identifier flavours that the vast majority of applications
// need: fully random version 4 UUIDs and name-based version 5 (SHA-1) UUIDs.
// A UUID is a 128-bit value canonically rendered as 36 lowercase hexadecimal
// characters in the grouping 8-4-4-4-12 (for example
// "6ba7b810-9dad-11d1-80b4-00c04fd430c8"), and every value produced here sets
// the RFC 4122 variant bits (10xx) and the appropriate version nibble.
//
// V4 is the workhorse: it reads 16 bytes from crypto/rand, forces the version
// nibble to 4 and the variant bits to the RFC 4122 layout, then formats the
// result. Because 122 of the 128 bits are random, collisions are astronomically
// unlikely and no coordination between generators is required. Use V4 whenever
// you need an opaque, unguessable identifier and have no reason to derive it
// from an existing name.
//
// V5 is deterministic and name-based. It hashes a namespace UUID concatenated
// with a name using SHA-1, truncates the digest to 16 bytes, and stamps in the
// version (5) and variant bits. The same namespace and name always yield the
// same UUID, which makes V5 ideal for stable identifiers derived from URLs,
// domain names, or other natural keys. Four standard namespaces from RFC 4122
// are provided as constants: NamespaceDNS, NamespaceURL, NamespaceOID, and
// NamespaceX500. (Note that this package intentionally implements only the
// random and SHA-1 name-based variants; the MD5-based version 3 is not
// included.)
//
// Parse, Format, and Validate round out the API. Parse converts the canonical
// string form into a raw [16]byte array, rejecting inputs of the wrong length,
// with misplaced dashes, or with non-hexadecimal digits. Format performs the
// inverse, rendering a [16]byte as the canonical lowercase dashed string.
// Validate is a convenience wrapper that reports whether a string parses
// cleanly. Parse is deliberately strict about the hyphenated 36-character
// layout and does not accept braces, urn: prefixes, or uppercase-only forms.
//
// Compared with the npm "uuid" package, the output format and the fixed
// version/variant bits match exactly, so identifiers are interchangeable across
// the two ecosystems: a V5 UUID computed here for a given namespace and name is
// byte-for-byte identical to one produced by uuid.v5 in Node, and V4 values
// validate under both. The main API-shape difference is idiomatic Go error
// returns instead of thrown exceptions, and the use of crypto/rand rather than
// Node's crypto.randomBytes as the entropy source.
package uuid

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"regexp"
)

// uuidRegexp mirrors the upstream npm "uuid" validation regex
// (src/regex.ts): a canonical 8-4-4-4-12 string whose version nibble is one
// of 1-8 and whose variant nibble is one of 8, 9, a, or b, plus the two
// special-case constants NIL (all zeros) and MAX (all fs). Matching is
// case-insensitive. Strings outside this grammar - including well-formed hex
// with an out-of-range version or variant - are rejected, exactly as upstream
// validate()/parse() reject them.
var uuidRegexp = regexp.MustCompile(`(?i)^(?:[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|00000000-0000-0000-0000-000000000000|ffffffff-ffff-ffff-ffff-ffffffffffff)$`)

// Predefined namespace UUIDs from RFC 4122, used with V3 and V5.
const (
	// NamespaceDNS is the namespace for fully-qualified domain names.
	NamespaceDNS = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	// NamespaceURL is the namespace for URLs.
	NamespaceURL = "6ba7b811-9dad-11d1-80b4-00c04fd430c8"
	// NamespaceOID is the namespace for ISO OIDs.
	NamespaceOID = "6ba7b812-9dad-11d1-80b4-00c04fd430c8"
	// NamespaceX500 is the namespace for X.500 DNs.
	NamespaceX500 = "6ba7b814-9dad-11d1-80b4-00c04fd430c8"
)

// V4 returns a randomly generated UUID version 4 string.
func V4() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// Set version to 4.
	b[6] = (b[6] & 0x0f) | 0x40
	// Set variant to RFC 4122 (10xx).
	b[8] = (b[8] & 0x3f) | 0x80
	return Format(b), nil
}

// V5 returns a name-based UUID version 5 (SHA-1) string.
func V5(namespace string, name string) (string, error) {
	ns, err := Parse(namespace)
	if err != nil {
		return "", err
	}
	h := sha1.New()
	h.Write(ns[:])
	h.Write([]byte(name))
	sum := h.Sum(nil)
	var b [16]byte
	copy(b[:], sum[:16])
	// Set version to 5.
	b[6] = (b[6] & 0x0f) | 0x50
	// Set variant to RFC 4122 (10xx).
	b[8] = (b[8] & 0x3f) | 0x80
	return Format(b), nil
}

// Parse parses a UUID string into a 16-byte array.
func Parse(s string) ([16]byte, error) {
	var b [16]byte
	if !uuidRegexp.MatchString(s) {
		return b, errors.New("uuid: invalid UUID")
	}
	stripped := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:36]
	decoded, err := hex.DecodeString(stripped)
	if err != nil {
		return b, errors.New("uuid: invalid hex")
	}
	copy(b[:], decoded)
	return b, nil
}

// Format renders a 16-byte array as a canonical lowercase UUID string.
func Format(b [16]byte) string {
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

// Validate reports whether s is a well-formed UUID string. It matches the
// upstream npm "uuid" validate(): the string must be a canonical 8-4-4-4-12
// form with a version nibble in 1-8 and a variant nibble in [89ab], or be the
// NIL or MAX constant. Out-of-range versions/variants are rejected.
func Validate(s string) bool {
	return uuidRegexp.MatchString(s)
}
