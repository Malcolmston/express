// Package ipblocklist provides middleware that rejects requests whose client IP
// address matches a configured blocklist of exact addresses or CIDR ranges,
// responding with 403 Forbidden, and lets every other request through. It
// mirrors the Node "express-ipfilter" package in its deny/blacklist mode and
// the common pattern of an IP-based denylist, reimplemented here on top of Go's
// standard net package with no third-party dependencies.
//
// Use this middleware when the default should be open access but a specific set
// of clients must be turned away — banning abusive addresses, blocking known
// bad networks or scanners, or dropping traffic from ranges you never intend to
// serve. It is a negative/deny model: anything not on the list is allowed,
// which is the complement of the sibling ipallowlist package. Choose blocklist
// when the exceptions are the untrusted parties; choose allowlist when the
// exceptions are the trusted ones.
//
// The handler is meant to run early in the chain, typically via app.Use before
// route handlers, or scoped in front of a particular router. On each request it
// reads the client IP from req.IP() and parses it with net.ParseIP. If the IP
// is valid and matches any blocklist entry, it short-circuits by writing
// res.Status(403) with the body "Forbidden" and does not call next(), so no
// downstream handler runs. In every other case — no match, or an IP that fails
// to parse — it calls next() and processing continues. It writes no additional
// response headers and mutates no request state.
//
// Matching semantics: each Block entry is parsed once at construction time.
// Entries that parse as CIDR (net.ParseCIDR, e.g. "10.0.0.0/8") are matched with
// IPNet.Contains; entries that parse as a bare IP (net.ParseIP, e.g. "1.2.3.4")
// are matched with IP.Equal. Both IPv4 and IPv6 are supported. Important edge
// cases: an empty or nil Block list blocks nothing, so all requests pass;
// malformed entries that parse as neither CIDR nor IP are silently dropped
// rather than raising an error; and a request whose IP cannot be parsed is
// allowed through (fail-open), the deliberate opposite of the allowlist's
// fail-closed behavior. Because the decision relies on req.IP(), correctness
// behind a proxy depends on how the framework derives the client address, so
// only treat this as a security boundary if that value is trustworthy.
//
// Compared to the Node express-ipfilter original, this port keeps the core
// deny-list contract — 403 for matching clients, pass-through otherwise — but is
// intentionally minimal. It does not implement allow mode (that is ipallowlist),
// configurable X-Forwarded-For trust, logging, custom forbidden messages, or
// range shorthand beyond standard CIDR notation. The rejection is a fixed 403
// "Forbidden" with no customization hook.
package ipblocklist

import (
	"net"

	"github.com/malcolmston/express"
)

// Options configures the IP blocklist middleware.
type Options struct {
	// Block is a list of denied client IPs or CIDR ranges (e.g.
	// "10.0.0.0/8"). Required.
	Block []string
}

// New returns middleware that responds with 403 when the request's client IP
// matches one of the configured blocklist entries, and otherwise calls next.
func New(opts Options) express.Handler {
	nets, ips := parseEntries(opts.Block)
	return func(req *express.Request, res *express.Response, next express.Next) {
		ip := net.ParseIP(req.IP())
		if ip != nil && matches(ip, nets, ips) {
			res.Status(403).Send("Forbidden")
			return
		}
		next()
	}
}

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
