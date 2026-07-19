// Package totp implements time-based one-time passwords as defined in RFC 6238,
// the short numeric codes that authenticator apps such as Google Authenticator,
// Authy, and 1Password display for two-factor authentication. It is a
// stdlib-only Go port of the TOTP functionality found in the popular npm
// libraries "otplib", "speakeasy", and "totp": generating the current code from
// a shared secret and verifying a user-supplied code against a small window of
// recent and upcoming time steps. It validates against the RFC 6238 Appendix B
// test vectors.
//
// TOTP is the time-based cousin of HOTP (RFC 4226). Both derive a code by
// applying HMAC to an 8-byte big-endian counter keyed by the shared secret and
// then performing the RFC 4226 "dynamic truncation" to extract a 31-bit number
// that is reduced modulo 10^digits and zero-padded to a fixed width. The one
// difference is what feeds the counter: HOTP uses an explicit counter that both
// sides advance on each use, while TOTP computes the counter from the clock as
// floor(unixTime / period). Because the server and the authenticator app share
// only the secret and read the same wall clock, they independently arrive at the
// same code with no network round trip.
//
// Behavior is controlled by Options: Digits (the code length, default 6),
// Period (the time step in seconds, default 30), and Algorithm (the HMAC hash,
// "SHA1", "SHA256", or "SHA512", default "SHA1"). Passing a nil *Options, or
// leaving any field at its zero value, selects that field's default, matching
// the conventional authenticator-app configuration. An unsupported algorithm
// name causes Generate and GenerateAt to return an error and Verify to report
// false. The 30-second period means a given code is valid for the remainder of
// its step; a period of 30 turns over on every multiple of 30 seconds of Unix
// time.
//
// The secret is supplied as a Base32 string, the encoding used by essentially
// every authenticator app and QR provisioning URI. Decoding is deliberately
// lenient: input is upper-cased, embedded spaces are stripped, and missing "="
// padding is added automatically, so a secret displayed to users as lowercase
// groups like "gezd gnbv gy3t qojq" decodes the same as its canonical form.
// Generate produces the code for the current time (via an overridable internal
// clock), while GenerateAt produces the code for a caller-supplied time.Time,
// which makes deterministic testing and code-at-a-past-instant use cases
// possible.
//
// Verify accepts a window parameter that tolerates clock skew between the server
// and the client's device. It recomputes the expected code for the current step
// and for every step from -window to +window around it, so a window of 1 accepts
// the previous, current, and next steps (roughly a 90-second span with the
// default 30-second period). Each candidate is compared to the supplied code
// with crypto/subtle.ConstantTimeCompare, so a match cannot be distinguished
// from a near-miss by timing, mirroring the constant-time comparison used by the
// reference JavaScript libraries. A window of 0 requires the code to match the
// current step exactly.
package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"strings"
	"time"
)

// timeNow is referenced instead of time.Now directly so tests can override it.
var timeNow = time.Now

// Options configures TOTP generation and verification.
type Options struct {
	Digits    int
	Period    int
	Algorithm string // "SHA1", "SHA256", or "SHA512"
}

func applyDefaults(opts *Options) Options {
	o := Options{Digits: 6, Period: 30, Algorithm: "SHA1"}
	if opts != nil {
		if opts.Digits > 0 {
			o.Digits = opts.Digits
		}
		if opts.Period > 0 {
			o.Period = opts.Period
		}
		if opts.Algorithm != "" {
			o.Algorithm = opts.Algorithm
		}
	}
	return o
}

func hashFor(algorithm string) (func() hash.Hash, error) {
	// Normalize like hectorm/otpauth: upper-case and drop separators so that
	// spellings such as "sha-512" and "SHA512" are treated identically.
	switch strings.ReplaceAll(strings.ToUpper(algorithm), "-", "") {
	case "SHA1", "":
		return sha1.New, nil
	case "SHA256":
		return sha256.New, nil
	case "SHA512":
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}
}

func decodeSecret(base32Secret string) ([]byte, error) {
	s := strings.ToUpper(strings.TrimSpace(base32Secret))
	s = strings.ReplaceAll(s, " ", "")
	// Tolerate missing padding.
	if pad := len(s) % 8; pad != 0 {
		s += strings.Repeat("=", 8-pad)
	}
	return base32.StdEncoding.DecodeString(s)
}

func generate(secret []byte, counter uint64, o Options) (string, error) {
	newHash, err := hashFor(o.Algorithm)
	if err != nil {
		return "", err
	}

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	mac := hmac.New(newHash, secret)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint32(sum[offset]&0x7f) << 24) |
		(uint32(sum[offset+1]) << 16) |
		(uint32(sum[offset+2]) << 8) |
		uint32(sum[offset+3])

	mod := uint32(math.Pow10(o.Digits))
	otp := binCode % mod
	return fmt.Sprintf("%0*d", o.Digits, otp), nil
}

// Generate returns the TOTP value for the current time.
func Generate(base32Secret string, opts *Options) (string, error) {
	return GenerateAt(base32Secret, timeNow(), opts)
}

// GenerateAt returns the TOTP value for the given time.
func GenerateAt(base32Secret string, t time.Time, opts *Options) (string, error) {
	o := applyDefaults(opts)
	secret, err := decodeSecret(base32Secret)
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix()) / uint64(o.Period)
	return generate(secret, counter, o)
}

// Verify reports whether code is valid within +/- window steps of the current time.
func Verify(base32Secret, code string, opts *Options, window int) bool {
	o := applyDefaults(opts)
	secret, err := decodeSecret(base32Secret)
	if err != nil {
		return false
	}
	now := timeNow().Unix()
	base := uint64(now) / uint64(o.Period)
	for i := -window; i <= window; i++ {
		counter := int64(base) + int64(i)
		if counter < 0 {
			continue
		}
		expected, err := generate(secret, uint64(counter), o)
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}
