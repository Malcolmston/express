// Package base32 encodes and decodes byte data using the RFC 4648 base32
// alphabet, providing a small convenience wrapper over the standard library's
// encoding/base32. It plays the role of the npm base32 helpers used in the
// wider Express port, offering the handful of encode/decode operations those
// libraries expose while delegating the actual bit-packing to Go's battle-tested
// implementation.
//
// Base32 represents arbitrary binary data as text drawn from a 32-character
// alphabet, so it is useful whenever data must survive systems that are
// case-insensitive, that mangle mixed case, or that only tolerate a limited set
// of letters and digits. Common uses include human-transcribable identifiers,
// TOTP/2FA secret keys, filenames on case-folding filesystems, and any token
// that might be spoken aloud or typed by hand. It trades compactness for
// robustness: base32 output is longer than base64, but its alphabet avoids the
// case sensitivity and punctuation that make base64 fragile in those settings.
//
// Encoding follows RFC 4648: the input bytes are treated as a bit stream, split
// into 5-bit groups, and each group is mapped to one of the uppercase letters A
// through Z and the digits 2 through 7. Because 5 bits do not divide evenly into
// 8-bit bytes, the final group is padded with zero bits and the output is padded
// with "=" characters so that its length is a multiple of eight. Encode produces
// this canonical padded form, while EncodeNoPadding produces the same characters
// with the trailing "=" run removed for contexts where padding is unwanted.
//
// Decode is deliberately lenient so it can consume output from a variety of
// producers. Before handing the string to the standard decoder it trims
// surrounding whitespace, upper-cases the text so lowercase input is accepted,
// and strips any trailing "=" so both padded and unpadded strings decode
// successfully. The empty string and empty byte slices round-trip to each other
// (an empty input encodes to "" and "" decodes back to an empty slice). Input
// containing characters outside the alphabet, or an invalid length after padding
// is removed, yields a non-nil error from encoding/base32 rather than partial
// output.
//
// Relative to the Node original the semantics are intentionally aligned:
// encoding yields the same RFC 4648 uppercase strings a JavaScript base32
// library would produce for the same bytes, and decoding accepts the same
// case-insensitive, optionally-unpadded inputs. The differences are idiomatic
// Go ones. Functions take and return []byte and string instead of Buffers, and
// decode failures are reported through Go's error return value instead of by
// throwing, so callers check err rather than catching an exception.
package base32

import (
	"encoding/base32"
	"strings"
)

// Encode encodes data using the uppercase RFC 4648 alphabet with padding.
func Encode(data []byte) string {
	return base32.StdEncoding.EncodeToString(data)
}

// EncodeNoPadding encodes data without the "=" padding characters.
func EncodeNoPadding(data []byte) string {
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data)
}

// Decode decodes a base32 string, case-insensitively and tolerating missing padding.
func Decode(s string) ([]byte, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.TrimRight(s, "=")
	return base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(s)
}
