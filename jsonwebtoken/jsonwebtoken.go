// Package jsonwebtoken signs and verifies JSON Web Tokens (JWTs) using HMAC
// (HS256/HS384/HS512). It is a Go port of the widely used npm "jsonwebtoken"
// package, covering the symmetric-key signing and verification path that most
// applications reach for, implemented entirely with the standard library
// (crypto/hmac, crypto/sha256, crypto/sha512, crypto/subtle and encoding/json).
//
// JWTs are the de facto token format for stateless authentication and for
// passing signed claims between services: a server signs a small JSON payload
// with a shared secret, hands the resulting string to a client, and can later
// trust that string because tampering would invalidate the signature. This port
// exists so Go programs can produce and check those tokens with the same mental
// model as the Node library — Sign to mint a token, Verify to validate one, and
// Decode to peek at the payload — without depending on a third-party module.
//
// A token is the three base64url segments header.payload.signature joined by
// dots. Sign builds a header of {"alg","typ":"JWT"}, copies the caller's Claims
// so the original map is never mutated, stamps an "iat" (issued-at) claim if one
// is not already present, and derives "exp", "nbf", "iss" and "sub" from
// SignOptions when those fields are set. The header and payload JSON are
// base64url-encoded (raw, unpadded) and joined; that "signing input" is run
// through HMAC with the SHA-256, SHA-384 or SHA-512 hash selected by the "alg"
// option; and the resulting MAC is base64url-encoded as the third segment. The
// default algorithm is HS256, so passing a nil *SignOptions signs an HS256 token
// with only an iat claim.
//
// Verify reverses the process: it splits the token, decodes the header, looks up
// the hash for the header's "alg", recomputes the expected signature over the
// header and payload segments, and compares it against the presented signature
// using crypto/subtle.ConstantTimeCompare so the check is not vulnerable to
// timing attacks. Only after the signature matches are the standard time claims
// enforced — a token whose "exp" is at or before the current time yields
// ErrTokenExpired, and a token whose "nbf" is still in the future yields
// ErrTokenNotValidYet. The clock is read through an overridable package-level
// timeNow, which the internal tests replace to exercise expiry deterministically.
// Malformed input (not exactly three segments, undecodable base64url, or invalid
// JSON) yields ErrInvalidToken, an unrecognised algorithm yields
// ErrUnsupportedAlg, and a wrong secret or altered token yields
// ErrInvalidSignature; all of these are exported sentinel errors suitable for
// errors.Is checks.
//
// Decode is the deliberately unauthenticated escape hatch: it returns the
// payload Claims WITHOUT verifying the signature or the time claims, mirroring
// the fact that a JWT payload is merely encoded, not encrypted. It is convenient
// for inspecting a token you have not yet validated, but its result must never
// be trusted for authorization — use Verify for that. Relative to Node, this
// port implements the HMAC family (HS256/HS384/HS512) and the exp, nbf, iss and
// sub registered claims; it does not implement RSA or ECDSA algorithms, the
// "none" algorithm, audience/issuer assertion during verification, clock-skew
// tolerance, or the full options surface of the original library, so callers
// needing those must layer them on top.
package jsonwebtoken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"strings"
	"time"
)

// Typed errors returned by Verify.
var (
	// ErrInvalidSignature is returned when the token signature does not match.
	ErrInvalidSignature = errors.New("jsonwebtoken: invalid signature")
	// ErrTokenExpired is returned when the token's exp claim is in the past.
	ErrTokenExpired = errors.New("jsonwebtoken: token expired")
	// ErrTokenNotValidYet is returned when the token's nbf claim is in the future.
	ErrTokenNotValidYet = errors.New("jsonwebtoken: token not valid yet")
	// ErrInvalidToken is returned when the token is malformed.
	ErrInvalidToken = errors.New("jsonwebtoken: invalid token")
	// ErrUnsupportedAlg is returned when the token uses an unsupported algorithm.
	ErrUnsupportedAlg = errors.New("jsonwebtoken: unsupported algorithm")
)

// timeNow is overridable in tests.
var timeNow = time.Now

// Claims represents the JWT payload.
type Claims map[string]interface{}

