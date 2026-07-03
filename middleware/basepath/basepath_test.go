package basepath

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler, path string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	app := express.New()
	app.Use(h)
	var seen string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		seen = req.Raw.URL.Path
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec, seen
}

func TestStripsPrefix(t *testing.T) {
	_, seen := do(t, New(Options{Prefix: "/app"}), "/app/users")
	if seen != "/users" {
		t.Fatalf("seen = %q", seen)
	}
}

func TestExactPrefixBecomesRoot(t *testing.T) {
	_, seen := do(t, New(Options{Prefix: "/app"}), "/app")
	if seen != "/" {
		t.Fatalf("seen = %q", seen)
	}
}

func TestNonStrictPassThrough(t *testing.T) {
	_, seen := do(t, New(Options{Prefix: "/app"}), "/other")
	if seen != "/other" {
		t.Fatalf("seen = %q", seen)
	}
}

func TestStrict404(t *testing.T) {
	rec, _ := do(t, New(Options{Prefix: "/app", Strict: true}), "/other")
	if rec.Code != http.StatusNotFound {
		t.Fatalf("code = %d, want 404", rec.Code)
	}
}
