package requireauth_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/requireauth"
)

func TestAuthenticated(t *testing.T) {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("user", "ada")
		next()
	})
	app.Use(requireauth.New(requireauth.Options{}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUnauthenticated(t *testing.T) {
	app := express.New()
	app.Use(requireauth.New(requireauth.Options{}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCustomKey(t *testing.T) {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("account", 7)
		next()
	})
	app.Use(requireauth.New(requireauth.Options{Key: "account"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
