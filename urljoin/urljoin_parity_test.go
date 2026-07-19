package urljoin

import "testing"

// Parity vectors transcribed verbatim from the upstream npm package
// jfromaniello/url-join (version 5.0.0). Each input/expected pair is copied
// directly from the reference test suite; none of these values are invented.
//
// Upstream test source:
//   https://raw.githubusercontent.com/jfromaniello/url-join/main/test/tests.js
// Upstream implementation reference:
//   https://raw.githubusercontent.com/jfromaniello/url-join/main/lib/url-join.js
//
// Vectors that exercise upstream's runtime type checking (throwing a TypeError
// for non-string segments) and the array-argument calling convention are
// intentionally omitted: the Go port's API is a variadic list of strings, so
// those behaviors have no representable analogue. The array-syntax cases that
// carry the same values as a variadic call are included as ordinary calls.

func TestParityUpstream(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		// joins a simple URL
		{"simple", []string{"http://www.google.com/", "foo/bar", "?test=123"}, "http://www.google.com/foo/bar?test=123"},
		// joins a hashbang
		{"hashbang", []string{"http://www.google.com", "#!", "foo/bar", "?test=123"}, "http://www.google.com/#!/foo/bar?test=123"},
		// joins a protocol
		{"protocol_colon_only", []string{"http:", "www.google.com/", "foo/bar", "?test=123"}, "http://www.google.com/foo/bar?test=123"},
		// joins a protocol with slashes
		{"protocol_slashes", []string{"http://", "www.google.com/", "foo/bar", "?test=123"}, "http://www.google.com/foo/bar?test=123"},
		// removes extra slashes
		{"extra_slashes", []string{"http:", "www.google.com///", "foo/bar", "?test=123"}, "http://www.google.com/foo/bar?test=123"},
		// removes extra slashes in an encoded URL
		{"encoded_url", []string{"http:", "www.google.com///", "foo/bar", "?url=http%3A//Ftest.com"}, "http://www.google.com/foo/bar?url=http%3A//Ftest.com"},
		{"encoded_url_2", []string{"http://a.com/23d04b3/", "/b/c.html"}, "http://a.com/23d04b3/b/c.html"},
		{"encoded_url_3", []string{"/foo", "/", "bar", "?test=123"}, "/foo/bar?test=123"},
		// joins anchors in URLs
		{"anchor", []string{"http:", "www.google.com///", "foo/bar", "?test=123", "#faaaaa"}, "http://www.google.com/foo/bar?test=123#faaaaa"},
		// joins protocol-relative URLs
		{"protocol_relative", []string{"//www.google.com", "foo/bar", "?test=123"}, "//www.google.com/foo/bar?test=123"},
		// joins file protocol URLs
		{"file_single_slash", []string{"file:/", "android_asset", "foo/bar"}, "file://android_asset/foo/bar"},
		{"file_colon_slash_seg", []string{"file:", "/android_asset", "foo/bar"}, "file://android_asset/foo/bar"},
		// joins absolute file protocol URLs
		{"file_triple_from_seg", []string{"file:", "///android_asset", "foo/bar"}, "file:///android_asset/foo/bar"},
		{"file_triple", []string{"file:///", "android_asset", "foo/bar"}, "file:///android_asset/foo/bar"},
		{"file_triple_double_seg", []string{"file:///", "//android_asset", "foo/bar"}, "file:///android_asset/foo/bar"},
		{"file_triple_inline", []string{"file:///android_asset", "foo/bar"}, "file:///android_asset/foo/bar"},
		// joins multiple query params
		{"multi_query_1", []string{"http:", "www.google.com///", "foo/bar", "?test=123", "?key=456"}, "http://www.google.com/foo/bar?test=123&key=456"},
		{"multi_query_2", []string{"http:", "www.google.com///", "foo/bar", "?test=123", "?boom=value", "&key=456"}, "http://www.google.com/foo/bar?test=123&boom=value&key=456"},
		{"multi_query_3", []string{"http://example.org/x", "?a=1", "?b=2", "?c=3", "?d=4"}, "http://example.org/x?a=1&b=2&c=3&d=4"},
		{"multi_query_4", []string{"http:", "www.google.com///", "foo/bar", "&test=123", "&key=456"}, "http://www.google.com/foo/bar?test=123&key=456"},
		{"multi_query_5", []string{"http:", "www.google.com///", "foo/bar", "&test=123", "?key=456"}, "http://www.google.com/foo/bar?test=123&key=456"},
		// filters out empty query parameters
		{"empty_query_q", []string{"http://google.com", "?"}, "http://google.com"},
		{"empty_query_amp", []string{"http://google.com", "&"}, "http://google.com"},
		// joins slashes in paths
		{"slashes_in_paths", []string{"http://example.org", "a//", "b//", "A//", "B//"}, "http://example.org/a/b/A/B/"},
		// joins colons in paths
		{"colons_in_paths", []string{"http://example.org/", ":foo:", "bar"}, "http://example.org/:foo:/bar"},
		// joins a simple path without a URL
		{"path_no_url", []string{"/", "test"}, "/test"},
		// joins a path with a colon
		{"path_with_colon", []string{"/users/:userId", "/cars/:carId"}, "/users/:userId/cars/:carId"},
		// joins slashes in the protocol
		{"proto_slashes_1", []string{"http://example.org", "a"}, "http://example.org/a"},
		{"proto_slashes_2", []string{"http:", "//example.org", "a"}, "http://example.org/a"},
		{"proto_slashes_3", []string{"http:///example.org", "a"}, "http://example.org/a"},
		{"proto_slashes_4", []string{"file:///example.org", "a"}, "file:///example.org/a"},
		{"proto_slashes_5", []string{"file:example.org", "a"}, "file://example.org/a"},
		{"proto_slashes_6", []string{"file:/", "example.org", "a"}, "file://example.org/a"},
		{"proto_slashes_7", []string{"file:", "/example.org", "a"}, "file://example.org/a"},
		{"proto_slashes_8", []string{"file:", "//example.org", "a"}, "file://example.org/a"},
		// skips empty strings
		{"skip_empty_1", []string{"http://foobar.com", "", "test"}, "http://foobar.com/test"},
		{"skip_empty_2", []string{"", "http://foobar.com", "", "test"}, "http://foobar.com/test"},
		// does not replace query params after the hash
		{"query_after_hash", []string{"http://example.com", "#a?b?c"}, "http://example.com#a?b?c"},
		// joins broken up query params
		{"broken_query", []string{"http://example.com", "/foo/bar?", "test=123"}, "http://example.com/foo/bar?test=123"},
		// joins broken up hash
		{"broken_hash", []string{"http://example.com", "/foo/bar#", "some-hash"}, "http://example.com/foo/bar#some-hash"},
		// joins leading empty string
		{"leading_empty", []string{"", "/test"}, "/test"},
		// joins a leading IPv6 hostname
		{"ipv6_leading", []string{"[2601:195:c381:3560::f42a]/", "/test"}, "[2601:195:c381:3560::f42a]/test"},
		// joins a leading IPv6 host with an IPv4 address in the least significant 32 bits
		{"ipv6_ipv4", []string{"[2601:195:c381:3560::0.0.244.42]", "/test"}, "[2601:195:c381:3560::0.0.244.42]/test"},
		// joins a protocol followed by an IPv6 host
		{"ipv6_with_protocol", []string{"https://", "[2601:195:c381:3560::f42a]/", "/test"}, "https://[2601:195:c381:3560::f42a]/test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := URLJoin(tt.parts...); got != tt.want {
				t.Errorf("URLJoin(%q) = %q, want %q", tt.parts, got, tt.want)
			}
		})
	}
}

// returns an empty string if no arguments are supplied
func TestParityEmpty(t *testing.T) {
	if got := URLJoin(); got != "" {
		t.Errorf("URLJoin() = %q, want %q", got, "")
	}
}
