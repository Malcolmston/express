// Package encodeurl encodes a URL to a percent-encoded form, excluding
// already-encoded sequences.
//
// It is a port of the npm package "encodeurl". The goal is to encode
// characters that are unsafe or invalid in a URL while leaving any existing
// valid percent-encoded sequences (such as "%20") untouched, so that a URL is
// never double-encoded.
package encodeurl

import (
	"strings"
)

// Encode encodes a URL leaving already percent-encoded sequences intact.
//
// Characters outside the safe set defined by the reference implementation are
// percent-encoded using UTF-8. Existing valid "%xx" escape sequences are
// preserved as-is, and a raw "%" that is not the start of a valid escape is
// encoded to "%25". Invalid UTF-8 is replaced with the Unicode replacement
// character before encoding.
func Encode(rawurl string) string {
	var b strings.Builder
	b.Grow(len(rawurl))

	runes := []rune(rawurl) // invalid UTF-8 becomes U+FFFD
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '%' {
			if validEscape(runes, i) {
				b.WriteByte('%')
				continue
			}
			b.WriteString("%25")
			continue
		}
		if safeChar(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteString(encodeURIChar(r))
	}
	return b.String()
}

// validEscape reports whether runes[i] ('%') begins a preserved escape.
//
// It follows the semantics of the reference ENCODE_CHARS_REGEXP: a '%' is left
// alone when it is followed by two hex digits, or when it is followed by a
// single hex digit at the very end of the string.
func validEscape(runes []rune, i int) bool {
	if i+1 >= len(runes) {
		return false // lone '%' at end
	}
	if !isHex(runes[i+1]) {
		return false // '%' followed by a non-hex char
	}
	if i+2 >= len(runes) {
		return true // '%x' at end of string is left intact
	}
	return isHex(runes[i+2])
}

func isHex(r rune) bool {
	return (r >= '0' && r <= '9') ||
		(r >= 'a' && r <= 'f') ||
		(r >= 'A' && r <= 'F')
}

// safeChar reports whether r is in the reference implementation's set of
// characters that are never encoded (excluding '%', handled separately).
func safeChar(r rune) bool {
	switch {
	case r == 0x21: // !
		return true
	case r >= 0x26 && r <= 0x3B: // & ' ( ) * + , - . / 0-9 : ;
		return true
	case r == 0x3D: // =
		return true
	case r >= 0x3F && r <= 0x5B: // ? @ A-Z [
		return true
	case r == 0x5D: // ]
		return true
	case r == 0x5F: // _
		return true
	case r >= 0x61 && r <= 0x7A: // a-z
		return true
	case r == 0x7E: // ~
		return true
	}
	return false
}

// encodeURIChar mirrors JavaScript's encodeURI for a single rune: reserved and
// unreserved URL characters are preserved, everything else is percent-encoded
// using its UTF-8 byte representation.
func encodeURIChar(r rune) string {
	if encodeURISafe(r) {
		return string(r)
	}
	const hexDigits = "0123456789ABCDEF"
	var b strings.Builder
	for _, by := range []byte(string(r)) {
		b.WriteByte('%')
		b.WriteByte(hexDigits[by>>4])
		b.WriteByte(hexDigits[by&0x0F])
	}
	return b.String()
}

// encodeURISafe reports whether r is a character that JavaScript's encodeURI
// does not encode.
func encodeURISafe(r rune) bool {
	if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
		return true
	}
	switch r {
	case '-', '_', '.', '!', '~', '*', '\'', '(', ')',
		';', ',', '/', '?', ':', '@', '&', '=', '+', '$', '#':
		return true
	}
	return false
}
