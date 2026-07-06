// Package panicjson provides middleware that recovers from panics raised by
// downstream handlers, logs them, and responds with a generic 500 JSON error so
// that a single failing request cannot crash the process. It is the express
// analogue of Node error-recovery middleware — Express's default error handler
// and packages such as Connect's errorhandler or Koa's try/catch bodies — but
// specialized to always answer with a compact JSON envelope.
//
// Use it as an outer safety net around routes that may panic, whether from
// programmer error, a failed type assertion, a nil dereference, or a
// deliberately raised value. In Go a panic that escapes an HTTP handler would
// otherwise unwind the goroutine serving the request; wrapping the chain in
// this middleware converts that fatal condition into a well-formed 500 response
// and a log line, keeping the server available for other clients.
//
// Mechanically it must run before the handlers it protects, so register it
// early via app.Use. It installs a deferred recover around the call to next: it
// invokes next() to run the rest of the chain, and if a downstream handler
// panics, the deferred function catches the value, logs it through the
// configured logger with the prefix "panicjson: recovered from panic:", and
// then attempts to write the error response. The response is only written when
// res.Written() reports that nothing has been sent yet, which avoids corrupting
// a reply whose headers or body were already partially flushed before the
// panic.
//
// The important semantics concern what is and is not exposed. The recovery
// response is a fixed 500 with Content-Type application/json and the exact body
// {"error":"Internal Server Error"}; the panic value itself is never sent to
// the client, only logged, so internal details are not leaked. Logging is
// controlled by Options.Logger, which defaults to a std logger writing to
// os.Stderr with timestamps when left nil. Note the recover only catches panics
// on the same goroutine as the request; a panic in a goroutine that a handler
// spawns is not recoverable here and will still crash the process.
//
// Parity with the Node original is behavioral. Like Express's built-in error
// handler this middleware turns an otherwise-fatal handler failure into a 500
// response, but it is opinionated where Express is generic: it always emits JSON
// with a fixed message rather than negotiating HTML/text or surfacing the error
// text, and it recovers from Go panics rather than from next(err) values, since
// this port uses panics rather than an error-forwarding convention.
package panicjson

import (
	"log"
	"net/http"
	"os"

	"github.com/malcolmston/express"
)

// Options configures the panic recovery middleware.
type Options struct {
	// Logger receives a message for each recovered panic. When nil a logger
	// writing to os.Stderr is used.
	Logger *log.Logger
}

// New returns middleware that recovers from panics raised by later handlers.
// It logs the panic value and, if the response has not been written, sends a
// 500 response with the body {"error":"Internal Server Error"}.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	logger := o.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("panicjson: recovered from panic: %v", r)
				if !res.Written() {
					res.Status(http.StatusInternalServerError).JSON(map[string]string{
						"error": "Internal Server Error",
					})
				}
			}
		}()
		next()
	}
}
