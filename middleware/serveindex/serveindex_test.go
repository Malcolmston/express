package serveindex

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func setup(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func do(t *testing.T, root, path string) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	app.Use(New(Options{Root: root}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("fallthrough")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestListsRoot(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/")
	body := rec.Body.String()
	if !strings.Contains(body, "a.txt") || !strings.Contains(body, "sub/") {
		t.Fatalf("listing missing entries: %s", body)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestListsSubdir(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/sub")
	if !strings.Contains(rec.Body.String(), "../") {
		t.Fatalf("subdir listing missing parent link: %s", rec.Body.String())
	}
}

func TestFileFallsThrough(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/a.txt")
	if rec.Body.String() != "fallthrough" {
		t.Fatalf("expected fall-through for file, got %q", rec.Body.String())
	}
}

func TestMissingFallsThrough(t *testing.T) {
	dir := setup(t)
	rec := do(t, dir, "/nope")
	if rec.Body.String() != "fallthrough" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
