// Package errorjson provides an express error handler that renders unhandled
// errors as a JSON object of the form {"error": "<message>"}.
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
