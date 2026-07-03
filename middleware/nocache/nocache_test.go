package nocache

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestNoCacheHeaders(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if got := rr.Header().Get("Cache-Control"); got != "no-store, no-cache, must-revalidate" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if got := rr.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma = %q", got)
	}
	if got := rr.Header().Get("Expires"); got != "0" {
		t.Fatalf("Expires = %q", got)
	}
}
