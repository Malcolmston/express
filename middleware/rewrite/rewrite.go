// Package rewrite provides URL-rewriting middleware. It transparently changes
// the request path before it reaches the routing layer, supporting regular
// expression matching with $1-style capture-group substitution. It is the
// express framework's Go analogue of Node URL rewriters such as
// express-urlrewrite and connect-modrewrite, and of the internal-rewrite subset
// of Apache's mod_rewrite: an express.Handler that quietly maps an incoming URL
// onto a different internal path without the client ever seeing a redirect.
//
// Reach for this middleware to decouple your public URL surface from your route
// definitions — serving legacy or vanity paths from a new handler, collapsing a
// family of URLs onto one route, versioning an API prefix, or normalizing paths
// — all without issuing 3xx redirects or teaching every handler about the old
// shape. Because the rewrite is internal, the browser's address bar and the
// client's view of the URL are unchanged; only the server's routing sees the new
// path. Use a redirect (res.Redirect) instead when you actually want the client
// to learn the canonical URL.
//
// Operationally the middleware belongs before the routes it should affect,
// typically among the first Use calls. On each request it reads req.Raw.URL.Path
// and walks its compiled rules in order. The first rule whose expression matches
// the path wins: the middleware computes the replacement with the rule's To
// template and calls req.SetPath, which updates the router's match path so the
// request genuinely re-routes rather than merely changing what handlers observe.
// It then stops scanning and calls next(); if no rule matches, the path is left
// untouched and next() is still called. The middleware never writes a response or
// short-circuits — it only rewrites and continues.
//
// Rules are configured through Options.Rules, an ordered slice of Rule. Each
// Rule supplies either a precompiled From (*regexp.Regexp) or a Pattern string
// compiled once at construction, plus a To replacement that may reference capture
// groups with the $1, $2, ... syntax accepted by regexp.ReplaceAllString. When
// both From and Pattern are set, From takes precedence. Ordering is significant
// because only the first match applies, so list more specific rules before
// broader ones. Any rule whose Pattern fails to compile — and any rule with
// neither From nor a non-empty Pattern — is silently skipped at construction
// time rather than causing a panic, so a single bad pattern never disables the
// rest.
//
// Compared with the Node originals this port is deliberately focused. It rewrites
// the URL path only: it does not match or rewrite the query string, method, or
// host, offers no [L]/[R]/[F] style flags, no proxy or external-redirect modes,
// and no conditional RewriteCond predicates. Matching is Go's RE2-based regexp,
// so backreferences and other PCRE-only features are unavailable, and replacement
// uses $1 rather than the \1 form some tools accept. For anything beyond
// first-match, path-only substitution, compose it with your own handler logic.
package rewrite

import (
	"regexp"

	"github.com/malcolmston/express"
)

// Rule describes a single rewrite. Exactly one of From or Pattern should be
// provided; if both are set From takes precedence.
type Rule struct {
	// From is a precompiled regular expression matched against the request
	// path.
	From *regexp.Regexp

	// Pattern is a regular-expression string used when From is nil. It is
	// compiled once when the middleware is constructed.
	Pattern string

	// To is the replacement path. It may reference capture groups using the
	// $1, $2, ... syntax accepted by regexp.ReplaceAllString.
	To string
}

// Options configures the rewrite middleware.
type Options struct {
	// Rules is the ordered list of rewrite rules. The first matching rule is
	// applied and rewriting stops.
	Rules []Rule
}

// New returns middleware that rewrites req.Raw.URL.Path according to the first
// matching rule, then calls next. Rules whose expressions fail to compile are
// skipped.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	type compiled struct {
		re *regexp.Regexp
		to string
	}
	rules := make([]compiled, 0, len(o.Rules))
	for _, r := range o.Rules {
		re := r.From
		if re == nil && r.Pattern != "" {
			c, err := regexp.Compile(r.Pattern)
			if err != nil {
				continue
			}
			re = c
		}
		if re == nil {
			continue
		}
		rules = append(rules, compiled{re: re, to: r.To})
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		path := req.Raw.URL.Path
		for _, r := range rules {
			if r.re.MatchString(path) {
				// SetPath updates the router's match path so the rewrite
				// actually re-routes, not just changes what handlers observe.
				req.SetPath(r.re.ReplaceAllString(path, r.to))
				break
			}
		}
		next()
	}
}
