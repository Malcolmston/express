package useragentblock_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/useragentblock"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(useragentblock.New(useragentblock.Options{
		Block: []string{"BadBot", "curl"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, ua string) int {
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", ua)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestBlockedAgent(t *testing.T) {
	if code := run(newApp(), "Mozilla BadBot/1.0"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestBlockedCaseInsensitive(t *testing.T) {
	if code := run(newApp(), "CURL/7.0"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestAllowedAgent(t *testing.T) {
	if code := run(newApp(), "Mozilla/5.0 Firefox"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}
