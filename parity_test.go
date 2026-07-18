package express

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// This file encodes known-answer vectors taken directly from the upstream
// Express.js test suite (expressjs/express, test/res.*.js and test/req.*.js) and
// asserts them against this Go port's public API. Each TestParity* function
// names the upstream describe/it it mirrors. Where the port makes a deliberate,
// documented behavioural choice that differs from a specific upstream vector
// (noted inline), that vector is intentionally omitted rather than asserted
// against the wrong expectation.

// run drives a single request through app and returns the recorder.
func runParity(app *Application, method, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	app.ServeHTTP(w, httptest.NewRequest(method, target, nil))
	return w
}

// TestParityResStatus mirrors test/res.status.js: valid codes in [100,999] are
// written verbatim; codes outside that range surface as a 500 carrying an
// "Invalid status code" message (Express raises a RangeError handled as 500).
func TestParityResStatus(t *testing.T) {
	valid := []int{101, 201, 302, 403, 501, 700, 800, 900}
	for _, code := range valid {
		app := New()
		c := code
		app.Use(func(req *Request, res *Response, next Next) { res.Status(c).End() })
		if w := runParity(app, "GET", "/"); w.Code != c {
			t.Errorf("Status(%d): got %d", c, w.Code)
		}
	}

	invalid := []int{99, 1000}
	for _, code := range invalid {
		app := New()
		c := code
		app.Use(func(req *Request, res *Response, next Next) { res.Status(c).End() })
		w := runParity(app, "GET", "/")
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status(%d): got %d, want 500", c, w.Code)
		}
		if body := w.Body.String(); !contains(body, "Invalid status code") {
			t.Errorf("Status(%d): body %q missing 'Invalid status code'", c, body)
		}
	}
}

// TestParityResType mirrors test/res.type.js: res.type resolves a filename or
// bare extension to a Content-Type via the mime database, defaulting to
// application/octet-stream for an unknown extension. The one documented
// divergence is the .js/.mjs family: this port pins mime-types v2 semantics
// (application/javascript) rather than upstream master's text/javascript, to
// stay consistent with the port's mimetypes package and existing tests.
func TestParityResType(t *testing.T) {
	cases := []struct{ in, want string }{
		{"foo.js", "application/javascript; charset=utf-8"}, // upstream v3: text/javascript
		{".json", "application/json; charset=utf-8"},
		{"file.tar.gz", "application/gzip"},
		{"FILE.JSON", "application/json; charset=utf-8"},
		{"file@test.json", "application/json; charset=utf-8"},
		{"application/vnd.amazon.ebook", "application/vnd.amazon.ebook"},
	}
	for _, c := range cases {
		app := New()
		in := c.in
		app.Use(func(req *Request, res *Response, next Next) { res.Type(in).End() })
		w := runParity(app, "GET", "/")
		if got := w.Header().Get("Content-Type"); got != c.want {
			t.Errorf("type(%q): got %q, want %q", in, got, c.want)
		}
	}
}

// TestParityResJSON mirrors test/res.json.js: primitives, arrays and objects
// serialise as JSON with an application/json; charset=utf-8 Content-Type.
func TestParityResJSON(t *testing.T) {
	cases := []struct {
		name string
		val  any
		want string
	}{
		{"null", nil, "null"},
		{"number", 300, "300"},
		{"string", "str", `"str"`},
		{"array", []string{"foo", "bar", "baz"}, `["foo","bar","baz"]`},
		{"object", map[string]string{"foo": "bar"}, `{"foo":"bar"}`},
	}
	for _, c := range cases {
		app := New()
		v := c.val
		app.Use(func(req *Request, res *Response, next Next) { res.JSON(v) })
		w := runParity(app, "GET", "/")
		if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
			t.Errorf("json(%s): Content-Type %q", c.name, ct)
		}
		if got := w.Body.String(); got != c.want {
			t.Errorf("json(%s): got %q, want %q", c.name, got, c.want)
		}
	}
}

