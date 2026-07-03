package featureflag

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestEnabled(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Flags: map[string]bool{"new-ui": true, "beta": false}}))

	var newUI, beta, missing bool
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		newUI = Enabled(req, "new-ui")
		beta = Enabled(req, "beta")
		missing = Enabled(req, "nope")
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !newUI {
		t.Errorf("new-ui should be enabled")
	}
	if beta {
		t.Errorf("beta should be disabled")
	}
	if missing {
		t.Errorf("unknown flag should be false")
	}
}

func TestEnabledWithoutMiddleware(t *testing.T) {
	app := express.New()
	var got bool
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = Enabled(req, "x")
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got {
		t.Fatalf("want false without middleware")
	}
}
