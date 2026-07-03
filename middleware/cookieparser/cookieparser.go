// Package cookieparser provides middleware that parses all cookies from the
// incoming request into a map[string]string and stores it on the request for
// convenient downstream access.
package cookieparser

import (
	"net/url"

	"github.com/malcolmston/express"
)

// contextKey is the key under which the parsed cookie map is stored on the
// request via req.Set.
const contextKey = "cookies"

// New returns middleware that parses every cookie on the request into a
// map[string]string and stores it via req.Set("cookies", m). Cookie values are
// URL-unescaped when possible. Retrieve the map with From.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		m := make(map[string]string)
		for _, c := range req.Raw.Cookies() {
			if v, err := url.QueryUnescape(c.Value); err == nil {
				m[c.Name] = v
			} else {
				m[c.Name] = c.Value
			}
		}
		req.Set(contextKey, m)
		next()
	}
}

// From returns the parsed cookie map previously stored by the middleware. If
// the middleware did not run it returns an empty, non-nil map.
func From(req *express.Request) map[string]string {
	if v, ok := req.Value(contextKey); ok {
		if m, ok := v.(map[string]string); ok {
			return m
		}
	}
	return map[string]string{}
}