// TestParityResSend mirrors test/res.send.js: a string body defaults to
// text/html; charset=utf-8, a []byte body to application/octet-stream, and a
// numeric value serialises as JSON.
func TestParityResSend(t *testing.T) {
	appHTML := New()
	appHTML.Use(func(req *Request, res *Response, next Next) { res.Send("<p>hey</p>") })
	if w := runParity(appHTML, "GET", "/"); w.Header().Get("Content-Type") != "text/html; charset=utf-8" || w.Body.String() != "<p>hey</p>" {
		t.Errorf("send(string): ct=%q body=%q", w.Header().Get("Content-Type"), w.Body.String())
	}

	appBuf := New()
	appBuf.Use(func(req *Request, res *Response, next Next) { res.Send([]byte("hello")) })
	if w := runParity(appBuf, "GET", "/"); w.Header().Get("Content-Type") != "application/octet-stream" || w.Body.String() != "hello" {
		t.Errorf("send([]byte): ct=%q body=%q", w.Header().Get("Content-Type"), w.Body.String())
	}

	appNum := New()
	appNum.Use(func(req *Request, res *Response, next Next) { res.Send(1000) })
	if w := runParity(appNum, "GET", "/"); w.Header().Get("Content-Type") != "application/json; charset=utf-8" || w.Body.String() != "1000" {
		t.Errorf("send(number): ct=%q body=%q", w.Header().Get("Content-Type"), w.Body.String())
	}
}

// TestParityResLocation mirrors test/res.location.js: the Location header is
// percent-encoded via encodeurl, leaving already-safe URLs untouched.
func TestParityResLocation(t *testing.T) {
	cases := []struct{ in, want string }{
		{"http://google.com/", "http://google.com/"},
		{"http://google.com", "http://google.com"},
		{"https://google.com?q=☃ §10", "https://google.com?q=%E2%98%83%20%C2%A710"},
		{"data:text/javascript,export default () => { }", "data:text/javascript,export%20default%20()%20=%3E%20%7B%20%7D"},
		{"", ""},
	}
	for _, c := range cases {
		app := New()
		in := c.in
		app.Use(func(req *Request, res *Response, next Next) { res.Location(in).End() })
		w := runParity(app, "GET", "/")
		if got := w.Header().Get("Location"); got != c.want {
			t.Errorf("location(%q): got %q, want %q", in, got, c.want)
		}
	}
}

// TestParityReqPath mirrors test/req.path.js: req.Path returns the parsed
// pathname without the query string.
func TestParityReqPath(t *testing.T) {
	app := New()
	var got string
	app.Use(func(req *Request, res *Response, next Next) {
		got = req.Path()
		res.End()
	})
	runParity(app, "GET", "/login?redirect=/post/1/comments")
	if got != "/login" {
		t.Errorf("req.Path() = %q, want /login", got)
	}
}

// TestParityReqQuery mirrors test/req.query.js: simple keys are parsed and the
// default is an empty set. (The port exposes query values via QueryValues
// rather than reflecting the whole query object, so vectors are asserted per
// key.)
func TestParityReqQuery(t *testing.T) {
	app := New()
	var name string
	var empty bool
	app.Get("/simple", func(req *Request, res *Response, next Next) {
		name = req.QueryValues().Get("user[name]")
		res.End()
	})
	app.Get("/empty", func(req *Request, res *Response, next Next) {
		empty = len(req.QueryValues()) == 0
		res.End()
	})
	runParity(app, "GET", "/simple?user[name]=tj")
	if name != "tj" {
		t.Errorf("query user[name] = %q, want tj", name)
	}
	runParity(app, "GET", "/empty")
	if !empty {
		t.Errorf("empty query should parse to no values")
	}
}

// TestParityRouteParams mirrors the route-parameter behaviour exercised
// throughout test/app.router.js: a ":name" segment is captured into req.Params.
func TestParityRouteParams(t *testing.T) {
	app := New()
	var id string
	app.Get("/users/:id", func(req *Request, res *Response, next Next) {
		id = req.Params("id")
		res.JSON(map[string]string{"id": id})
	})
	w := runParity(app, "GET", "/users/42")
	if id != "42" {
		t.Errorf("params id = %q, want 42", id)
	}
	if w.Body.String() != `{"id":"42"}` {
		t.Errorf("body = %q", w.Body.String())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
