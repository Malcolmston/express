// Package ulid generates and decodes Universally Unique Lexicographically
// Sortable Identifiers (ULIDs). It is a standard-library-only Go port of the
// npm "ulid" package. A ULID packs a millisecond-precision timestamp together
// with random entropy into a single 128-bit value that is both globally unique
// and naturally ordered by creation time, giving you the k-sortability of an
// auto-increment key without a central coordinator.
//
// The 128 bits are split into two fields: the high 48 bits hold a Unix
// timestamp in milliseconds, and the low 80 bits hold random entropy. New reads
// those 10 entropy bytes from crypto/rand, while NewWithEntropy lets the caller
// supply them for deterministic or reproducible generation. The timestamp field
// is 48 bits wide, so New/NewWithEntropy reject any ms value greater than
// 2^48-1 (roughly the year 10889), and the entropy slice must be exactly ten
// bytes long.
//
// The 16 raw bytes are rendered as a fixed 26-character string using Crockford
// base32 (Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"), which omits the
// ambiguous letters I, L, O, and U. The first ten characters encode the
// timestamp and the remaining sixteen encode the entropy; because base32 does
// not divide 128 evenly, the leading character carries only the top three bits,
// so a valid ULID never begins with a digit greater than 7. Decode is
// case-insensitive and also accepts the Crockford aliases i/I and l/L for 1 and
// o/O for 0.
//
// Sortability is the defining property. Since the timestamp occupies the most
// significant bits and base32 preserves byte order, ULIDs generated at
// different milliseconds sort correctly both as raw bytes and as ASCII strings
// (string comparison and time order agree). Within a single millisecond,
// ordering is determined by the random entropy and is therefore arbitrary
// unless the caller supplies monotonically increasing entropy. Note that,
// unlike the npm package's optional monotonicFactory, this port does not
// implement automatic monotonic entropy bumping: two ULIDs created in the same
// millisecond via New have independent random tails and no guaranteed relative
// order.
//
// Decode reverses the encoding into the raw [16]byte, and Timestamp extracts
// just the embedded millisecond value without a full decode round trip.
// Compared with the Node package, the binary layout, Crockford alphabet, and
// 26-character canonical form are identical, so a ULID string produced here can
// be decoded by the JavaScript library and vice versa; the differences are
// idiomatic Go error returns and the absence of the JS monotonic and
// pseudo-random PRNG helpers.
package ulid

import (
	"crypto/rand"
	"errors"
)

// Alphabet is the Crockford base32 alphabet used by ULID.
const Alphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// dec maps a byte value to its Crockford base32 digit, or 0xff if invalid.
var dec [256]byte

func init() {
	for i := range dec {
		dec[i] = 0xff
	}
	for i := 0; i < len(Alphabet); i++ {
		dec[Alphabet[i]] = byte(i)
	}
	// Crockford aliases and lowercase.
	dec['l'], dec['L'] = 1, 1
	dec['i'], dec['I'] = 1, 1
	dec['o'], dec['O'] = 0, 0
	for i := 0; i < len(Alphabet); i++ {
		c := Alphabet[i]
		if c >= 'A' && c <= 'Z' {
			dec[c+32] = byte(i)
		}
	}
}

// New generates a ULID for the given millisecond timestamp with random entropy.
func New(ms uint64) (string, error) {
	var entropy [10]byte
	if _, err := rand.Read(entropy[:]); err != nil {
		return "", err
	}
	return NewWithEntropy(ms, entropy[:])
}

// NewWithEntropy generates a ULID deterministically from ms and 10-byte entropy.
func NewWithEntropy(ms uint64, entropy []byte) (string, error) {
	if len(entropy) != 10 {
		return "", errors.New("ulid: entropy must be 10 bytes")
	}
	if ms > (1<<48)-1 {
		return "", errors.New("ulid: timestamp overflow")
	}
	var b [16]byte
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	copy(b[6:], entropy)
	return encode(b), nil
}

