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
// is not already present, and derives "exp", "nbf", "aud", "iss", "sub" and
// "jti" from SignOptions when those fields are set — refusing, as the Node
// library does, any option that would overwrite a claim the caller already
// supplied. "exp" and "nbf" are measured from the payload's own "iat" when it
// carries one, so a token's lifetime is relative to when it says it was issued.
// The header and payload JSON are
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
// be trusted for authorization — use Verify for that. DecodeComplete is the
// same escape hatch returning header, payload and signature separately, the
// equivalent of Node's jwt.decode(token, {complete: true}).
//
// Verify accepts an optional *VerifyOptions mirroring the Node verify options:
// Algorithms pins the acceptable "alg" values, Audience, Issuer, Subject and
// JwtID assert the corresponding registered claims, ClockTolerance allows
// leeway on the time comparisons, ClockTimestamp fixes the instant they are
// compared against, MaxAge bounds how old an "iat" may be, and
// IgnoreExpiration / IgnoreNotBefore disable the respective checks. Pinning
// Algorithms is the important one: it stops a token from selecting the
// algorithm it is validated with, the shape of the classic JWT
// algorithm-confusion attacks. An Algorithms slice that is non-nil but empty
// permits nothing and therefore rejects every token, so an allowlist computed
// from configuration fails closed rather than open. SignOptions likewise covers
// Audience, JwtID, NoTimestamp, KeyID and extra Header parameters; "alg" and
// "typ" are always written by Sign itself so a caller-supplied header can never
// disagree with the MAC that was computed.
//
// Because every algorithm here is an HMAC, both Sign and Verify refuse two kinds
// of key that cannot be legitimate: an empty secret (ErrMissingKey), which is a
// secret an attacker also has, and PEM-encoded asymmetric key material
// (ErrAsymmetricKey). The latter is the algorithm-confusion attack in its
// simplest form — a public key is public, so anyone could use its bytes as a
// shared secret and mint a token that would otherwise verify.
//
// Relative to Node, this port implements the HMAC family
// (HS256/HS384/HS512) and the exp, nbf, iat, iss, sub, aud and jti registered
// claims; it does not implement RSA or ECDSA algorithms or the "none"
// algorithm, so callers needing those must layer them on top.
package jsonwebtoken

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"sort"
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
	// ErrInvalidAudience is returned when VerifyOptions.Audience is set and the
	// token's "aud" claim does not contain it.
	ErrInvalidAudience = errors.New("jsonwebtoken: invalid audience")
	// ErrInvalidIssuer is returned when VerifyOptions.Issuer is set and the
	// token's "iss" claim does not match any accepted issuer.
	ErrInvalidIssuer = errors.New("jsonwebtoken: invalid issuer")
	// ErrInvalidSubject is returned when VerifyOptions.Subject is set and the
	// token's "sub" claim does not match it.
	ErrInvalidSubject = errors.New("jsonwebtoken: invalid subject")
	// ErrInvalidJwtID is returned when VerifyOptions.JwtID is set and the token's
	// "jti" claim does not match it.
	ErrInvalidJwtID = errors.New("jsonwebtoken: invalid jwt id")
	// ErrTokenTooOld is returned when VerifyOptions.MaxAge is set and the token's
	// "iat" claim is older than that age.
	ErrTokenTooOld = errors.New("jsonwebtoken: maxAge exceeded")
	// ErrMissingIssuedAt is returned when VerifyOptions.MaxAge is set but the
	// token carries no usable "iat" claim, so its age cannot be established.
	ErrMissingIssuedAt = errors.New("jsonwebtoken: iat required when maxAge is set")
	// ErrMissingKey is returned when Sign or Verify is called with no secret. A
	// zero-length HMAC key is a key an attacker also knows, so it is refused
	// rather than used, matching the Node library's "secretOrPrivateKey must
	// have a value".
	ErrMissingKey = errors.New("jsonwebtoken: secret must have a value")
	// ErrAsymmetricKey is returned when an HMAC-signed token is verified against
	// asymmetric key material (a PEM-encoded public or private key, or a
	// certificate). Accepting it is the classic JWT algorithm-confusion attack:
	// the "public" key is public, so anyone could use its bytes as the HMAC
	// secret and mint a token. Node reports "secretOrPublicKey must be a
	// symmetric key when using HS256".
	ErrAsymmetricKey = errors.New("jsonwebtoken: a symmetric key is required for the HMAC algorithms")
	// ErrClaimConflict is returned by Sign when a SignOptions field would
	// overwrite a claim the caller already supplied, for example ExpiresIn
	// against an explicit "exp". Node refuses the same combinations rather than
	// silently preferring one.
	ErrClaimConflict = errors.New("jsonwebtoken: option conflicts with a claim already in the payload")
	// ErrInvalidClaimType is returned when a time claim that must be numeric
	// ("exp" or "nbf") carries some other JSON type. Ignoring it would let a
	// token whose "exp" is the string "0" bypass expiry entirely.
	ErrInvalidClaimType = errors.New("jsonwebtoken: exp and nbf must be numbers")
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
	// Audience, when non-empty, sets the "aud" claim. A single entry is written
	// as a JSON string and several entries as a JSON array, matching how the
	// Node library serialises its audience option.
	Audience []string
	// JwtID, when non-empty, sets the "jti" claim (Node's jwtid option).
	JwtID string
	// NoTimestamp suppresses the automatic "iat" claim (Node's noTimestamp
	// option). An "iat" supplied explicitly in claims is still honoured.
	NoTimestamp bool
	// KeyID, when non-empty, sets the "kid" header parameter (Node's keyid
	// option) so a verifier can select the right key from a key set.
	KeyID string
	// Header holds extra header parameters merged into the JOSE header. The
	// "alg" and "typ" parameters are always written by Sign and cannot be
	// overridden, so a caller cannot accidentally mint a token whose advertised
	// algorithm disagrees with the MAC that was actually computed.
	Header map[string]interface{}
}

