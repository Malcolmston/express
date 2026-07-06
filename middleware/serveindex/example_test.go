package serveindex_test

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/serveindex"
)

// Example demonstrates serveindex rendering an HTML directory listing for a
// GET request whose path resolves to a directory beneath the configured root.
// A temporary directory is populated with a file and a sub-directory, and
// serveindex.New is mounted with that directory as its Root. A request for "/"
// resolves to the root directory, so the middleware writes an escaped "Index of"
// page linking each entry, with directories listed first and given a trailing
// slash. Requests that resolve to a file or a missing path would instead fall
// through to the next handler untouched.
func Example() {
	dir, _ := os.MkdirTemp("", "serveindex-example")
	defer os.RemoveAll(dir)
	_ = os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644)
	_ = os.Mkdir(filepath.Join(dir, "sub"), 0o755)

	app := express.New()
	app.Use(serveindex.New(serveindex.Options{Root: dir}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	body := w.Body.String()
	fmt.Println(strings.HasPrefix(w.Header().Get("Content-Type"), "text/html"))
	fmt.Println(strings.Contains(body, `<a href="/sub/">sub/</a>`))
	fmt.Println(strings.Contains(body, `<a href="/a.txt">a.txt</a>`))
	// Output:
	// true
	// true
	// true
}
