// Package notfound provides a terminal handler that responds with 404 Not
// Found for requests that fall through the router unmatched. It plays the role
// of Express's built-in final 404 handler and of common Node snippets like the
// "catch-all" app.use((req, res) => res.status(404)...) pattern and the
// express-error-handler style notFound helper, giving unmatched routes one
// consistent response instead of the framework default.
//
// Use it as the last entry you register, after every route and every other
// piece of middleware, so it only runs when nothing earlier produced a response.
// This is the idiomatic way to control the body, content type, and shape of your
// "not found" answer application-wide: a plain-text message for humans, or a JSON
// error object for API clients that expect machine-readable errors.
//
// The handler is intentionally terminal. When invoked it first checks
// res.Written() and returns immediately if a response has already been committed
// by an upstream handler, which prevents duplicate writes and superfluous status
// changes. Otherwise it sets the status to 404 via res.Status(http.StatusNotFound)
// and writes the body, and it deliberately does NOT call next() — the chain stops
// here. That is why chain position matters: register it too early and it would
// swallow requests that a later route should have handled.
//
// Behavior is controlled by an optional Options value. Message sets the response
// body and defaults to "Not Found" when empty. JSON, when true, sends
// {"error": message} through res.JSON (which sets an application/json content
// type); when false the message is sent as text/plain via res.Type("text").Send.
// Passing no Options at all yields the plain-text default. The written-guard is
// the key edge case to remember: if an earlier handler already flushed headers or
// body, this middleware becomes a no-op rather than corrupting that response.
//
// Parity with the Express/Node original is close in spirit and cleaner in
// practice. Express's default 404 emits an HTML "Cannot GET /path" page; this
// port defaults to a terse "Not Found" and, unlike the bare framework default,
// offers first-class JSON output and a guard against double-writing. It does not
// attempt to render stack traces or environment-dependent HTML, keeping the 404
// response predictable for both browsers and API consumers.
package notfound

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the notfound handler.
type Options struct {
	// Message is the body sent with the 404 response. When empty a default
	// message is used.
	Message string

	// JSON, when true, sends the message as a JSON object {"error": message}
	// instead of plain text.
	JSON bool
}

// New returns a terminal handler that always responds 404. It does not call
// next.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Message == "" {
		o.Message = "Not Found"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if res.Written() {
			return
		}
		res.Status(http.StatusNotFound)
		if o.JSON {
			res.JSON(map[string]string{"error": o.Message})
			return
		}
		res.Type("text").Send(o.Message)
	}
}
