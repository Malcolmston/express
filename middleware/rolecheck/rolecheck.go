// Package rolecheck provides middleware that authorizes requests only when the
// caller holds at least one of the required roles.
package rolecheck

import "github.com/malcolmston/express"

// Options configures the role-check middleware.
type Options struct {
	// Roles lists the roles that satisfy the check; a request is authorized
	// when it holds any one of them. Required.
	Roles []string
	// Getter extracts the roles associated with a request. Required.
	Getter func(req *express.Request) []string
}

// New returns middleware that responds with 403 unless the request holds at
// least one of the required roles.
func New(opts Options) express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		if opts.Getter == nil {
			res.Status(403).Send("Forbidden")
			return
		}
		have := opts.Getter(req)
		for _, want := range opts.Roles {
			for _, h := range have {
				if h == want {
					next()
					return
				}
			}
		}
		res.Status(403).Send("Forbidden")
	}
}
