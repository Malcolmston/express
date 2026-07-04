package rawbody

import (
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

// errReadCloser returns an error on Read to exercise the ReadAll failure path.
type errReadCloser struct{}

func (errReadCloser) Read(p []byte) (int, error) { return 0, errors.New("boom") }
func (errReadCloser) Close() error               { return nil }

func TestRawBodyReadError(t *testing.T) {
	app := express.New()
	app.Use(New())
	reached := false
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		reached = true
		res.Send("ok")
	})

	r := httptest.NewRequest("POST", "/", strings.NewReader("ignored"))
	r.Body = errReadCloser{}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	if reached {
		t.Fatal("handler should not run when body read fails")
	}
	if w.Code != 500 {
		t.Fatalf("read error code = %d, want 500", w.Code)
	}
}

func TestRawBodyNilRawBody(t *testing.T) {
	app := express.New()
	app.Use(New())
	var stored []byte
	var ok bool
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		stored, ok = req.Body().([]byte)
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Body = nil
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if !ok || len(stored) != 0 {
		t.Fatalf("nil body should store empty []byte, got %v ok=%v", stored, ok)
	}
}

func TestRawBodyEmptyBodyRestores(t *testing.T) {
	app := express.New()
	app.Use(New())
	var stored []byte
	var reread string
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		stored, _ = req.Body().([]byte)
		data, _ := io.ReadAll(req.Raw.Body)
		reread = string(data)
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if len(stored) != 0 {
		t.Fatalf("stored = %q, want empty", stored)
	}
	if reread != "" {
		t.Fatalf("reread = %q, want empty", reread)
	}
}
