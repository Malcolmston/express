// Package methodoverride provides middleware that overrides the HTTP request
// method using either a request header or a query-string parameter. This lets
// HTML forms (which can only issue GET and POST) emulate verbs such as PUT,
// PATCH, and DELETE. It is the Go port of the Node method-override middleware
// (the method-override npm package used with connect/express), reproducing its
// core header- and query-based override behavior with only the standard library.
//
// Use this middleware in front of a route table that is built around REST verbs
// but must be reachable from plain HTML forms or clients that can only send POST.
// By letting a form carry the intended verb in a hidden _method field or in the
// X-HTTP-Method-Override header, you can keep RESTful routes (app.Put,
// app.Delete, app.Patch) while the browser still submits an ordinary POST. It is
// equally useful for constrained HTTP clients or proxies that strip non-standard
// verbs.
//
// Register it early — before the router matches — so the rewritten method is
// used for route dispatch. On each request the handler first checks whether the
// incoming method is POST; only then does it look for an override. It reads the
// configured header via req.Get, and if that is empty falls back to the
// configured query parameter via req.Query. If a value is found it upper-cases
// it and writes it back to req.Raw.Method, mutating the underlying
// *http.Request in place so every downstream handler and the router observe the
// new verb. In all cases the handler calls next() and never writes a response
// itself; it only rewrites request state.
//
// Two options control the source names: Header (default DefaultHeader,
// "X-HTTP-Method-Override") and Query (default DefaultQuery, "_method"). The
// header takes precedence over the query parameter when both are present. The
// override is deliberately applied only to POST requests — matching the common
// convention that overrides ride on a form submission — so a GET carrying
// _method=delete is left untouched and keeps its original verb. Values are
// upper-cased but otherwise unvalidated, so a bogus override such as _method=foo
// will set the method to "FOO" and typically fall through to no matching route;
// callers that accept untrusted input should constrain the allowed verbs
// upstream.
//
// The chief security consideration is that this middleware lets a client change
// the effective HTTP method, which can bypass method-based assumptions in later
// handlers or upstream access rules — only enable it where form-driven verb
// emulation is actually wanted, and be mindful that CSRF protections must
// account for the overridden verb. Compared with the Node original, this port
// keeps the header-then-query precedence and the POST-only rule while adopting
// Go idioms: exported DefaultHeader/DefaultQuery constants, a struct of Options
// instead of a callback-style getter, and a direct write to req.Raw.Method
// rather than stashing the original method on the request object.
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
