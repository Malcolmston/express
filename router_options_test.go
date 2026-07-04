package express

import "testing"

func TestCaseInsensitiveByDefault(t *testing.T) {
	app := New()
	app.Get("/foo", func(req *Request, res *Response, next Next) { res.Send("ok") })
	if c := do(app, "GET", "/FOO", "").Code; c != 200 {
		t.Fatalf("case-insensitive default: /FOO got %d, want 200", c)
	}
}

func TestCaseSensitiveRouter(t *testing.T) {
	r := NewRouter(RouterOptions{CaseSensitive: true})
	r.Get("/foo", func(req *Request, res *Response, next Next) { res.Send("ok") })
	app := New()
	app.Use(r)
	if c := do(app, "GET", "/FOO", "").Code; c != 404 {
		t.Fatalf("case-sensitive: /FOO got %d, want 404", c)
	}
	if c := do(app, "GET", "/foo", "").Code; c != 200 {
		t.Fatalf("case-sensitive: /foo got %d, want 200", c)
	}
}

func TestStrictRouting(t *testing.T) {
	r := NewRouter(RouterOptions{Strict: true})
	r.Get("/foo", func(req *Request, res *Response, next Next) { res.Send("ok") })
	app := New()
	app.Use(r)
	if c := do(app, "GET", "/foo/", "").Code; c != 404 {
		t.Fatalf("strict: /foo/ got %d, want 404", c)
	}
	// Non-strict app default tolerates the trailing slash.
	app2 := New()
	app2.Get("/foo", func(req *Request, res *Response, next Next) { res.Send("ok") })
	if c := do(app2, "GET", "/foo/", "").Code; c != 200 {
		t.Fatalf("non-strict: /foo/ got %d, want 200", c)
	}
}

func TestMountAtParamPath(t *testing.T) {
	// A sub-router mounted at /users/:userId should receive the residual path
	// ("/profile") with the dynamic segment stripped.
	child := NewRouter(RouterOptions{MergeParams: true})
	child.Get("/profile", func(req *Request, res *Response, next Next) {
		res.Send("profile of " + req.Params("userId"))
	})
	app := New()
	app.Use("/users/:userId", child)

	w := do(app, "GET", "/users/42/profile", "")
	if w.Code != 200 || w.Body.String() != "profile of 42" {
		t.Fatalf("mergeParams mount: code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestMergeParamsDisabledScopesParams(t *testing.T) {
	// Without MergeParams the child does not see the parent's :userId.
	child := NewRouter() // MergeParams false
	child.Get("/profile", func(req *Request, res *Response, next Next) {
		res.Send("id=[" + req.Params("userId") + "]")
	})
	app := New()
	app.Use("/users/:userId", child)

	w := do(app, "GET", "/users/42/profile", "")
	if w.Body.String() != "id=[]" {
		t.Fatalf("expected scoped (empty) userId, got %q", w.Body.String())
	}
}

func TestNestedRouters(t *testing.T) {
	inner := NewRouter()
	inner.Get("/deep", func(req *Request, res *Response, next Next) { res.Send("deep") })
	mid := NewRouter()
	mid.Use("/inner", inner)
	app := New()
	app.Use("/api", mid)

	if b := do(app, "GET", "/api/inner/deep", "").Body.String(); b != "deep" {
		t.Fatalf("nested routers: %q", b)
	}
}
