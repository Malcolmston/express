// Package csp provides middleware that builds and sets a Content-Security-Policy
// response header from a directive map. It is the Go analogue of Helmet's
// contentSecurityPolicy middleware from the Node ecosystem, packaged as a
// drop-in express.Handler. Where Helmet accepts a directives object and renders
// it into the header string, this port accepts a Go map of directive names to
// source lists and does the same rendering, then attaches the value to every
// response that flows through it.
//
// Use this middleware to instruct browsers about which origins are legitimate
// sources of scripts, styles, images, fonts, frames, and other resources, which
// is the single most effective defense against cross-site scripting and
// resource-injection attacks. A policy of "default-src 'self'" alone already
// blocks inline scripts and third-party code; tighten it further by adding
// directives such as "script-src", "style-src", or "img-src". Mount it globally
// with app.Use so the header is present on every response, or attach it to a
// specific router or path prefix when only part of the application needs a
// policy.
//
// Operationally the middleware belongs near the front of the chain, before any
// handler that writes the response body, because a Content-Security-Policy
// header is only honored if it is sent with the response. On each request it
// writes a single header — Content-Security-Policy by default, or
// Content-Security-Policy-Report-Only when Options.ReportOnly is set — and then
// always calls next() so the request proceeds untouched. The header value is
// computed once when New builds the handler, not per request, so the cost at
// request time is a single res.Set. No request headers or request state are
// read; the middleware is purely additive to the response.
//
// The policy string is produced by Build, which sorts directive names
// alphabetically for a deterministic, stable header value, joins each
// directive's sources with spaces, and separates directives with "; ". A
// directive whose source list is empty or nil is emitted as a bare keyword
// (for example "upgrade-insecure-requests") with no trailing value. When
// Options.Directives is empty the middleware falls back to the safe default
// {"default-src": {"'self'"}}, so the zero-value Options is usable and yields
// "default-src 'self'". Report-only mode sends the policy for monitoring without
// enforcing it, letting you collect violation reports before switching to
// enforcement.
//
// Compared with Helmet's contentSecurityPolicy, this port keeps the same
// header-building semantics — directive map in, policy string out, enforce or
// report-only — but is deliberately minimal. It does not ship Helmet's baked-in
// default directive set beyond the single "default-src 'self'" fallback, does
// not generate per-request nonces (see the sibling cspnonce package for that),
// and does not offer per-directive functions for dynamic values; every source
// is a static string supplied at construction time. Callers who need nonces or
// hashes should compose this package with their own logic or use cspnonce.
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
