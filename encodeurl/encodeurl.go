// Package encodeurl encodes a URL to a percent-encoded form while leaving any
// already-encoded sequences intact, a port of the npm "encodeurl" package that
// Express and the send/serve-static middleware use when reflecting a request
// path back into a header such as Location. It exposes the single Encode
// function and is built on only the Go standard library.
//
// The problem it solves is double-encoding. When a server takes a URL that may
// already contain percent-escapes — most commonly the request URL itself — and
// needs to place it somewhere that requires a valid URL (a redirect target, an
// error message, a link), naively percent-encoding it would turn an existing
// "%20" into "%2520". Encode instead encodes only the characters that are unsafe
// or invalid in a URL and passes through sequences that already look like valid
// escapes, so the output is safe to emit exactly once without corrupting escapes
// the caller had already applied.
//
// Encoding follows the reference implementation's fixed character set. A byte is
// left untouched when it is in encodeURI's reserved-plus-unreserved set (letters,
// digits, and the punctuation such as ! # $ & ' ( ) * + , - . / : ; = ? @ _ ~ [ ]
// that make up URL structure); every other character is percent-encoded using its
// UTF-8 bytes with upper-case hex digits. The '%' character is special-cased: a
// '%' that begins a valid escape — two hex digits, or a single hex digit at the
// very end of the string — is preserved as a literal '%', while a '%' that is not
// the start of a valid escape is itself encoded to "%25", which is what prevents
// double-encoding while still repairing stray percent signs.
//
// A few edge cases are worth noting. The input is decoded into runes before
// scanning, so invalid UTF-8 bytes are replaced with the Unicode replacement
// character U+FFFD (which then encodes to its UTF-8 form) rather than being
// passed through raw. The empty string encodes to the empty string, and an input
// that contains only safe characters and valid escapes is returned effectively
// unchanged. Encode does not parse or validate URL structure — it does not know
// about schemes, hosts, or query boundaries — it is purely a character-level
// transform, so it can be applied to a full URL or to any fragment of one.
//
// Parity with the Node original is intentional: the same safe-character set, the
// same "preserve valid escapes, encode stray %" rule, and the same upper-case
// UTF-8 percent-encoding, so Encode reproduces encodeurl's output for typical
// paths and URLs. The differences are idiomatic — Encode takes and returns a Go
// string and uses Go's UTF-8 handling instead of a JavaScript regular-expression
// replace over a JavaScript string.
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
