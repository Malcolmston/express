// Package xssfilter provides middleware that sets the legacy X-XSS-Protection
// response header. Modern guidance (and this package's default) is to disable
// the buggy browser auditor by sending "0".
package xssfilter

import "github.com/malcolmston/express"

// Options configures the xssfilter middleware. The zero value is usable and
// yields X-XSS-Protection: 0.
type Options struct {
	// Value overrides the header value. When empty, "0" is used.
	Value string
}

// New returns middleware that sets the X-XSS-Protection header.
func New(opts ...Options) express.Handler {
	value := "0"
	if len(opts) > 0 && opts[0].Value != "" {
		value = opts[0].Value
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-XSS-Protection", value)
		next()
	}
}
