package express

import (
	"regexp"
	"testing"
)

// These tests live in the core package so they can exercise SetPath's effect on
// route matching directly (the router matches on the internal path).

func TestSetPathReroutes(t *testing.T) {
	app := New()
	// A middleware that rewrites /old/* to /new/* via SetPath.
	re := regexp.MustCompile(`^/old/(.*)$`)
	app.Use(func(req *Request, res *Response, next Next) {
		if re.MatchString(req.Path()) {
			req.SetPath(re.ReplaceAllString(req.Path(), "/new/$1"))
		}
		next()
	})
	app.Get("/new/thing", func(req *Request, res *Response, next Next) {
		res.Send("matched new")
	})

	// Requesting the OLD path must now match the NEW route.
	w := do(app, "GET", "/old/thing", "")
	if w.Code != 200 || w.Body.String() != "matched new" {
		t.Fatalf("rewrite did not re-route: code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestSetPathBasePathReroutes(t *testing.T) {
	app := New()
	// Strip a "/app" prefix, then a root-relative route must match.
	app.Use(func(req *Request, res *Response, next Next) {
		if p := req.Path(); len(p) >= 4 && p[:4] == "/app" {
			req.SetPath(p[4:])
		}
		next()
	})
	app.Get("/dashboard", func(req *Request, res *Response, next Next) {
		res.Send("dashboard")
	})

	w := do(app, "GET", "/app/dashboard", "")
	if w.Code != 200 || w.Body.String() != "dashboard" {
		t.Fatalf("basepath did not re-route: code=%d body=%q", w.Code, w.Body.String())
	}
}
