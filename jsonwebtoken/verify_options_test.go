package jsonwebtoken

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

// seg base64url-encodes a JSON value as one token segment.
func seg(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ---- adversarial: forged and downgraded tokens ------------------------------

// TestAlgNoneRejected covers the classic unsigned-token attack: an attacker
// rewrites the header to {"alg":"none"} and drops the signature.
func TestAlgNoneRejected(t *testing.T) {
	secret := []byte("secret")
	for _, alg := range []string{"none", "None", "NONE", ""} {
		header := seg(t, map[string]interface{}{"alg": alg, "typ": "JWT"})
		payload := seg(t, Claims{"sub": "admin"})
		for _, sig := range []string{"", "AAAA"} {
			token := header + "." + payload + "." + sig
			if _, err := Verify(token, secret); !errors.Is(err, ErrUnsupportedAlg) {
				t.Errorf("alg=%q sig=%q: err = %v, want ErrUnsupportedAlg", alg, sig, err)
			}
		}
	}
}

// TestAlgorithmsPinningRejectsOtherAlgorithms shows that a token cannot pick
// the algorithm it is checked with once VerifyOptions.Algorithms is set, even
// though the substituted algorithm is one the package implements.
func TestAlgorithmsPinningRejectsOtherAlgorithms(t *testing.T) {
	secret := []byte("shared-secret")
	token, err := Sign(Claims{"sub": "alice"}, secret, &SignOptions{Alg: "HS512"})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	// Without pinning the HS512 token verifies fine.
	if _, err := Verify(token, secret); err != nil {
		t.Fatalf("unpinned Verify: %v", err)
	}
	// Pinned to HS256, the same token is refused before any MAC is computed.
	_, err = Verify(token, secret, &VerifyOptions{Algorithms: []string{"HS256"}})
	if !errors.Is(err, ErrUnsupportedAlg) {
		t.Fatalf("pinned Verify: err = %v, want ErrUnsupportedAlg", err)
	}
	// And it still verifies when the pin names its own algorithm.
	if _, err := Verify(token, secret, &VerifyOptions{Algorithms: []string{"HS384", "HS512"}}); err != nil {
		t.Fatalf("matching pin: %v", err)
	}
}

// TestTamperedPayloadRejected changes a claim while keeping the signature.
func TestTamperedPayloadRejected(t *testing.T) {
	secret := []byte("secret")
	token, err := Sign(Claims{"role": "user"}, secret, nil)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	parts := strings.Split(token, ".")
	forged := parts[0] + "." + seg(t, Claims{"role": "admin"}) + "." + parts[2]

	if _, err := Verify(forged, secret); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("err = %v, want ErrInvalidSignature", err)
	}
	// Decode still reads it — which is exactly why Decode must not be trusted.
	claims, err := Decode(forged)
	if err != nil || claims["role"] != "admin" {
		t.Fatalf("Decode(forged) = %v, %v", claims, err)
	}
}

// TestSignatureFromDifferentSecretRejected is the wrong-key case.
func TestSignatureFromDifferentSecretRejected(t *testing.T) {
	token, _ := Sign(Claims{"sub": "alice"}, []byte("key-a"), nil)
	if _, err := Verify(token, []byte("key-b")); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("err = %v, want ErrInvalidSignature", err)
	}
}

// TestVerifyReturnsNilClaimsOnEveryFailure makes sure no failure path leaks a
// usable claims map that a careless caller might act on.
func TestVerifyReturnsNilClaimsOnEveryFailure(t *testing.T) {
	secret := []byte("secret")
	good, _ := Sign(Claims{"sub": "a", "exp": time.Now().Add(-time.Hour).Unix()}, secret, nil)
	tokens := []string{
		"",
		"a.b",
		"a.b.c.d",
		"!!!." + seg(t, Claims{}) + ".x",
		good, // expired
	}
	for _, tok := range tokens {
		claims, err := Verify(tok, secret)
		if err == nil {
			t.Errorf("Verify(%q) unexpectedly succeeded", tok)
		}
		if claims != nil {
			t.Errorf("Verify(%q) returned claims %v alongside an error", tok, claims)
		}
	}
}

// ---- registered claim assertions --------------------------------------------

