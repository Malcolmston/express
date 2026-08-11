// Package otpauth builds and parses otpauth:// key URIs, a stdlib-only Go port
// of the URI portion of the npm "otpauth" library
// (https://www.npmjs.com/package/otpauth). One-time passwords are the short
// numeric codes used for two-factor authentication; the otpauth:// URI is the
// de-facto standard, originally defined by Google Authenticator, for handing a
// shared secret and its parameters to an authenticator app, most commonly by
// encoding the URI in a QR code that the user scans during account setup.
//
// This package concerns itself with the transport format — constructing a
// correct otpauth:// URI from a Config and parsing one back into a Config. It
// covers both one-time-password families defined by the standard: HOTP (HMAC
// based one-time passwords, RFC 4226), which are counter-based and advance each
// time a code is consumed, and TOTP (time-based one-time passwords, RFC 6238),
// which derive the code from the current time divided into fixed-length steps.
// The Type field selects between them, defaulting to "totp" when left empty.
//
// A URI has the shape otpauth://TYPE/LABEL?PARAMETERS. The LABEL identifies the
// account and is conventionally "Issuer:Account" (for example
// "ACME Co:alice@example.com"); URL builds this from the Issuer and Account
// fields, and Parse splits it back apart, also honoring an explicit issuer query
// parameter which takes precedence when present. The query string carries the
// secret plus optional metadata: secret (a base32-encoded shared key such as
// "JBSWY3DPEHPK3PXP"), issuer, algorithm (SHA1, SHA256 or SHA512), digits (the
// code length, typically 6), and either period (the TOTP time step in seconds,
// typically 30) for TOTP or counter (the initial HOTP counter) for HOTP.
//
// URL emits parameters in a fixed, spec-friendly order and percent-encodes them
// so that spaces appear as %20 rather than "+", matching the otpauth
// specification and the expectations of common authenticator apps; a period is
// only written for TOTP and a counter only for HOTP. Parse is tolerant on
// input: it accepts any parameter ordering, treats a wrong scheme as an error,
// returns descriptive errors when digits, period or counter are not valid
// integers, and leaves optional fields at their zero values when absent. URL
// followed by Parse round-trips the meaningful fields, as the package tests
// demonstrate.
//
// Beyond the URI layer the package also computes and checks the numeric codes,
// matching the npm library's TOTP#generate / HOTP#generate and validate:
// Config.Generate returns the code for a given time (or, for HOTP, for
// Config.Counter), Config.GenerateCounter exposes the underlying RFC 4226
// primitive for an explicit counter, and Config.Validate checks a submitted
// code against a window of steps either side of the expected one, returning
// the step offset so a caller can detect clock drift or advance an HOTP
// counter. Secrets are decoded from base32 by DecodeSecret, leniently enough to
// accept the spaced, lower-case forms users copy out of enrollment pages. All
// three OATH hash algorithms (SHA1, SHA256, SHA512) are supported and the
// implementation is checked against the RFC 4226 Appendix D and RFC 6238
// Appendix B test vectors. Comparison in Validate is constant time.
//
// Everything is built on the standard library (net/url, encoding/base32,
// crypto/hmac and the crypto hash packages), so the package covers the whole
// enrollment-to-verification path — build a URI for a QR code, parse the URI a
// client scanned, and check the codes it produces — with no dependencies.
package otpauth

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Config describes an otpauth key URI.
type Config struct {
	Type      string // "totp" or "hotp"
	Issuer    string
	Account   string
	Secret    string
	Digits    int
	Period    int
	Algorithm string
	Counter   uint64
	// OmitIssuerFromLabel keeps the issuer in the query string only, leaving it
	// out of the "Issuer:Account" label. It mirrors the npm library's
	// issuerInLabel flag, which Parse clears when it reads a URI whose issuer
	// appeared only as a query parameter. The zero value puts the issuer in the
	// label, which is what authenticator apps expect.
	OmitIssuerFromLabel bool
}

