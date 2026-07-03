package redirectmap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("fell-through")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestRedirect(t *testing.T) {
	rec := do(t, New(Options{Map: map[string]string{"/old": "/new"}}), "/old")
	if rec.Code != http.StatusFound {
		t.Fatalf("code = %d, want 302", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/new" {
		t.Fatalf("location = %q", loc)
	}
}

func TestCustomStatus(t *testing.T) {
	rec := do(t, New(Options{Map: map[string]string{"/a": "/b"}, Status: 301}), "/a")
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("code = %d, want 301", rec.Code)
	}
}

func TestFallThrough(t *testing.T) {
	rec := do(t, New(Options{Map: map[string]string{"/a": "/b"}}), "/other")
	if rec.Body.String() != "fell-through" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
