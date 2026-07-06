// Package rolecheck provides middleware that authorizes a request only when the
// caller holds at least one of a configured set of roles. It is the express
// framework's Go analogue of the small role-guard helpers that Node developers
// hand-roll on top of Passport or connect-roles, and of libraries such as
// express-authorization or the role checks baked into access-control kits:
// authentication has already established who the caller is, and this middleware
// answers the coarser question of whether that caller is allowed in at all.
//
// Reach for this middleware to gate a route or a whole router behind a role,
// after some earlier handler has attached the caller's roles to the request.
// Typical uses are an admin console that only "admin" may reach, an editorial
// area open to "editor" or "admin", or a moderation endpoint reserved for
// "moderator". Because the roles are plain strings compared for equality, the
// middleware is agnostic about where they come from: a session, a decoded JWT,
// a database lookup, or a header can all feed the same guard through the
// Getter callback.
//
// Operationally the middleware belongs after authentication and after whatever
// step populates the caller's roles, but before the protected handler. On each
// request it invokes Options.Getter to obtain the roles the caller currently
// holds, then scans Options.Roles looking for any overlap. The first role in
// Options.Roles that also appears in the caller's set satisfies the check; the
// middleware calls next() exactly once and the request proceeds untouched. The
// guard never writes to the response on the success path and stores nothing new
// on the request — it is a pure gate.
//
// The check is a logical OR: holding any single required role is enough, which
// is what distinguishes it from the sibling scopecheck package, whose check is
// a logical AND over every required scope. Matching is case-sensitive and exact,
// so "Admin" does not satisfy a requirement of "admin". When the caller holds
// none of the required roles the request is short-circuited with 403 Forbidden
// and a plain "Forbidden" body, and next() is never called. The same 403 is
// returned when Options.Getter is nil, which is treated as a misconfiguration
// rather than a panic; an empty or nil Options.Roles can never be satisfied and
// therefore rejects every request.
//
// Compared with the sprawling Node middlewares it stands in for, this port is
// deliberately minimal. It performs no authentication of its own, resolves no
// role hierarchies or wildcards, supports no per-action or resource-scoped
// permissions, and offers no customisation of the status code or body — every
// denial collapses to a bare 403. All policy beyond "hold one of these roles"
// belongs in the caller's own handlers, and the meaning and source of the role
// strings is defined entirely by the Getter callback.
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
