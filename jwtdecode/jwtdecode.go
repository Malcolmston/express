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
// The functions validate structure but not content: a token that does not split
// into exactly three dot-separated segments, or whose selected segment is not
// valid base64url or not valid JSON, returns ErrInvalidToken. Empty input is
// therefore rejected as malformed. Compared with the Node package, this port
// covers the header and payload decode paths and returns a typed error instead
// of throwing an InvalidTokenError; it does not accept configuration options
// such as jwt-decode's { header: true } flag, offering the dedicated
// DecodeHeader function instead.
package jwtdecode

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// ErrInvalidToken is returned when the token is malformed.
var ErrInvalidToken = errors.New("jwtdecode: invalid token")

func decodeSegment(seg string) (map[string]interface{}, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "="))
	if err != nil {
		return nil, ErrInvalidToken
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, ErrInvalidToken
	}
	return m, nil
}

// Decode decodes the payload (second segment) of a JWT without verifying it.
func Decode(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodeSegment(parts[1])
}

// DecodeHeader decodes the header (first segment) of a JWT without verifying it.
func DecodeHeader(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodeSegment(parts[0])
}
