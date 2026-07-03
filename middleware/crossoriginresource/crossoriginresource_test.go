package crossoriginresource

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Header().Get("Cross-Origin-Resource-Policy")
}

func TestDefault(t *testing.T) {
	if got := do(t, New()); got != "same-origin" {
		t.Fatalf("got %q", got)
	}
}

func TestCustom(t *testing.T) {
	if got := do(t, New(Options{Policy: "cross-origin"})); got != "cross-origin" {
		t.Fatalf("got %q", got)
	}
}
