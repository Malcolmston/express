// Package realip provides middleware that determines the true client IP
// address from the X-Forwarded-For and X-Real-IP headers, taking an optional
// list of trusted proxies into account, and rewrites req.Raw.RemoteAddr to the
// resolved value. It ports the "trust proxy" resolution that Express performs
// on req.ip together with the header-walking logic of the Node request-ip and
// proxy-addr packages, collapsing them into a single handler that leaves the
// rest of the stack seeing the originating client rather than the nearest hop.
//
// Use it whenever the application sits behind a reverse proxy, load balancer,
// CDN, or ingress controller that terminates the client connection and forwards
// the real address in a header. Without it, req.Raw.RemoteAddr — and therefore
// anything derived from it such as req.IP(), logging, and IP-keyed rate limiting
// — would report the proxy's address instead of the end user's. Placing this
// middleware first lets every downstream consumer read a corrected RemoteAddr
// transparently, which is exactly what the RewritesRemoteAddr test asserts.
//
// Mechanically the middleware resolves the address, and when the result is
// non-empty stores it on the request under an internal key and overwrites
// req.Raw.RemoteAddr, then always calls next(); it never short-circuits or
// writes a response. Resolution reads X-Forwarded-For first: its comma-separated
// list is split and trimmed, and with no trusted proxies configured the
// left-most entry (the original client) is returned. When TrustedProxies is
// non-empty the list is walked from the right, returning the first address that
// is not a known proxy, which defends against a client spoofing extra hops on
// the left. If X-Forwarded-For is absent it falls back to a trimmed X-Real-IP,
// and finally to the host portion of RemoteAddr. Register it via app.Use ahead
// of any middleware that inspects the client IP.
//
// The single option, TrustedProxies, selects the resolution strategy: empty
// means "trust the left-most XFF entry" (simplest, appropriate when the header
// is written by an infrastructure you control and clients cannot inject it),
// while a populated list switches to the safer right-most-untrusted walk. Note
// the stored value is written verbatim into RemoteAddr without a port, so callers
// that later net.SplitHostPort it should tolerate a bare host; the ClientIP
// helper handles that fallback for you. Comparison against TrustedProxies is an
// exact string match on trimmed entries, so it expects addresses rather than
// CIDR ranges — a deliberate simplification over proxy-addr's subnet support.
//
// Parity with the Node originals is behavioral on the common path and simplified
// on the edges. Like Express's trust-proxy and request-ip it prefers
// X-Forwarded-For, understands X-Real-IP, and returns the client rather than the
// proxy; unlike proxy-addr it does not parse CIDR notation, recognize the
// "loopback"/"linklocal"/"uniquelocal" preset subnets, or consult the many
// vendor-specific headers (CF-Connecting-IP, True-Client-IP, and friends). The
// resolved address is exposed both by mutating RemoteAddr and via the exported
// ClientIP accessor so downstream code can opt into either integration style.
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
