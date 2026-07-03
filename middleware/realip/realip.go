// Package realip provides middleware that determines the true client IP
// address from the X-Forwarded-For and X-Real-IP headers, taking an optional
// list of trusted proxies into account, and rewrites req.Raw.RemoteAddr to it.
package realip

import (
	"net"
	"strings"

	"github.com/malcolmston/express"
)

const contextKey = "clientip"

// Options configures the real-IP middleware.
type Options struct {
	// TrustedProxies is an optional list of proxy IP addresses. When set, the
	// client IP is chosen as the right-most entry in X-Forwarded-For that is
	// not one of these proxies. When empty, the left-most X-Forwarded-For entry
	// is used.
	TrustedProxies []string
}

// New returns middleware that resolves the client IP from forwarding headers
// and stores it on the request, also updating req.Raw.RemoteAddr. Retrieve the
// resolved value with ClientIP.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	trusted := make(map[string]bool, len(o.TrustedProxies))
	for _, p := range o.TrustedProxies {
		trusted[strings.TrimSpace(p)] = true
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		ip := resolve(req, trusted)
		if ip != "" {
			req.Set(contextKey, ip)
			req.Raw.RemoteAddr = ip
		}
		next()
	}
}

// ClientIP returns the resolved client IP for the request. If the middleware
// did not run it falls back to the host portion of req.Raw.RemoteAddr.
func ClientIP(req *express.Request) string {
	if v, ok := req.Value(contextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if host, _, err := net.SplitHostPort(req.Raw.RemoteAddr); err == nil {
		return host
	}
	return req.Raw.RemoteAddr
}

func resolve(req *express.Request, trusted map[string]bool) string {
	if xff := req.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		if len(trusted) == 0 {
			return parts[0]
		}
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] != "" && !trusted[parts[i]] {
				return parts[i]
			}
		}
		return parts[0]
	}
	if xr := strings.TrimSpace(req.Get("X-Real-IP")); xr != "" {
		return xr
	}
	if host, _, err := net.SplitHostPort(req.Raw.RemoteAddr); err == nil {
		return host
	}
	return req.Raw.RemoteAddr
}
