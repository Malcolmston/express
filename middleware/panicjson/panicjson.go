// Package panicjson provides middleware that recovers from panics in
// downstream handlers, logs them, and responds with a generic 500 JSON error
// so a single failing request cannot crash the server.
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
