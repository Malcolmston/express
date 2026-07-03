// Package ipblocklist provides middleware that rejects requests whose client
// IP matches a configured blocklist of addresses or CIDR ranges.
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
