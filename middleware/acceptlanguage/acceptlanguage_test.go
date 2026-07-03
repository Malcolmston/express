package acceptlanguage

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func lang(t *testing.T, h express.Handler, header string) string {
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
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if header != "" {
		r.Header.Set("Accept-Language", header)
	}
	app.ServeHTTP(httptest.NewRecorder(), r)
	return got
}

func TestHighestQ(t *testing.T) {
	if got := lang(t, New(), "fr;q=0.5,en;q=0.9,de;q=0.1"); got != "en" {
		t.Fatalf("got %q", got)
	}
}

func TestNegotiateSupported(t *testing.T) {
	h := New(Options{Supported: []string{"en", "fr"}, Default: "en"})
	if got := lang(t, h, "de,fr;q=0.9"); got != "fr" {
		t.Fatalf("got %q", got)
	}
}

func TestPrimarySubtagMatch(t *testing.T) {
	h := New(Options{Supported: []string{"en", "fr"}, Default: "en"})
	if got := lang(t, h, "en-US,en;q=0.9"); got != "en" {
		t.Fatalf("got %q", got)
	}
}

func TestFallsBackToDefault(t *testing.T) {
	h := New(Options{Supported: []string{"en", "fr"}, Default: "en"})
	if got := lang(t, h, "de,es;q=0.9"); got != "en" {
		t.Fatalf("got %q", got)
	}
}

func TestEmptyHeaderDefault(t *testing.T) {
	h := New(Options{Supported: []string{"en"}, Default: "en"})
	if got := lang(t, h, ""); got != "en" {
		t.Fatalf("got %q", got)
	}
}
