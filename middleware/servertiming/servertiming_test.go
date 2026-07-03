package servertiming

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/malcolmston/express"
)

func TestServerTimingEmitted(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		m := From(req)
		m.Add("db", 12500*time.Microsecond) // 12.5ms
		m.AddWithDesc("render", "template render", 3*time.Millisecond)
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	got := rr.Header().Get(HeaderName)
	if !strings.Contains(got, "db;dur=12.50") {
		t.Fatalf("Server-Timing = %q, missing db metric", got)
	}
	if !strings.Contains(got, `render;desc="template render";dur=3.00`) {
		t.Fatalf("Server-Timing = %q, missing render metric", got)
	}
}

func TestServerTimingNoMetrics(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get(HeaderName); got != "" {
		t.Fatalf("Server-Timing = %q, want empty", got)
	}
}

func TestFromWithoutMiddleware(t *testing.T) {
	// From must not panic when middleware isn't installed.
	app := express.New()
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		From(req).Add("x", time.Millisecond)
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if rr.Code != 200 {
		t.Fatalf("code = %d", rr.Code)
	}
}
