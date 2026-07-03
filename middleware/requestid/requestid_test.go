package requestid

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestRequestIDGenerated(t *testing.T) {
	app := express.New()
	app.Use(New())
	var stored any
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		stored, _ = req.Value(ContextKey)
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	id := rr.Header().Get(DefaultHeader)
	if len(id) != 32 {
		t.Fatalf("generated id = %q (len %d), want 32 hex chars", id, len(id))
	}
	if stored != id {
		t.Fatalf("stored id %v != header id %v", stored, id)
	}
}

func TestRequestIDReused(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(DefaultHeader, "abc123")
	app.ServeHTTP(rr, r)
	if got := rr.Header().Get(DefaultHeader); got != "abc123" {
		t.Fatalf("id = %q, want reused abc123", got)
	}
}

func TestRequestIDCustomHeader(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Header: "X-Trace"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Header().Get("X-Trace") == "" {
		t.Fatalf("expected X-Trace header set")
	}
}
