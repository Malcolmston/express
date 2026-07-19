package forwarded

import (
	"net/http"
	"reflect"
	"testing"
)

func TestForwardedNoHeader(t *testing.T) {
	got := Forwarded("127.0.0.1", "")
	want := []string{"127.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedReverseOrder(t *testing.T) {
	got := Forwarded("127.0.0.1", "10.0.0.1, 10.0.0.2, 192.168.0.1")
	want := []string{"127.0.0.1", "192.168.0.1", "10.0.0.2", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedSingleEntry(t *testing.T) {
	got := Forwarded("127.0.0.1", "10.0.0.1")
	want := []string{"127.0.0.1", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedTrimsWhitespace(t *testing.T) {
	got := Forwarded("127.0.0.1", "  10.0.0.1 ,  10.0.0.2  ")
	want := []string{"127.0.0.1", "10.0.0.2", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedStripsPortIPv4(t *testing.T) {
	got := Forwarded("127.0.0.1:8080", "")
	want := []string{"127.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedStripsPortIPv6Bracketed(t *testing.T) {
	got := Forwarded("[::1]:8080", "")
	want := []string{"::1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedBareIPv6NoPort(t *testing.T) {
	got := Forwarded("::1", "")
	want := []string{"::1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedFullIPv6(t *testing.T) {
	got := Forwarded("2001:db8::1", "")
	want := []string{"2001:db8::1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestForwardedBlankEntrySkipped(t *testing.T) {
	got := Forwarded("127.0.0.1", "10.0.0.1,,10.0.0.2")
	want := []string{"127.0.0.1", "10.0.0.2", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFromRequest(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	r.RemoteAddr = "127.0.0.1:1234"
	r.Header.Set("X-Forwarded-For", "10.0.0.1, 10.0.0.2")

	got := FromRequest(r)
	want := []string{"127.0.0.1", "10.0.0.2", "10.0.0.1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFromRequestNoHeader(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.com", nil)
	r.RemoteAddr = "[2001:db8::1]:443"

	got := FromRequest(r)
	want := []string{"2001:db8::1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
