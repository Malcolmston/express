package jsonwebtoken

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

var testSecret = []byte("my-secret-key")

func TestSignVerifyRoundTrip(t *testing.T) {
	tok, err := Sign(Claims{"name": "John"}, testSecret, &SignOptions{ExpiresIn: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Verify(tok, testSecret)
	if err != nil {
		t.Fatal(err)
	}
	if claims["name"] != "John" {
		t.Errorf("name = %v", claims["name"])
	}
	if _, ok := claims["iat"]; !ok {
		t.Error("expected iat set")
	}
	if _, ok := claims["exp"]; !ok {
		t.Error("expected exp set")
	}
}

func TestAlgs(t *testing.T) {
	for _, alg := range []string{"HS256", "HS384", "HS512"} {
		tok, err := Sign(Claims{"a": 1}, testSecret, &SignOptions{Alg: alg})
		if err != nil {
			t.Fatalf("%s sign: %v", alg, err)
		}
		if _, err := Verify(tok, testSecret); err != nil {
			t.Fatalf("%s verify: %v", alg, err)
		}
	}
}

func TestExpired(t *testing.T) {
	base := time.Unix(1000000000, 0)
	timeNow = func() time.Time { return base }
	defer func() { timeNow = time.Now }()

	tok, err := Sign(Claims{"x": 1}, testSecret, &SignOptions{ExpiresIn: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	// Move clock past expiry.
	timeNow = func() time.Time { return base.Add(time.Hour) }
	_, err = Verify(tok, testSecret)
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestNotBefore(t *testing.T) {
	base := time.Unix(1000000000, 0)
	timeNow = func() time.Time { return base }
	defer func() { timeNow = time.Now }()

	tok, err := Sign(Claims{"x": 1}, testSecret, &SignOptions{NotBefore: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Verify(tok, testSecret)
	if !errors.Is(err, ErrTokenNotValidYet) {
		t.Errorf("expected ErrTokenNotValidYet, got %v", err)
	}

	// After nbf it should verify.
	timeNow = func() time.Time { return base.Add(2 * time.Hour) }
	if _, err := Verify(tok, testSecret); err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestTamperedSignature(t *testing.T) {
	tok, err := Sign(Claims{"x": 1}, testSecret, nil)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	tampered := parts[0] + "." + parts[1] + "." + parts[2] + "abc"
	if _, err := Verify(tampered, testSecret); !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("expected ErrInvalidSignature, got %v", err)
	}
	// Wrong secret.
	if _, err := Verify(tok, []byte("wrong")); !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("expected ErrInvalidSignature for wrong secret, got %v", err)
	}
}

func TestDecodeWithoutVerify(t *testing.T) {
	tok, err := Sign(Claims{"role": "admin"}, testSecret, nil)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := Decode(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims["role"] != "admin" {
		t.Errorf("role = %v", claims["role"])
	}
	if _, err := Decode("bad.token"); !errors.Is(err, ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

// Simulate jwtdecode reading a jsonwebtoken-produced token by manually
// base64url-decoding the payload segment (no cross-package import).
func TestManualPayloadDecode(t *testing.T) {
	tok, err := Sign(Claims{"sub": "42", "name": "Alice"}, testSecret, nil)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatal("expected 3 parts")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["sub"] != "42" || m["name"] != "Alice" {
		t.Errorf("payload = %v", m)
	}
}

func TestUnsupportedAlg(t *testing.T) {
	if _, err := Sign(Claims{}, testSecret, &SignOptions{Alg: "RS256"}); !errors.Is(err, ErrUnsupportedAlg) {
		t.Errorf("expected ErrUnsupportedAlg, got %v", err)
	}
}
