package favicon

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
		res.Send("app")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestNoContentWhenEmpty(t *testing.T) {
	rec := do(t, New(), "/favicon.ico")
	if rec.Code != http.StatusNoContent {
		t.Fatalf("code = %d, want 204", rec.Code)
	}
}

func TestServesData(t *testing.T) {
	rec := do(t, New(Options{Data: []byte{0x00, 0x01, 0x02}, MaxAge: 60}), "/favicon.ico")
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/x-icon" {
		t.Fatalf("content-type = %q", ct)
	}
	if rec.Body.Len() != 3 {
		t.Fatalf("body len = %d", rec.Body.Len())
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=60" {
		t.Fatalf("cache-control = %q", cc)
	}
}

func TestPassThrough(t *testing.T) {
	rec := do(t, New(), "/index.html")
	if rec.Body.String() != "app" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}