// URL builds an otpauth:// URI from the config.
//
// The output matches the canonical form of the npm library's toString(): the
// parameters appear in the order issuer, secret, algorithm, digits and then
// period (TOTP) or counter (HOTP), with the effective defaults written out
// rather than omitted, and every component percent-encoded the way JavaScript's
// encodeURIComponent does — spaces as %20, and "!'()*-._~" left alone.
func URL(c Config) string {
	typ := strings.ToLower(strings.TrimSpace(c.Type))
	if typ == "" {
		typ = "totp"
	}

	var b strings.Builder
	b.WriteString("otpauth://")
	b.WriteString(typ)
	b.WriteString("/")
	if c.Issuer != "" {
		if !c.OmitIssuerFromLabel {
			b.WriteString(encode(c.Issuer))
			b.WriteString(":")
		}
		b.WriteString(encode(c.Account))
		b.WriteString("?issuer=")
		b.WriteString(encode(c.Issuer))
		b.WriteString("&")
	} else {
		b.WriteString(encode(c.Account))
		b.WriteString("?")
	}
	b.WriteString("secret=")
	b.WriteString(encode(NormalizeSecret(c.Secret)))
	b.WriteString("&algorithm=")
	b.WriteString(encode(c.algorithm()))
	b.WriteString("&digits=")
	b.WriteString(strconv.Itoa(c.digits()))
	if typ == "hotp" {
		b.WriteString("&counter=")
		b.WriteString(strconv.FormatUint(c.Counter, 10))
	} else {
		b.WriteString("&period=")
		b.WriteString(strconv.Itoa(c.period()))
	}
	return b.String()
}

// encode percent-encodes one URI component exactly as JavaScript's
// encodeURIComponent does: unreserved characters plus !'()*-._~ pass through,
// a space becomes %20 rather than '+', and everything else is escaped as its
// UTF-8 bytes in upper-case hex.
func encode(s string) string {
	const keep = "-_.!~*'()"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'A' && ch <= 'Z', ch >= 'a' && ch <= 'z', ch >= '0' && ch <= '9',
			strings.IndexByte(keep, ch) >= 0:
			b.WriteByte(ch)
		default:
			b.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}
	return b.String()
}

// Errors returned by Parse. They mirror the failures the npm library's
// URI.parse throws, which are all reported before any code is generated.
var (
	// ErrInvalidURI is returned when the string is not an otpauth:// URI of the
	// form otpauth://TYPE/LABEL?PARAMETERS.
	ErrInvalidURI = errors.New("otpauth: invalid URI format")
	// ErrMissingSecret is returned when the "secret" parameter is absent or is
	// not base32.
	ErrMissingSecret = errors.New("otpauth: missing or invalid 'secret' parameter")
	// ErrMissingCounter is returned for an hotp URI without a valid "counter"
	// parameter, which RFC 4226 makes mandatory.
	ErrMissingCounter = errors.New("otpauth: missing or invalid 'counter' parameter")
	// ErrInvalidDigits is returned when "digits" is not a positive integer.
	ErrInvalidDigits = errors.New("otpauth: invalid 'digits' parameter")
	// ErrInvalidPeriod is returned when "period" is not a positive integer.
	ErrInvalidPeriod = errors.New("otpauth: invalid 'period' parameter")
	// ErrInvalidAlgorithm is returned when "algorithm" does not name one of the
	// SHA algorithms the otpauth URI scheme allows.
	ErrInvalidAlgorithm = errors.New("otpauth: invalid 'algorithm' parameter")
)

// base32Chars reports whether s is a non-empty base32 string, optionally
// followed by '=' padding, the same shape the npm library requires of a
// "secret" parameter.
func base32Chars(s string) bool {
	if s == "" {
		return false
	}
	body := strings.TrimRight(s, "=")
	if body == "" {
		return false
	}
	for _, r := range body {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
			// Only A-Z are base32 digits; the npm regex is case-insensitive
			// over [2-7A-Z].
			u := r
			if u >= 'a' && u <= 'z' {
				u -= 'a' - 'A'
			}
			if u < 'A' || u > 'Z' {
				return false
			}
		case r >= '2' && r <= '7':
		default:
			return false
		}
	}
	return true
}

// NormalizeSecret canonicalises a base32 secret the way the npm library's
// Secret does: whitespace removed, upper-cased and '=' padding stripped. It
// leaves an undecodable secret alone rather than reporting an error; use
// DecodeSecret to validate one.
func NormalizeSecret(secret string) string {
	return strings.TrimRight(strings.ToUpper(strings.Join(strings.Fields(secret), "")), "=")
}

// algorithm returns the effective HMAC algorithm name, upper-cased.
func (c Config) algorithm() string {
	if a := strings.ToUpper(strings.TrimSpace(c.Algorithm)); a != "" {
		return a
	}
	return DefaultAlgorithm
}

// validAlgorithmName reports whether name is one of the algorithms the otpauth
// URI scheme allows. SHA-3 variants are accepted here — they are legal in a URI
// — even though Generate cannot compute them without a non-stdlib dependency.
func validAlgorithmName(name string) bool {
	switch strings.ToUpper(name) {
	case "SHA1", "SHA224", "SHA256", "SHA384", "SHA512",
		"SHA3-224", "SHA3-256", "SHA3-384", "SHA3-512":
		return true
	}
	return false
}

