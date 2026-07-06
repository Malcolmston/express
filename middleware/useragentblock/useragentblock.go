// Package useragentblock provides middleware that rejects requests whose
// User-Agent header contains any configured blocked substring. It is a small,
// dependency-free equivalent of the Node "express-useragent-blocker" family of
// middleware and the hand-rolled "block by user-agent" guards commonly written
// on top of connect: a deny-list applied to the client-supplied User-Agent
// header. There is no upstream npm package this maps to one-for-one; it
// captures the widespread pattern of denying obvious scraper, bot and CLI
// tooling agents by name.
//
// Use it to shed unwanted traffic cheaply — blocking greedy crawlers, curl or
// wget probes, headless scrapers, or a specific misbehaving bot — before that
// traffic reaches expensive downstream handlers. Because the User-Agent header
// is trivially spoofed, treat this as traffic hygiene and coarse abuse
// mitigation rather than a real security boundary; a determined client can
// always send a different header. For richer, non-blocking classification of
// the same header see the sibling useragent package.
//
// Mount it early in the chain with app.Use so it runs before your routes. On
// each request it lower-cases the "User-Agent" header (read via req.Get) and
// tests it against the pre-lowered Block substrings. On the first match it
// short-circuits the chain by writing res.Status(403).Send("Forbidden") and
// returning without calling next, so no downstream handler runs. When nothing
// matches it calls next() and the request proceeds normally. Matching is
// case-insensitive substring containment, not a full parse, so "curl" also
// matches "libcurl" and "CURL/7.0".
//
// The single required option is Options.Block, the list of substrings to deny.
// Empty strings in the list are ignored so an accidental "" entry cannot block
// every request. If Block is nil or empty the middleware becomes a pass-through
// that blocks nothing. Matching stops at the first hit, and the 403 body is
// always the plain text "Forbidden" with no configurable message or status.
//
// Parity notes: this port is intentionally minimal. It does not support glob
// or regular-expression patterns, allow-lists, per-route exemptions, custom
// status codes or response bodies, or logging of blocked requests. It only
// inspects the User-Agent header — not IP, referer, or other signals — and
// performs plain substring matching. If you need pattern matching or custom
// responses, compose your own handler around req.Get("User-Agent").
package useragentblock

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the user-agent blocking middleware.
type Options struct {
	// Block lists case-insensitive substrings; a request whose User-Agent
	// contains any of them is rejected. Required.
	Block []string
}

// New returns middleware that responds with 403 when the request's User-Agent
// contains one of the configured blocked substrings.
func New(opts Options) express.Handler {
	blocked := make([]string, len(opts.Block))
	for i, b := range opts.Block {
		blocked[i] = strings.ToLower(b)
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		ua := strings.ToLower(req.Get("User-Agent"))
		for _, b := range blocked {
			if b != "" && strings.Contains(ua, b) {
				res.Status(403).Send("Forbidden")
				return
			}
		}
		next()
	}
}
