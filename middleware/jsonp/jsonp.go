// Package jsonp provides Express middleware that wraps JSON responses in a
// JavaScript callback invocation when the request carries a callback query
// parameter, enabling classic cross-origin JSONP requests. It reproduces the
// JSONP behavior built into Express's own res.jsonp() (which reads the callback
// name from the "callback" query parameter by default) as a standalone,
// stdlib-only handler that transparently upgrades any JSON body produced
// downstream.
//
// JSONP ("JSON with padding") predates CORS: a page loads data from another
// origin by including it as a <script src="...?callback=fn">, and the server
// replies with executable JavaScript — fn({...}) — that calls a function the
// page defined. Use this middleware when you must support legacy browsers or
// legacy clients that still rely on that <script>-tag technique and cannot use
// CORS or fetch. For modern applications CORS is the correct choice; JSONP is
// inherently a script-injection channel and should be treated with care.
//
// The handler is designed to sit in front of the JSON-producing handlers,
// typically mounted with app.Use. On each request it reads the callback name
// from the configured query parameter (req.Query, default "callback"). If that
// parameter is absent or fails validation, it calls next() immediately and does
// nothing else, so the response passes through byte-for-byte unchanged. When a
// valid callback is present it temporarily swaps res.Writer for a buffering
// captureWriter, calls next() to let downstream handlers render their JSON into
// the buffer, then restores the original writer and flushes the transformed
// result. It thus both reads request query state and rewrites the response body
// and headers.
//
// The transformation turns a buffered body of `<json>` into `callback(<json>);`,
// sets Content-Type to "application/javascript; charset=utf-8", adds
// X-Content-Type-Options: nosniff, deletes any stale Content-Length, and writes
// the status code the downstream handler chose (defaulting to 200 if none was
// set). Callback names are validated against the pattern
// ^[A-Za-z_$][\w$.]*$, which permits identifiers and dotted member access such
// as "window.cb" while rejecting anything — parentheses, spaces, operators —
// that could break out of the function-call context; an invalid name is treated
// as no callback and the raw JSON is returned untouched. This validation plus
// the nosniff header are the package's core defenses against reflected-XSS style
// abuse of the callback parameter, though JSONP endpoints should still only
// expose data safe to hand to arbitrary third-party pages.
//
// Parity with the Express original is close but not identical. Like Express it
// defaults to the "callback" query parameter, wraps the body in
// callback(...), and serves it as JavaScript. It differs in that it operates as
// a body-rewriting wrapper over any JSON writer (rather than a dedicated
// res.jsonp method), it does not emit Express's typeof-guard/anti-CSRF prelude,
// it hardens the callback with strict name validation plus a nosniff header,
// and the callback query key is configurable only through Options.Param rather
// than an app "jsonp callback name" setting.
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
