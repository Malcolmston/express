// Package useragentblock provides middleware that rejects requests whose
// User-Agent header contains any configured blocked substring.
package useragentblock

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the user-agent blocking middleware.
type Options struct {
	// Block lists case-insensitive substrings; a request whose User-Agent
	// contains any of them is rejected. Required.
	Block []string
}

// New returns middleware that responds with 403 when the request's User-Agent
// contains one of the configured blocked substrings.
func New(opts Options) express.Handler {
	blocked := make([]string, len(opts.Block))
	for i, b := range opts.Block {
		blocked[i] = strings.ToLower(b)
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		ua := strings.ToLower(req.Get("User-Agent"))
		for _, b := range blocked {
			if b != "" && strings.Contains(ua, b) {
				res.Status(403).Send("Forbidden")
				return
			}
		}
		next()
	}
}
