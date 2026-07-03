// Package notfound provides a terminal handler that responds with 404 Not
// Found. Register it last, after all routes and other middleware, to produce a
// consistent "not found" response for unmatched requests.
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
