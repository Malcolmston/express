// Package cachecontrol provides express middleware that sets a Cache-Control
// header assembled from a set of caching options.
package cachecontrol

import (
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the Cache-Control directive.
type Options struct {
	// MaxAge is the max-age directive in seconds. It is emitted when greater
	// than zero (or when SetMaxAge is true, allowing an explicit max-age=0).
	MaxAge int
	// SetMaxAge forces emission of max-age even when MaxAge is 0.
	SetMaxAge bool
	// Public adds the "public" directive.
	Public bool
	// Private adds the "private" directive.
	Private bool
	// NoStore adds the "no-store" directive.
	NoStore bool
}

// New returns middleware that sets the Cache-Control response header built from
// the provided options. Directives are emitted in a stable order.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	value := build(o)
	return func(req *express.Request, res *express.Response, next express.Next) {
		if value != "" {
			res.Set("Cache-Control", value)
		}
		next()
	}
}

func build(o Options) string {
	var parts []string
	if o.Public {
		parts = append(parts, "public")
	}
	if o.Private {
		parts = append(parts, "private")
	}
	if o.NoStore {
		parts = append(parts, "no-store")
	}
	if o.MaxAge > 0 || o.SetMaxAge {
		parts = append(parts, "max-age="+strconv.Itoa(o.MaxAge))
	}
	return strings.Join(parts, ", ")
}
