// Package version provides middleware that exposes an application's version.
// It sets an X-Version header on every response and, when a request targets the
// configured path, replies with a small JSON document reporting the version.
package version

import "github.com/malcolmston/express"

// Options configures the version middleware.
type Options struct {
	// Version is the version string reported. Defaults to "unknown".
	Version string
	// Path is the endpoint that returns the version as JSON. Defaults to
	// "/version".
	Path string
	// Header is the response header name. Defaults to "X-Version".
	Header string
}

type payload struct {
	Version string `json:"version"`
}

// New returns version middleware configured by opts.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Version == "" {
		o.Version = "unknown"
	}
	if o.Path == "" {
		o.Path = "/version"
	}
	if o.Header == "" {
		o.Header = "X-Version"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set(o.Header, o.Version)
		if o.Path != "" && req.Path() == o.Path {
			res.Status(200).JSON(payload{Version: o.Version})
			return
		}
		next()
	}
}
