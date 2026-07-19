package proxyaddr

// Parity tests transcribed from the upstream npm module "jshttp/proxy-addr".
//
// Every input -> expected-output vector below is taken verbatim from the
// upstream test suite (the createReq(socketAddr, {"x-forwarded-for": ...})
// requests and their asserted results):
//
//	https://raw.githubusercontent.com/jshttp/proxy-addr/master/test/test.js
//
// The reference implementation used to confirm matching semantics is:
//
//	https://raw.githubusercontent.com/jshttp/proxy-addr/master/index.js
//
// Upstream helper trust functions are reproduced here:
//
//	function all ()      { return true }
//	function none ()     { return false }
//	function trust10x(a) { return /^10\./.test(a) }
//
// The Go port operates on the raw socket remote-address string and the raw
// X-Forwarded-For header string rather than a Node request object, so each
// createReq(socket, header) is expressed as (socket, xff).

import (
	"reflect"
	"strings"
	"testing"
)

func parityAll(addr string, i int) bool  { return true }
func parityNone(addr string, i int) bool { return false }
func parityTrust10x(addr string, i int) bool {
	return strings.HasPrefix(addr, "10.")
}

// mustCompile compiles a trust value list, failing the test on error.
func mustCompile(t *testing.T, vals ...string) func(string, int) bool {
	t.Helper()
	trust, err := Compile(vals)
	if err != nil {
		t.Fatalf("Compile(%v): %v", vals, err)
	}
	return trust
}

// TestParityProxyAddrFunctions covers the function-trust vectors from the
// upstream "with all trusted", "with none trusted", and "with some trusted"
// describe blocks.
func TestParityProxyAddrFunctions(t *testing.T) {
	cases := []struct {
		name   string
		socket string
		xff    string
		trust  func(string, int) bool
		want   string
	}{
		// with all trusted
		{"all/no-headers", "127.0.0.1", "", parityAll, "127.0.0.1"},
		{"all/header-value", "127.0.0.1", "10.0.0.1", parityAll, "10.0.0.1"},
		{"all/furthest-header", "127.0.0.1", "10.0.0.1, 10.0.0.2", parityAll, "10.0.0.1"},
		// with none trusted
		{"none/no-headers", "127.0.0.1", "", parityNone, "127.0.0.1"},
		{"none/with-headers", "127.0.0.1", "10.0.0.1, 10.0.0.2", parityNone, "127.0.0.1"},
		// with some trusted (trust10x)
		{"some/no-headers", "127.0.0.1", "", parityTrust10x, "127.0.0.1"},
		{"some/not-trusted", "127.0.0.1", "10.0.0.1, 10.0.0.2", parityTrust10x, "127.0.0.1"},
		{"some/socket-trusted", "10.0.0.1", "192.168.0.1", parityTrust10x, "192.168.0.1"},
		{"some/first-untrusted-after-trusted", "10.0.0.1", "192.168.0.1, 10.0.0.2", parityTrust10x, "192.168.0.1"},
		{"some/not-skip-untrusted", "10.0.0.1", "10.0.0.3, 192.168.0.1, 10.0.0.2", parityTrust10x, "192.168.0.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ProxyAddr(c.socket, c.xff, c.trust); got != c.want {
				t.Errorf("ProxyAddr(%q, %q) = %q, want %q", c.socket, c.xff, got, c.want)
			}
		})
	}
}

