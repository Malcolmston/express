// Package querynormalize provides middleware that normalizes the request query
// string: it lower-cases parameter keys, trims surrounding whitespace from
// values, and rebuilds the raw query in a deterministic (key-sorted) order.
package querynormalize

import (
	"net/url"
	"strings"

	"github.com/malcolmston/express"
)

// New returns middleware that rewrites req.Raw.URL.RawQuery in normalized form.
// Requests without a query string are left untouched.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		raw := req.Raw.URL.RawQuery
		if raw == "" {
			next()
			return
		}
		values, err := url.ParseQuery(raw)
		if err != nil {
			// Leave a malformed query string as-is rather than dropping data.
			next()
			return
		}

		normalized := make(url.Values, len(values))
		for key, vals := range values {
			lk := strings.ToLower(key)
			for _, v := range vals {
				normalized[lk] = append(normalized[lk], strings.TrimSpace(v))
			}
		}

		// url.Values.Encode sorts keys, producing deterministic output.
		req.Raw.URL.RawQuery = normalized.Encode()
		next()
	}
}
