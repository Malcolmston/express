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
	switch strings.ToUpper(algorithm) {
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
