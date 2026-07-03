// Package featureflag provides middleware that attaches a set of boolean
// feature flags to each request, letting handlers gate behavior on named
// features.
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
