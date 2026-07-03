package rolecheck_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/rolecheck"
)

func newApp(roles []string) *express.Application {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("roles", roles)
		next()
	})
	app.Use(rolecheck.New(rolecheck.Options{
		Roles: []string{"admin", "editor"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("roles")
			r, _ := v.([]string)
			return r
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

func TestHasRole(t *testing.T) {
	if code := run(newApp([]string{"editor"})); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestMissingRole(t *testing.T) {
	if code := run(newApp([]string{"viewer"})); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestNoRoles(t *testing.T) {
	if code := run(newApp(nil)); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}
