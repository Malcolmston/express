package express

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionPersistsAcrossRequests(t *testing.T) {
	app := New()
	app.Use(Session())

	app.Get("/set", func(req *Request, res *Response, next Next) {
		req.Session().Set("count", 1)
		res.Send("set")
	})
	app.Get("/get", func(req *Request, res *Response, next Next) {
		v, _ := req.Session().Get("count")
		res.JSON(map[string]any{"count": v})
	})

	// First request sets the value and receives a cookie.
	w := do(app, "GET", "/set", "")
	cookies := w.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected a session cookie")
	}

	// Second request presents the cookie and reads the value back.
	r2 := httptest.NewRequest("GET", "/get", nil)
	for _, c := range cookies {
		r2.AddCookie(c)
	}
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	if !strings.Contains(w2.Body.String(), `"count":1`) {
		t.Fatalf("session not persisted: %q", w2.Body.String())
	}
}

func TestSessionDestroy(t *testing.T) {
	app := New()
	app.Use(Session())
	app.Get("/set", func(req *Request, res *Response, next Next) {
		req.Session().Set("user", "ada")
		res.Send("ok")
	})
	app.Get("/logout", func(req *Request, res *Response, next Next) {
		req.Session().Destroy()
		res.Send("bye")
	})
	app.Get("/whoami", func(req *Request, res *Response, next Next) {
		res.Send(req.Session().GetString("user"))
	})

	w := do(app, "GET", "/set", "")
	cookies := w.Result().Cookies()

	// Destroy.
	r2 := httptest.NewRequest("GET", "/logout", nil)
	for _, c := range cookies {
		r2.AddCookie(c)
	}
	app.ServeHTTP(httptest.NewRecorder(), r2)

	// The old cookie should no longer resolve to a user.
	r3 := httptest.NewRequest("GET", "/whoami", nil)
	for _, c := range cookies {
		r3.AddCookie(c)
	}
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, r3)
	if w3.Body.String() != "" {
		t.Fatalf("expected empty user after destroy, got %q", w3.Body.String())
	}
}

func TestMultipartFileUpload(t *testing.T) {
	app := New()
	app.Use(Multipart(0))
	app.Post("/upload", func(req *Request, res *Response, next Next) {
		f, hdr, err := req.FormFile("file")
		if err != nil {
			next(err)
			return
		}
		defer f.Close()
		buf := new(bytes.Buffer)
		buf.ReadFrom(f)
		res.JSON(map[string]any{"name": hdr.Filename, "size": buf.Len(), "field": req.FormValue("title")})
	})

	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	mw.WriteField("title", "hello")
	fw, _ := mw.CreateFormFile("file", "greeting.txt")
	fw.Write([]byte("hi there"))
	mw.Close()

	r := httptest.NewRequest("POST", "/upload", body)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	got := w.Body.String()
	if !strings.Contains(got, `"name":"greeting.txt"`) || !strings.Contains(got, `"size":8`) || !strings.Contains(got, `"field":"hello"`) {
		t.Fatalf("unexpected upload result: %s", got)
	}
}

func TestTextBodyParser(t *testing.T) {
	app := New()
	app.Use(Text())
	app.Post("/echo", func(req *Request, res *Response, next Next) {
		res.Send(req.Body().(string))
	})
	r := httptest.NewRequest("POST", "/echo", strings.NewReader("plain words"))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "plain words" {
		t.Fatalf("text body = %q", w.Body.String())
	}
}

func TestBeforeWriteHookRunsOnce(t *testing.T) {
	app := New()
	calls := 0
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.OnBeforeWrite(func() { calls++ })
		res.Send("a")
	})
	do(app, "GET", "/", "")
	if calls != 1 {
		t.Fatalf("before-write hook ran %d times, want 1", calls)
	}
}
