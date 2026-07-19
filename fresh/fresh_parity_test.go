package fresh

import (
	"net/http"
	"testing"
)

// Upstream parity tests transcribed verbatim from the jshttp/fresh test suite.
//
// Source (fetched 2026-07-19):
//   https://raw.githubusercontent.com/jshttp/fresh/master/test/fresh.js
//   https://raw.githubusercontent.com/jshttp/fresh/master/index.js
//
// Each vector uses the exact request/response header values and expected
// boolean from upstream's describe/it blocks. Upstream uses lowercased header
// keys against plain objects; here we build http.Header via Set, which is the
// port's public contract.

func parityHeader(kv map[string]string) http.Header {
	h := http.Header{}
	for k, v := range kv {
		h.Set(k, v)
	}
	return h
}

func TestParityUpstream(t *testing.T) {
	cases := []struct {
		name string
		req  map[string]string
		res  map[string]string
		want bool
	}{
		// non-conditional GET
		{"non-conditional get is stale", map[string]string{}, map[string]string{}, false},

		// If-None-Match
		{"etags match is fresh", map[string]string{"If-None-Match": `"foo"`}, map[string]string{"ETag": `"foo"`}, true},
		{"etags mismatch is stale", map[string]string{"If-None-Match": `"foo"`}, map[string]string{"ETag": `"bar"`}, false},
		{"at least one matches is fresh", map[string]string{"If-None-Match": ` "bar" , "foo"`}, map[string]string{"ETag": `"foo"`}, true},
		{"list separated with tabs is fresh", map[string]string{"If-None-Match": "\"bar\",\t\"foo\""}, map[string]string{"ETag": `"foo"`}, true},
		{"etag missing is stale", map[string]string{"If-None-Match": `"foo"`}, map[string]string{}, false},

		// weak / strong
		{"weak etag exact match is fresh", map[string]string{"If-None-Match": `W/"foo"`}, map[string]string{"ETag": `W/"foo"`}, true},
		{"weak req strong res is fresh", map[string]string{"If-None-Match": `W/"foo"`}, map[string]string{"ETag": `"foo"`}, true},
		{"strong etag exact match is fresh", map[string]string{"If-None-Match": `"foo"`}, map[string]string{"ETag": `"foo"`}, true},
		{"strong req weak res is fresh", map[string]string{"If-None-Match": `"foo"`}, map[string]string{"ETag": `W/"foo"`}, true},

		// star
		{"star alone is fresh", map[string]string{"If-None-Match": "*"}, map[string]string{"ETag": `"foo"`}, true},
		{"star ignored when not only value", map[string]string{"If-None-Match": `*, "bar"`}, map[string]string{"ETag": `"foo"`}, false},

		// If-Modified-Since
		{"modified since date is stale", map[string]string{"If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"}, map[string]string{"Last-Modified": "Sat, 01 Jan 2000 01:00:00 GMT"}, false},
		{"unmodified since date is fresh", map[string]string{"If-Modified-Since": "Sat, 01 Jan 2000 01:00:00 GMT"}, map[string]string{"Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"}, true},
		{"last-modified missing is stale", map[string]string{"If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"}, map[string]string{}, false},
		{"invalid if-modified-since is stale", map[string]string{"If-Modified-Since": "foo"}, map[string]string{"Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"}, false},
		{"invalid last-modified is stale", map[string]string{"If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"}, map[string]string{"Last-Modified": "foo"}, false},

		// combined If-None-Match and If-Modified-Since
		{"both match is fresh", map[string]string{"If-None-Match": `"foo"`, "If-Modified-Since": "Sat, 01 Jan 2000 01:00:00 GMT"}, map[string]string{"ETag": `"foo"`, "Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"}, true},
		{"only etag matches is fresh", map[string]string{"If-None-Match": `"foo"`, "If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"}, map[string]string{"ETag": `"foo"`, "Last-Modified": "Sat, 01 Jan 2000 01:00:00 GMT"}, true},
		{"only last-modified matches is stale", map[string]string{"If-None-Match": `"foo"`, "If-Modified-Since": "Sat, 01 Jan 2000 01:00:00 GMT"}, map[string]string{"ETag": `"bar"`, "Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"}, false},
		{"none match is stale", map[string]string{"If-None-Match": `"foo"`, "If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"}, map[string]string{"ETag": `"bar"`, "Last-Modified": "Sat, 01 Jan 2000 01:00:00 GMT"}, false},

		// Cache-Control: no-cache
		{"no-cache alone is stale", map[string]string{"Cache-Control": " no-cache"}, map[string]string{}, false},
		{"no-cache with matching etag is stale", map[string]string{"Cache-Control": " no-cache", "If-None-Match": `"foo"`}, map[string]string{"ETag": `"foo"`}, false},
		{"no-cache with unmodified date is stale", map[string]string{"Cache-Control": " no-cache", "If-Modified-Since": "Sat, 01 Jan 2000 01:00:00 GMT"}, map[string]string{"Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Fresh(parityHeader(tc.req), parityHeader(tc.res))
			if got != tc.want {
				t.Fatalf("Fresh(%v, %v) = %v, want %v", tc.req, tc.res, got, tc.want)
			}
		})
	}
}
