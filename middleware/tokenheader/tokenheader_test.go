package tokenheader_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/tokenheader"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(tokenheader.New(tokenheader.Options{
		Header: "X-Session-Token",
		Verify: func(token string) bool { return token == "s3cr3t" },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func run(app *express.Application, token string) int {
	r := httptest.NewRequest("GET", "/", nil)
	if token != "" {
		r.Header.Set("X-Session-Token", token)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestValidToken(t *testing.T) {
	if code := run(newApp(), "s3cr3t"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestInvalidToken(t *testing.T) {
	if code := run(newApp(), "wrong"); code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
}

func TestMissingToken(t *testing.T) {
	if code := run(newApp(), ""); code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
}
