package spa

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/malcolmston/express"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("INDEX"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("JS"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func do(t *testing.T, root, path string) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(New(Options{Root: root}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(404).Send("missing-asset")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestServesRealFile(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/app.js")
	if rec.Body.String() != "JS" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestFallbackForRoute(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/dashboard/settings")
	if rec.Body.String() != "INDEX" {
		t.Fatalf("expected index fallback, got %q", rec.Body.String())
	}
}

func TestMissingAssetFallsThrough(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/missing.css")
	if rec.Code != 404 || rec.Body.String() != "missing-asset" {
		t.Fatalf("code=%d body=%q", rec.Code, rec.Body.String())
	}
}
