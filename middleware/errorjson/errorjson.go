// Package errorjson provides an express error-handling middleware that renders
// unhandled errors as a JSON object of the form {"error": "<message>"}. It
// plays the role of a terminal JSON error handler, comparable to the default
// error responder in Express plus a small custom middleware such as the ones
// commonly written on top of the Node "http-errors" package: instead of
// Express's HTML error page, every error that propagates to the end of the
// chain is serialized as a compact JSON document suited to API clients.
//
// Use this handler when the application is an API (or any client that expects
// JSON) and you want a single, consistent error envelope regardless of where an
// error originates. Register it last, after all routes and other middleware, so
// that it acts as the catch-all: express dispatches to error handlers only once
// a handler calls next(err), so anything that forwards an error eventually
// lands here.
//
// Unlike the other middleware in this tree, New returns an express.ErrorHandler
// (signature func(err error, req, res, next)), not a plain Handler. It is still
// registered with app.Use, whose variadic any argument accepts error handlers
// as well as ordinary handlers. Its position in the chain is the tail: express
// skips it during normal request flow and invokes it only when an upstream
// handler passes an error to next. On invocation it first checks res.Written():
// if the response has already been committed (headers or body flushed) it
// returns immediately and does not call next, avoiding a duplicate write and a
// superfluous log line. Otherwise it sets the status and writes the JSON body,
// terminating the request; it deliberately does not call next, since it is the
// final responder.
//
// The status code comes from Options.Status, defaulting to 500
// (http.StatusInternalServerError) when unset or non-positive. The response
// body is map[string]string{"error": msg}, where msg is err.Error(); if err
// happens to be nil the message falls back to "internal server error". Writing
// the body goes through res.JSON, so the Content-Type is set to
// application/json and the payload is JSON-encoded (with HTML-safe escaping
// applied by encoding/json). Note the same status is used for every error: this
// handler does not read a status off typed errors, so if you need per-error
// codes you should map them upstream or wrap this handler.
//
// Regarding parity with the Node original: the intent matches the common
// Express pattern of a final JSON error middleware — catch propagated errors,
// emit {"error": message} with a fixed status. It intentionally exposes only
// the error's message string and no stack trace, which is safer for production
// but means it does not replicate Express's development-mode stack output. It
// also does not attempt content negotiation (it always answers JSON) and does
// not honor a per-error statusCode field the way an http-errors-aware handler
// would, keeping the port minimal and dependency-free.
package errorjson

import (
	"net/http"

	"github.com/malcolmston/express"
)

// Options configures the JSON error handler.
type Options struct {
	// Status is the HTTP status code used for the response (default 500).
	Status int
}

// New returns an express.ErrorHandler that writes the error's message as a JSON
// document. Register it with app.Use(errorjson.New()). If the response has
// already been committed it does nothing.
func New(opts ...Options) express.ErrorHandler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Status <= 0 {
		o.Status = http.StatusInternalServerError
	}

	return func(err error, req *express.Request, res *express.Response, next express.Next) {
		if res.Written() {
			return
		}
		msg := "internal server error"
		if err != nil {
			msg = err.Error()
		}
		res.Status(o.Status).JSON(map[string]string{"error": msg})
	}
}
