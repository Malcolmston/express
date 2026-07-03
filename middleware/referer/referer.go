// Package referer provides middleware that captures the Referer request header,
// parses out its host, and stores the result on the request under the key
// "referer" for downstream handlers (analytics, anti-CSRF checks, etc.).
package referer

import (
	"net/url"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the Referer info is stored.
const Key = "referer"

// Referer holds the raw Referer header and its parsed host.
type Referer struct {
	// URL is the raw Referer header value (may be empty).
	URL string

	// Host is the host component parsed from URL, or "" when absent or
	// unparseable.
	Host string
}

// New returns middleware that stores a Referer via req.Set(Key, ref). Both the
// standard "Referer" and the (rare) correct spelling "Referrer" are consulted.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		raw := req.Get("Referer")
		if raw == "" {
			raw = req.Get("Referrer")
		}
		ref := Referer{URL: raw}
		if raw != "" {
			if u, err := url.Parse(raw); err == nil {
				ref.Host = u.Host
			}
		}
		req.Set(Key, ref)
		next()
	}
}

// From retrieves the Referer stored by this middleware.
func From(req *express.Request) (Referer, bool) {
	v, ok := req.Value(Key)
	if !ok {
		return Referer{}, false
	}
	r, ok := v.(Referer)
	return r, ok
}
