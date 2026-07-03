package poweredby

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestPoweredByCustom(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Value: "MyApp/2.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("X-Powered-By"); got != "MyApp/2.0" {
		t.Fatalf("X-Powered-By = %q", got)
	}
}

func TestPoweredByDefault(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("X-Powered-By"); got != DefaultValue {
		t.Fatalf("X-Powered-By = %q", got)
	}
}
