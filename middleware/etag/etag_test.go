package etag

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func expectedTag(body string) string {
	sum := sha1.Sum([]byte(body))
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

func TestETagSet(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	if got := rr.Header().Get("ETag"); got != expectedTag("hello") {
		t.Fatalf("ETag = %q, want %q", got, expectedTag("hello"))
	}
	if rr.Body.String() != "hello" {
		t.Fatalf("body = %q", rr.Body.String())
	}
}

func TestETagNotModified(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("If-None-Match", expectedTag("hello"))
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)

	if rr.Code != 304 {
		t.Fatalf("status = %d, want 304", rr.Code)
	}
	if rr.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rr.Body.String())
	}
}

func TestETagMismatchWritesBody(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("hello")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("If-None-Match", `"stale"`)
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)

	if rr.Code != 200 || rr.Body.String() != "hello" {
		t.Fatalf("code=%d body=%q", rr.Code, rr.Body.String())
	}
}
