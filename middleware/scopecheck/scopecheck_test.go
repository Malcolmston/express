package scopecheck_test

import (
	"fmt"
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

// Example demonstrates guarding an endpoint that demands multiple scopes. An
// upstream handler stashes the scopes granted to the caller's access token
// (here "read", "write", and "delete", standing in for a decoded OAuth "scope"
// claim), and scopecheck.New requires both "read" and "write". Because the
// check is a logical AND, the request passes only when every required scope is
// present; here all are, so the protected handler runs and returns 200. Had the
// token carried only "read", the missing "write" scope would short-circuit the
// request with a 403 "Forbidden" response.
func Example() {
	app := express.New()

	// Upstream: attach the scopes granted by the caller's token.
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("scopes", []string{"read", "write", "delete"})
		next()
	})

	// Require both "read" and "write" (logical AND).
	app.Use(scopecheck.New(scopecheck.Options{
		Required: []string{"read", "write"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("scopes")
			scopes, _ := v.([]string)
			return scopes
		},
	}))

	app.Get("/documents", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("granted")
	})

	r := httptest.NewRequest("GET", "/documents", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
	// Output:
	// 200
	// granted
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
