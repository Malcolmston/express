// Package referrerpolicy provides middleware that sets the Referrer-Policy
// response header, controlling how much referrer information is included with
// requests.
package referrerpolicy

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the referrerpolicy middleware. The zero value is usable
// and yields Referrer-Policy: no-referrer.
type Options struct {
	// Policy is one or more Referrer-Policy tokens. When empty, "no-referrer"
	// is used. Multiple tokens are joined with ", ".
	Policy []string
}

// New returns middleware that sets the Referrer-Policy header.
func New(opts ...Options) express.Handler {
	value := "no-referrer"
	if len(opts) > 0 && len(opts[0].Policy) > 0 {
		value = strings.Join(opts[0].Policy, ", ")
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Referrer-Policy", value)
		next()
	}
}
