package ienoopen

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestNoOpen(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Download-Options"); got != "noopen" {
		t.Fatalf("got %q, want noopen", got)
	}
}
