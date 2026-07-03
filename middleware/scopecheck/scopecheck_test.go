package scopecheck_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/scopecheck"
)

func newApp(scopes []string) *express.Application {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("scopes", scopes)
		next()
	})
	app.Use(scopecheck.New(scopecheck.Options{
		Required: []string{"read", "write"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("scopes")
			s, _ := v.([]string)
			return s
		},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application) int {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestAllScopes(t *testing.T) {
	if code := run(newApp([]string{"read", "write", "delete"})); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestPartialScopes(t *testing.T) {
	if code := run(newApp([]string{"read"})); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestNoScopes(t *testing.T) {
	if code := run(newApp(nil)); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}
