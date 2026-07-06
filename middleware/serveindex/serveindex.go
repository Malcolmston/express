// Package serveindex provides middleware that renders an HTML directory listing
// for requests that resolve to a directory beneath a configured root. It is the
// express framework's Go analogue of the Node connect/express serve-index
// middleware, which produces the familiar "Index of /path" pages, reduced here
// to a compact, dependency-free HTML renderer built on the standard library's
// os.ReadDir and html.EscapeString.
//
// Reach for this middleware to expose a browsable view of a directory tree —
// a downloads area, a build-artifact folder, static documentation, or a quick
// file share — where users benefit from clickable links to sub-directories and
// files. It is designed to sit next to a static file server: serveindex renders
// the folder pages while the file server delivers the actual file contents, so
// the two together reproduce the classic auto-index behaviour of a plain web
// server.
//
// Operationally the middleware is typically mounted before a static handler. It
// acts only on GET and HEAD requests; any other method calls next() immediately.
// For an eligible request it takes req.Path(), cleans it, and joins it onto the
// cleaned Options.Root, then confirms the result is still contained within Root
// before touching the filesystem. It calls os.Stat on the resolved path and
// proceeds only when the target exists and is a directory; on success it reads
// the entries, sets a "text/html; charset=utf-8" Content-Type, and writes the
// rendered listing via res.Send. In every other case — wrong method, path
// escaping the root, stat error, a non-directory target, or a read error — it
// calls next() and lets the request fall through to the next handler.
//
// The rendered page lists directories before files, each group sorted by name,
// and links directories with a trailing slash. Every path that is not the root
// gets a "../" parent link, and all displayed names and hrefs are passed through
// html.EscapeString so entry names cannot inject markup into the page. Path
// traversal is contained defensively: the joined path is checked against Root
// with an os.PathSeparator-aware prefix test, so requests such as /../etc are
// rejected and fall through rather than escaping the configured root. Because a
// listing is written with res.Send, serveindex short-circuits the chain only on
// the success path; the fall-through cases leave the response entirely to
// downstream handlers.
//
// Compared with the Node serve-index original, this port is deliberately
// minimal. There is a single Option, Root; there is no support for the
// alternative JSON or plain-text views, no pluggable HTML/stylesheet templates,
// no icons, file sizes, or modification-time columns, no hidden-file filtering
// or custom sort options, and no ETag/caching negotiation. It renders one plain,
// escaped HTML listing and otherwise gets out of the way, leaving file delivery,
// access control, and richer presentation to other middleware.
package serveindex

import (
	"html"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the serveindex middleware.
type Options struct {
	// Root is the base directory whose contents are listed.
	Root string
}

// New returns middleware that serves directory listings from Options.Root.
func New(opts Options) express.Handler {
	root := filepath.Clean(opts.Root)

	return func(req *express.Request, res *express.Response, next express.Next) {
		if m := req.Method(); m != http.MethodGet && m != http.MethodHead {
			next()
			return
		}
		urlPath := req.Path()
		// Confine the resolved path to root to prevent traversal.
		rel := filepath.Clean("/" + urlPath)
		full := filepath.Join(root, rel)
		if full != root && !strings.HasPrefix(full, root+string(os.PathSeparator)) {
			next()
			return
		}
		info, err := os.Stat(full)
		if err != nil || !info.IsDir() {
			next()
			return
		}
		entries, err := os.ReadDir(full)
		if err != nil {
			next()
			return
		}
		res.Set("Content-Type", "text/html; charset=utf-8").Send(render(urlPath, entries))
	}
}

func render(urlPath string, entries []os.DirEntry) string {
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	if !strings.HasSuffix(urlPath, "/") {
		urlPath += "/"
	}

	var b strings.Builder
	title := html.EscapeString(urlPath)
	b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\"><title>Index of ")
	b.WriteString(title)
	b.WriteString("</title></head><body>\n<h1>Index of ")
	b.WriteString(title)
	b.WriteString("</h1>\n<ul>\n")
	if urlPath != "/" {
		b.WriteString("<li><a href=\"../\">../</a></li>\n")
	}
	for _, e := range entries {
		name := e.Name()
		link := name
		display := name
		if e.IsDir() {
			link += "/"
			display += "/"
		}
		href := path.Join(urlPath, link)
		if e.IsDir() && !strings.HasSuffix(href, "/") {
			href += "/"
		}
		b.WriteString("<li><a href=\"")
		b.WriteString(html.EscapeString(href))
		b.WriteString("\">")
		b.WriteString(html.EscapeString(display))
		b.WriteString("</a></li>\n")
	}
	b.WriteString("</ul>\n</body></html>\n")
	return b.String()
}