func TestVerifyAudience(t *testing.T) {
	secret := []byte("secret")

	single, _ := Sign(Claims{}, secret, &SignOptions{Audience: []string{"api"}})
	multi, _ := Sign(Claims{}, secret, &SignOptions{Audience: []string{"api", "admin"}})

	// Sign writes one audience as a string and several as an array, like Node.
	c, _ := Decode(single)
	if _, ok := c["aud"].(string); !ok {
		t.Errorf("single audience should serialise as a string, got %T", c["aud"])
	}
	c, _ = Decode(multi)
	if _, ok := c["aud"].([]interface{}); !ok {
		t.Errorf("multiple audiences should serialise as an array, got %T", c["aud"])
	}

	for _, tok := range []string{single, multi} {
		if _, err := Verify(tok, secret, &VerifyOptions{Audience: "api"}); err != nil {
			t.Errorf("Verify with matching audience: %v", err)
		}
		if _, err := Verify(tok, secret, &VerifyOptions{Audience: "other"}); !errors.Is(err, ErrInvalidAudience) {
			t.Errorf("Verify with wrong audience: err = %v, want ErrInvalidAudience", err)
		}
	}

	// A token with no aud claim at all fails an audience assertion.
	none, _ := Sign(Claims{}, secret, nil)
	if _, err := Verify(none, secret, &VerifyOptions{Audience: "api"}); !errors.Is(err, ErrInvalidAudience) {
		t.Errorf("audience assertion on a token without aud: err = %v", err)
	}
}

