// Package ipallowlist provides middleware that only permits requests whose
// client IP address matches a configured allowlist of exact addresses or CIDR
// ranges, rejecting everything else with 403 Forbidden. It fills the same role
// as the Node "express-ipfilter" package (used in "allow"/whitelist mode) and
// the general pattern of an IP-based access gate, reimplemented here on top of
// Go's standard net package with no third-party dependencies.
//
// Reach for this middleware when a route or an entire application must only be
// reachable from a known set of clients — an internal admin panel, a metrics or
// health endpoint scraped by fixed collectors, a webhook receiver locked to a
// provider's published ranges, or a service fronted by a load balancer whose
// egress IPs are known. It is a positive/allow model: anything not explicitly
// listed is denied, which is the safer default for restricting access. For the
// inverse (deny a few, allow the rest) use the sibling ipblocklist package.
//
// The handler is intended to run early in the chain, usually via app.Use before
// any route handlers or, scoped, in front of a specific router. On each request
// it reads the client IP from req.IP() (which reflects the server's view of the
// remote address / proxy handling) and parses it with net.ParseIP. If the IP is
// valid and matches any configured entry it calls next() and processing
// continues normally; otherwise it short-circuits by writing res.Status(403)
// with the body "Forbidden" and never calls next(), so no downstream handler
// runs. It writes no other response headers and mutates no request state.
//
// Matching semantics: each Allow entry is parsed once at construction time.
// Entries that parse as CIDR (net.ParseCIDR, e.g. "192.168.0.0/24") are matched
// with IPNet.Contains; entries that parse as a bare IP (net.ParseIP, e.g.
// "10.0.0.5") are matched with IP.Equal. Both IPv4 and IPv6 are supported.
// Important edge cases to be aware of: Allow is effectively required — an empty
// or nil list matches nothing and therefore denies every request; malformed
// entries that parse as neither CIDR nor IP are silently dropped rather than
// causing an error; and a request whose IP cannot be parsed is denied. Because
// the decision is based on req.IP(), correctness behind a proxy depends on how
// the framework derives the client address, so trust that value accordingly
// before relying on the allowlist as a security boundary.
//
// Compared to the Node express-ipfilter original, this port keeps the core
// contract — allow-list mode returning 403 for non-matching clients — but is
// deliberately minimal. It does not implement blacklist mode (that is
// ipblocklist), configurable X-Forwarded-For trust, per-entry logging, custom
// forbidden messages, or IP-range shorthand beyond standard CIDR notation. The
// response is a fixed 403 "Forbidden" with no customization hook.
package ipallowlist

import (
	"net"

	"github.com/malcolmston/express"
)

// Options configures the IP allowlist middleware.
type Options struct {
	// Allow is a list of permitted client IPs or CIDR ranges (e.g.
	// "10.0.0.0/8"). Required.
	Allow []string
}

// New returns middleware that responds with 403 unless the request's client IP
// matches one of the configured allowlist entries.
func New(opts Options) express.Handler {
	nets, ips := parseEntries(opts.Allow)
	return func(req *express.Request, res *express.Response, next express.Next) {
		ip := net.ParseIP(req.IP())
		if ip != nil && matches(ip, nets, ips) {
			next()
			return
		}
		res.Status(403).Send("Forbidden")
	}
}

// parseEntries splits allowlist/blocklist entries into CIDR networks and single
// IP addresses.
func parseEntries(entries []string) ([]*net.IPNet, []net.IP) {
	var nets []*net.IPNet
	var ips []net.IP
	for _, e := range entries {
		if _, n, err := net.ParseCIDR(e); err == nil {
			nets = append(nets, n)
			continue
		}
		if ip := net.ParseIP(e); ip != nil {
			ips = append(ips, ip)
		}
	}
	return nets, ips
}

// matches reports whether ip is covered by any of the networks or equals any of
// the single IPs.
func matches(ip net.IP, nets []*net.IPNet, ips []net.IP) bool {
	for _, n := range nets {
		if n.Contains(ip) {
			return true
		}
	}
	for _, a := range ips {
		if a.Equal(ip) {
			return true
		}
	}
	return false
}
