package proxyaddr_test

import (
	"fmt"

	"github.com/malcolmston/express/proxyaddr"
)

// ExampleProxyAddr resolves the real client address of a request that arrived
// through a trusted reverse proxy. Compile builds a trust predicate from the
// "loopback" preset, which trusts 127.0.0.1 and ::1. The socket's remote address
// is the loopback proxy, and the X-Forwarded-For header carries the original
// client address 203.0.113.1. Because the socket hop is trusted, the walk steps
// past it to the header entry, and since that address is not itself trusted it is
// returned as the client. This is the logic behind Express's req.ip under a
// "trust proxy" setting.
func ExampleProxyAddr() {
	trust, err := proxyaddr.Compile([]string{"loopback"})
	if err != nil {
		panic(err)
	}
	fmt.Println(proxyaddr.ProxyAddr("127.0.0.1", "203.0.113.1", trust))
	// Output: 203.0.113.1
}

// ExampleAll returns the full forwarded address chain without applying any trust
// decision, which is useful for logging or for building the equivalent of
// Express's req.ips. The chain always begins with the socket remote address,
// followed by the X-Forwarded-For entries read from right (nearest proxy) to left
// (original client). Here the socket address 127.0.0.1 comes first, then the
// rightmost header value 198.51.100.2, then the leftmost 203.0.113.1. Callers can
// pair this ordering with a trust predicate to reason about each hop.
func ExampleAll() {
	fmt.Println(proxyaddr.All("127.0.0.1", "203.0.113.1, 198.51.100.2"))
	// Output: [127.0.0.1 198.51.100.2 203.0.113.1]
}
