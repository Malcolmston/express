package apikey_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/apikey"
)

func appWithKeys() *express.Application {
	app := express.New()
	app.Use(apikey.New(apikey.Options{
		Query: "api_key",
		Keys:  []string{"key1", "key2"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func TestValidHeaderKey(t *testing.T) {
	app := appWithKeys()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", "key2")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestValidQueryKey(t *testing.T) {
	app := appWithKeys()
	r := httptest.NewRequest("GET", "/?api_key=key1", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMissingKey(t *testing.T) {
	app := appWithKeys()
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestInvalidKey(t *testing.T) {
	app := appWithKeys()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-API-Key", "nope")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestVerifyFunc(t *testing.T) {
	app := express.New()
	app.Use(apikey.New(apikey.Options{
		Header: "X-Token",
		Verify: func(key string) bool { return key == "abc" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Token", "abc")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
