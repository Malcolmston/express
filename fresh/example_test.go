package fresh_test

import (
	"fmt"
	"net/http"

	"github.com/malcolmston/express/fresh"
)

// ExampleFresh demonstrates a fresh response validated by ETag. The request
// carries an If-None-Match header, and the response advertises a matching ETag,
// so the client's cached copy is still current. When Fresh reports true, a
// server would typically answer 304 Not Modified and omit the body. The
// comparison ignores a leading "W/" weak-validator prefix. Here the tags are
// identical, so the response is fresh.
func ExampleFresh() {
	reqHeaders := http.Header{}
	reqHeaders.Set("If-None-Match", `"abc123"`)

	resHeaders := http.Header{}
	resHeaders.Set("ETag", `"abc123"`)

	fmt.Println(fresh.Fresh(reqHeaders, resHeaders))
	// Output: true
}

// ExampleFresh_stale shows a stale result caused by a mismatched ETag. The
// request's If-None-Match validator does not equal the response's current ETag,
// meaning the cached representation is out of date. Fresh returns false, and a
// server would send the full, updated response body. Any single non-matching
// validator is enough to make the response stale. This is the counterpart to
// the matching-ETag case.
func ExampleFresh_stale() {
	reqHeaders := http.Header{}
	reqHeaders.Set("If-None-Match", `"old"`)

	resHeaders := http.Header{}
	resHeaders.Set("ETag", `"new"`)

	fmt.Println(fresh.Fresh(reqHeaders, resHeaders))
	// Output: false
}

// ExampleFresh_noConditionalHeaders illustrates an unconditional request. When
// neither If-None-Match nor If-Modified-Since is present, there is nothing to
// validate the cached copy against. In that situation Fresh always reports
// false, treating the response as stale so the full body is sent. This is the
// default outcome for a plain GET without cache validators. It matches the
// behavior of the npm "fresh" module.
func ExampleFresh_noConditionalHeaders() {
	fmt.Println(fresh.Fresh(http.Header{}, http.Header{}))
	// Output: false
}

// ExampleFresh_noCache shows how a Cache-Control: no-cache directive forces
// revalidation. Even though the request's If-None-Match matches the response
// ETag, the no-cache directive instructs caches to treat the response as stale.
// Fresh therefore returns false, short-circuiting the validator checks. This
// lets a client demand a fresh response regardless of its stored validators. It
// mirrors the no-cache handling of the original library.
func ExampleFresh_noCache() {
	reqHeaders := http.Header{}
	reqHeaders.Set("If-None-Match", `"abc123"`)
	reqHeaders.Set("Cache-Control", "no-cache")

	resHeaders := http.Header{}
	resHeaders.Set("ETag", `"abc123"`)

	fmt.Println(fresh.Fresh(reqHeaders, resHeaders))
	// Output: false
}

// ExampleFresh_modifiedSince demonstrates freshness by modification date. The
// request's If-Modified-Since is later than the response's Last-Modified time,
// so the resource has not changed since the client last fetched it. Fresh
// reports true, allowing a 304 Not Modified response. Dates are parsed in the
// standard HTTP formats. If the Last-Modified time were strictly after the
// If-Modified-Since time, the response would instead be stale.
func ExampleFresh_modifiedSince() {
	reqHeaders := http.Header{}
	reqHeaders.Set("If-Modified-Since", "Sat, 01 Jan 2000 00:00:00 GMT")

	resHeaders := http.Header{}
	resHeaders.Set("Last-Modified", "Fri, 31 Dec 1999 00:00:00 GMT")

	fmt.Println(fresh.Fresh(reqHeaders, resHeaders))
	// Output: true
}
