package proxyaddr

import (
	"reflect"
	"testing"
)

func trustNone(addr string, i int) bool { return false }
func trustAll(addr string, i int) bool  { return true }

func TestProxyAddrNoTrust(t *testing.T) {
	// Nothing trusted: real address is the socket address.
	got := ProxyAddr("127.0.0.1", "10.0.0.1, 10.0.0.2", trustNone)
	if got != "127.0.0.1" {
		t.Errorf("got %q, want 127.0.0.1", got)
	}
}

func TestProxyAddrTrustAll(t *testing.T) {
	// Everything trusted: climbs to the leftmost header address.
	got := ProxyAddr("127.0.0.1", "10.0.0.1, 10.0.0.2", trustAll)
	if got != "10.0.0.1" {
		t.Errorf("got %q, want 10.0.0.1", got)
	}
}

func TestProxyAddrCompileLoopback(t *testing.T) {
	trust, err := Compile([]string{"loopback"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// Socket is loopback (trusted), so move up to first XFF entry.
	got := ProxyAddr("127.0.0.1", "10.0.0.1, 192.168.0.1", trust)
	if got != "192.168.0.1" {
		t.Errorf("got %q, want 192.168.0.1", got)
	}
}

func TestProxyAddrCompileMixedTrust(t *testing.T) {
	trust, err := Compile([]string{"loopback", "10.0.0.0/8"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// Chain: [127.0.0.1, 192.168.0.1, 10.0.0.2, 10.0.0.1]
	// 127.0.0.1 trusted -> 192.168.0.1 NOT trusted -> return it.
	got := ProxyAddr("127.0.0.1", "10.0.0.1, 10.0.0.2, 192.168.0.1", trust)
	if got != "192.168.0.1" {
		t.Errorf("got %q, want 192.168.0.1", got)
	}
}

func TestProxyAddrCIDR(t *testing.T) {
	trust, err := Compile([]string{"127.0.0.1/8", "10.0.0.0/8"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	got := ProxyAddr("127.0.0.1", "1.2.3.4, 10.0.0.1", trust)
	if got != "1.2.3.4" {
		t.Errorf("got %q, want 1.2.3.4", got)
	}
}

func TestProxyAddrSingleAddress(t *testing.T) {
	trust, _ := Compile([]string{"loopback"})
	got := ProxyAddr("127.0.0.1", "", trust)
	if got != "127.0.0.1" {
		t.Errorf("got %q, want 127.0.0.1", got)
	}
}

func TestProxyAddrStripsPort(t *testing.T) {
	trust, _ := Compile([]string{"loopback"})
	got := ProxyAddr("127.0.0.1:8080", "1.2.3.4", trust)
	if got != "1.2.3.4" {
		t.Errorf("got %q, want 1.2.3.4", got)
	}
}

func TestAll(t *testing.T) {
	got := All("127.0.0.1", "10.0.0.1, 10.0.0.2")
	want := []string{"127.0.0.1", "10.0.0.2", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompilePresets(t *testing.T) {
	trust, err := Compile([]string{"loopback", "linklocal", "uniquelocal"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	cases := map[string]bool{
		"127.0.0.1":   true,
		"::1":         true,
		"169.254.1.1": true,
		"fe80::1":     true,
		"10.0.0.1":    true,
		"192.168.1.1": true,
		"fc00::1":     true,
		"8.8.8.8":     false,
	}
	for addr, want := range cases {
		if got := trust(addr, 0); got != want {
			t.Errorf("trust(%q) = %v, want %v", addr, got, want)
		}
	}
}

func TestCompileInvalid(t *testing.T) {
	if _, err := Compile([]string{"not-a-cidr!!"}); err == nil {
		t.Error("expected error for invalid CIDR")
	}
	if _, err := Compile([]string{""}); err == nil {
		t.Error("expected error for empty value")
	}
}

func TestCompileBareIP(t *testing.T) {
	trust, err := Compile([]string{"192.168.0.1"})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !trust("192.168.0.1", 0) {
		t.Error("expected exact IP to be trusted")
	}
	if trust("192.168.0.2", 0) {
		t.Error("expected different IP to be untrusted")
	}
}
