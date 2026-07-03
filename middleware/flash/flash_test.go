package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestAddAndGet(t *testing.T) {
	app := express.New()
	app.Use(express.Session())
	app.Use(New())

	var got []Message
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		Add(req, "info", "hello")
		Add(req, "error", "boom")
		got = Get(req)
		// A second Get should now be empty (messages cleared).
		if again := Get(req); again != nil {
			t.Errorf("expected messages cleared, got %v", again)
		}
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if len(got) != 2 {
		t.Fatalf("got %d messages, want 2: %v", len(got), got)
	}
	if got[0] != (Message{"info", "hello"}) || got[1] != (Message{"error", "boom"}) {
		t.Fatalf("unexpected messages: %v", got)
	}
}

func TestNilSessionSafe(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		Add(req, "info", "x") // must not panic
		if got := Get(req); got != nil {
			t.Errorf("want nil, got %v", got)
		}
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
}
