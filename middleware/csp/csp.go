// Package csp provides middleware that builds and sets a Content-Security-Policy
// response header from a directive map.
package csp

import (
	"sort"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the csp middleware. The zero value is usable and yields
// the default policy "default-src 'self'".
type Options struct {
	// Directives maps a CSP directive name (e.g. "default-src", "script-src")
	// to its list of source expressions. When empty, a default of
	// {"default-src": {"'self'"}} is used. Directive names are emitted in
	// sorted order for a deterministic header value.
	Directives map[string][]string

	// ReportOnly, when true, sets Content-Security-Policy-Report-Only instead
	// of Content-Security-Policy.
	ReportOnly bool
}

// Build renders a directive map into a Content-Security-Policy header value.
// Directives are sorted by name and each directive's sources are joined with
// spaces; directives are separated by "; ".
func Build(directives map[string][]string) string {
	names := make([]string, 0, len(directives))
	for name := range directives {
		names = append(names, name)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names))
	for _, name := range names {
		sources := directives[name]
		if len(sources) == 0 {
			parts = append(parts, name)
			continue
		}
		parts = append(parts, name+" "+strings.Join(sources, " "))
	}
	return strings.Join(parts, "; ")
}

// New returns middleware that sets a Content-Security-Policy header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	directives := o.Directives
	if len(directives) == 0 {
		directives = map[string][]string{"default-src": {"'self'"}}
	}
	value := Build(directives)

	header := "Content-Security-Policy"
	if o.ReportOnly {
		header = "Content-Security-Policy-Report-Only"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set(header, value)
		next()
	}
}
