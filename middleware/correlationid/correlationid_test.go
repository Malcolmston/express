package correlationid

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestCorrelationIDDefault(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if len(rr.Header().Get(DefaultHeader)) != 32 {
		t.Fatalf("correlation id = %q", rr.Header().Get(DefaultHeader))
	}
}

func TestCorrelationIDPreserved(t *testing.T) {
	app := express.New()
	app.Use(New())
	var stored any
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		stored, _ = req.Value(ContextKey)
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(DefaultHeader, "corr-42")
	app.ServeHTTP(rr, r)
	if rr.Header().Get(DefaultHeader) != "corr-42" || stored != "corr-42" {
		t.Fatalf("id not preserved: header=%q stored=%v", rr.Header().Get(DefaultHeader), stored)
	}
}
