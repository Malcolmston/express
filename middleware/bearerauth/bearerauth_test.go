package bearerauth_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/bearerauth"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(bearerauth.New(bearerauth.Options{
		Verify: func(token string) bool { return token == "good-token" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func TestValidToken(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer good-token")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("expected 200 ok, got %d %q", w.Code, w.Body.String())
	}
}

func TestInvalidToken(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer bad")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMissingHeader(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
