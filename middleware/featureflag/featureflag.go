// Package featureflag provides middleware that attaches a fixed set of boolean
// feature flags to every request, letting downstream handlers gate behavior on
// named features. It is a small, dependency-free analogue of the request-scoped
// flag lookup exposed by Node feature-flag libraries such as
// "express-feature-flags" or the flag helpers bundled with LaunchDarkly's and
// Unleash's Express SDKs: rather than talking to a remote flag service, it
// carries a caller-supplied map of names to enabled states and makes it
// available for the lifetime of the request.
//
// Use this middleware when you want to toggle code paths (a new UI, an
// experimental endpoint, a gradual rollout) without redeploying or littering
// handlers with global configuration reads. Because the flag set is captured
// once when New is called, it is best for process-wide flags resolved at
// startup; if you need per-user or dynamically refreshed flags, build the map
// per request in your own middleware and store it under the same request state,
// or wrap New so it recomputes the map on each call.
//
// Mechanically the middleware sits anywhere in the chain before the handlers
// that read flags. On each request it calls req.Set with the internal context
// key to stash the flag map on the request, then immediately calls next() to
// continue the chain. It never writes a response, never sets a header, and
// never short-circuits, so it is safe to mount globally with app.Use. Handlers
// (or nested middleware) retrieve a flag by calling the package function
// Enabled, which reads the map back out of the request state.
//
// The stored map is shared by reference across all requests: New copies the
// Options.Flags pointer, so mutating the original map after New returns will be
// observed by in-flight requests and is not safe for concurrent writes. Treat
// the map as read-only once passed in. Enabled is defensively written to
// degrade gracefully — if the middleware never ran, if the stored value has an
// unexpected type, or if the flag name is unknown, it returns false rather than
// panicking, so an unregistered flag and a flag explicitly set to false are
// indistinguishable to callers.
//
// Compared with the Node originals this port is intentionally minimal: there is
// no percentage rollout, no user targeting, no remote evaluation, and no
// streaming updates — just a static name-to-bool table resolved locally. That
// keeps it stdlib-only and allocation-light, at the cost of the dynamic
// evaluation features a full flag platform provides. If Options.Flags is nil an
// empty map is substituted, so Enabled always reports false and the middleware
// is a harmless no-op.
package featureflag

import "github.com/malcolmston/express"

const contextKey = "featureflags"

// Options configures the feature flag middleware.
type Options struct {
	// Flags maps feature names to their enabled state.
	Flags map[string]bool
}

// New returns middleware that stores the configured flags on each request.
// Query them with Enabled.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	flags := o.Flags
	if flags == nil {
		flags = map[string]bool{}
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		req.Set(contextKey, flags)
		next()
	}
}

// Enabled reports whether the named feature flag is enabled for the request.
// It returns false if the middleware did not run or the flag is unknown.
func Enabled(req *express.Request, name string) bool {
	if v, ok := req.Value(contextKey); ok {
		if m, ok := v.(map[string]bool); ok {
			return m[name]
		}
	}
	return false
}
