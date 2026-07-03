// Package favicon provides middleware that answers requests for /favicon.ico,
// short-circuiting them before they reach application routes or logging. It can
// serve icon bytes supplied in Options or reply with 204 No Content.
package favicon

import (
	"net/http"
	"strconv"

	"github.com/malcolmston/express"
)

// Path is the request path handled by this middleware.
const Path = "/favicon.ico"

// Options configures the favicon middleware.
type Options struct {
	// Data holds the raw bytes of the icon. When empty the middleware replies
	// with 204 No Content instead of a body.
	Data []byte

	// ContentType is the Content-Type sent with the icon. When empty it
	// defaults to "image/x-icon".
	ContentType string

	// MaxAge is the Cache-Control max-age in seconds. When positive a
	// Cache-Control header is added.
	MaxAge int
}

// New returns middleware that handles GET/HEAD requests for /favicon.ico and
// passes every other request through to next.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.ContentType == "" {
		o.ContentType = "image/x-icon"
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		if req.Path() != Path {
			next()
			return
		}
		if m := req.Method(); m != http.MethodGet && m != http.MethodHead {
			res.Set("Allow", "GET, HEAD").SendStatus(http.StatusMethodNotAllowed)
			return
		}
		if o.MaxAge > 0 {
			res.Set("Cache-Control", "public, max-age="+strconv.Itoa(o.MaxAge))
		}
		if len(o.Data) == 0 {
			res.SendStatus(http.StatusNoContent)
			return
		}
		res.Set("Content-Type", o.ContentType).Send(o.Data)
	}
}
