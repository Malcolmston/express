// Package rewrite provides URL-rewriting middleware. It transparently changes
// the request path before it reaches the routing layer, supporting regular
// expression matching with $1-style capture-group substitution.
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
