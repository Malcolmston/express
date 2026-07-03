package express

import (
	"net/http"
	"strings"
	"time"
)

// Fresh reports whether the client's cached response is still fresh for this
// request, based on the response's ETag / Last-Modified headers and the
// request's If-None-Match / If-Modified-Since headers. When true, a handler may
// send 304 Not Modified instead of a body. Fresh only applies to GET and HEAD.
func (req *Request) Fresh(res *Response) bool {
	if req.Method() != http.MethodGet && req.Method() != http.MethodHead {
		return false
	}

	inm := req.Get("If-None-Match")
	ims := req.Get("If-Modified-Since")
	if inm == "" && ims == "" {
		return false
	}

	// If-None-Match takes precedence over If-Modified-Since.
	if inm != "" {
		etag := res.GetHeader("ETag")
		if inm == "*" {
			return true
		}
		if etag == "" {
			return false
		}
		for _, tag := range strings.Split(inm, ",") {
			if etagMatch(strings.TrimSpace(tag), etag) {
				return true
			}
		}
		return false
	}

	// If-Modified-Since.
	lastMod := res.GetHeader("Last-Modified")
	if lastMod == "" {
		return false
	}
	since, err1 := http.ParseTime(ims)
	modified, err2 := http.ParseTime(lastMod)
	if err1 != nil || err2 != nil {
		return false
	}
	return !modified.After(since)
}

// Stale is the negation of Fresh.
func (req *Request) Stale(res *Response) bool { return !req.Fresh(res) }

// Fresh reports whether the current request is fresh against the headers set on
// this response (convenience wrapper around req.Fresh).
func (res *Response) Fresh() bool { return res.req.Fresh(res) }

// NotModified sends a 304 Not Modified with no body. Use it after setting ETag
// or Last-Modified when the request is fresh.
func (res *Response) NotModified() {
	res.Status(http.StatusNotModified)
	// A 304 must not include entity headers describing a body.
	h := res.Writer.Header()
	h.Del("Content-Type")
	h.Del("Content-Length")
	h.Del("Transfer-Encoding")
	res.End()
}

// ETag sets a (strong) ETag header from an already-computed tag. The value is
// quoted if not already.
func (res *Response) ETag(tag string) *Response {
	if !strings.HasPrefix(tag, "\"") && !strings.HasPrefix(tag, "W/") {
		tag = "\"" + tag + "\""
	}
	return res.Set("ETag", tag)
}

// LastModified sets the Last-Modified header from t.
func (res *Response) LastModified(t time.Time) *Response {
	return res.Set("Last-Modified", t.UTC().Format(http.TimeFormat))
}

// etagMatch compares two ETags, treating weak and strong forms as equal for
// If-None-Match (which uses the weak comparison function).
func etagMatch(a, b string) bool {
	a = strings.TrimPrefix(a, "W/")
	b = strings.TrimPrefix(b, "W/")
	return a == b
}
