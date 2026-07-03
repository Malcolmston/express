package hostcheck_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hostcheck"
)

func newApp(status int) *express.Application {
	app := express.New()
	app.Use(hostcheck.New(hostcheck.Options{
		Allow:  []string{"example.com", "www.example.com"},
		Status: status,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, host string) int {
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = host
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestAllowedHost(t *testing.T) {
	if code := run(newApp(0), "example.com"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestAllowedHostWithPort(t *testing.T) {
	if code := run(newApp(0), "www.example.com:8080"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestDisallowedHostDefault421(t *testing.T) {
	if code := run(newApp(0), "evil.com"); code != 421 {
		t.Fatalf("expected 421, got %d", code)
	}
}

func TestDisallowedHostCustom400(t *testing.T) {
	if code := run(newApp(400), "evil.com"); code != 400 {
		t.Fatalf("expected 400, got %d", code)
	}
}
