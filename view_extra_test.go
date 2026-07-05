package express

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Multiple "views" directories are searched in order.
func TestRenderMultipleViewDirs(t *testing.T) {
	dir1, dir2 := t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(dir2, "only2.html"), []byte(`from-two`), 0o644)

	app := New()
	app.Set("views", []string{dir1, dir2})
	app.Get("/p", func(req *Request, res *Response, next Next) { res.Render("only2") })
	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "from-two") {
		t.Fatalf("multi-dir render = %q", b)
	}
}

// A view naming a directory resolves to its index file.
func TestRenderIndexFile(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "users"), 0o755)
	os.WriteFile(filepath.Join(dir, "users", "index.html"), []byte(`user-index`), 0o644)

	app := New()
	app.Set("views", dir)
	app.Get("/p", func(req *Request, res *Response, next Next) { res.Render("users") })
	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "user-index") {
		t.Fatalf("index render = %q", b)
	}
}

// res.Render merges app.locals + res.Locals under a map argument (argument wins).
func TestRenderLocalsMerge(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "t.html"),
		[]byte(`{{.Site}}|{{.User}}|{{.Page}}`), 0o644)

	app := New()
	app.Set("views", dir)
	app.Locals()["Site"] = "acme"
	app.Locals()["User"] = "app-default"
	app.Get("/p", func(req *Request, res *Response, next Next) {
		res.Locals["User"] = "ada" // request-scoped overrides app-level
		res.Render("t", map[string]any{"Page": "home"})
	})
	if b := do(app, "GET", "/p", "").Body.String(); b != "acme|ada|home" {
		t.Fatalf("locals merge = %q, want %q", b, "acme|ada|home")
	}
}

// The "view cache" setting compiles a template once; edits are ignored while on.
func TestViewCacheCompilesOnce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "c.html")
	os.WriteFile(path, []byte(`v1`), 0o644)

	app := New()
	app.Set("views", dir).Set("view cache", true)
	app.Get("/p", func(req *Request, res *Response, next Next) { res.Render("c") })

	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "v1") {
		t.Fatalf("first render = %q", b)
	}
	os.WriteFile(path, []byte(`v2`), 0o644)
	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "v1") {
		t.Fatalf("cached render should still be v1, got %q", b)
	}

	// With the cache off, the edit is picked up.
	app.Set("view cache", false)
	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "v2") {
		t.Fatalf("uncached render should be v2, got %q", b)
	}
}

// app.Render is the app-level render returning the string.
func TestAppRender(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.html"), []byte(`hi {{.N}}`), 0o644)
	app := New()
	app.Set("views", dir)
	out, err := app.Render("a", map[string]any{"N": "bob"})
	if err != nil || out != "hi bob" {
		t.Fatalf("app.Render = %q, %v", out, err)
	}
}

// A missing view yields Express's "failed to lookup view" error.
func TestLookupViewError(t *testing.T) {
	app := New()
	app.Set("views", t.TempDir())
	_, err := app.Render("nope", nil)
	if err == nil || !strings.Contains(err.Error(), "failed to lookup view") {
		t.Fatalf("expected lookup error, got %v", err)
	}
}

// An absolute view path is honoured directly.
func TestRenderAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	abs := filepath.Join(dir, "abs.html")
	os.WriteFile(abs, []byte(`absolute`), 0o644)
	app := New()
	out, err := app.Render(abs, nil)
	if err != nil || out != "absolute" {
		t.Fatalf("absolute render = %q, %v", out, err)
	}
}

// A non-map data argument is passed through untouched (struct rendering).
func TestRenderStructData(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "s.html"), []byte(`{{.Title}}`), 0o644)
	app := New()
	app.Set("views", dir)
	app.Get("/p", func(req *Request, res *Response, next Next) {
		res.Render("s", struct{ Title string }{"Struct"})
	})
	if b := do(app, "GET", "/p", "").Body.String(); !strings.Contains(b, "Struct") {
		t.Fatalf("struct render = %q", b)
	}
}
