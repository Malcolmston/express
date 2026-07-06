package signedcookies_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/signedcookies"
)

// ExampleNew demonstrates gating a route behind an HMAC-signed cookie. The
// middleware is built with a Secret and the cookie Name to verify, then mounted
// with app.Use so it runs before the protected handler. We mint a cookie value
// with signedcookies.Sign, attach it to a request, and confirm the handler
// reads the verified user id back out via req.Value. A second request whose
// cookie has been tampered with fails the constant-time signature check and is
// rejected with 401 before the handler ever runs. The example drives the app
// entirely in memory with net/http/httptest.
func ExampleNew() {
	secret := []byte("s3cr3t")

	app := express.New()
	app.Use(signedcookies.New(signedcookies.Options{
		Secret: secret,
		Name:   "uid",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		v, _ := req.Value("uid")
		res.Send("hello " + v.(string))
	})

	// A valid, signed cookie is accepted.
	good := httptest.NewRequest(http.MethodGet, "/", nil)
	good.AddCookie(&http.Cookie{Name: "uid", Value: signedcookies.Sign(secret, "alice")})
	gw := httptest.NewRecorder()
	app.ServeHTTP(gw, good)
	fmt.Printf("valid:    %d %s\n", gw.Code, gw.Body.String())

	// A tampered cookie is rejected before the handler runs.
	bad := httptest.NewRequest(http.MethodGet, "/", nil)
	bad.AddCookie(&http.Cookie{Name: "uid", Value: signedcookies.Sign(secret, "alice") + "x"})
	bw := httptest.NewRecorder()
	app.ServeHTTP(bw, bad)
	fmt.Printf("tampered: %d %s\n", bw.Code, bw.Body.String())

	// Output:
	// valid:    200 hello alice
	// tampered: 401 Unauthorized
}

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
