// Package methodallow provides middleware that restricts requests to a
// configured set of HTTP methods, responding with 405 otherwise.
package methodallow

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the method-allow middleware.
type Options struct {
	// Methods lists the permitted HTTP methods (case-insensitive). Required.
	Methods []string
}

// New returns middleware that responds with 405 Method Not Allowed, including
// an Allow header, unless the request method is in the permitted set.
func New(opts Options) express.Handler {
	allowed := make(map[string]struct{}, len(opts.Methods))
	canonical := make([]string, 0, len(opts.Methods))
	for _, m := range opts.Methods {
		up := strings.ToUpper(m)
		allowed[up] = struct{}{}
		canonical = append(canonical, up)
	}
	allowHeader := strings.Join(canonical, ", ")
	return func(req *express.Request, res *express.Response, next express.Next) {
		if _, ok := allowed[strings.ToUpper(req.Method())]; ok {
			next()
			return
		}
		res.Set("Allow", allowHeader)
		res.Status(405).Send("Method Not Allowed")
	}
}
