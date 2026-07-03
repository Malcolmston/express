// Package scopecheck provides middleware that authorizes requests only when the
// caller holds all of the required scopes.
package scopecheck

import "github.com/malcolmston/express"

// Options configures the scope-check middleware.
type Options struct {
	// Required lists the scopes that must all be present. Required.
	Required []string
	// Getter extracts the scopes associated with a request. Required.
	Getter func(req *express.Request) []string
}

// New returns middleware that responds with 403 unless every required scope is
// present on the request.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		var have []string
		if opts.Getter != nil {
			have = opts.Getter(req)
		}
		set := make(map[string]struct{}, len(have))
		for _, s := range have {
			set[s] = struct{}{}
		}
		for _, want := range opts.Required {
			if _, ok := set[want]; !ok {
				res.Status(403).Send("Forbidden")
				return
			}
		}
		next()
	}
}
