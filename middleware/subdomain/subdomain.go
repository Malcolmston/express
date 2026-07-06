// Package subdomain provides middleware that extracts the subdomain portion of
// the request's hostname and stores it on the request for downstream handlers
// under the key "subdomain". It is the express framework's Go analogue of
// Express's built-in req.subdomains array and its "subdomain offset" setting,
// packaged as an explicit middleware so multi-tenant routing can key off the
// host without every handler re-parsing req.Hostname itself.
//
// Reach for this middleware when a single application serves many tenants or
// sections addressed by host — acme.example.com and globex.example.com routed to
// different customers, or api.example.com, www.example.com, and admin.example.com
// dispatched to different handler trees. Running it once at the front of the
// chain computes the subdomain a single time and hands every downstream handler
// a ready-made value to switch on, avoiding duplicated and inconsistent host
// parsing scattered through the codebase.
//
// Operationally the middleware belongs early, before any handler that needs the
// tenant. On each request it reads the host with req.Hostname (already stripped
// of any port), lowercases it, computes the subdomain, and stashes it with
// req.Set(Key, sub) where Key is the exported constant "subdomain"; it then
// always calls next(). It never reads or writes response headers and never
// short-circuits, so it is transparent to the client — a handler retrieves the
// value with req.Value(subdomain.Key), which is always present and is the empty
// string when there is no subdomain.
//
// Extraction follows one of two rules. When Options.BaseHost is set (for
// example "example.com") it is lowercased and trimmed of surrounding dots, and
// the middleware strips a trailing ".BaseHost" suffix and returns everything
// before it: "api.example.com" yields "api" and "a.b.example.com" yields the
// full multi-level label "a.b". A host equal to BaseHost, or one that does not
// end in BaseHost at all, yields "". When BaseHost is empty a heuristic is used
// instead: the host is split on dots and, only if it has more than two labels,
// all labels before the final two are joined as the subdomain — so
// "shop.example.com" yields "shop" while "example.com" yields "". An empty host
// always yields "".
//
// The heuristic's two-label assumption is the port's main limitation and the
// reason BaseHost exists: multi-label public suffixes such as
// "shop.example.co.uk" would treat "example" as part of the subdomain, and an
// IP-address or localhost host has no meaningful subdomain. Set BaseHost
// explicitly (this port has no Public Suffix List and no configurable offset
// beyond it) for any real deployment. Compared with Express, which exposes
// req.subdomains as a reversed array of labels governed by "subdomain offset",
// this middleware stores a single joined string under a request value and leaves
// any further splitting to the caller.
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
