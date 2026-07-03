package errorjson

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestRendersError(t *testing.T) {
	app := express.New()
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		next(errors.New("boom"))
	})
	app.Use(New())

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"boom"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestCustomStatus(t *testing.T) {
	app := express.New()
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		next(errors.New("bad"))
	})
	app.Use(New(Options{Status: http.StatusBadGateway}))

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
