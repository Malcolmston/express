package ipblocklist_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ipblocklist"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(ipblocklist.New(ipblocklist.Options{
		Block: []string{"1.2.3.4", "10.0.0.0/8"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	return app
}

func request(app *express.Application, ip string) int {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = ip + ":1234"
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w.Code
}

func TestBlockedExact(t *testing.T) {
	if code := request(newApp(), "1.2.3.4"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestBlockedCIDR(t *testing.T) {
	if code := request(newApp(), "10.5.5.5"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}

func TestAllowed(t *testing.T) {
	if code := request(newApp(), "8.8.8.8"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}
