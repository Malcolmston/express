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

// Typed errors.
var (
	ErrInvalidSignature = errors.New("jsonwebtoken: invalid signature")
	ErrTokenExpired     = errors.New("jsonwebtoken: token expired")
	ErrTokenNotValidYet = errors.New("jsonwebtoken: token not valid yet")
	ErrInvalidToken     = errors.New("jsonwebtoken: invalid token")
	ErrUnsupportedAlg   = errors.New("jsonwebtoken: unsupported algorithm")
)

// timeNow is overridable in tests.
var timeNow = time.Now

// Claims represents the JWT payload.
type Claims map[string]interface{}

// SignOptions configures signing.
type SignOptions struct {
	Alg       string
	ExpiresIn time.Duration
	NotBefore time.Duration
	Issuer    string
	Subject   string
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
