package methodallow_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/methodallow"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(methodallow.New(methodallow.Options{Methods: []string{"GET", "POST"}}))
	app.All("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func TestAllowedMethod(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDisallowedMethod(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("DELETE", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 405 {
		t.Fatalf("expected 405, got %d", w.Code)
	}
	if got := w.Header().Get("Allow"); got != "GET, POST" {
		t.Fatalf("unexpected Allow header: %q", got)
	}
}
