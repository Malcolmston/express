// Package ipallowlist provides middleware that only permits requests whose
// client IP matches a configured allowlist of addresses or CIDR ranges.
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