// query is a parsed otpauth:// query string. Lookups are case-insensitive on
// the parameter name, and the first occurrence of a name wins, as the npm
// library's reduce over the '&'-separated pairs does.
type query map[string]string

// Get returns the value of the named parameter, or "" when it is absent.
func (q query) Get(name string) string { return q[strings.ToLower(name)] }

// parseQuery splits a raw otpauth:// query string the way the npm library does:
// on '&', then on the first '=', percent-decoding each half. Unlike
// net/url.ParseQuery it does NOT treat '+' as a space — decodeURIComponent
// leaves '+' alone, and a '+' is legal in an issuer or label.
func parseQuery(raw string) query {
	q := query{}
	for _, pair := range strings.Split(raw, "&") {
		if pair == "" {
			continue
		}
		name, value := pair, ""
		if i := strings.Index(pair, "="); i >= 0 {
			name, value = pair[:i], pair[i+1:]
		}
		key := strings.ToLower(percentDecode(name))
		q[key] = percentDecode(value)
	}
	return q
}

// percentDecode undoes percent-encoding, leaving anything that is not a valid
// escape untouched, which is how decodeURIComponent behaves for the inputs an
// otpauth URI can carry.
func percentDecode(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	if out, err := url.PathUnescape(s); err == nil {
		return out
	}
	return s
}

// positiveInt parses a strictly positive decimal integer, optionally signed
// with '+', as the npm library's POSITIVE_INTEGER_REGEX requires.
func positiveInt(s string) (int, bool) {
	t := strings.TrimPrefix(s, "+")
	if t == "" || t[0] == '0' {
		return 0, false
	}
	n, err := strconv.Atoi(t)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}

// Parse parses an otpauth:// URI into a Config.
//
// It validates the same things the npm library's URI.parse does, and fails
// rather than guessing: the scheme must be otpauth and the type "totp" or
// "hotp"; "secret" is mandatory and must be base32; an hotp URI must carry a
// "counter"; and "digits" and "period" must be positive integers. Parameters
// that are absent are filled in with the scheme's defaults (SHA1, 6 digits and,
// for TOTP, a 30-second period), so the returned Config describes the same
// one-time passwords the URI does.
func Parse(rawurl string) (Config, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return Config{}, ErrInvalidURI
	}
	if !strings.EqualFold(u.Scheme, "otpauth") {
		return Config{}, ErrInvalidURI
	}
	typ := strings.ToLower(u.Host)
	if typ != "totp" && typ != "hotp" {
		return Config{}, ErrInvalidURI
	}
	if u.RawQuery == "" {
		return Config{}, ErrInvalidURI
	}

	c := Config{Type: typ}

	label := strings.TrimPrefix(u.Path, "/")
	if label == "" {
		// The npm library's URI regex requires a non-empty label.
		return Config{}, ErrInvalidURI
	}
	labelHasIssuer := false
	if idx := strings.Index(label, ":"); idx >= 0 {
		labelHasIssuer = true
		c.Issuer = label[:idx]
		c.Account = strings.TrimLeft(label[idx+1:], " ")
	} else {
		c.Account = label
	}

	q := parseQuery(u.RawQuery)
	if v := q.Get("issuer"); v != "" {
		c.Issuer = v
		// An issuer that appears only in the query string stays there when the
		// config is rendered back to a URI, matching issuerInLabel=false.
		c.OmitIssuerFromLabel = !labelHasIssuer
	}

	secret := q.Get("secret")
	if !base32Chars(secret) {
		return Config{}, ErrMissingSecret
	}
	// Canonicalise as the npm library's Secret does, so the Config carries the
	// same base32 string a URI built from it would.
	c.Secret = NormalizeSecret(secret)

	if v := q.Get("algorithm"); v != "" {
		if !validAlgorithmName(v) {
			return Config{}, ErrInvalidAlgorithm
		}
		c.Algorithm = strings.ToUpper(v)
	} else {
		c.Algorithm = DefaultAlgorithm
	}

	if v := q.Get("digits"); v != "" {
		n, ok := positiveInt(v)
		if !ok {
			return Config{}, ErrInvalidDigits
		}
		c.Digits = n
	} else {
		c.Digits = DefaultDigits
	}

	if typ == "hotp" {
		v := q.Get("counter")
		n, err := strconv.ParseInt(v, 10, 64)
		if v == "" || err != nil || n < 0 {
			return Config{}, ErrMissingCounter
		}
		c.Counter = uint64(n)
	} else if v := q.Get("period"); v != "" {
		n, ok := positiveInt(v)
		if !ok {
			return Config{}, ErrInvalidPeriod
		}
		c.Period = n
	} else {
		c.Period = DefaultPeriod
	}
	return c, nil
}