// TestParityProxyAddrCompiled covers the compiled-trust vectors: arrays, empty
// arrays, IPv4/IPv6 literals and CIDR, netmask notation, mixed IP versions,
// IPv4-mapped IPv6, pre-defined names, and non-IP header values.
func TestParityProxyAddrCompiled(t *testing.T) {
	cases := []struct {
		name   string
		socket string
		xff    string
		trust  []string
		want   string
	}{
		// when given array
		{"array/literal", "10.0.0.1", "192.168.0.1, 10.0.0.2", []string{"10.0.0.1", "10.0.0.2"}, "192.168.0.1"},
		{"array/non-ip-not-trusted", "10.0.0.1", "192.168.0.1, 10.0.0.2, localhost", []string{"10.0.0.1", "10.0.0.2"}, "localhost"},
		{"array/none-match", "10.0.0.1", "192.168.0.1, 10.0.0.2", []string{"127.0.0.1", "192.168.0.100"}, "10.0.0.1"},
		{"array/empty-no-headers", "127.0.0.1", "", []string{}, "127.0.0.1"},
		{"array/empty-with-headers", "127.0.0.1", "10.0.0.1, 10.0.0.2", []string{}, "127.0.0.1"},
		// when given IPv4 addresses
		{"ipv4/literal", "10.0.0.1", "192.168.0.1, 10.0.0.2", []string{"10.0.0.1", "10.0.0.2"}, "192.168.0.1"},
		{"ipv4/cidr", "10.0.0.1", "192.168.0.1, 10.0.0.200", []string{"10.0.0.2/26"}, "10.0.0.200"},
		{"ipv4/netmask", "10.0.0.1", "192.168.0.1, 10.0.0.200", []string{"10.0.0.2/255.255.255.192"}, "10.0.0.200"},
		// when given IPv6 addresses
		{"ipv6/literal", "fe80::1", "2002:c000:203::1, fe80::2", []string{"fe80::1", "fe80::2"}, "2002:c000:203::1"},
		{"ipv6/cidr", "fe80::1", "2002:c000:203::1, fe80::ff00", []string{"fe80::/125"}, "fe80::ff00"},
		// when IP versions mixed
		{"mixed/match-respective", "::1", "2002:c000:203::1", []string{"127.0.0.1", "::1"}, "2002:c000:203::1"},
		{"mixed/no-cross-version", "::1", "2002:c000:203::1", []string{"127.0.0.1"}, "::1"},
		// when IPv4-mapped IPv6 addresses
		{"mapped/ipv4-trust-ipv6-req", "::ffff:a00:1", "192.168.0.1, 10.0.0.2", []string{"10.0.0.1", "10.0.0.2"}, "192.168.0.1"},
		{"mapped/ipv4-netmask-trust-ipv6-req", "::ffff:a00:1", "192.168.0.1, 10.0.0.2", []string{"10.0.0.1/16"}, "192.168.0.1"},
		{"mapped/ipv6-trust-ipv4-req", "10.0.0.1", "192.168.0.1, 10.0.0.2", []string{"::ffff:a00:1", "::ffff:a00:2"}, "192.168.0.1"},
		{"mapped/cidr", "10.0.0.1", "192.168.0.1, 10.0.0.200", []string{"::ffff:a00:2/122"}, "10.0.0.200"},
		{"mapped/cidr-mixed-ipv6-cidr", "10.0.0.1", "192.168.0.1, 10.0.0.200", []string{"::ffff:a00:2/122", "fe80::/125"}, "10.0.0.200"},
		{"mapped/cidr-mixed-ipv4", "10.0.0.1", "192.168.0.1, 10.0.0.200", []string{"::ffff:a00:2/122", "127.0.0.1"}, "10.0.0.200"},
		// when given pre-defined names
		{"names/single", "fe80::1", "2002:c000:203::1, fe80::2", []string{"linklocal"}, "2002:c000:203::1"},
		{"names/multiple", "::1", "2002:c000:203::1, fe80::2", []string{"loopback", "linklocal"}, "2002:c000:203::1"},
		// when header contains non-ip addresses
		{"nonip/stop-first-nonip", "127.0.0.1", "myrouter, 127.0.0.1, proxy", []string{"127.0.0.1"}, "proxy"},
		{"nonip/stop-first-malformed", "127.0.0.1", "myrouter, 127.0.0.1, ::8:8:8:8:8:8:8:8:8", []string{"127.0.0.1"}, "::8:8:8:8:8:8:8:8:8"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			trust := mustCompile(t, c.trust...)
			if got := ProxyAddr(c.socket, c.xff, trust); got != c.want {
				t.Errorf("ProxyAddr(%q, %q, %v) = %q, want %q", c.socket, c.xff, c.trust, got, c.want)
			}
		})
	}
}

