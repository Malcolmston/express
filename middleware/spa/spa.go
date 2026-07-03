// Package spa provides middleware for serving single-page applications. It
// serves static files from a root directory and, for navigation requests that
// do not resolve to a real file, falls back to the SPA's index document so that
// client-side routing works on deep links and page reloads.
package spa

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/malcolmston/express"
)

// DefaultIndex is the fallback document served for client-side routes.
const DefaultIndex = "index.html"

// Options configures the spa middleware.
type Options struct {
	// Root is the directory containing the built SPA assets.
	Root string

	// Index is the fallback file served for extension-less paths that do not
	// map to an existing file. When empty it defaults to "index.html".
	Index string
}

// New returns middleware that serves files from Root, falling back to Index.
func New(opts Options) express.Handler {
	root := filepath.Clean(opts.Root)
	index := opts.Index
	if index == "" {
		index = DefaultIndex
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if m := req.Method(); m != http.MethodGet && m != http.MethodHead {
			next()
			return
		}
		rel := filepath.Clean("/" + req.Path())
		full := filepath.Join(root, rel)
		if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
			next()
			return
		}

		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			http.ServeFile(res.Writer, req.Raw, full)
			return
		}

		// No matching file. For extension-less "navigation" requests, serve the
		// SPA index so client-side routing can take over.
		if filepath.Ext(rel) == "" {
			idx := filepath.Join(root, index)
			if info, err := os.Stat(idx); err == nil && !info.IsDir() {
				http.ServeFile(res.Writer, req.Raw, idx)
				return
			}
		}
		next()
	}
}
