package rewrite

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/malcolmston/express"
)

func pathAfter(t *testing.T, h express.Handler, url string) string {
	t.Helper()
	app := express.New()
	app.Use(h)
	var got string
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		got = req.Raw.URL.Path
		res.Send("ok")
	})
	app.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, url, nil))
	return got
}

func TestPatternRewrite(t *testing.T) {
	h := New(Options{Rules: []Rule{{Pattern: `^/old/(.*)$`, To: "/new/$1"}}})
	if got := pathAfter(t, h, "/old/thing"); got != "/new/thing" {
		t.Fatalf("got %q", got)
	}
}

func TestCompiledRegexp(t *testing.T) {
	h := New(Options{Rules: []Rule{{From: regexp.MustCompile(`^/a/(\d+)$`), To: "/b/$1"}}})
	if got := pathAfter(t, h, "/a/42"); got != "/b/42" {
		t.Fatalf("got %q", got)
	}
}

func TestNoMatch(t *testing.T) {
	h := New(Options{Rules: []Rule{{Pattern: `^/x$`, To: "/y"}}})
	if got := pathAfter(t, h, "/keep"); got != "/keep" {
		t.Fatalf("got %q", got)
	}
}

func TestFirstMatchWins(t *testing.T) {
	h := New(Options{Rules: []Rule{
		{Pattern: `^/p/(.*)$`, To: "/first/$1"},
		{Pattern: `^/p/(.*)$`, To: "/second/$1"},
	}})
	if got := pathAfter(t, h, "/p/z"); got != "/first/z" {
		t.Fatalf("got %q", got)
	}
}

func TestBadPatternSkipped(t *testing.T) {
	h := New(Options{Rules: []Rule{{Pattern: `(`, To: "/x"}}})
	if got := pathAfter(t, h, "/keep"); got != "/keep" {
		t.Fatalf("got %q", got)
	}
}
