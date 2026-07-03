package express

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newReq(method, target string) *http.Request {
	return httptest.NewRequest(method, target, nil)
}

func serve(app *Application, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

// ---- routing: optional & regex params, Route, Param -------------------------

func TestOptionalParam(t *testing.T) {
	app := New()
	app.Get("/users/:id?", func(req *Request, res *Response, next Next) {
		if id := req.Params("id"); id != "" {
			res.Send("user " + id)
		} else {
			res.Send("all users")
		}
	})
	if b := do(app, "GET", "/users/42", "").Body.String(); b != "user 42" {
		t.Fatalf("with id: %q", b)
	}
	if b := do(app, "GET", "/users", "").Body.String(); b != "all users" {
		t.Fatalf("without id: %q", b)
	}
}

func TestRegexParam(t *testing.T) {
	app := New()
	app.Get(`/items/:id(\d+)`, func(req *Request, res *Response, next Next) {
		res.Send("item " + req.Params("id"))
	})
	if c := do(app, "GET", "/items/7", "").Code; c != 200 {
		t.Fatalf("numeric id: code=%d", c)
	}
	// Non-numeric must not match -> 404.
	if c := do(app, "GET", "/items/abc", "").Code; c != 404 {
		t.Fatalf("non-numeric id: code=%d, want 404", c)
	}
}

func TestRouteChaining(t *testing.T) {
	app := New()
	app.Route("/things").
		Get(func(req *Request, res *Response, next Next) { res.Send("list") }).
		Post(func(req *Request, res *Response, next Next) { res.Status(201).Send("created") })
	if b := do(app, "GET", "/things", "").Body.String(); b != "list" {
		t.Fatalf("GET: %q", b)
	}
	if c := do(app, "POST", "/things", "").Code; c != 201 {
		t.Fatalf("POST: code=%d", c)
	}
}

func TestParamCallback(t *testing.T) {
	app := New()
	app.Param("id", func(req *Request, res *Response, next Next, value string) {
		req.Set("loadedID", "loaded-"+value)
		next()
	})
	app.Get("/u/:id", func(req *Request, res *Response, next Next) {
		v, _ := req.Value("loadedID")
		res.Send(v.(string))
	})
	if b := do(app, "GET", "/u/9", "").Body.String(); b != "loaded-9" {
		t.Fatalf("param callback: %q", b)
	}
}

// ---- file responses ---------------------------------------------------------

func TestSendFileAndDownload(t *testing.T) {
	dir := t.TempDir()
	fpath := filepath.Join(dir, "hello.txt")
	os.WriteFile(fpath, []byte("file body"), 0o644)

	app := New()
	app.Get("/f", func(req *Request, res *Response, next Next) { res.SendFile(fpath) })
	app.Get("/d", func(req *Request, res *Response, next Next) { res.Download(fpath, "renamed.txt") })

	w := do(app, "GET", "/f", "")
	if w.Body.String() != "file body" {
		t.Fatalf("sendfile body = %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/plain") {
		t.Fatalf("sendfile content-type = %q", ct)
	}

	w2 := do(app, "GET", "/d", "")
	if cd := w2.Header().Get("Content-Disposition"); !strings.Contains(cd, `filename="renamed.txt"`) {
		t.Fatalf("download disposition = %q", cd)
	}
}

// ---- content negotiation ----------------------------------------------------

func TestAccepts(t *testing.T) {
	app := New()
	app.Get("/n", func(req *Request, res *Response, next Next) {
		res.Send(req.Accepts("html", "json"))
	})
	r := newReq("GET", "/n")
	r.Header.Set("Accept", "application/json, text/html;q=0.9")
	if b := serve(app, r).Body.String(); b != "json" {
		t.Fatalf("accepts = %q, want json", b)
	}
}

func TestFormat(t *testing.T) {
	app := New()
	app.Get("/res", func(req *Request, res *Response, next Next) {
		res.Format(map[string]func(){
			"html": func() { res.Send("<b>hi</b>") },
			"json": func() { res.JSON(map[string]string{"msg": "hi"}) },
		})
	})
	r := newReq("GET", "/res")
	r.Header.Set("Accept", "application/json")
	if b := serve(app, r).Body.String(); !strings.Contains(b, `"msg":"hi"`) {
		t.Fatalf("format json = %q", b)
	}
	r2 := newReq("GET", "/res")
	r2.Header.Set("Accept", "text/html")
	if b := serve(app, r2).Body.String(); b != "<b>hi</b>" {
		t.Fatalf("format html = %q", b)
	}
}

func TestRanges(t *testing.T) {
	app := New()
	app.Get("/r", func(req *Request, res *Response, next Next) {
		ranges, ok := req.Ranges(100)
		if !ok || len(ranges) != 1 {
			res.Send("none")
			return
		}
		res.Send(itoaRange(ranges[0]))
	})
	r := newReq("GET", "/r")
	r.Header.Set("Range", "bytes=10-19")
	if b := serve(app, r).Body.String(); b != "10-19" {
		t.Fatalf("range = %q", b)
	}
}

// ---- conditional GET --------------------------------------------------------

func TestFreshETag(t *testing.T) {
	app := New()
	app.Get("/c", func(req *Request, res *Response, next Next) {
		res.ETag("abc123")
		if req.Fresh(res) {
			res.NotModified()
			return
		}
		res.Send("full body")
	})
	// No validators -> full body.
	if b := do(app, "GET", "/c", "").Body.String(); b != "full body" {
		t.Fatalf("first request = %q", b)
	}
	// Matching If-None-Match -> 304.
	r := newReq("GET", "/c")
	r.Header.Set("If-None-Match", `"abc123"`)
	w := serve(app, r)
	if w.Code != 304 {
		t.Fatalf("conditional = %d, want 304", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("304 should have empty body, got %q", w.Body.String())
	}
}

// ---- views ------------------------------------------------------------------

func TestRenderView(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.html"), []byte(`<h1>Hi {{.Name}}</h1>`), 0o644)

	app := New()
	app.Set("views", dir)
	app.Get("/page", func(req *Request, res *Response, next Next) {
		res.Render("hello", map[string]string{"Name": "Ada"})
	})
	w := do(app, "GET", "/page", "")
	if !strings.Contains(w.Body.String(), "<h1>Hi Ada</h1>") {
		t.Fatalf("render = %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("render content-type = %q", ct)
	}
}

// helpers
func itoaRange(r Range) string { return itoa(int(r.Start)) + "-" + itoa(int(r.End)) }
