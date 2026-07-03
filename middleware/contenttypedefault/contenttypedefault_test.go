package contenttypedefault

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestDefaultApplied(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.End()
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Content-Type"); got != DefaultType {
		t.Fatalf("Content-Type = %q, want %q", got, DefaultType)
	}
}

func TestDefaultNotOverriding(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Type: "application/pdf"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Type("json").End()
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want json", got)
	}
}
