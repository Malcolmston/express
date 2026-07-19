// Package parseurl parses the URL of an incoming HTTP request into its
// components, a stdlib-only Go port of the npm package "parseurl"
// (https://www.npmjs.com/package/parseurl). In the Node/Express world parseurl
// is the tiny utility that turns a request's raw target string (req.url) into a
// structured URL object with fields such as path, query and pathname, and it
// underpins routing, static file serving and any middleware that needs to reason
// about the request path independently of the query string.
//
// The reason parseurl exists at all in Node is performance and correctness of
// caching. Parsing a URL is not free, and a single request typically has its URL
// inspected by many layers of middleware, so the original library memoizes the
// parsed result on the request object and reparses only when the underlying
// req.url string has changed. It also distinguishes the current URL (which
// Express may rewrite as it strips mounted route prefixes) from the original,
// unmodified request target, exposing the latter through parseurl.original.
//
// This Go port maps those two concepts onto net/http and net/url. Parse returns
// the request's effective URL as a *url.URL, and OriginalURL returns the
// original request target; both read from req.RequestURI, the raw target line as
// received, and fall back to the request's already-parsed req.URL field when
// RequestURI is empty or unparseable. A nil request yields a nil result rather
// than panicking, mirroring parseurl's tolerance of a missing url. ParseString
// is a thin convenience wrapper over url.Parse for parsing an arbitrary URL
// string directly.
//
// Because Go's http.Request already carries a parsed URL and an immutable
// RequestURI, this port does not need parseurl's memoization machinery: parsing
// RequestURI with net/url is cheap and the raw target does not mutate underneath
// you the way req.url can in Express, so each call simply parses afresh and no
// per-request cache is stored. Parse first tries url.ParseRequestURI, which is
// the correct strict parser for an HTTP request target, and only falls back to
// the more lenient url.Parse (and finally to req.URL) if that fails, so
// well-formed absolute paths and absolute-form targets are both handled.
//
// The resulting *url.URL exposes the same information the Node object does, just
// under Go's field names: Path holds the decoded path, EscapedPath preserves the
// original percent-encoding, and RawQuery (or the parsed Query map) holds the
// query string. Parity with Node is therefore behavioral rather than literal —
// you get the request's path and query parsed the way net/url parses them, with
// the original request target available separately — which is exactly what the
// downstream routing and static-serving ports in this module consume.
package parseurl

import (
	"net/http"
	"net/url"
	"strings"
)

// Parse parses the request's effective URL, using req.RequestURI and falling
// back to req.URL when RequestURI is empty or cannot be parsed. It returns nil
// when req is nil.
func Parse(req *http.Request) *url.URL {
	if req == nil {
		return nil
	}
	raw := req.RequestURI
	if raw == "" {
		return req.URL
	}
	// Match upstream parseurl, which delegates to Node's url.parse for any
	// target containing whitespace or a fragment: url.parse trims surrounding
	// whitespace and splits off the #fragment so it is excluded from pathname,
	// query and search. Go's url.ParseRequestURI does neither (it treats "#"
	// and spaces as ordinary path/query bytes), so normalise those here. This
	// also matches url.ParseRequestURI's own documented assumption that the
	// input carries no "#fragment" suffix.
	raw = strings.TrimSpace(raw)
	frag := ""
	if i := strings.IndexByte(raw, '#'); i >= 0 {
		frag = raw[i+1:]
		raw = raw[:i]
	}
	if u, err := url.ParseRequestURI(raw); err == nil {
		u.Fragment = frag
		return u
	}
	if u, err := url.Parse(raw); err == nil {
		if frag != "" {
			u.Fragment = frag
		}
		return u
	}
	return req.URL
}

// OriginalURL parses the request's original URL. Like Parse, it uses
// req.RequestURI as the source of the original, unmodified request target.
func OriginalURL(req *http.Request) *url.URL {
	return Parse(req)
}

// ParseString parses a raw URL string into a *url.URL.
func ParseString(rawurl string) (*url.URL, error) {
	return url.Parse(rawurl)
}
