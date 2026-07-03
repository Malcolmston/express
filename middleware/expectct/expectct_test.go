package expectct

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func do(t *testing.T, h express.Handler) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec.Header().Get("Expect-CT")
}

func TestDefault(t *testing.T) {
	if got := do(t, New()); got != "max-age=0" {
		t.Fatalf("got %q", got)
	}
}

func TestFull(t *testing.T) {
	want := `max-age=86400, enforce, report-uri="https://example.com/report"`
	got := do(t, New(Options{MaxAge: 86400, Enforce: true, ReportURI: "https://example.com/report"}))
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
