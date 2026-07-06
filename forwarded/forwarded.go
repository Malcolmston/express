// Package forwarded parses the X-Forwarded-For header along with a socket
// remote address to produce the chain of addresses that a request traversed.
// It is a Go port of the npm module "forwarded" (the same helper that Express
// and the "proxy-addr" module build on) and reproduces that module's ordering
// and trimming rules exactly.
//
// Use this package when your application sits behind one or more reverse
// proxies or load balancers and you need to recover the client's original
// address, or the full list of intermediaries a request passed through. The
// socket's immediate peer (the nearest proxy, or the client itself when there
// is no proxy) is always known reliably; the X-Forwarded-For header records the
// addresses each proxy observed and appended as the request was relayed onward.
//
// The algorithm is deliberately simple and does no validation. Forwarded takes
// the socket remote address and the raw X-Forwarded-For value, strips any port
// from the remote address, and returns a slice whose first element is that
// remote address. It then splits the header on commas and appends the entries
// from rightmost to leftmost, so the resulting slice reads from the proxy
// closest to the server outward toward the original client. FromRequest is a
// convenience wrapper that pulls r.RemoteAddr and the X-Forwarded-For header
// from an *http.Request before delegating to Forwarded.
//
// Several edge cases are handled to preserve parity with the Node original.
// An empty or missing header yields a single-element slice containing only the
// remote address. Each header entry is trimmed of surrounding whitespace, but
// empty entries (for example from a "a,,b" value) are preserved as empty
// strings rather than being dropped. Port stripping understands IPv4
// "host:port", bracketed IPv6 "[::1]:8080" and "[::1]" forms, and bare IPv6
// literals such as "::1" or "2001:db8::1", which are returned unchanged because
// they carry no port. Malformed input is never rejected; it is returned as-is.
//
// The one intentional deviation from the JavaScript module is ergonomic rather
// than behavioral: because Go's *http.Request exposes RemoteAddr as a string
// instead of a socket object, the core logic lives in Forwarded, which accepts
// the remote address and header value directly, and FromRequest adapts a
// request to that signature. The addresses returned are plain strings and are
// not parsed into net.IP values, matching the string-based results of the npm
// package.
package forwarded

import (
	"net"
	"net/http"
	"strings"
)

// Forwarded returns the list of addresses in the request chain.
//
// The first element is remoteAddr (with any port stripped). The remaining
// elements are the comma-separated entries of the X-Forwarded-For value (xff),
// trimmed of surrounding whitespace and appended from the rightmost entry to
// the leftmost entry. Empty entries in the header are preserved as empty
// strings, matching the behavior of the npm "forwarded" module.
func Forwarded(remoteAddr string, xff string) []string {
	addrs := []string{stripPort(remoteAddr)}

	if xff == "" {
		return addrs
	}

	parts := strings.Split(xff, ",")
	// Push entries from rightmost to leftmost (closest proxy first).
	for i := len(parts) - 1; i >= 0; i-- {
		addrs = append(addrs, strings.TrimSpace(parts[i]))
	}

	return addrs
}

// FromRequest returns the address chain for an *http.Request, reading
// r.RemoteAddr (with the port stripped) and the X-Forwarded-For header.
func FromRequest(r *http.Request) []string {
	return Forwarded(r.RemoteAddr, r.Header.Get("X-Forwarded-For"))
}

// stripPort removes a trailing port from an address if present, correctly
// handling bracketed and unbracketed IPv6 literals as well as IPv4 addresses.
func stripPort(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return addr
	}

	// Bracketed IPv6 form, optionally with a port: [::1]:8080 or [::1].
	if addr[0] == '[' {
		if host, _, err := net.SplitHostPort(addr); err == nil {
			return host
		}
		// No port; strip the brackets.
		if end := strings.IndexByte(addr, ']'); end != -1 {
			return addr[1:end]
		}
		return addr
	}

	// If it parses as a bare IP (including unbracketed IPv6), leave it as is.
	if net.ParseIP(addr) != nil {
		return addr
	}

	// Otherwise, attempt to split host:port (IPv4 with port, or host:port).
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}

	return addr
}
