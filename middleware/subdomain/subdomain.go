// Package subdomain provides middleware that extracts the subdomain portion of
// the request's hostname and stores it on the request for downstream handlers
// under the key "subdomain".
package subdomain

import (
	"strings"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the extracted subdomain is stored.
const Key = "subdomain"

// Options configures the subdomain middleware.
type Options struct {
	// BaseHost is the site's base domain (e.g. "example.com"). When set, it is
	// stripped from the hostname and everything before it is treated as the
	// subdomain. When empty, a simple heuristic is used: for a hostname with
	// more than two dotted labels, all labels before the final two form the
	// subdomain.
	BaseHost string
}

// New returns middleware that computes the subdomain and stores it via
// req.Set(Key, sub). When there is no subdomain the stored value is "".
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	base := strings.ToLower(strings.Trim(o.BaseHost, "."))

	return func(req *express.Request, res *express.Response, next express.Next) {
		req.Set(Key, extract(strings.ToLower(req.Hostname()), base))
		next()
	}
}

func extract(host, base string) string {
	if host == "" {
		return ""
	}
	if base != "" {
		if host == base {
			return ""
		}
		if strings.HasSuffix(host, "."+base) {
			return strings.TrimSuffix(host, "."+base)
		}
		return ""
	}
	labels := strings.Split(host, ".")
	if len(labels) <= 2 {
		return ""
	}
	return strings.Join(labels[:len(labels)-2], ".")
}
