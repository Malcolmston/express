package rolecheck_test

import (
	"fmt"
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

// Example demonstrates gating a route behind a role requirement. An upstream
// handler stashes the caller's roles on the request (here a fixed "editor",
// standing in for a value pulled from a session or decoded token), and
// rolecheck.New is configured to admit anyone holding "admin" or "editor".
// Because the check is a logical OR, the single "editor" role is enough to
// satisfy the guard, so the request reaches the protected handler and returns
// 200. A caller holding neither role would instead be short-circuited with a
// 403 "Forbidden" response before the handler ever runs.
func Example() {
	app := express.New()

	// Upstream: attach the caller's roles (from a session, JWT, etc.).
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("roles", []string{"editor"})
		next()
	})

	// Admit callers holding either "admin" or "editor".
	app.Use(rolecheck.New(rolecheck.Options{
		Roles: []string{"admin", "editor"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("roles")
			roles, _ := v.([]string)
			return roles
		},
	}))

	app.Get("/admin", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	r := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// welcome
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
