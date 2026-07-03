package querynormalize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func normalized(t *testing.T, url string) string {
	t.Helper()
	app := express.New()
	app.Use(New())
	var got string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		got = req.Raw.URL.RawQuery
		res.Send("ok")
	})
	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, url, nil))
	return got
}

func TestLowercaseKeysSorted(t *testing.T) {
	if got := normalized(t, "/?B=2&A=1"); got != "a=1&b=2" {
		t.Fatalf("got %q", got)
	}
}

func TestTrimsValues(t *testing.T) {
	// %20 is a space; trimming should remove it.
	if got := normalized(t, "/?name=%20bob%20"); got != "name=bob" {
		t.Fatalf("got %q", got)
	}
}

func TestNoQuery(t *testing.T) {
	if got := normalized(t, "/plain"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestDeterministic(t *testing.T) {
	a := normalized(t, "/?z=1&y=2&x=3")
	b := normalized(t, "/?x=3&y=2&z=1")
	if a != b {
		t.Fatalf("not deterministic: %q vs %q", a, b)
	}
	if a != "x=3&y=2&z=1" {
		t.Fatalf("got %q", a)
	}
}
