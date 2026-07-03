// Package serveindex provides middleware that renders an HTML directory
// listing for requests that resolve to a directory beneath a configured root.
// Requests that do not map to a directory fall through to the next handler,
// making it easy to combine with a static file server.
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
