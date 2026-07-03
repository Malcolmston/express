package permittedcrossdomain

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
	return rec.Header().Get("X-Permitted-Cross-Domain-Policies")
}

func TestDefault(t *testing.T) {
	if got := do(t, New()); got != "none" {
		t.Fatalf("got %q", got)
	}
}

func TestCustom(t *testing.T) {
	if got := do(t, New(Options{Policy: "master-only"})); got != "master-only" {
		t.Fatalf("got %q", got)
	}
}
