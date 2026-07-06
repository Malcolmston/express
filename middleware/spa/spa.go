// Package spa provides middleware for serving single-page applications. It
// serves static files from a root directory and, for navigation requests that
// do not resolve to a real file, falls back to the SPA's index document so that
// client-side routing works on deep links and page reloads. It is the express
// framework's Go analogue of the connect-history-api-fallback middleware
// combined with a static file server (express.static / serve-static): the
// classic "serve the build output, and rewrite unknown routes to index.html"
// pattern used by React, Vue, Angular, and Svelte production bundles.
//
// Reach for this middleware to host a compiled front end from the same Go
// process that serves its API. Without it, a user who reloads the page at
// /dashboard/settings or shares a deep link would hit the server with a path
// that has no matching file and receive a 404, because the routing for that
// path lives in JavaScript, not on disk. This middleware bridges that gap by
// returning index.html for such navigation requests so the client router can
// take over and render the right view.
//
// Operationally the middleware runs near the front of the chain, typically after
// the API routes so real endpoints win and the SPA acts as the catch-all. It
// handles only GET and HEAD requests, calling next() for any other method. For a
// handled request it cleans the request path, joins it under the cleaned Root,
// and guards against path traversal by requiring the result to stay within Root
// (an escaping path falls through via next()). If that path names an existing,
// non-directory file it is streamed with http.ServeFile against res.Writer and
// req.Raw, which supplies Content-Type, Last-Modified, ETag, and Range support
// for free — and next() is not called.
//
// When no real file matches, the fallback rule is keyed on the file extension:
// only an extension-less path (treated as a client-side "navigation" route) is
// eligible, and the middleware serves Options.Index — DefaultIndex, "index.html",
// when Index is empty — from Root if that index file exists. A request for a
// missing asset that does have an extension, such as /styles/missing.css, is
// deliberately not rewritten to index.html; instead the middleware calls next()
// so a downstream handler can return a genuine 404. This keeps a broken
// stylesheet or script from silently receiving HTML.
//
// Compared with connect-history-api-fallback this port keeps the core deep-link
// fallback but is intentionally lean. The "navigation" test is the presence of a
// file extension rather than the configurable Accept-header and rewrite-rule
// machinery of the Node package, there is no per-route rewrite table or verbose
// logging option, and directory index resolution is limited to the single
// configured Index file. Static delivery, MIME typing, and caching headers are
// delegated wholesale to net/http's http.ServeFile rather than reimplemented.
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
