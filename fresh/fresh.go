// Package fresh implements HTTP response freshness testing, a port of the npm
// "fresh" package. It determines whether a cached response is still fresh
// relative to a request's conditional headers, and is the same primitive that
// Express uses to answer 304 Not Modified requests.
//
// Use this package when a server holds a representation of a resource together
// with its validators (an ETag and/or a Last-Modified time) and needs to decide
// whether the client's cached copy is still current. When Fresh reports true,
// the caller can skip sending the body and respond with 304 Not Modified;
// when it reports false, the resource is considered stale and the full response
// should be sent. This mirrors the conditional-request handling described in
// RFC 7232.
//
// Freshness is evaluated from two request headers, If-None-Match and
// If-Modified-Since, checked against the response's ETag and Last-Modified
// headers. An unconditional request (neither If-None-Match nor
// If-Modified-Since present) is always treated as stale, because there is
// nothing to validate against. A Cache-Control: no-cache directive on the
// request forces revalidation and is likewise reported as stale, regardless of
// the validators.
//
// When both conditional headers are present, both must pass for the response to
// be fresh. If-None-Match succeeds when the response ETag is listed in the
// header, using a weak comparison in which a leading "W/" prefix is ignored; a
// bare "*" matches any ETag. A missing response ETag fails the check unless the
// request sent "*". If-Modified-Since succeeds when the response's
// Last-Modified time is not strictly after the requested time; a missing
// Last-Modified, or a date in either header that cannot be parsed in any of the
// standard HTTP date formats, fails the check.
//
// The behavior tracks the npm "fresh" module closely, including the "*"
// handling, weak ETag comparison, no-cache short-circuit, and the rule that a
// request with no conditional headers is stale. The signature is adapted to Go:
// rather than accepting two plain header maps by convention, Fresh takes two
// http.Header values, and nil headers are tolerated and treated as empty. Date
// parsing accepts the RFC 1123, RFC 1123 with numeric zone, RFC 850, and ANSI C
// asctime formats in addition to the preferred http.TimeFormat.
package fresh

import (
	"net/http"
	"strings"
	"time"
)

// Fresh reports whether a response is fresh given the request headers and the
// response headers.
//
// It returns true when the response's validators satisfy the request's
// conditional headers (If-None-Match / If-Modified-Since). If neither
// conditional header is present, or the request contains a Cache-Control
// no-cache directive, the response is considered stale (false).
func Fresh(reqHeaders, resHeaders http.Header) bool {
	if reqHeaders == nil {
		reqHeaders = http.Header{}
	}
	if resHeaders == nil {
		resHeaders = http.Header{}
	}

	modifiedSince := reqHeaders.Get("If-Modified-Since")
	noneMatch := reqHeaders.Get("If-None-Match")

	// Unconditional request: not fresh.
	if modifiedSince == "" && noneMatch == "" {
		return false
	}

	// Cache-Control: no-cache forces revalidation.
	cacheControl := reqHeaders.Get("Cache-Control")
	if cacheControl != "" && cacheControlNoCache(cacheControl) {
		return false
	}

	// If-None-Match takes precedence over If-Modified-Since. When present, it
	// alone determines freshness; If-Modified-Since is not consulted. Only a
	// lone "*" (the entire header value) is a wildcard that matches any ETag.
	if noneMatch != "" {
		if noneMatch == "*" {
			return true
		}
		etag := resHeaders.Get("ETag")
		if etag == "" {
			return false
		}
		return etagMatches(noneMatch, etag)
	}

	// If-Modified-Since
	if modifiedSince != "" {
		lastModified := resHeaders.Get("Last-Modified")
		if lastModified == "" {
			return false
		}
		lm, err1 := parseHTTPDate(lastModified)
		ms, err2 := parseHTTPDate(modifiedSince)
		if err1 != nil || err2 != nil {
			return false
		}
		if lm.After(ms) {
			return false
		}
	}

	return true
}

// cacheControlNoCache reports whether the Cache-Control value has a no-cache
// directive.
func cacheControlNoCache(value string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), "no-cache") {
			return true
		}
	}
	return false
}

// etagMatches reports whether the response ETag matches any tag in the
// If-None-Match token list. The comparison mirrors upstream jshttp/fresh: a
// tag matches when it equals the ETag exactly, or when either side is the weak
// ("W/") form of the other. A bare "*" inside a list is not a wildcard here;
// only a lone "*" header value (handled by the caller) matches any ETag.
func etagMatches(noneMatch, etag string) bool {
	for _, tag := range strings.Split(noneMatch, ",") {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if tag == etag || tag == "W/"+etag || "W/"+tag == etag {
			return true
		}
	}
	return false
}

// parseHTTPDate parses an HTTP date in any of the standard formats.
func parseHTTPDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	var err error
	for _, layout := range []string{http.TimeFormat, time.RFC1123, time.RFC1123Z, time.RFC850, time.ANSIC} {
		var t time.Time
		t, err = time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, err
}
