// Package jwtauth provides middleware that verifies HS256-signed JSON Web
// Tokens supplied via the Authorization header. It depends only on the
// standard library.
package jwtauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/malcolmston/express"
)

// Options configures the JWT middleware.
type Options struct {
	// Secret is the HMAC key used to verify the HS256 signature. Required.
	Secret []byte
	// Key is the request value name under which the verified claims are
	// stored. Defaults to "claims".
	Key string
}

// Claims represents the decoded JWT payload.
type Claims map[string]any

// New returns middleware that verifies an HS256 JWT from the
// "Authorization: Bearer <token>" header. On success it stores the decoded
// claims on the request and calls next; otherwise it responds with 401.
func New(opts Options) express.Handler {
	key := opts.Key
	if key == "" {
		key = "claims"
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		token, ok := bearerToken(req.Get("Authorization"))
		if !ok {
			unauthorized(res)
			return
		}
		claims, err := Verify(token, opts.Secret)
		if err != nil {
			unauthorized(res)
			return
		}
		req.Set(key, claims)
		next()
	}
}

// Verify validates an HS256 JWT against secret and returns its claims. It
// checks the signature and, when present, the exp expiration claim.
func Verify(token string, secret []byte) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errInvalid
	}

	// Verify the header declares HS256.
	headerBytes, err := decodeSegment(parts[0])
	if err != nil {
		return nil, errInvalid
	}
	var header struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, errInvalid
	}
	if header.Alg != "HS256" {
		return nil, errInvalid
	}

	// Verify the signature over "header.payload".
	signing := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signing))
	expected := mac.Sum(nil)
	got, err := decodeSegment(parts[2])
	if err != nil {
		return nil, errInvalid
	}
	if !hmac.Equal(expected, got) {
		return nil, errInvalid
	}

	// Decode claims and validate expiration if present.
	payload, err := decodeSegment(parts[1])
	if err != nil {
		return nil, errInvalid
	}
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, errInvalid
	}
	if exp, ok := claims["exp"]; ok {
		if expF, ok := exp.(float64); ok {
			if time.Now().Unix() >= int64(expF) {
				return nil, errExpired
			}
		}
	}
	return claims, nil
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) < len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	t := strings.TrimSpace(header[len(prefix):])
	return t, t != ""
}

// decodeSegment decodes a base64url JWT segment, tolerating missing padding.
func decodeSegment(seg string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(seg, "="))
}

func unauthorized(res *express.Response) {
	res.Set("WWW-Authenticate", "Bearer")
	res.Status(401).Send("Unauthorized")
}

type jwtError string

func (e jwtError) Error() string { return string(e) }

const (
	errInvalid jwtError = "jwtauth: invalid token"
	errExpired jwtError = "jwtauth: token expired"
)
