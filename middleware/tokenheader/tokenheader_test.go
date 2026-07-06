package tokenheader_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/tokenheader"
)

// ExampleNew demonstrates guarding a route with a custom API-key header. It
// configures the middleware to read X-API-Key and accept only the value
// "s3cr3t", mounts it on an express.Application ahead of a "GET /" handler, and
// then drives two requests through httptest: one carrying the correct key and
// one carrying a wrong key. The valid request passes verification, next() runs,
// and the handler replies 200 with its body; the invalid request is
// short-circuited with a bare 401 Unauthorized and next() is never called. The
// same 401 would result from a missing header, an empty token, or a nil Verify,
// since all failures collapse to one indistinguishable response. Because both
// outcomes are fully determined by the inputs, the Output block is
// deterministic.
func ExampleNew() {
	app := express.New()
	app.Use(tokenheader.New(tokenheader.Options{
		Header: "X-API-Key",
		Verify: func(token string) bool { return token == "s3cr3t" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	call := func(key string) {
		r := httptest.NewRequest("GET", "/", nil)
		if key != "" {
			r.Header.Set("X-API-Key", key)
		}
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		fmt.Printf("key=%-8q -> %d %s\n", key, w.Code, w.Body.String())
	}
	call("s3cr3t")
	call("nope")
	// Output:
	// key="s3cr3t" -> 200 welcome
	// key="nope"   -> 401 Unauthorized
}

func newApp() *express.Application {
	app := express.New()
	app.Use(tokenheader.New(tokenheader.Options{
		Header: "X-Session-Token",
		Verify: func(token string) bool { return token == "s3cr3t" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, token string) int {
	r := httptest.NewRequest("GET", "/", nil)
	if token != "" {
		r.Header.Set("X-Session-Token", token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestValidToken(t *testing.T) {
	if code := run(newApp(), "s3cr3t"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestInvalidToken(t *testing.T) {
	if code := run(newApp(), "wrong"); code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestMissingToken(t *testing.T) {
	if code := run(newApp(), ""); code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
}
