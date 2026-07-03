package hmacauth_test

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hmacauth"
)

var secret = []byte("shh")

func newApp() *express.Application {
	app := express.New()
	app.Use(hmacauth.New(hmacauth.Options{Secret: secret}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		// Confirm the body is still readable downstream.
		b, _ := io.ReadAll(req.Raw.Body)
		res.Send("got:" + string(b))
	})
	return app
}

func TestValidSignature(t *testing.T) {
	app := newApp()
	body := "hello world"
	sig := hmacauth.Sign(secret, []byte(body))
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("X-Signature", sig)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "got:"+body {
		t.Fatalf("expected 200 with body echoed, got %d %q", w.Code, w.Body.String())
	}
}

func TestInvalidSignature(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("POST", "/", strings.NewReader("hello world"))
	r.Header.Set("X-Signature", "deadbeef")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestMissingSignature(t *testing.T) {
	app := newApp()
	r := httptest.NewRequest("POST", "/", strings.NewReader("data"))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestTamperedBody(t *testing.T) {
	app := newApp()
	sig := hmacauth.Sign(secret, []byte("original"))
	r := httptest.NewRequest("POST", "/", strings.NewReader("tampered"))
	r.Header.Set("X-Signature", sig)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401 for tampered body, got %d", w.Code)
	}
}
