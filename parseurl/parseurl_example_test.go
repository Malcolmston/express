package parseurl_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express/parseurl"
)

// ExampleParse builds an HTTP request with a path and query string and parses
// its effective URL into structured components. Parse reads the request's raw
// target and returns a *url.URL, from which the decoded path and the raw query
// string can be read directly. The query can also be decoded into individual
// parameters via the returned URL's Query method. This mirrors how Node's
// parseurl turns req.url into a structured object for routing and middleware.
func ExampleParse() {
	req := httptest.NewRequest("GET", "/foo/bar?baz=1&qux=2", nil)

	u := parseurl.Parse(req)
	fmt.Println("path:", u.Path)
	fmt.Println("query:", u.RawQuery)
	fmt.Println("baz:", u.Query().Get("baz"))
	// Output:
	// path: /foo/bar
	// query: baz=1&qux=2
	// baz: 1
}

// ExampleParse_encoded shows that percent-encoded path segments are handled
// correctly. The request target contains a space encoded as %20 in both the
// path and a query value. Parse returns a URL whose Path field holds the
// decoded value while EscapedPath preserves the original encoding. This lets
// callers compare against decoded paths yet still reconstruct the exact
// on-the-wire form when needed.
func ExampleParse_encoded() {
	req := httptest.NewRequest("GET", "/foo%20bar?a=b%20c", nil)

	u := parseurl.Parse(req)
	fmt.Println("decoded path:", u.Path)
	fmt.Println("escaped path:", u.EscapedPath())
	// Output:
	// decoded path: /foo bar
	// escaped path: /foo%20bar
}

// ExampleParseString parses an arbitrary absolute URL string directly, without
// needing an *http.Request. It is a thin convenience wrapper over the standard
// library's url.Parse. The returned URL exposes the host, path and query as
// separate fields. This is handy when you already hold a URL string rather
// than a request, for example when parsing a configured upstream target.
func ExampleParseString() {
	u, err := parseurl.ParseString("http://example.com/a/b?c=d")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("host:", u.Host)
	fmt.Println("path:", u.Path)
	fmt.Println("query:", u.RawQuery)
	// Output:
	// host: example.com
	// path: /a/b
	// query: c=d
}
