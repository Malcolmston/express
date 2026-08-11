// Package jwtdecode decodes the payload of a JSON Web Token without verifying
// its signature. It is a Go port of the npm "jwt-decode" package, implemented
// with only the standard library (encoding/base64 and encoding/json).
//
// The jwt-decode library is a tiny, browser-friendly helper whose single job is
// to turn a JWT string into a plain object so an application can read the claims
// it carries — a user id, a display name, an expiry timestamp — for display or
// routing decisions. It exists precisely because full JWT libraries are heavier
// than needed when all you want is to look inside a token you already hold. This
// port serves the same purpose in Go: it exposes Decode for the payload and
// DecodeHeader for the header, each returning a map[string]interface{} shaped
// exactly like the decoded JSON.
//
// A JWT is three base64url segments — header, payload and signature — joined by
// dots. Decoding splits on the dot, selects the requested segment, base64url-
// decodes it (tolerating stray "=" padding by trimming it before decoding), and
// unmarshals the resulting bytes as JSON. Because JSON numbers unmarshal into
// Go float64 values, numeric claims such as "iat" or "exp" come back as float64
// and string claims as string. No cryptography is performed at any point.
//
// This is intentionally an UNVERIFIED read. The signature segment is never
// examined and no secret or public key is consulted, so a decoded payload proves
// nothing about the token's authenticity — anyone can craft a token with any
// claims they like. Never trust the result for authorization or authentication;
// use the jsonwebtoken package's Verify to validate a signed token before acting
// on its claims. jwtdecode is for reading, not trusting.
//
// The functions validate structure but not content, and only as far as the
// segment they are asked for: like the Node package they select segment #1
// (DecodeHeader) or #2 (Decode) from the dot-separated string and never look at
// the rest, so a two-segment string still yields a payload and a fourth segment
// is ignored. A missing segment, a segment that is not valid base64 or not a
// valid JSON object returns ErrInvalidToken; empty input is therefore rejected
// as malformed. Decoding follows the Node implementation exactly: '-' and '_'
// are rewritten to '+' and '/', the length is padded up to a multiple of four
// (a length that is one more than a multiple of four cannot be base64 and is
// rejected), and both base64 alphabets are consequently accepted. Compared with
// the Node package, this port returns a typed error instead of throwing an
// InvalidTokenError, and offers the dedicated DecodeHeader function rather than
// jwt-decode's { header: true } flag; because it returns a
// map[string]interface{}, a payload that is valid JSON but not an object (a
// bare number, string, array or boolean) is an error here where jwt-decode
// returns the value.
package jwtdecode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidToken is returned when the token is malformed.
var ErrInvalidToken = errors.New("jwtdecode: invalid token")

// b64Decode mirrors the base64UrlDecode of jwt-decode: rewrite the URL-safe
// alphabet onto the standard one, restore the padding the encoder stripped, and
// reject a length that no base64 string can have.
func b64Decode(seg string) ([]byte, error) {
	s := strings.ReplaceAll(strings.ReplaceAll(seg, "-", "+"), "_", "/")
	switch len(s) % 4 {
	case 0:
	case 2:
		s += "=="
	case 3:
		s += "="
	default:
		return nil, ErrInvalidToken
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return raw, nil
}

func decodeSegment(token string, index int) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if index >= len(parts) {
		return nil, ErrInvalidToken
	}
	raw, err := b64Decode(parts[index])
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, ErrInvalidToken
	}
	return m, nil
}

// Decode decodes the payload (second segment) of a JWT without verifying it.
func Decode(token string) (map[string]interface{}, error) {
	return decodeSegment(token, 1)
}

// DecodeHeader decodes the header (first segment) of a JWT without verifying it.
func DecodeHeader(token string) (map[string]interface{}, error) {
	return decodeSegment(token, 0)
}
