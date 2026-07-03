package hidepoweredby

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

// TestRemove verifies the header is stripped. The framework only re-adds a
// default "Express" value when the x-powered-by setting is enabled, so it is
// disabled here to observe the deletion performed by the middleware.
func TestRemove(t *testing.T) {
	app := express.New()
	app.Disable("x-powered-by")
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Powered-By", "Express")
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Powered-By"); got != "" {
		t.Fatalf("X-Powered-By = %q, want empty", got)
	}
}

// TestDecoy verifies a non-empty decoy value survives, overriding the
// framework's default.
func TestDecoy(t *testing.T) {
	app := express.New()
	app.Use(New(Options{SetTo: "PHP/8.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Powered-By"); got != "PHP/8.0" {
		t.Fatalf("X-Powered-By = %q, want PHP/8.0", got)
	}
}
