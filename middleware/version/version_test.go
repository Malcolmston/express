package version_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/version"
)

func TestVersionEndpoint(t *testing.T) {
	app := express.New()
	app.Use(version.New(version.Options{Version: "1.2.3"}))
	r := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != `{"version":"1.2.3"}` {
		t.Fatalf("unexpected body %q", w.Body.String())
	}
	if w.Header().Get("X-Version") != "1.2.3" {
		t.Fatalf("expected X-Version header, got %q", w.Header().Get("X-Version"))
	}
}

func TestHeaderOnOtherRoutes(t *testing.T) {
	app := express.New()
	app.Use(version.New(version.Options{Version: "9.9"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) { res.Send("home") })
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "home" {
		t.Fatalf("expected fall-through body, got %q", w.Body.String())
	}
	if w.Header().Get("X-Version") != "9.9" {
		t.Fatalf("expected X-Version on all responses, got %q", w.Header().Get("X-Version"))
	}
}

func TestCustomPathAndHeader(t *testing.T) {
	app := express.New()
	app.Use(version.New(version.Options{Version: "2.0", Path: "/v", Header: "X-App-Version"}))
	r := httptest.NewRequest("GET", "/v", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Header().Get("X-App-Version") != "2.0" {
		t.Fatalf("expected custom header, got %q", w.Header().Get("X-App-Version"))
	}
}
