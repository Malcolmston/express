package jwtauth_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/jwtauth"
)

var secret = []byte("test-secret")

// makeToken builds an HS256 JWT with the given claims for testing.
func makeToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	header := map[string]any{"alg": "HS256", "typ": "JWT"}
	enc := func(v any) string {
		b, _ := json.Marshal(v)
		return base64.RawURLEncoding.EncodeToString(b)
	}
	signing := enc(header) + "." + enc(claims)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(signing))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signing + "." + sig
}

func newApp() *express.Application {
	app := express.New()
	app.Use(jwtauth.New(jwtauth.Options{Secret: secret}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		claims, _ := req.Value("claims")
		res.JSON(claims)
	})
	return app
}

func TestValidToken(t *testing.T) {
	app := newApp()
	tok := makeToken(t, map[string]any{"sub": "42"})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if out["sub"] != "42" {
		t.Fatalf("unexpected claims: %v", out)
	}
}

func TestExpiredToken(t *testing.T) {
	app := newApp()
	tok := makeToken(t, map[string]any{"exp": float64(time.Now().Add(-time.Hour).Unix())})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 for expired, got %d", w.Code)
	}
}

func TestBadSignature(t *testing.T) {
	app := newApp()
	tok := makeToken(t, map[string]any{"sub": "42"}) + "tamper"
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 for bad signature, got %d", w.Code)
	}
}

func TestMissingToken(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestValidExp(t *testing.T) {
	app := newApp()
	tok := makeToken(t, map[string]any{"exp": float64(time.Now().Add(time.Hour).Unix())})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200 for valid exp, got %d", w.Code)
	}
}