func TestVerifyIssuerSubjectAndJwtID(t *testing.T) {
	secret := []byte("secret")
	token, err := Sign(Claims{}, secret, &SignOptions{
		Issuer:  "auth.example.com",
		Subject: "user-42",
		JwtID:   "abc123",
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	ok := &VerifyOptions{
		Issuer:  []string{"other.example.com", "auth.example.com"},
		Subject: "user-42",
		JwtID:   "abc123",
	}
	if _, err := Verify(token, secret, ok); err != nil {
		t.Fatalf("Verify with matching claims: %v", err)
	}

	tests := []struct {
		name string
		opts *VerifyOptions
		want error
	}{
		{"issuer", &VerifyOptions{Issuer: []string{"evil.example.com"}}, ErrInvalidIssuer},
		{"subject", &VerifyOptions{Subject: "user-43"}, ErrInvalidSubject},
		{"jwtid", &VerifyOptions{JwtID: "zzz"}, ErrInvalidJwtID},
	}
	for _, tc := range tests {
		if _, err := Verify(token, secret, tc.opts); !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
}

// ---- time claims -------------------------------------------------------------

func TestClockToleranceAndIgnoreFlags(t *testing.T) {
	secret := []byte("secret")
	base := time.Unix(1_700_000_000, 0)
	timeNow = func() time.Time { return base }
	defer func() { timeNow = time.Now }()

	token, _ := Sign(Claims{}, secret, &SignOptions{ExpiresIn: 60 * time.Second, NotBefore: 30 * time.Second})

	// nbf is 30s out: rejected now, accepted with 30s of leeway.
	if _, err := Verify(token, secret); !errors.Is(err, ErrTokenNotValidYet) {
		t.Fatalf("err = %v, want ErrTokenNotValidYet", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{ClockTolerance: 30 * time.Second}); err != nil {
		t.Fatalf("with tolerance: %v", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{IgnoreNotBefore: true}); err != nil {
		t.Fatalf("with IgnoreNotBefore: %v", err)
	}

	// Jump past expiry.
	timeNow = func() time.Time { return base.Add(120 * time.Second) }
	if _, err := Verify(token, secret, &VerifyOptions{IgnoreNotBefore: true}); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("err = %v, want ErrTokenExpired", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{IgnoreNotBefore: true, IgnoreExpiration: true}); err != nil {
		t.Fatalf("with IgnoreExpiration: %v", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{IgnoreNotBefore: true, ClockTolerance: 90 * time.Second}); err != nil {
		t.Fatalf("expiry within tolerance: %v", err)
	}
}

func TestMaxAge(t *testing.T) {
	secret := []byte("secret")
	base := time.Unix(1_700_000_000, 0)
	timeNow = func() time.Time { return base }
	defer func() { timeNow = time.Now }()

	// A long-lived token that is nonetheless old.
	token, _ := Sign(Claims{}, secret, &SignOptions{ExpiresIn: 365 * 24 * time.Hour})

	timeNow = func() time.Time { return base.Add(2 * time.Hour) }
	if _, err := Verify(token, secret); err != nil {
		t.Fatalf("without MaxAge the token is still valid: %v", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{MaxAge: time.Hour}); !errors.Is(err, ErrTokenTooOld) {
		t.Fatalf("err = %v, want ErrTokenTooOld", err)
	}
	if _, err := Verify(token, secret, &VerifyOptions{MaxAge: 3 * time.Hour}); err != nil {
		t.Fatalf("within MaxAge: %v", err)
	}

	// MaxAge is unenforceable without an iat claim, so it must fail closed.
	noIat, _ := Sign(Claims{}, secret, &SignOptions{NoTimestamp: true})
	if _, err := Verify(noIat, secret, &VerifyOptions{MaxAge: time.Hour}); !errors.Is(err, ErrMissingIssuedAt) {
		t.Fatalf("err = %v, want ErrMissingIssuedAt", err)
	}
}

// ---- sign options -------------------------------------------------------------

func TestNoTimestampAndKeyIDAndHeader(t *testing.T) {
	secret := []byte("secret")
	token, err := Sign(Claims{"sub": "a"}, secret, &SignOptions{
		NoTimestamp: true,
		KeyID:       "key-1",
		Header:      map[string]interface{}{"cty": "JWT", "alg": "none", "typ": "evil"},
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}

	c, err := DecodeComplete(token)
	if err != nil {
		t.Fatalf("DecodeComplete: %v", err)
	}
	if _, ok := c.Payload["iat"]; ok {
		t.Error("NoTimestamp should suppress iat")
	}
	if c.Header["kid"] != "key-1" {
		t.Errorf("kid = %v", c.Header["kid"])
	}
	if c.Header["cty"] != "JWT" {
		t.Errorf("extra header parameters should survive, cty = %v", c.Header["cty"])
	}
	// A caller-supplied alg/typ must never override the ones Sign computed,
	// otherwise the advertised algorithm could disagree with the actual MAC.
	if c.Header["alg"] != "HS256" || c.Header["typ"] != "JWT" {
		t.Errorf("alg/typ were overridable: %v", c.Header)
	}
	if c.Signature == "" {
		t.Error("signature segment missing")
	}
	if _, err := Verify(token, secret, &VerifyOptions{Algorithms: []string{"HS256"}}); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

// NoTimestamp removes the "iat" claim even when the caller supplied one
// explicitly, as the Node library's noTimestamp option does — but the expiry is
// still measured from that iat.
func TestNoTimestampDropsExplicitIat(t *testing.T) {
	token, err := Sign(Claims{"iat": int64(42)}, []byte("s"), &SignOptions{
		NoTimestamp: true,
		ExpiresIn:   time.Hour,
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	c, _ := Decode(token)
	if _, present := c["iat"]; present {
		t.Errorf("iat = %v, want it dropped by NoTimestamp", c["iat"])
	}
	if got, _ := toInt64(c["exp"]); got != 42+3600 {
		t.Errorf("exp = %v, want the supplied iat plus an hour", c["exp"])
	}
}

func TestDecodeCompleteRejectsMalformed(t *testing.T) {
	for _, tok := range []string{"", "a.b", "a.b.c.d", "!!!.e30.x"} {
		if _, err := DecodeComplete(tok); err == nil {
			t.Errorf("DecodeComplete(%q) unexpectedly succeeded", tok)
		}
	}
}

func TestSignRejectsUnsupportedAlgorithm(t *testing.T) {
	for _, alg := range []string{"none", "RS256", "ES256", "hs256"} {
		if _, err := Sign(Claims{}, []byte("s"), &SignOptions{Alg: alg}); !errors.Is(err, ErrUnsupportedAlg) {
			t.Errorf("Sign(alg=%q): err = %v, want ErrUnsupportedAlg", alg, err)
		}
	}
}

func TestSignDoesNotMutateCallerClaims(t *testing.T) {
	claims := Claims{"sub": "a"}
	if _, err := Sign(claims, []byte("s"), &SignOptions{Issuer: "iss", Audience: []string{"aud"}}); err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(claims) != 1 {
		t.Fatalf("caller claims were mutated: %v", claims)
	}
}
