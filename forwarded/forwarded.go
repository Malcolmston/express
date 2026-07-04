// Package forwarded parses the X-Forwarded-For header along with a socket
// remote address to produce the chain of addresses that a request traversed.
//
// It is a Go port of the npm module "forwarded". The returned slice always
// begins with the socket remote address, followed by the addresses parsed from
// the X-Forwarded-For header value in reverse order (that is, the proxy closest
// to the server first).
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
