package nonce

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestGeneratesNonce(t *testing.T) {
	app := express.New()
	app.Use(New())
	var fromReq, fromLocals string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fromReq = Nonce(req)
		fromLocals, _ = res.Locals[ContextKey].(string)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if fromReq == "" {
		t.Fatalf("expected a nonce on the request")
	}
	if fromReq != fromLocals {
		t.Fatalf("req nonce %q != locals nonce %q", fromReq, fromLocals)
	}
}

func TestNoncesDiffer(t *testing.T) {
	app := express.New()
	app.Use(New())
	seen := map[string]bool{}
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		seen[Nonce(req)] = true
		res.Send("ok")
	})
	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 distinct nonces, got %d", len(seen))
	}
}
