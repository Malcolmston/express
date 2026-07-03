package realip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func resolveWith(t *testing.T, header, value string, opts ...Options) string {
	t.Helper()
	app := express.New()
	app.Use(New(opts...))
	var got string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = ClientIP(req)
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		req.Header.Set(header, value)
	}
	app.ServeHTTP(rec, req)
	return got
}

func TestXForwardedForLeftmost(t *testing.T) {
	if got := resolveWith(t, "X-Forwarded-For", "1.2.3.4, 5.6.7.8"); got != "1.2.3.4" {
		t.Fatalf("got %q, want 1.2.3.4", got)
	}
}

func TestTrustedProxySkipped(t *testing.T) {
	got := resolveWith(t, "X-Forwarded-For", "9.9.9.9, 5.6.7.8",
		Options{TrustedProxies: []string{"5.6.7.8"}})
	if got != "9.9.9.9" {
		t.Fatalf("got %q, want 9.9.9.9", got)
	}
}

func TestXRealIP(t *testing.T) {
	if got := resolveWith(t, "X-Real-IP", "8.8.8.8"); got != "8.8.8.8" {
		t.Fatalf("got %q, want 8.8.8.8", got)
	}
}

func TestRewritesRemoteAddr(t *testing.T) {
	app := express.New()
	app.Use(New())
	var remote string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		remote = req.Raw.RemoteAddr
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	app.ServeHTTP(rec, req)
	if remote != "1.1.1.1" {
		t.Fatalf("RemoteAddr = %q, want 1.1.1.1", remote)
	}
}