// SignOptions configures signing. The zero value signs an HS256 token with only
// an iat (issued-at) claim.
type SignOptions struct {
	// Alg selects the HMAC algorithm: "HS256", "HS384" or "HS512". When empty
	// it defaults to "HS256"; any other value makes Sign return ErrUnsupportedAlg.
	Alg string
	// ExpiresIn, when greater than zero, sets an "exp" claim to the signing time
	// plus this duration. Zero omits the claim so the token never expires.
	ExpiresIn time.Duration
	// NotBefore, when greater than zero, sets an "nbf" claim to the signing time
	// plus this duration, making the token invalid until then. Zero omits it.
	NotBefore time.Duration
	// Issuer, when non-empty, sets the "iss" claim.
	Issuer string
	// Subject, when non-empty, sets the "sub" claim.
	Subject string
}

func hasherFor(alg string) (func() hash.Hash, error) {
	switch alg {
	case "HS256":
		return sha256.New, nil
	case "HS384":
		return sha512.New384, nil
	case "HS512":
		return sha512.New, nil
	default:
		return nil, ErrUnsupportedAlg
	}
}

func sign(signingInput string, secret []byte, alg string) (string, error) {
	h, err := hasherFor(alg)
	if err != nil {
		return "", err
	}
	mac := hmac.New(h, secret)
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Sign builds and signs a JWT.
func Sign(claims Claims, secret []byte, opts *SignOptions) (string, error) {
	if opts == nil {
		opts = &SignOptions{}
	}
	alg := opts.Alg
	if alg == "" {
		alg = "HS256"
	}
	if _, err := hasherFor(alg); err != nil {
		return "", err
	}

	// Copy claims so we don't mutate the caller's map.
	payload := Claims{}
	for k, v := range claims {
		payload[k] = v
	}

	now := timeNow()
	if _, ok := payload["iat"]; !ok {
		payload["iat"] = now.Unix()
	}
	if opts.ExpiresIn > 0 {
		payload["exp"] = now.Add(opts.ExpiresIn).Unix()
	}
	if opts.NotBefore > 0 {
		payload["nbf"] = now.Add(opts.NotBefore).Unix()
	}
	if opts.Issuer != "" {
		payload["iss"] = opts.Issuer
	}
	if opts.Subject != "" {
		payload["sub"] = opts.Subject
	}

	header := map[string]interface{}{"alg": alg, "typ": "JWT"}
	hb, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	pb, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)
	sig, err := sign(signingInput, secret, alg)
	if err != nil {
		return "", err
	}
	return signingInput + "." + sig, nil
}

func decodePayload(seg string) (Claims, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "="))
	if err != nil {
		return nil, ErrInvalidToken
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, ErrInvalidToken
	}
	return c, nil
}

func decodeHeader(seg string) (map[string]interface{}, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "="))
	if err != nil {
		return nil, ErrInvalidToken
	}
	var h map[string]interface{}
	if err := json.Unmarshal(raw, &h); err != nil {
		return nil, ErrInvalidToken
	}
	return h, nil
}

// Decode returns the payload claims without verifying the signature.
func Decode(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodePayload(parts[1])
}

// Verify checks the signature and standard time claims, returning the claims.
func Verify(token string, secret []byte) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	header, err := decodeHeader(parts[0])
	if err != nil {
		return nil, err
	}
	alg, _ := header["alg"].(string)
	if _, err := hasherFor(alg); err != nil {
		return nil, err
	}
	expected, err := sign(parts[0]+"."+parts[1], secret, alg)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(expected), []byte(parts[2])) != 1 {
		return nil, ErrInvalidSignature
	}

	claims, err := decodePayload(parts[1])
	if err != nil {
		return nil, err
	}

	now := timeNow().Unix()
	if exp, ok := toInt64(claims["exp"]); ok {
		if now >= exp {
			return nil, ErrTokenExpired
		}
	}
	if nbf, ok := toInt64(claims["nbf"]); ok {
		if now < nbf {
			return nil, ErrTokenNotValidYet
		}
	}
	return claims, nil
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	case json.Number:
		i, err := n.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}
