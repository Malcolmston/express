// Package parseurl parses the URL of an incoming HTTP request into its
// components.
//
// It is a port of the npm package "parseurl". The reference implementation
// memoizes the parsed URL on the request object; this port simply parses the
// request's RequestURI (falling back to the request's URL field) on each call.
package parseurl

import (
	"net/http"
	"net/url"
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
	if u, err := url.ParseRequestURI(raw); err == nil {
		return u
	}
	if u, err := url.Parse(raw); err == nil {
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
