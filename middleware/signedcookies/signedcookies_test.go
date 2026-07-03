package signedcookies_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/signedcookies"
)

var secret = []byte("cookie-secret")

func newApp(optional bool) *express.Application {
	app := express.New()
	app.Use(signedcookies.New(signedcookies.Options{
		Secret:   secret,
		Name:     "session",
		Optional: optional,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		v, _ := req.Value("session")
		if v == nil {
			res.Send("anon")
			return
		}
		res.Send(v.(string))
	})
	return app
}

func TestValidCookie(t *testing.T) {
	app := newApp(false)
	signed := signedcookies.Sign(secret, "user42")
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: signed})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "user42" {
		t.Fatalf("expected user42, got %d %q", w.Code, w.Body.String())
	}
}

func TestTamperedCookie(t *testing.T) {
	app := newApp(false)
	signed := signedcookies.Sign(secret, "user42")
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: signed + "x"})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 for tampered cookie, got %d", w.Code)
	}
}

func TestMissingCookieRejected(t *testing.T) {
	app := newApp(false)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMissingCookieOptional(t *testing.T) {
	app := newApp(true)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "anon" {
		t.Fatalf("expected anon, got %d %q", w.Code, w.Body.String())
	}
}

func TestVerifyHelper(t *testing.T) {
	signed := signedcookies.Sign(secret, "abc")
	v, ok := signedcookies.Verify(secret, signed)
	if !ok || v != "abc" {
		t.Fatalf("verify failed: %q %v", v, ok)
	}
}
