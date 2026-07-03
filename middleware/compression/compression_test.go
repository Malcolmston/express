package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestCompressesLargeBody(t *testing.T) {
	payload := strings.Repeat("hello world ", 100)
	app := express.New()
	app.Use(New(Options{MinLength: 10}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(payload)
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip, deflate")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", rr.Header().Get("Content-Encoding"))
	}
	gr, err := gzip.NewReader(bytes.NewReader(rr.Body.Bytes()))
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	got, _ := io.ReadAll(gr)
	if string(got) != payload {
		t.Fatalf("decompressed body mismatch")
	}
}

func TestNoGzipWhenNotAccepted(t *testing.T) {
	app := express.New()
	app.Use(New(Options{MinLength: 1}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("some body content")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Header().Get("Content-Encoding") != "" {
		t.Fatalf("unexpected Content-Encoding")
	}
	if rr.Body.String() != "some body content" {
		t.Fatalf("body = %q", rr.Body.String())
	}
}

func TestNoGzipBelowMinLength(t *testing.T) {
	app := express.New()
	app.Use(New(Options{MinLength: 1024}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("tiny")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)
	if rr.Header().Get("Content-Encoding") != "" {
		t.Fatalf("should not compress below min length")
	}
	if rr.Body.String() != "tiny" {
		t.Fatalf("body = %q", rr.Body.String())
	}
}
