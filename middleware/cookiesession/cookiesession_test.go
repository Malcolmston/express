package cookiesession

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(New(Options{Secret: "s3cr3t"}))
	return app
}

func TestRoundTrip(t *testing.T) {
	app := newApp()
	app.Get("/set", func(req *express.Request, res *express.Response, next express.Next) {
		Set(req, "user", "alice")
		res.Send("set")
	})
	app.Get("/get", func(req *express.Request, res *express.Response, next express.Next) {
		v, _ := Get(req, "user")
		res.Send(v)
	})

	// First request sets the value and returns a signed cookie.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/set", nil))
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected a session cookie to be set")
	}

	// Second request presents the cookie and reads the value back.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/get", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	app.ServeHTTP(rec2, req2)
	if rec2.Body.String() != "alice" {
		t.Fatalf("body = %q, want alice", rec2.Body.String())
	}
}

func TestTamperedCookieIgnored(t *testing.T) {
	app := newApp()
	app.Get("/get", func(req *express.Request, res *express.Response, next express.Next) {
		if _, ok := Get(req, "user"); ok {
			t.Errorf("tampered cookie should not decode")
		}
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/get", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc.def"})
	app.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestUnmodifiedSessionSetsNoCookie(t *testing.T) {
	app := newApp()
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if len(rec.Result().Cookies()) != 0 {
		t.Fatalf("expected no cookie when session unchanged")
	}
}
