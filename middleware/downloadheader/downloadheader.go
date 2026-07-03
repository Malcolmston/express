// Package downloadheader provides express middleware that marks responses as
// file downloads via the Content-Disposition header.
package downloadheader

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the middleware.
type Options struct {
	// Filename is the suggested download filename. When empty, a bare
	// "attachment" disposition is sent.
	Filename string
}

// New returns middleware that sets Content-Disposition: attachment so browsers
// offer to save the response rather than render it. If a Filename is given it
// is included (with quotes escaped) as the suggested name.
func New(opts ...Options) express.Handler {
	var filename string
	if len(opts) > 0 {
		filename = opts[0].Filename
	}
	value := "attachment"
	if filename != "" {
		value = `attachment; filename="` + escapeQuotes(filename) + `"`
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Content-Disposition", value)
		next()
	}
}

func escapeQuotes(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
