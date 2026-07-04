package decompress

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

// TestDecompressClose reads and closes the wrapped body, exercising gzipBody.Close.
func TestDecompressClose(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	var closeErr error
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		got = string(data)
		closeErr = req.Raw.Body.Close()
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(gzipBytes(t, "hello world"))))
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got != "hello world" {
		t.Fatalf("decompressed = %q", got)
	}
	if closeErr != nil {
		t.Fatalf("Close error = %v", closeErr)
	}
}

// TestDecompressXGzip covers the x-gzip alias.
func TestDecompressXGzip(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		got = string(data)
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(gzipBytes(t, "xg"))))
	r.Header.Set("Content-Encoding", "x-gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got != "xg" {
		t.Fatalf("x-gzip decompressed = %q", got)
	}
}

// TestDecompressInvalidGzip covers the gzip.NewReader error branch.
func TestDecompressInvalidGzip(t *testing.T) {
	app := express.New()
	app.Use(New())
	reached := false
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		reached = true
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader("not gzip data"))
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if reached {
		t.Fatal("handler should not run on invalid gzip")
	}
	if w.Code != 500 {
		t.Fatalf("invalid gzip code = %d, want 500", w.Code)
	}
}

// TestDecompressNilBody covers the nil-body branch.
func TestDecompressNilBody(t *testing.T) {
	app := express.New()
	app.Use(New())
	reached := false
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		reached = true
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", nil)
	r.Body = nil
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if !reached {
		t.Fatal("nil-body request should pass through to handler")
	}
}

// TestDecompressHeaderClearedAndLength verifies side effects on gzip requests.
func TestDecompressHeaderCleared(t *testing.T) {
	app := express.New()
	app.Use(New())
	var enc string
	var length int64
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		enc = req.Get("Content-Encoding")
		length = req.Raw.ContentLength
		io.ReadAll(req.Raw.Body)
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(string(gzipBytes(t, "data"))))
	r.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if enc != "" {
		t.Fatalf("Content-Encoding should be cleared, got %q", enc)
	}
	if length != -1 {
		t.Fatalf("ContentLength = %d, want -1", length)
	}
}
