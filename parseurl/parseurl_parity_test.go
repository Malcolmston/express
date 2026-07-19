package parseurl

import (
	"net/http"
	"testing"
)

// Upstream-parity tests for the pillarjs/parseurl port.
//
// Every input -> expected-output vector below is transcribed from the real
// upstream test suite, not invented:
//
//	https://raw.githubusercontent.com/pillarjs/parseurl/master/test/test.js
//
// Behavioral reference (fastparse semantics) from:
//
//	https://raw.githubusercontent.com/pillarjs/parseurl/master/index.js
//
// Upstream returns an object with {pathname, query, search, href, host,
// hostname, port}. This port returns a *url.URL, so the fields map as:
//
//	pathname -> u.Path
//	query    -> u.RawQuery              (upstream null  <=> "")
//	search   -> "?"+u.RawQuery or ""    (upstream null  <=> "")
//	href     -> u.String()
//	host     -> u.Host                  (upstream empty <=> "")
//	hostname -> u.Hostname()
//	port     -> u.Port()
//
// Because net/http exposes a single request target (RequestURI) with no
// separate originalUrl, the port's OriginalURL delegates to Parse. Upstream
// vectors that assert on a distinct originalUrl value therefore have no
// representable analog and are covered only where originalUrl == url.

// search derives the upstream "search" field from a parsed query string.
func search(rawQuery string) string {
	if rawQuery == "" {
		return ""
	}
	return "?" + rawQuery
}

// reqWith builds a request whose effective target is the given raw string,
// mirroring upstream's createReq(url) where req.url is parsed.
func reqWith(target string) *http.Request {
	return &http.Request{RequestURI: target}
}

type parityCase struct {
	name     string
	target   string
	pathname string
	query    string // "" == upstream null
	href     string // "" == skip href assertion (e.g. auth-looking path)
	host     string
	hostname string
	port     string
}

// parityCases mirrors the assertions in describe('parseurl(req)') of test.js.
var parityCases = []parityCase{
	{
		name:     "parse the request URL",
		target:   "/foo/bar",
		pathname: "/foo/bar",
		href:     "/foo/bar",
	},
	{
		name:     "parse with query string",
		target:   "/foo/bar?fizz=buzz",
		pathname: "/foo/bar",
		query:    "fizz=buzz",
		href:     "/foo/bar?fizz=buzz",
	},
	{
		name:     "parse with hash",
		target:   "/foo/bar#bazz",
		pathname: "/foo/bar",
		href:     "/foo/bar#bazz",
	},
	{
		name:     "parse with query string and hash",
		target:   "/foo/bar?fizz=buzz#bazz",
		pathname: "/foo/bar",
		query:    "fizz=buzz",
		href:     "/foo/bar?fizz=buzz#bazz",
	},
	{
		name:     "parse a full URL",
		target:   "http://localhost:8888/foo/bar",
		pathname: "/foo/bar",
		href:     "http://localhost:8888/foo/bar",
		host:     "localhost:8888",
		hostname: "localhost",
		port:     "8888",
	},
	{
		// Upstream only asserts pathname here; url.parse would misread the
		// leading "//" as an authority, so both libraries keep it as a path.
		name:     "not choke on auth-looking URL",
		target:   "//todo@txt",
		pathname: "//todo@txt",
	},
	{
		// Upstream trims surrounding whitespace via url.parse; upstream only
		// asserts pathname for this caching vector.
		name:     "trailing whitespace trimmed",
		target:   "/foo/bar ",
		pathname: "/foo/bar",
	},
}

func TestParityParse(t *testing.T) {
	for _, tc := range parityCases {
		t.Run(tc.name, func(t *testing.T) {
			u := Parse(reqWith(tc.target))
			if u == nil {
				t.Fatalf("Parse(%q) returned nil", tc.target)
			}
			if u.Path != tc.pathname {
				t.Errorf("pathname = %q, want %q", u.Path, tc.pathname)
			}
			if u.RawQuery != tc.query {
				t.Errorf("query = %q, want %q", u.RawQuery, tc.query)
			}
			if got := search(u.RawQuery); got != search(tc.query) {
				t.Errorf("search = %q, want %q", got, search(tc.query))
			}
			if tc.href != "" {
				if got := u.String(); got != tc.href {
					t.Errorf("href = %q, want %q", got, tc.href)
				}
			}
			if u.Host != tc.host {
				t.Errorf("host = %q, want %q", u.Host, tc.host)
			}
			if got := u.Hostname(); got != tc.hostname {
				t.Errorf("hostname = %q, want %q", got, tc.hostname)
			}
			if got := u.Port(); got != tc.port {
				t.Errorf("port = %q, want %q", got, tc.port)
			}
		})
	}
}

// TestParityMissingURL mirrors 'should return undefined missing url': with no
// target the port returns req.URL (here nil), the closest analog to undefined.
func TestParityMissingURL(t *testing.T) {
	if u := Parse(&http.Request{}); u != nil {
		t.Fatalf("Parse with empty target = %v, want nil", u)
	}
}

// TestParityMultipleTimes mirrors 'should parse multiple times': repeated
// parses of the same request yield a stable pathname.
func TestParityMultipleTimes(t *testing.T) {
	req := reqWith("/foo/bar")
	for i := 0; i < 3; i++ {
		if u := Parse(req); u == nil || u.Path != "/foo/bar" {
			t.Fatalf("parse #%d = %v, want pathname /foo/bar", i, u)
		}
	}
}

// TestParityOriginalFallback mirrors 'should parse req.url when originalUrl
// missing': OriginalURL falls back to the request target. In the Go port
// OriginalURL always mirrors Parse because http.Request has no separate
// originalUrl field.
func TestParityOriginalFallback(t *testing.T) {
	u := OriginalURL(reqWith("/foo/bar"))
	if u == nil || u.Path != "/foo/bar" || u.String() != "/foo/bar" {
		t.Fatalf("OriginalURL = %v, want pathname/href /foo/bar", u)
	}
}

// TestParityOriginalMissing mirrors 'should return undefined missing req.url
// and originalUrl'.
func TestParityOriginalMissing(t *testing.T) {
	if u := OriginalURL(&http.Request{}); u != nil {
		t.Fatalf("OriginalURL with empty target = %v, want nil", u)
	}
}
