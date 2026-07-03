package vary

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestVaryAppends(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Fields: []string{"Accept-Encoding", "Origin"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	vals := rr.Header().Values("Vary")
	if len(vals) != 2 || vals[0] != "Accept-Encoding" || vals[1] != "Origin" {
		t.Fatalf("Vary = %v", vals)
	}
}

func TestVaryNoDuplicate(t *testing.T) {
	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Vary", "Accept-Encoding")
		next()
	})
	app.Use(New(Options{Fields: []string{"accept-encoding", "Origin"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	vals := rr.Header().Values("Vary")
	// Accept-Encoding present once, Origin appended.
	if len(vals) != 2 {
		t.Fatalf("Vary = %v, want 2 entries", vals)
	}
}
