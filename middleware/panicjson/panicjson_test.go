package panicjson

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestRecoversPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	app := express.New()
	app.Use(New(Options{Logger: logger}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		panic("kaboom")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Internal Server Error") {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if !strings.Contains(buf.String(), "kaboom") {
		t.Fatalf("log = %q, want it to mention the panic", buf.String())
	}
}

func TestNoPanicPassesThrough(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.String() != "ok" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
