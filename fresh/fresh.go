// Package fresh implements HTTP response freshness testing, a port of the npm
// "fresh" package. It determines whether a cached response is still fresh
// relative to a request's conditional headers.
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

	// If-None-Match
	if noneMatch != "" {
		if strings.TrimSpace(noneMatch) == "*" {
			// still need If-Modified-Since to also pass if present
		} else {
			etag := resHeaders.Get("ETag")
			if etag == "" {
				return false
			}
			if !etagMatches(noneMatch, etag) {
				return false
			}
		}
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

// etagMatches reports whether the ETag matches any tag in the If-None-Match
// list, using weak comparison (the "W/" prefix is ignored).
func etagMatches(noneMatch, etag string) bool {
	target := stripWeak(strings.TrimSpace(etag))
	for _, tag := range strings.Split(noneMatch, ",") {
		tag = stripWeak(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if tag == "*" || tag == target {
			return true
		}
	}
	return false
}

func stripWeak(tag string) string {
	if strings.HasPrefix(tag, "W/") {
		return tag[2:]
	}
	return tag
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
