package jsonp

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestJSONPWrapsBody(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]int{"a": 1})
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/?callback=cb", nil))

	if got := rr.Body.String(); got != `cb({"a":1});` {
		t.Fatalf("body = %q", got)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/javascript; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
}

func TestJSONPNoCallbackPassthrough(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]int{"a": 1})
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if got := rr.Body.String(); got != `{"a":1}` {
		t.Fatalf("body = %q", got)
	}
}

func TestJSONPInvalidCallbackPassthrough(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON(map[string]int{"a": 1})
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/?callback=alert(1)", nil))

	if got := rr.Body.String(); got != `{"a":1}` {
		t.Fatalf("body = %q, want raw json (invalid callback rejected)", got)
	}
}

func TestJSONPCustomParam(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Param: "cb"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.JSON([]int{1, 2})
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/?cb=fn", nil))
	if got := rr.Body.String(); got != `fn([1,2]);` {
		t.Fatalf("body = %q", got)
	}
}
