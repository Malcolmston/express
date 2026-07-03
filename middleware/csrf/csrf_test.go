package csrf

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestIssuesCookieOnGet(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(Token(req))
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	sc := rec.Header().Get("Set-Cookie")
	if !strings.HasPrefix(sc, "csrf=") {
		t.Fatalf("expected csrf cookie, got %q", sc)
	}
	if rec.Body.Len() == 0 {
		t.Fatalf("expected token in body")
	}
}

func TestValidPostPasses(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "tok123"})
	req.Header.Set("X-CSRF-Token", "tok123")
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestMismatchRejected(t *testing.T) {
	app := express.New()
	app.Use(New())
	called := false
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		called = true
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "tok123"})
	req.Header.Set("X-CSRF-Token", "wrong")
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if called {
		t.Fatalf("handler must not run on mismatch")
	}
}

func TestMissingTokenRejected(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "tok123"})
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
