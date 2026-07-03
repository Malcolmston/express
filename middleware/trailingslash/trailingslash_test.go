package trailingslash

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
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestEnforceAdds(t *testing.T) {
	rec := do(t, New(Options{Enforce: true}), "/about")
	if rec.Code != 301 || rec.Header().Get("Location") != "/about/" {
		t.Fatalf("code=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestEnforceKeepsSlashed(t *testing.T) {
	rec := do(t, New(Options{Enforce: true}), "/about/")
	if rec.Body.String() != "ok" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestStripRemoves(t *testing.T) {
	rec := do(t, New(Options{Strip: true}), "/about/")
	if rec.Code != 301 || rec.Header().Get("Location") != "/about" {
		t.Fatalf("code=%d loc=%q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestRootSkipped(t *testing.T) {
	rec := do(t, New(Options{Strip: true}), "/")
	if rec.Body.String() != "ok" {
		t.Fatalf("root should pass through, body=%q", rec.Body.String())
	}
}

func TestPreservesQuery(t *testing.T) {
	rec := do(t, New(Options{Enforce: true}), "/p?a=1")
	if rec.Header().Get("Location") != "/p/?a=1" {
		t.Fatalf("loc=%q", rec.Header().Get("Location"))
	}
}
