package otpauth

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"
)

// Errors returned by the code-generation helpers.
var (
	// ErrInvalidSecret is returned when Config.Secret is empty or is not valid
	// base32.
	ErrInvalidSecret = errors.New("otpauth: invalid base32 secret")
	// ErrUnsupportedAlgorithm is returned for an Algorithm other than SHA1,
	// SHA256 or SHA512.
	ErrUnsupportedAlgorithm = errors.New("otpauth: unsupported algorithm")
	// ErrUnsupportedType is returned for a Type other than "totp" or "hotp".
	ErrUnsupportedType = errors.New("otpauth: unsupported type")
)

// Defaults applied when a Config leaves a field at its zero value. They match
// the defaults of the npm otpauth library and of Google Authenticator.
const (
	// DefaultDigits is the code length used when Config.Digits is zero.
	DefaultDigits = 6
	// DefaultPeriod is the TOTP time step, in seconds, used when Config.Period
	// is zero.
	DefaultPeriod = 30
	// DefaultAlgorithm is the HMAC hash used when Config.Algorithm is empty.
	DefaultAlgorithm = "SHA1"
	// MaxDigits is the code length beyond which dynamic truncation can no
	// longer add information: it yields a 31-bit value, which is at most ten
	// decimal digits. Longer codes are still produced, zero padded on the left,
	// matching the npm library.
	MaxDigits = 10
)

// base32NoPad decodes the unpadded, upper-case base32 alphabet that
// authenticator secrets use.
var base32NoPad = base32.StdEncoding.WithPadding(base32.NoPadding)

// DecodeSecret decodes a base32 shared secret such as "JBSWY3DPEHPK3PXP" into
// its raw bytes.
//
// It is lenient in the way authenticator apps are: surrounding and embedded
// whitespace is removed, the input is upper-cased, and both padded and unpadded
// forms are accepted. An empty or undecodable secret yields ErrInvalidSecret.
func DecodeSecret(secret string) ([]byte, error) {
	s := strings.ToUpper(strings.Join(strings.Fields(secret), ""))
	s = strings.TrimRight(s, "=")
	if s == "" {
		return nil, ErrInvalidSecret
	}
	b, err := base32NoPad.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidSecret
	}
	return b, nil
}

// hasherFor maps an otpauth "algorithm" parameter to its hash constructor. The
// SHA-2 family the npm library accepts is covered; its SHA-3 variants are not,
// since they are outside the standard library.
func hasherFor(alg string) (func() hash.Hash, error) {
	switch strings.ToUpper(strings.TrimSpace(alg)) {
	case "", "SHA1":
		return sha1.New, nil
	case "SHA224":
		return sha256.New224, nil
	case "SHA256":
		return sha256.New, nil
	case "SHA384":
		return sha512.New384, nil
	case "SHA512":
		return sha512.New, nil
	default:
		return nil, ErrUnsupportedAlgorithm
	}
}

// digits returns the effective code length for the config.
func (c Config) digits() int {
	if c.Digits <= 0 {
		return DefaultDigits
	}
	return c.Digits
}

// period returns the effective TOTP time step for the config.
func (c Config) period() int {
	if c.Period > 0 {
		return c.Period
	}
	return DefaultPeriod
}

// kind returns the normalised, lower-case OTP type ("totp" or "hotp"),
// defaulting to "totp" for an empty Type as URL does.
func (c Config) kind() (string, error) {
	switch strings.ToLower(strings.TrimSpace(c.Type)) {
	case "", "totp":
		return "totp", nil
	case "hotp":
		return "hotp", nil
	default:
		return "", ErrUnsupportedType
	}
}

// GenerateCounter computes the one-time password for an explicit counter value,
// the RFC 4226 primitive that both HOTP and TOTP are built on.
//
// The secret is taken from Config.Secret and base32-decoded, the HMAC hash from
// Config.Algorithm (SHA1, SHA256 or SHA512, defaulting to SHA1) and the code
// length from Config.Digits (defaulting to 6). The 8-byte big-endian counter is
// run through HMAC, the result is dynamically truncated per RFC 4226 §5.3, and
// the 31-bit remainder is reduced modulo 10^digits and zero padded.
func (c Config) GenerateCounter(counter uint64) (string, error) {
	key, err := DecodeSecret(c.Secret)
	if err != nil {
		return "", err
	}
	newHash, err := hasherFor(c.Algorithm)
	if err != nil {
		return "", err
	}

	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	mac := hmac.New(newHash, key)
	mac.Write(buf[:])
	sum := mac.Sum(nil)

	// Dynamic truncation (RFC 4226 section 5.3).
	offset := sum[len(sum)-1] & 0x0f
	binCode := (uint64(sum[offset]&0x7f) << 24) |
		(uint64(sum[offset+1]) << 16) |
		(uint64(sum[offset+2]) << 8) |
		uint64(sum[offset+3])

	// A 31-bit truncation is at most ten digits, so beyond MaxDigits the
	// modulus would be a no-op (and would overflow uint64 past 19 digits): the
	// code is simply zero padded, as the npm library's padStart does.
	d := c.digits()
	if d < MaxDigits {
		mod := uint64(1)
		for i := 0; i < d; i++ {
			mod *= 10
		}
		binCode %= mod
	}
	return fmt.Sprintf("%0*d", d, binCode), nil
}

// CounterAt returns the counter value the config uses at time t: the fixed
// Config.Counter for an HOTP config, or the RFC 6238 time step
// floor(unix(t) / Period) for a TOTP one.
func (c Config) CounterAt(t time.Time) (uint64, error) {
	kind, err := c.kind()
	if err != nil {
		return 0, err
	}
	if kind == "hotp" {
		return c.Counter, nil
	}
	sec := t.Unix()
	if sec < 0 {
		sec = 0
	}
	return uint64(sec) / uint64(c.period()), nil
}

// Generate returns the one-time password for the config at time t.
//
// For a TOTP config (the default when Type is empty) the counter is the RFC
// 6238 time step floor(unix(t) / Period); for an HOTP config t is ignored and
// Config.Counter is used directly. This is the equivalent of the npm otpauth
// library's TOTP#generate({timestamp}) and HOTP#generate({counter}).
func (c Config) Generate(t time.Time) (string, error) {
	counter, err := c.CounterAt(t)
	if err != nil {
		return "", err
	}
	return c.GenerateCounter(counter)
}

// Validate checks code against the config and reports the delta, in counter
// steps, between the code and the expected value.
//
// window is the number of steps to search on either side of the expected
// counter, mirroring the npm otpauth validate({window}) option; the returned
// delta is negative when the code is behind (an old code, or a slow clock) and
// positive when it is ahead. A window of zero accepts only the exact step, and
// the default in the Node library is 1. ok is false when nothing in the window
// matches.
//
// Every candidate is compared in constant time and the scan does not stop
// early, so the running time does not reveal which step matched. Callers using
// HOTP must advance Config.Counter past the matched step, otherwise the same
// code can be replayed.
func (c Config) Validate(code string, t time.Time, window int) (delta int, ok bool) {
	if window < 0 {
		window = 0
	}
	base, err := c.CounterAt(t)
	if err != nil {
		return 0, false
	}
	for d := -window; d <= window; d++ {
		var candidate uint64
		switch {
		case d < 0:
			u := uint64(-d)
			if u > base {
				continue
			}
			candidate = base - u
		default:
			candidate = base + uint64(d)
			if candidate < base {
				continue
			}
		}
		expected, err := c.GenerateCounter(candidate)
		if err != nil {
			return 0, false
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 && !ok {
			delta, ok = d, true
		}
	}
	return delta, ok
}