// encode renders 16 bytes as a 26-char Crockford base32 ULID string.
func encode(b [16]byte) string {
	// 128 bits -> 26 base32 chars. The first char encodes only the top 3 bits.
	s := make([]byte, 26)
	s[0] = Alphabet[(b[0]&0xe0)>>5]
	s[1] = Alphabet[b[0]&0x1f]
	s[2] = Alphabet[(b[1]&0xf8)>>3]
	s[3] = Alphabet[((b[1]&0x07)<<2)|((b[2]&0xc0)>>6)]
	s[4] = Alphabet[(b[2]&0x3e)>>1]
	s[5] = Alphabet[((b[2]&0x01)<<4)|((b[3]&0xf0)>>4)]
	s[6] = Alphabet[((b[3]&0x0f)<<1)|((b[4]&0x80)>>7)]
	s[7] = Alphabet[(b[4]&0x7c)>>2]
	s[8] = Alphabet[((b[4]&0x03)<<3)|((b[5]&0xe0)>>5)]
	s[9] = Alphabet[b[5]&0x1f]

	s[10] = Alphabet[(b[6]&0xf8)>>3]
	s[11] = Alphabet[((b[6]&0x07)<<2)|((b[7]&0xc0)>>6)]
	s[12] = Alphabet[(b[7]&0x3e)>>1]
	s[13] = Alphabet[((b[7]&0x01)<<4)|((b[8]&0xf0)>>4)]
	s[14] = Alphabet[((b[8]&0x0f)<<1)|((b[9]&0x80)>>7)]
	s[15] = Alphabet[(b[9]&0x7c)>>2]
	s[16] = Alphabet[((b[9]&0x03)<<3)|((b[10]&0xe0)>>5)]
	s[17] = Alphabet[b[10]&0x1f]
	s[18] = Alphabet[(b[11]&0xf8)>>3]
	s[19] = Alphabet[((b[11]&0x07)<<2)|((b[12]&0xc0)>>6)]
	s[20] = Alphabet[(b[12]&0x3e)>>1]
	s[21] = Alphabet[((b[12]&0x01)<<4)|((b[13]&0xf0)>>4)]
	s[22] = Alphabet[((b[13]&0x0f)<<1)|((b[14]&0x80)>>7)]
	s[23] = Alphabet[(b[14]&0x7c)>>2]
	s[24] = Alphabet[((b[14]&0x03)<<3)|((b[15]&0xe0)>>5)]
	s[25] = Alphabet[b[15]&0x1f]
	return string(s)
}

// Decode parses a 26-char ULID string into its 16 raw bytes.
func Decode(id string) ([16]byte, error) {
	var b [16]byte
	if len(id) != 26 {
		return b, errors.New("ulid: invalid length")
	}
	var v [26]byte
	for i := 0; i < 26; i++ {
		d := dec[id[i]]
		if d == 0xff {
			return b, errors.New("ulid: invalid character")
		}
		v[i] = d
	}
	// First char must fit in 3 bits (top of 128-bit value).
	if v[0] > 7 {
		return b, errors.New("ulid: overflow")
	}
	b[0] = (v[0] << 5) | v[1]
	b[1] = (v[2] << 3) | (v[3] >> 2)
	b[2] = (v[3] << 6) | (v[4] << 1) | (v[5] >> 4)
	b[3] = (v[5] << 4) | (v[6] >> 1)
	b[4] = (v[6] << 7) | (v[7] << 2) | (v[8] >> 3)
	b[5] = (v[8] << 5) | v[9]
	b[6] = (v[10] << 3) | (v[11] >> 2)
	b[7] = (v[11] << 6) | (v[12] << 1) | (v[13] >> 4)
	b[8] = (v[13] << 4) | (v[14] >> 1)
	b[9] = (v[14] << 7) | (v[15] << 2) | (v[16] >> 3)
	b[10] = (v[16] << 5) | v[17]
	b[11] = (v[18] << 3) | (v[19] >> 2)
	b[12] = (v[19] << 6) | (v[20] << 1) | (v[21] >> 4)
	b[13] = (v[21] << 4) | (v[22] >> 1)
	b[14] = (v[22] << 7) | (v[23] << 2) | (v[24] >> 3)
	b[15] = (v[24] << 5) | v[25]
	return b, nil
}

// Timestamp decodes the millisecond timestamp encoded in a ULID string.
func Timestamp(id string) (uint64, error) {
	b, err := Decode(id)
	if err != nil {
		return 0, err
	}
	ms := uint64(b[0])<<40 | uint64(b[1])<<32 | uint64(b[2])<<24 |
		uint64(b[3])<<16 | uint64(b[4])<<8 | uint64(b[5])
	return ms, nil
}
