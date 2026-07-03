// Package jsonp provides express middleware that wraps JSON responses in a
// JavaScript callback invocation when the request carries a callback query
// parameter, enabling cross-origin JSONP requests.
package jsonp

import (
	"bytes"
	"net/http"
	"regexp"

	"github.com/malcolmston/express"
)

// DefaultParam is the query-string parameter inspected for the callback name.
const DefaultParam = "callback"

// callbackPattern validates callback names, permitting identifiers with dotted
// member access (e.g. "window.cb") while rejecting anything that could break
// out of the function-call context.
var callbackPattern = regexp.MustCompile(`^[A-Za-z_$][\w$.]*$`)

// Options configures the jsonp middleware.
type Options struct {
	// Param is the query parameter naming the callback. Defaults to "callback".
	Param string
}

// captureWriter buffers the response so a JSON body can be rewritten as a
// JavaScript callback invocation.
type captureWriter struct {
	http.ResponseWriter
	buf         bytes.Buffer
	status      int
	wroteHeader bool
}

func (w *captureWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
}

func (w *captureWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.buf.Write(p)
}

// New returns middleware that, when a valid callback query parameter is present,
// transforms the response body from `<json>` into `callback(<json>);` and sets
// the Content-Type to application/javascript. Requests without a callback (or
// with an invalid callback name) are passed through unchanged.
func New(opts ...Options) express.Handler {
	param := DefaultParam
	if len(opts) > 0 && opts[0].Param != "" {
		param = opts[0].Param
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		callback := req.Query(param)
		if callback == "" || !callbackPattern.MatchString(callback) {
			next()
			return
		}

		orig := res.Writer
		cw := &captureWriter{ResponseWriter: orig}
		res.Writer = cw
		next()
		res.Writer = orig

		status := cw.status
		if !cw.wroteHeader {
			status = http.StatusOK
		}

		var out bytes.Buffer
		out.WriteString(callback)
		out.WriteByte('(')
		out.Write(cw.buf.Bytes())
		out.WriteString(");")

		h := orig.Header()
		h.Set("Content-Type", "application/javascript; charset=utf-8")
		// Prevent content-type sniffing of the script payload.
		h.Set("X-Content-Type-Options", "nosniff")
		h.Del("Content-Length")
		orig.WriteHeader(status)
		_, _ = orig.Write(out.Bytes())
	}
}
