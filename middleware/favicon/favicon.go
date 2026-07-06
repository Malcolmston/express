// Package favicon provides express middleware that answers requests for
// /favicon.ico, short-circuiting them before they reach application routes or
// logging. It ports the Node "serve-favicon" middleware in spirit: browsers
// request /favicon.ico automatically and often, and this handler deals with
// those requests in one place instead of letting them fall through to routers,
// loggers, and not-found handlers. Unlike serve-favicon it can also operate
// without any icon at all, replying 204 No Content.
//
// Use this middleware to keep favicon noise out of application logic and access
// logs, to cheaply serve a small embedded icon, or to explicitly answer favicon
// probes with 204 when you have no icon to serve. Because it short-circuits,
// mount it early — typically before your logging middleware and route
// handlers — so favicon requests are handled and terminated before any of that
// runs. The request path it claims is exported as the Path constant,
// "/favicon.ico".
//
// The handler inspects every request but acts on only one path. If req.Path()
// is not Path it immediately calls next() and does nothing else, so all other
// requests pass straight through untouched. For the favicon path it enforces
// the HTTP method: any method other than GET or HEAD is rejected with an Allow:
// GET, HEAD header and a 405 Method Not Allowed, terminating the request. For a
// valid GET/HEAD it optionally sets Cache-Control, then either replies 204 No
// Content when no icon bytes are configured, or sets the Content-Type and sends
// the icon bytes. In every favicon-path branch the middleware writes the
// response itself and does not call next, so downstream handlers never see the
// request.
//
// Behavior is configured through Options. Data holds the raw icon bytes; when
// empty the middleware answers 204 rather than sending a body. ContentType is
// the type sent with the icon and defaults to "image/x-icon" when left blank.
// MaxAge is a Cache-Control max-age in seconds and, when positive, adds
// Cache-Control: public, max-age=<n>; a zero or negative MaxAge omits the
// Cache-Control header entirely. All options are resolved once at New time. Note
// the icon is served straight from memory with no ETag, Last-Modified, or
// conditional-request handling, so pair it with an etag or caching middleware if
// you want revalidation.
//
// Regarding parity with the Node original: like serve-favicon this middleware
// intercepts /favicon.ico, restricts it to GET/HEAD, returns 405 with an Allow
// header for other methods, and supports a Cache-Control max-age. The
// differences are that it accepts the icon as in-memory bytes (or none) rather
// than a filesystem path, that with no bytes it degrades to a 204 response
// instead of requiring an icon, and that it omits serve-favicon's ETag and
// Last-Modified validators. The Content-Type default of image/x-icon matches
// serve-favicon's default.
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
