package origincheck_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/origincheck"
)

func newApp(optional bool) *express.Application {
	app := express.New()
	app.Use(origincheck.New(origincheck.Options{
		Allow:    []string{"example.com", "app.example.com:8443"},
		Optional: optional,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, origin string) int {
	r := httptest.NewRequest("GET", "/", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestAllowedOrigin(t *testing.T) {
	if code := run(newApp(false), "https://example.com"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestAllowedOriginWithPort(t *testing.T) {
	if code := run(newApp(false), "https://app.example.com:8443"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestDisallowedOrigin(t *testing.T) {
	if code := run(newApp(false), "https://evil.com"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestMissingOriginRejected(t *testing.T) {
	if code := run(newApp(false), ""); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestMissingOriginOptional(t *testing.T) {
	if code := run(newApp(true), ""); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}
