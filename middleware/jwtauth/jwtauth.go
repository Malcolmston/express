// Package jwtauth provides Express middleware that verifies HS256-signed JSON
// Web Tokens supplied via the Authorization header, gating access to protected
// routes. It fills the same role as the Node "express-jwt" middleware (commonly
// paired with "jsonwebtoken"), but is a deliberately small, stdlib-only
// reimplementation: HMAC-SHA256 verification is done with crypto/hmac and
// crypto/sha256, and JSON decoding with encoding/json, so it pulls in no
// third-party dependencies.
//
// Use this middleware to protect API endpoints with stateless bearer-token
// authentication. A client obtains an HS256 JWT elsewhere (a login endpoint,
// an identity provider) and presents it on each request as
// "Authorization: Bearer <token>"; the middleware verifies the signature and
// exposes the decoded claims to downstream handlers. It is appropriate when the
// signer and verifier share a symmetric secret. It does not mint tokens — only
// verify them — and it only supports the HS256 algorithm, so RS256/ES256 or
// asymmetric-key setups are out of scope.
//
// The handler is meant to run before the handlers it protects, either mounted
// globally with app.Use or scoped to a protected router. On each request it
// reads the Authorization header via req.Get, extracts the Bearer token
// (matching the "Bearer " prefix case-insensitively and trimming surrounding
// space), and verifies it with Verify. On success it stores the decoded Claims
// on the request via req.Set under the configured key (default "claims"), so a
// later handler can retrieve them with req.Value(key), and then calls next() to
// continue. On any failure it short-circuits: it sets the WWW-Authenticate:
// Bearer header and responds with 401 Unauthorized, without calling next(), so
// no protected handler runs.
//
// Verification is strict. A token must have exactly three dot-separated
// segments; the header must declare "alg":"HS256"; the HMAC-SHA256 signature
// over "header.payload" must match using a constant-time hmac.Equal comparison;
// and segments are base64url-decoded tolerating missing padding. If the payload
// contains an "exp" claim that decodes as a JSON number, the token is rejected
// once the current time reaches or passes that Unix timestamp. Missing header,
// wrong prefix, empty token, malformed base64, non-HS256 algorithm, bad
// signature, unparseable claims, or an elapsed exp all resolve to a 401. Note
// that "exp" is the only registered claim checked — nbf, iat, iss, aud, and
// friends are decoded into the claims map but not validated — and any error is
// collapsed into the same generic 401 rather than being surfaced to the client.
//
// Compared to the Node express-jwt original, this port keeps the essential
// contract — verify a bearer JWT and attach its claims to the request — but is
// intentionally minimal. It supports only HS256 (no algorithms allowlist,
// asymmetric keys, or JWKS), validates only exp (no nbf/aud/iss options), always
// requires a token (there is no "credentialsRequired: false" optional mode),
// exposes claims under a request key of your choice rather than req.auth/req.user,
// and reports failures as a fixed 401 "Unauthorized" instead of forwarding a
// typed error to an error-handling middleware.
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

// Error implements the error interface; it returns the underlying string.
func (e jwtError) Error() string { return string(e) }

const (
	errInvalid jwtError = "jwtauth: invalid token"
	errExpired jwtError = "jwtauth: token expired"
)
