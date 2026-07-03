package abtest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestAssignsBucket(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = Bucket(req)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got != "A" && got != "B" {
		t.Fatalf("bucket = %q, want A or B", got)
	}
	if len(rec.Result().Cookies()) == 0 {
		t.Fatalf("expected an abtest cookie to be set")
	}
}

func TestStableAcrossRequests(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Buckets: []string{"A", "B", "C"}}))
	var got string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = Bucket(req)
		res.Send("ok")
	})

	// First request assigns and sets a cookie.
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	first := got
	cookies := rec.Result().Cookies()

	// Second request presents the cookie; bucket must be unchanged and no new
	// cookie should be issued.
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	app.ServeHTTP(rec2, req2)

	if got != first {
		t.Fatalf("bucket changed: %q -> %q", first, got)
	}
	if len(rec2.Result().Cookies()) != 0 {
		t.Fatalf("expected no new cookie on return visit")
	}
}
