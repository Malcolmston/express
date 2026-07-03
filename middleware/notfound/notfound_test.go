package notfound

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(h)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	return rec
}

func TestDefaultText(t *testing.T) {
	rec := do(t, New())
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d", rec.Code)
	}
	if rec.Body.String() != "Not Found" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestCustomMessage(t *testing.T) {
	rec := do(t, New(Options{Message: "gone away"}))
	if rec.Body.String() != "gone away" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestJSON(t *testing.T) {
	rec := do(t, New(Options{Message: "missing", JSON: true}))
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}
	if !strings.Contains(rec.Body.String(), `"error":"missing"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
