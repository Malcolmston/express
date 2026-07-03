package subdomain

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func sub(t *testing.T, h express.Handler, host string) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	var got string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		if v, ok := req.Value(Key); ok {
			got, _ = v.(string)
		}
		res.Send("ok")
	})
	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "http://"+host+"/", nil))
	return got
}

func TestBaseHost(t *testing.T) {
	if got := sub(t, New(Options{BaseHost: "example.com"}), "api.example.com"); got != "api" {
		t.Fatalf("got %q", got)
	}
}

func TestMultiLevelBaseHost(t *testing.T) {
	if got := sub(t, New(Options{BaseHost: "example.com"}), "a.b.example.com"); got != "a.b" {
		t.Fatalf("got %q", got)
	}
}

func TestBareBaseHost(t *testing.T) {
	if got := sub(t, New(Options{BaseHost: "example.com"}), "example.com"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestHeuristic(t *testing.T) {
	if got := sub(t, New(), "shop.example.com"); got != "shop" {
		t.Fatalf("got %q", got)
	}
}

func TestHeuristicNoSub(t *testing.T) {
	if got := sub(t, New(), "example.com"); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}
