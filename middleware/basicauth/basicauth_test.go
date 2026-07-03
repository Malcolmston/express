package basicauth_test

import (
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/basicauth"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(basicauth.New(basicauth.Options{
		Realm: "test",
		Verify: func(user, pass string) bool {
			return user == "admin" && pass == "secret"
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func basicHeader(user, pass string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
}

func TestValidCredentials(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", basicHeader("admin", "secret"))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("expected 200 ok, got %d %q", w.Code, w.Body.String())
	}
}

func TestMissingCredentials(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if got := w.Header().Get("WWW-Authenticate"); got != `Basic realm="test"` {
		t.Fatalf("unexpected challenge: %q", got)
	}
}

func TestWrongCredentials(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", basicHeader("admin", "nope"))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestDefaultRealm(t *testing.T) {
	app := express.New()
	app.Use(basicauth.New(basicauth.Options{Verify: func(u, p string) bool { return false }}))
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got := w.Header().Get("WWW-Authenticate"); got != `Basic realm="Restricted"` {
		t.Fatalf("unexpected default realm: %q", got)
	}
}
