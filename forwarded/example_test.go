package forwarded_test

import (
	"fmt"
	"net/http"

	"github.com/malcolmston/express/forwarded"
)

// ExampleForwarded shows the core parsing behavior of the package. The first
// argument is the socket remote address, and the second is the raw
// X-Forwarded-For header value that the proxies appended as the request was
// relayed. The returned slice starts with the remote address (its port
// stripped) and then lists the header entries from rightmost to leftmost, so
// the proxy nearest the server appears first and the original client last.
// Surrounding whitespace on each header entry is trimmed away.
func ExampleForwarded() {
	addrs := forwarded.Forwarded("127.0.0.1:1234", "10.0.0.1, 10.0.0.2, 192.168.0.1")
	fmt.Println(addrs)
	// Output: [127.0.0.1 192.168.0.1 10.0.0.2 10.0.0.1]
}

// ExampleForwarded_noHeader demonstrates the behavior when no X-Forwarded-For
// header is present. With an empty header value the result is a single-element
// slice containing just the remote address. Any port on the remote address is
// removed first, so "127.0.0.1:8080" becomes "127.0.0.1". This is the common
// case for a request that reached the server directly without passing through a
// proxy. The result is always non-empty.
func ExampleForwarded_noHeader() {
	addrs := forwarded.Forwarded("127.0.0.1:8080", "")
	fmt.Println(addrs)
	// Output: [127.0.0.1]
}

// ExampleForwarded_ipv6 illustrates that IPv6 remote addresses are handled in
// both bracketed and bare forms. A bracketed literal with a port such as
// "[::1]:8080" has its port and brackets removed to yield "::1". A bare IPv6
// literal carries no port and is returned unchanged. This matches the port
// stripping performed by the npm "forwarded" module. Only the remote address is
// normalized this way; header entries are only trimmed.
func ExampleForwarded_ipv6() {
	fmt.Println(forwarded.Forwarded("[::1]:8080", ""))
	fmt.Println(forwarded.Forwarded("2001:db8::1", ""))
	// Output:
	// [::1]
	// [2001:db8::1]
}

// ExampleFromRequest shows the convenience wrapper for an *http.Request. It
// reads RemoteAddr and the X-Forwarded-For header from the request and delegates
// to Forwarded, so you do not have to extract those values yourself. Here the
// request arrived from socket 127.0.0.1 after passing through two proxies. The
// resulting chain lists the closest proxy first and the original client last.
// This is the form you would typically call from an HTTP handler.
func ExampleFromRequest() {
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	fmt.Println(forwarded.FromRequest(r))
	// Output: [127.0.0.1 10.0.0.2 10.0.0.1]
}
