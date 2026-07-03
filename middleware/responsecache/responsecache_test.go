package responsecache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
)

func TestCachesGet(t *testing.T) {
	app := express.New()
	app.Use(New(Options{TTL: time.Minute}))
	calls := 0
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		calls++
		res.Type("text").Send("hello")
	})

	do := func() *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		return rec
	}

	r1 := do()
	if r1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("first request X-Cache = %q, want MISS", r1.Header().Get("X-Cache"))
	}
	r2 := do()
	if r2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("second request X-Cache = %q, want HIT", r2.Header().Get("X-Cache"))
	}
	if calls != 1 {
		t.Fatalf("handler called %d times, want 1", calls)
	}
	if r2.Body.String() != "hello" {
		t.Fatalf("cached body = %q", r2.Body.String())
	}
	if r2.Header().Get("Content-Type") == "" {
		t.Fatalf("expected content-type to be replayed")
	}
}

func TestSkipsNonGet(t *testing.T) {
	app := express.New()
	app.Use(New())
	calls := 0
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		calls++
		res.Send("ok")
	})
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	}
	if calls != 2 {
		t.Fatalf("POST handler called %d times, want 2 (not cached)", calls)
	}
}

func TestSkipsNon200(t *testing.T) {
	app := express.New()
	app.Use(New())
	calls := 0
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		calls++
		res.Status(500).Send("err")
	})
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	}
	if calls != 2 {
		t.Fatalf("500 response should not be cached, calls = %d", calls)
	}
}
