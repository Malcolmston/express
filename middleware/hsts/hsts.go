// Package hsts provides middleware that sets the HTTP
// Strict-Transport-Security (HSTS) response header, instructing browsers to
// only access the site over HTTPS.
package hsts

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// DefaultMaxAge is the default max-age in seconds (180 days).
const DefaultMaxAge = 15552000

// Options configures the HSTS middleware. The zero value is usable and uses
// DefaultMaxAge with neither includeSubDomains nor preload.
type Options struct {
	// MaxAge is the number of seconds browsers should remember to prefer
	// HTTPS. When zero, DefaultMaxAge is used. Use a negative value to send
	// max-age=0.
	MaxAge int

	// IncludeSubDomains adds the includeSubDomains directive.
	IncludeSubDomains bool

	// Preload adds the preload directive for inclusion in browser preload
	// lists.
	Preload bool
}

// New returns middleware that sets the Strict-Transport-Security header.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	maxAge := o.MaxAge
	switch {
	case maxAge == 0:
		maxAge = DefaultMaxAge
	case maxAge < 0:
		maxAge = 0
	}

	parts := []string{"max-age=" + strconv.Itoa(maxAge)}
	if o.IncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	if o.Preload {
		parts = append(parts, "preload")
	}
	value := strings.Join(parts, "; ")

	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Strict-Transport-Security", value)
		next()
	}
}
