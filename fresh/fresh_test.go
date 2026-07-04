package fresh

import (
	"net/http"
	"testing"
)

func req(kv map[string]string) http.Header {
	h := http.Header{}
	for k, v := range kv {
		h.Set(k, v)
	}
	return h
}

func TestNoConditionalHeaders(t *testing.T) {
	if Fresh(http.Header{}, http.Header{}) {
		t.Fatal("no conditional headers should be stale")
	}
}

func TestMatchingETag(t *testing.T) {
	reqH := req(map[string]string{"If-None-Match": `"foo"`})
	resH := req(map[string]string{"ETag": `"foo"`})
	if !Fresh(reqH, resH) {
		t.Fatal("matching etag should be fresh")
	}
}

func TestNonMatchingETag(t *testing.T) {
	reqH := req(map[string]string{"If-None-Match": `"foo"`})
	resH := req(map[string]string{"ETag": `"bar"`})
	if Fresh(reqH, resH) {
		t.Fatal("non-matching etag should be stale")
	}
}

func TestETagListMatch(t *testing.T) {
	reqH := req(map[string]string{"If-None-Match": `"a", "b", "c"`})
	resH := req(map[string]string{"ETag": `"b"`})
	if !Fresh(reqH, resH) {
		t.Fatal("etag in list should be fresh")
	}
}

func TestWeakETagComparison(t *testing.T) {
	reqH := req(map[string]string{"If-None-Match": `W/"foo"`})
	resH := req(map[string]string{"ETag": `"foo"`})
	if !Fresh(reqH, resH) {
		t.Fatal("weak comparison should match")
	}
}

func TestStar(t *testing.T) {
	reqH := req(map[string]string{"If-None-Match": "*"})
	resH := req(map[string]string{"ETag": `"anything"`})
	if !Fresh(reqH, resH) {
		t.Fatal("* should be fresh")
	}
}

func TestModifiedSinceFresh(t *testing.T) {
	reqH := req(map[string]string{"If-Modified-Since": "Sat, 01 Jan 2000 00:00:00 GMT"})
	resH := req(map[string]string{"Last-Modified": "Fri, 31 Dec 1999 00:00:00 GMT"})
	if !Fresh(reqH, resH) {
		t.Fatal("last-modified before if-modified-since should be fresh")
	}
}

func TestModifiedSinceStale(t *testing.T) {
	reqH := req(map[string]string{"If-Modified-Since": "Fri, 31 Dec 1999 00:00:00 GMT"})
	resH := req(map[string]string{"Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT"})
	if Fresh(reqH, resH) {
		t.Fatal("last-modified after if-modified-since should be stale")
	}
}

func TestNoCache(t *testing.T) {
	reqH := req(map[string]string{
		"If-None-Match": `"foo"`,
		"Cache-Control": "no-cache",
	})
	resH := req(map[string]string{"ETag": `"foo"`})
	if Fresh(reqH, resH) {
		t.Fatal("no-cache should force stale")
	}
}

func TestBothConditionsMustPass(t *testing.T) {
	// etag matches but modified-since is stale -> not fresh
	reqH := req(map[string]string{
		"If-None-Match":     `"foo"`,
		"If-Modified-Since": "Fri, 31 Dec 1999 00:00:00 GMT",
	})
	resH := req(map[string]string{
		"ETag":          `"foo"`,
		"Last-Modified": "Sat, 01 Jan 2000 00:00:00 GMT",
	})
	if Fresh(reqH, resH) {
		t.Fatal("both conditions must pass")
	}
}
