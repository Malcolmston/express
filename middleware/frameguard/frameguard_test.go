package frameguard

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
	return rec.Header().Get("X-Frame-Options")
}

func TestDefault(t *testing.T) {
	if got := do(t, New()); got != "SAMEORIGIN" {
		t.Fatalf("got %q", got)
	}
}

func TestDeny(t *testing.T) {
	if got := do(t, New(Options{Action: "deny"})); got != "DENY" {
		t.Fatalf("got %q", got)
	}
}
