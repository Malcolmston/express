package rawbody

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestRawBodyStoresAndRestores(t *testing.T) {
	app := express.New()
	app.Use(New())

	var stored []byte
	var reread string
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		if b, ok := req.Body().([]byte); ok {
			stored = b
		}
		data, _ := io.ReadAll(req.Raw.Body)
		reread = string(data)
		res.Send("ok")
	})

	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("hello world"))
	app.ServeHTTP(rr, r)

	if string(stored) != "hello world" {
		t.Fatalf("stored body = %q, want %q", stored, "hello world")
	}
	if reread != "hello world" {
		t.Fatalf("re-read body = %q, want %q", reread, "hello world")
	}
}

func TestRawBodyNilBody(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	app.ServeHTTP(rr, r)
	if rr.Body.String() != "ok" {
		t.Fatalf("body = %q", rr.Body.String())
	}
}
