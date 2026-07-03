package expires

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
)

func TestExpiresHeader(t *testing.T) {
	fixed := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	now = func() time.Time { return fixed }
	defer func() { now = time.Now }()

	app := express.New()
	app.Use(New(Options{Duration: time.Hour}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	got := rr.Header().Get("Expires")
	want := fixed.Add(time.Hour).Format(http.TimeFormat)
	if got != want {
		t.Fatalf("Expires = %q, want %q", got, want)
	}
}
