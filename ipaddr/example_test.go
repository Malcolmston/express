package ipaddr_test

import (
	"fmt"

	"github.com/malcolmston/express/ipaddr"
)

// ExampleParse parses a textual IPv4 address into an *Address and inspects it.
// Parse trims surrounding whitespace and rejects anything net.ParseIP cannot
// read, returning an error in that case. The Kind method reports the detected
// family, either "ipv4" or "ipv6". The String method returns the canonical form
// of the address, preserving the IPv4 family rather than widening it to an
// IPv4-mapped IPv6 literal. Here a dotted-quad address is recognised as IPv4 and
// printed back in its original family.
func ExampleParse() {
	a, err := ipaddr.Parse("192.168.1.1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(a.Kind())
	fmt.Println(a.String())
	// Output:
	// ipv4
	// 192.168.1.1
}

// ExampleIsValid demonstrates the boolean validity helper. IsValid is a thin
// wrapper over Parse that reports only whether the string is a syntactically
// valid IPv4 or IPv6 address. A loopback IPv4 address is valid, an IPv6 literal
// is valid, but an octet above 255 is not, and plain non-address text is not.
// It never returns an error, making it convenient for guard clauses. Use Parse
// instead when you need the parsed value or the specific failure reason.
func ExampleIsValid() {
	fmt.Println(ipaddr.IsValid("127.0.0.1"))
	fmt.Println(ipaddr.IsValid("::1"))
	fmt.Println(ipaddr.IsValid("10.0.0.256"))
	fmt.Println(ipaddr.IsValid("garbage"))
	// Output:
	// true
	// false
	// false
	// false
}

// ExampleAddress_Match tests CIDR containment. Match parses the given CIDR block
// with net.ParseCIDR and reports whether the address falls inside that network.
// A malformed CIDR yields a non-nil error rather than a false negative. Here the
// address 192.168.5.5 is inside 192.168.0.0/16 but outside 10.0.0.0/8. The same
// method works uniformly for IPv6 ranges. This mirrors the match method of the
// ipaddr.js library.
func ExampleAddress_Match() {
	a, _ := ipaddr.Parse("192.168.5.5")
	in, _ := a.Match("192.168.0.0/16")
	out, _ := a.Match("10.0.0.0/8")
	fmt.Println(in)
	fmt.Println(out)
	// Output:
	// true
	// false
}

// ExampleAddress_Range classifies addresses into category names. Range applies
// the special-range tables from ipaddr.js in order and returns the first match.
// A 10.x address is "private", a public address such as 8.8.8.8 is "unicast",
// and the IPv6 loopback ::1 is "loopback". IPv4 and IPv6 use overlapping but
// distinct category sets, with "unicast" as the fallthrough for ordinary public
// addresses. This is useful for deciding whether to trust or filter an address.
func ExampleAddress_Range() {
	for _, s := range []string{"10.1.2.3", "8.8.8.8", "::1"} {
		a, _ := ipaddr.Parse(s)
		fmt.Printf("%s -> %s\n", s, a.Range())
	}
	// Output:
	// 10.1.2.3 -> private
	// 8.8.8.8 -> unicast
	// ::1 -> loopback
}
