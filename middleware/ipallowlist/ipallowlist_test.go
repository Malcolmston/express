package ipallowlist_test

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/ipallowlist"
)

func newApp() *express.Application {
	app := express.New()
	app.Use(ipallowlist.New(ipallowlist.Options{
		Allow: []string{"10.0.0.5", "192.168.0.0/24"},
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

func TestAllowedExactIP(t *testing.T) {
	if code := request(newApp(), "10.0.0.5"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestAllowedCIDR(t *testing.T) {
	if code := request(newApp(), "192.168.0.42"); code != 200 {
		t.Fatalf("expected 200, got %d", code)
	}
}

func TestBlocked(t *testing.T) {
	if code := request(newApp(), "8.8.8.8"); code != 403 {
		t.Fatalf("expected 403, got %d", code)
	}
}