// VerifyOptions constrains what Verify will accept. The zero value accepts any
// algorithm this package implements and asserts only the "exp" and "nbf" time
// claims, which is what Verify does when called without options.
//
// Setting Algorithms is strongly recommended. It pins the acceptable "alg"
// values so a token cannot select the algorithm it is checked with, which is
// the shape of the classic JWT algorithm-confusion attacks; it mirrors the
// algorithms option that the Node library documents as mandatory for
// non-symmetric keys.
type VerifyOptions struct {
	// Algorithms is the allowlist of acceptable "alg" header values, for
	// example []string{"HS256"}. A nil slice means "no allowlist configured"
	// and accepts any algorithm this package implements; a non-nil but EMPTY
	// slice permits nothing and therefore rejects every token, which is what
	// the Node library does for algorithms: []. That distinction matters
	// because allowlists are often computed — from configuration, from an
	// environment variable, from an intersection — and an accidentally empty
	// one must fail closed. A token whose alg is not listed yields
	// ErrUnsupportedAlg, checked before the signature is computed.
	Algorithms []string
	// Audience, when non-empty, requires the token's "aud" claim (a string or
	// an array of strings) to contain this value, else ErrInvalidAudience.
	Audience string
	// Issuer, when non-empty, lists the acceptable "iss" values. A token whose
	// issuer is not among them yields ErrInvalidIssuer.
	Issuer []string
	// Subject, when non-empty, requires an exactly matching "sub" claim, else
	// ErrInvalidSubject.
	Subject string
	// JwtID, when non-empty, requires an exactly matching "jti" claim, else
	// ErrInvalidJwtID.
	JwtID string
	// ClockTolerance is the leeway allowed when comparing "exp" and "nbf"
	// against the current time, for clocks that are slightly out of step. It
	// mirrors Node's clockTolerance option (which is expressed in seconds).
	ClockTolerance time.Duration
	// ClockTimestamp overrides the current time used for the "exp", "nbf" and
	// MaxAge comparisons, mirroring Node's clockTimestamp option. The zero
	// value means "read the clock". Setting it makes verification reproducible,
	// which is what tests and audit replays need.
	ClockTimestamp time.Time
	// MaxAge, when greater than zero, rejects a token whose "iat" is further in
	// the past than this duration, even if "exp" has not been reached. A token
	// without a usable "iat" yields ErrMissingIssuedAt.
	MaxAge time.Duration
	// IgnoreExpiration skips the "exp" check. Use it only to inspect an expired
	// token deliberately — never on an authorization path.
	IgnoreExpiration bool
	// IgnoreNotBefore skips the "nbf" check.
	IgnoreNotBefore bool
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

// member is one key of a JSON object, kept with its neighbours in a slice so
// the object can be emitted in a chosen order rather than the arbitrary order a
// Go map would give. The JOSE header and the payload are both signed as exact
// bytes, so their key order is part of the token.
type member struct {
	key   string
	value interface{}
}

// marshalObject encodes members as a JSON object, preserving their order.
func marshalObject(members []member) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, m := range members {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(m.key)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		v, err := json.Marshal(m.value)
		if err != nil {
			return nil, err
		}
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Sign builds and signs a JWT.
//
// A SignOptions field that would overwrite a claim already present in claims —
// ExpiresIn against "exp", NotBefore against "nbf", Audience against "aud",
// Issuer against "iss", Subject against "sub", JwtID against "jti" — is refused
// with ErrClaimConflict rather than silently winning, as the Node library
// refuses it. ExpiresIn and NotBefore are measured from the "iat" the payload
// carries when it has one, and from the signing time otherwise.
//
// The caller's claims are copied, never mutated, and are emitted in sorted key
// order followed by the claims derived from options ("iat", "nbf", "exp", "aud",
// "iss", "sub", "jti") — Go maps have no insertion order to preserve, so a
// deterministic one is chosen.
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
	if len(secret) == 0 {
		return "", ErrMissingKey
	}
	if isAsymmetricKeyMaterial(secret) {
		return "", ErrAsymmetricKey
	}

	// Copy the claims so the caller's map is never mutated.
	payload := Claims{}
	for k, v := range claims {
		payload[k] = v
	}
	for opt, claim := range map[string]string{
		"ExpiresIn": "exp", "NotBefore": "nbf", "Audience": "aud",
		"Issuer": "iss", "Subject": "sub", "JwtID": "jti",
	} {
		set := false
		switch opt {
		case "ExpiresIn":
			set = opts.ExpiresIn > 0
		case "NotBefore":
			set = opts.NotBefore > 0
		case "Audience":
			set = len(opts.Audience) > 0
		case "Issuer":
			set = opts.Issuer != ""
		case "Subject":
			set = opts.Subject != ""
		case "JwtID":
			set = opts.JwtID != ""
		}
		if _, present := payload[claim]; set && present {
			return "", fmt.Errorf("%w: %s and %q", ErrClaimConflict, opt, claim)
		}
	}

	// The reference point for ExpiresIn and NotBefore is the token's own "iat"
	// when the caller supplied one, exactly as Node's timespan() does.
	now := timeNow().Unix()
	base := now
	if iat, ok := toInt64(payload["iat"]); ok {
		base = iat
	}

	var derived []member
	if opts.NoTimestamp {
		// noTimestamp removes an "iat" the caller supplied, as Node does; the
		// expiry is still measured from it.
		delete(payload, "iat")
	} else if _, ok := payload["iat"]; !ok {
		derived = append(derived, member{"iat", now})
	}
	if opts.NotBefore > 0 {
		derived = append(derived, member{"nbf", base + int64(opts.NotBefore/time.Second)})
	}
	if opts.ExpiresIn > 0 {
		derived = append(derived, member{"exp", base + int64(opts.ExpiresIn/time.Second)})
	}
	switch len(opts.Audience) {
	case 0:
	case 1:
		derived = append(derived, member{"aud", opts.Audience[0]})
	default:
		aud := make([]interface{}, len(opts.Audience))
		for i, a := range opts.Audience {
			aud[i] = a
		}
		derived = append(derived, member{"aud", aud})
	}
	if opts.Issuer != "" {
		derived = append(derived, member{"iss", opts.Issuer})
	}
	if opts.Subject != "" {
		derived = append(derived, member{"sub", opts.Subject})
	}
	if opts.JwtID != "" {
		derived = append(derived, member{"jti", opts.JwtID})
	}

	payloadMembers := make([]member, 0, len(payload)+len(derived))
	for _, k := range sortedKeys(payload) {
		payloadMembers = append(payloadMembers, member{k, payload[k]})
	}
	payloadMembers = append(payloadMembers, derived...)

	// "alg" and "typ" come first and are written by Sign itself, so a
	// caller-supplied header parameter can never contradict the algorithm the
	// MAC was actually computed with.
	headerMembers := []member{{"alg", alg}, {"typ", "JWT"}}
	if opts.KeyID != "" {
		headerMembers = append(headerMembers, member{"kid", opts.KeyID})
	}
	for _, k := range sortedKeys(opts.Header) {
		switch k {
		case "alg", "typ":
			continue
		case "kid":
			if opts.KeyID != "" {
				continue
			}
		}
		headerMembers = append(headerMembers, member{k, opts.Header[k]})
	}

	hb, err := marshalObject(headerMembers)
	if err != nil {
		return "", err
	}
	pb, err := marshalObject(payloadMembers)
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

// isAsymmetricKeyMaterial reports whether secret is a PEM-encoded key or
// certificate rather than a symmetric secret. Such bytes are never a legitimate
// HMAC key: a public key is public, so treating it as a shared secret is the
// algorithm-confusion attack.
func isAsymmetricKeyMaterial(secret []byte) bool {
	block, _ := pem.Decode(secret)
	if block == nil {
		return false
	}
	t := strings.ToUpper(block.Type)
	return strings.Contains(t, "KEY") || strings.Contains(t, "CERTIFICATE")
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
//
// The result proves nothing about the token's authenticity — anyone can mint a
// token with arbitrary claims. Never use it for authorization; use Verify.
func Decode(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	return decodePayload(parts[1])
}

// Complete is the fully decomposed form of a token, the equivalent of Node's
// jwt.decode(token, {complete: true}).
type Complete struct {
	// Header holds the decoded JOSE header ("alg", "typ", "kid", ...).
	Header map[string]interface{}
	// Payload holds the decoded claims.
	Payload Claims
	// Signature is the third segment, still base64url-encoded.
	Signature string
}

// DecodeComplete returns the token's header, payload and signature WITHOUT
// verifying anything, mirroring jwt.decode(token, {complete: true}).
//
// It is useful for reading the "kid" header to pick a key before calling
// Verify. As with Decode, nothing it returns is trustworthy on its own.
func DecodeComplete(token string) (*Complete, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	header, err := decodeHeader(parts[0])
	if err != nil {
		return nil, err
	}
	payload, err := decodePayload(parts[1])
	if err != nil {
		return nil, err
	}
	return &Complete{Header: header, Payload: payload, Signature: parts[2]}, nil
}

// Verify checks the signature and standard time claims, returning the claims.
//
// An optional *VerifyOptions pins the acceptable algorithms and asserts the
// registered claims (aud, iss, sub, jti) and token age. Passing no options, or a
// nil one, keeps the historical behaviour: any algorithm this package implements
// is accepted and only "exp" and "nbf" are enforced. Prefer supplying
// VerifyOptions.Algorithms on any authorization path so a token cannot choose
// the algorithm it is validated with.
//
// Claims are only returned once the signature has been confirmed with a
// constant-time comparison; every failure mode returns a nil Claims map, so a
// caller cannot accidentally act on unverified data.
func Verify(token string, secret []byte, opts ...*VerifyOptions) (Claims, error) {
	var o VerifyOptions
	if len(opts) > 0 && opts[0] != nil {
		o = *opts[0]
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	if len(secret) == 0 {
		return nil, ErrMissingKey
	}
	header, err := decodeHeader(parts[0])
	if err != nil {
		return nil, err
	}
	alg, _ := header["alg"].(string)
	if _, err := hasherFor(alg); err != nil {
		return nil, err
	}
	// A nil Algorithms means "not configured"; a non-nil empty one permits
	// nothing, so it must reject rather than fall through to "accept anything".
	if o.Algorithms != nil && !containsString(o.Algorithms, alg) {
		return nil, ErrUnsupportedAlg
	}
	// Every algorithm this package implements is an HMAC, so asymmetric key
	// material can never be the right key: refuse it instead of using its bytes
	// as a shared secret.
	if isAsymmetricKeyMaterial(secret) {
		return nil, ErrAsymmetricKey
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
	if err := checkClaims(claims, &o); err != nil {
		return nil, err
	}
	return claims, nil
}

// checkClaims enforces the time and registered-claim assertions of o against an
// already signature-verified payload.
func checkClaims(claims Claims, o *VerifyOptions) error {
	nowSec := timeNow().Unix()
	if !o.ClockTimestamp.IsZero() {
		nowSec = o.ClockTimestamp.Unix()
	}
	leeway := int64(o.ClockTolerance / time.Second)
	if leeway < 0 {
		leeway = 0
	}

	// A time claim that is present but not a number is a malformed token, not a
	// missing constraint: ignoring it would let an "exp" of "0" (a string) or
	// null defeat the expiry check altogether.
	if !o.IgnoreNotBefore {
		if raw, present := claims["nbf"]; present {
			nbf, ok := toInt64(raw)
			if !ok {
				return ErrInvalidClaimType
			}
			if nowSec+leeway < nbf {
				return ErrTokenNotValidYet
			}
		}
	}
	if !o.IgnoreExpiration {
		if raw, present := claims["exp"]; present {
			exp, ok := toInt64(raw)
			if !ok {
				return ErrInvalidClaimType
			}
			if nowSec-leeway >= exp {
				return ErrTokenExpired
			}
		}
	}
	if o.MaxAge > 0 {
		iat, ok := toInt64(claims["iat"])
		if !ok {
			return ErrMissingIssuedAt
		}
		if nowSec-leeway >= iat+int64(o.MaxAge/time.Second) {
			return ErrTokenTooOld
		}
	}
	if o.Audience != "" && !audienceContains(claims["aud"], o.Audience) {
		return ErrInvalidAudience
	}
	// As with Algorithms, a non-nil but empty Issuer list accepts nothing.
	if o.Issuer != nil {
		iss, _ := claims["iss"].(string)
		if !containsString(o.Issuer, iss) {
			return ErrInvalidIssuer
		}
	}
	if o.Subject != "" {
		if sub, _ := claims["sub"].(string); sub != o.Subject {
			return ErrInvalidSubject
		}
	}
	if o.JwtID != "" {
		if jti, _ := claims["jti"].(string); jti != o.JwtID {
			return ErrInvalidJwtID
		}
	}
	return nil
}

// audienceContains reports whether an "aud" claim — which RFC 7519 allows to be
// either a single string or an array of strings — includes want.
func audienceContains(aud interface{}, want string) bool {
	switch v := aud.(type) {
	case string:
		return v == want
	case []interface{}:
		for _, e := range v {
			if s, ok := e.(string); ok && s == want {
				return true
			}
		}
	case []string:
		return containsString(v, want)
	}
	return false
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
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