// TestParityAll covers proxyaddr.all(req) without a trust argument: the full
// ordered chain of socket address plus reversed X-Forwarded-For entries.
func TestParityAll(t *testing.T) {
	cases := []struct {
		name   string
		socket string
		xff    string
		want   []string
	}{
		{"no-headers", "127.0.0.1", "", []string{"127.0.0.1"}},
		{"single-header", "127.0.0.1", "10.0.0.1", []string{"127.0.0.1", "10.0.0.1"}},
		{"order", "127.0.0.1", "10.0.0.1, 10.0.0.2", []string{"127.0.0.1", "10.0.0.2", "10.0.0.1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := All(c.socket, c.xff); !reflect.DeepEqual(got, c.want) {
				t.Errorf("All(%q, %q) = %v, want %v", c.socket, c.xff, got, c.want)
			}
		})
	}
}

// TestParityTrustInvocation verifies the trust predicate is invoked as
// trust(addr, i) over the chain, and is not called for the final (topmost)
// address. Vectors are from the upstream "should be invoked as trust(addr, i)"
// and "should provide all values to function" tests, both using an
// always-truthy trust function.
func TestParityTrustInvocation(t *testing.T) {
	type call struct {
		addr string
		i    int
	}

	cases := []struct {
		name   string
		socket string
		xff    string
		want   []call
	}{
		{
			name:   "invoked-as-addr-i",
			socket: "127.0.0.1",
			xff:    "192.168.0.1, 10.0.0.1",
			want:   []call{{"127.0.0.1", 0}, {"10.0.0.1", 1}},
		},
		{
			name:   "all-values",
			socket: "127.0.0.1",
			xff:    "myrouter, 127.0.0.1, proxy",
			want:   []call{{"127.0.0.1", 0}, {"proxy", 1}, {"127.0.0.1", 2}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var log []call
			ProxyAddr(c.socket, c.xff, func(addr string, i int) bool {
				log = append(log, call{addr, i})
				return true
			})
			if !reflect.DeepEqual(log, c.want) {
				t.Errorf("trust calls = %v, want %v", log, c.want)
			}
		})
	}
}

// TestParityCompileValidation covers the upstream compile() argument-validation
// vectors: accepted forms return a usable function, rejected forms return an
// error ("invalid IP address" / "invalid range on address").
func TestParityCompileValidation(t *testing.T) {
	accept := []string{
		"127.0.0.1",
		"::1",
		"::ffff:127.0.0.1",
		"loopback",
		"10.0.0.2/26",
		"10.0.0.2/255.255.255.192",
		"fe80::/125",
		"::ffff:a00:2/122",
	}
	for _, v := range accept {
		if _, err := Compile([]string{v}); err != nil {
			t.Errorf("Compile([%q]) unexpected error: %v", v, err)
		}
	}
	if _, err := Compile([]string{"loopback", "10.0.0.1"}); err != nil {
		t.Errorf("Compile preset+ip unexpected error: %v", err)
	}

	reject := []string{
		"blargh",
		"10.0.300.1",
		"::ffff:30.168.1.9000",
		"-1",
		"10.0.0.1/internet",
		"10.0.0.1/6000",
		"::1/6000",
		"::ffff:a00:2/136",
		"::ffff:a00:2/-1",
		"10.0.0.1/255.0.255.0",
		"10.0.0.1/ffc0::",
		"fe80::/ffc0::",
		"fe80::/255.255.255.0",
		"::ffff:a00:2/255.255.255.0",
	}
	for _, v := range reject {
		if _, err := Compile([]string{v}); err == nil {
			t.Errorf("Compile([%q]) expected error, got nil", v)
		}
	}
}
