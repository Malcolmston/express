// Package frameguard provides middleware that sets the X-Frame-Options response
// header to control whether the page may be rendered inside a frame, helping to
// mitigate clickjacking attacks.
package frameguard

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the frameguard middleware. The zero value is usable and
// yields X-Frame-Options: SAMEORIGIN.
type Options struct {
	// Action is the X-Frame-Options directive to send. Supported values are
	// "SAMEORIGIN" (default) and "DENY" (case-insensitive).
	Action string
}

// New returns middleware that sets the X-Frame-Options header.
func New(opts ...Options) express.Handler {
	action := "SAMEORIGIN"
	if len(opts) > 0 && opts[0].Action != "" {
		if strings.EqualFold(opts[0].Action, "DENY") {
			action = "DENY"
		} else {
			action = "SAMEORIGIN"
		}
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Frame-Options", action)
		next()
	}
}
