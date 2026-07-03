package pagination

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func run(t *testing.T, url string, opts ...Options) Pagination {
	t.Helper()
	app := express.New()
	app.Use(New(opts...))
	var got Pagination
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = From(req)
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	return got
}

func TestDefaults(t *testing.T) {
	p := run(t, "/")
	if p.Page != 1 || p.Limit != 20 || p.Offset != 0 {
		t.Fatalf("got %+v", p)
	}
}

func TestParsedValues(t *testing.T) {
	p := run(t, "/?page=3&limit=10")
	if p.Page != 3 || p.Limit != 10 || p.Offset != 20 {
		t.Fatalf("got %+v", p)
	}
}

func TestBounds(t *testing.T) {
	p := run(t, "/?page=-5&limit=9999", Options{DefaultLimit: 25, MaxLimit: 50})
	if p.Page != 1 {
		t.Fatalf("page = %d, want 1", p.Page)
	}
	if p.Limit != 50 {
		t.Fatalf("limit = %d, want 50 (capped)", p.Limit)
	}
}

func TestInvalidLimitFallsBack(t *testing.T) {
	p := run(t, "/?limit=abc", Options{DefaultLimit: 15})
	if p.Limit != 15 {
		t.Fatalf("limit = %d, want 15", p.Limit)
	}
}
