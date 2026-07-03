// Package methodoverride provides middleware that overrides the HTTP request
// method using either a request header or a query-string parameter. This lets
// HTML forms (which can only issue GET and POST) emulate verbs such as PUT,
// PATCH, and DELETE.
package methodoverride

import (
	"strings"

	"github.com/malcolmston/express"
)

// DefaultHeader is the header consulted for the overriding method.
const DefaultHeader = "X-HTTP-Method-Override"

// DefaultQuery is the query-string parameter consulted for the overriding
// method.
const DefaultQuery = "_method"

// Options configures the method-override middleware.
type Options struct {
	// Header is the request header inspected for an override. When empty,
	// DefaultHeader is used.
	Header string

	// Query is the query-string parameter inspected for an override. When
	// empty, DefaultQuery is used.
	Query string
}

// New returns middleware that rewrites req.Raw.Method from the configured
// header or query parameter. The header takes precedence over the query
// parameter. The override is only applied to POST requests, matching the
// common convention that overrides ride on a form submission.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Header == "" {
		o.Header = DefaultHeader
	}
	if o.Query == "" {
		o.Query = DefaultQuery
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		// Overrides only make sense on POST, the verb HTML forms can send.
		if strings.EqualFold(req.Method(), "POST") {
			method := req.Get(o.Header)
			if method == "" {
				method = req.Query(o.Query)
			}
			if method != "" {
				req.Raw.Method = strings.ToUpper(method)
			}
		}
		next()
	}
}
