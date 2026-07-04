// Package otpauth builds and parses otpauth:// key URIs.
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
	add("secret", c.Secret)
	if c.Issuer != "" {
		add("issuer", c.Issuer)
	}
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
