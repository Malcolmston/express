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
// Parity with the Node library is intentionally scoped to the URI layer: this
// port generates and reads otpauth:// URIs and their parameters but does not
// itself compute HOTP/TOTP codes or validate secrets, whereas the full npm
// otpauth module also produces and verifies the numeric passwords. Keeping the
// dependency surface to the standard library (net/url, strconv, strings) makes
// the package a small, focused helper for the common task of generating an
// enrollment URI for a QR code, or extracting the configured parameters from a
// URI a client supplied. Callers that need the actual six-digit code can pair
// the parsed Config with a standard HOTP/TOTP implementation.
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
}

// URL builds an otpauth:// URI from the config.
func URL(c Config) string {
	typ := c.Type
	if typ == "" {
		typ = "totp"
	}

	label := c.Account
	if c.Issuer != "" {
		label = c.Issuer + ":" + c.Account
	}

	// Build the query manually so spaces encode as %20 (not +), matching the
	// otpauth spec and common authenticator apps.
	var params []string
	add := func(k, v string) {
		params = append(params, encode(k)+"="+encode(v))
	}
	// Upstream (hectorm/otpauth) emits the issuer parameter before the secret
	// in its canonical toString() output; match that ordering for parity.
	if c.Issuer != "" {
		add("issuer", c.Issuer)
	}
	add("secret", c.Secret)
	if c.Algorithm != "" {
		add("algorithm", c.Algorithm)
	}
	if c.Digits > 0 {
		add("digits", strconv.Itoa(c.Digits))
	}
	if typ == "hotp" {
		add("counter", strconv.FormatUint(c.Counter, 10))
	} else if c.Period > 0 {
		add("period", strconv.Itoa(c.Period))
	}

	u := url.URL{
		Scheme:   "otpauth",
		Host:     typ,
		Path:     "/" + label,
		RawQuery: strings.Join(params, "&"),
	}
	return u.String()
}

// encode percent-encodes a query component, using %20 for spaces.
func encode(s string) string {
	return strings.ReplaceAll(url.QueryEscape(s), "+", "%20")
}

// Parse parses an otpauth:// URI into a Config.
func Parse(rawurl string) (Config, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return Config{}, err
	}
	if u.Scheme != "otpauth" {
		return Config{}, errors.New("not an otpauth URI")
	}

	c := Config{Type: u.Host}

	label := strings.TrimPrefix(u.Path, "/")
	if idx := strings.Index(label, ":"); idx >= 0 {
		c.Issuer = label[:idx]
		c.Account = strings.TrimLeft(label[idx+1:], " ")
	} else {
		c.Account = label
	}

	q := u.Query()
	c.Secret = q.Get("secret")
	if v := q.Get("issuer"); v != "" {
		c.Issuer = v
	}
	c.Algorithm = q.Get("algorithm")
	if v := q.Get("digits"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid digits: %w", err)
		}
		c.Digits = n
	}
	if v := q.Get("period"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid period: %w", err)
		}
		c.Period = n
	}
	if v := q.Get("counter"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid counter: %w", err)
		}
		c.Counter = n
	}
	return c, nil
}
