package spa_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/spa"
)

// ExampleNew demonstrates serving a single-page application build with a
// deep-link fallback. We stage a throwaway build directory containing an
// index.html and a real asset, then mount spa.New pointed at that Root followed
// by a catch-all 404 handler. A request for the existing app.js is streamed
// straight from disk, while a request for the client-side route
// /dashboard/settings resolves to no file and is rewritten to index.html so the
// browser router can render it. Finally a request for a missing, extensioned
// asset is intentionally not rewritten and falls through to the 404 handler.
// The three cases are driven in memory with net/http/httptest.
func ExampleNew() {
	root, _ := os.MkdirTemp("", "spa-demo")
	defer os.RemoveAll(root)
	os.WriteFile(filepath.Join(root, "index.html"), []byte("APP-SHELL"), 0o644)
	os.WriteFile(filepath.Join(root, "app.js"), []byte("console.log(1)"), 0o644)

	app := express.New()
	app.Use(spa.New(spa.Options{Root: root}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(404).Send("not found")
	})

	get := func(path string) (int, string) {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w.Code, w.Body.String()
	}

	c1, b1 := get("/app.js")
	c2, b2 := get("/dashboard/settings")
	c3, b3 := get("/missing.css")
	fmt.Printf("asset:    %d %s\n", c1, b1)
	fmt.Printf("deeplink: %d %s\n", c2, b2)
	fmt.Printf("missing:  %d %s\n", c3, b3)

	// Output:
	// asset:    200 console.log(1)
	// deeplink: 200 APP-SHELL
	// missing:  404 not found
}
