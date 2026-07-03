package requestcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestAttachesCtxAndHeader(t *testing.T) {
	app := express.New()
	app.Use(New())
	var id string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		c := From(req)
		if c == nil {
			t.Fatal("no ctx attached")
		}
		id = c.ID
		if c.Start.IsZero() {
			t.Fatal("start not set")
		}
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if id == "" {
		t.Fatal("empty id")
	}
	if got := rec.Header().Get(HeaderName); got != id {
		t.Fatalf("header %q != ctx id %q", got, id)
	}
}

func TestUniqueIDs(t *testing.T) {
	h := New()
	app := express.New()
	app.Use(h)
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(From(req).ID)
	})
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		id := rec.Body.String()
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
}

func TestTrustHeader(t *testing.T) {
	app := express.New()
	app.Use(New(Options{TrustHeader: true}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(From(req).ID)
	})
	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderName, "abc-123")
	app.ServeHTTP(rec, r)
	if rec.Body.String() != "abc-123" {
		t.Fatalf("got %q", rec.Body.String())
	}
}

func TestCustomGenerator(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Generator: func() string { return "fixed" }}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send(From(req).ID)
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Body.String() != "fixed" {
		t.Fatalf("got %q", rec.Body.String())
	}
}
