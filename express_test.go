package express

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func do(app *Application, method, target string, body string) *httptest.ResponseRecorder {
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

func TestBasicRouting(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Send("hello")
	})
	w := do(app, "GET", "/", "")
	if w.Code != 200 {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if w.Body.String() != "hello" {
		t.Fatalf("body = %q, want hello", w.Body.String())
	}
}

func TestRouteParams(t *testing.T) {
	app := New()
	app.Get("/users/:id/books/:book", func(req *Request, res *Response, next Next) {
		res.JSON(map[string]string{"id": req.Params("id"), "book": req.Params("book")})
	})
	w := do(app, "GET", "/users/42/books/go", "")
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["id"] != "42" || got["book"] != "go" {
		t.Fatalf("params = %v", got)
	}
}

func TestMethodMismatchIs404(t *testing.T) {
	app := New()
	app.Get("/only-get", func(req *Request, res *Response, next Next) { res.Send("ok") })
	w := do(app, "POST", "/only-get", "")
	if w.Code != 404 {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestMiddlewareChain(t *testing.T) {
	app := New()
	order := []string{}
	app.Use(func(req *Request, res *Response, next Next) {
		order = append(order, "a")
		next()
	})
	app.Use(func(req *Request, res *Response, next Next) {
		order = append(order, "b")
		next()
	})
	app.Get("/", func(req *Request, res *Response, next Next) {
		order = append(order, "handler")
		res.Send("done")
	})
	do(app, "GET", "/", "")
	if strings.Join(order, ",") != "a,b,handler" {
		t.Fatalf("order = %v", order)
	}
}

func TestErrorHandling(t *testing.T) {
	app := New()
	app.Get("/boom", func(req *Request, res *Response, next Next) {
		next(http.ErrBodyNotAllowed)
	})
	app.Use(func(err error, req *Request, res *Response, next Next) {
		res.Status(500).Send("caught: " + err.Error())
	})
	w := do(app, "GET", "/boom", "")
	if w.Code != 500 || !strings.HasPrefix(w.Body.String(), "caught:") {
		t.Fatalf("code=%d body=%q", w.Code, w.Body.String())
	}
}

func TestMountedRouter(t *testing.T) {
	app := New()
	api := NewRouter()
	api.Get("/users/:id", func(req *Request, res *Response, next Next) {
		res.Send("user " + req.Params("id"))
	})
	app.Use("/api", api)
	w := do(app, "GET", "/api/users/7", "")
	if w.Body.String() != "user 7" {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestQueryAndJSONBody(t *testing.T) {
	app := New()
	app.Use(JSON())
	app.Post("/echo", func(req *Request, res *Response, next Next) {
		res.JSON(map[string]any{
			"q":    req.Query("q"),
			"body": req.Body(),
		})
	})
	r := httptest.NewRequest("POST", "/echo?q=hi", strings.NewReader(`{"name":"go"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["q"] != "hi" {
		t.Fatalf("query not parsed: %v", got)
	}
	body, ok := got["body"].(map[string]any)
	if !ok || body["name"] != "go" {
		t.Fatalf("json body not parsed: %v", got["body"])
	}
}

func TestStatusAndJSON(t *testing.T) {
	app := New()
	app.Get("/created", func(req *Request, res *Response, next Next) {
		res.Status(201).JSON(map[string]int{"ok": 1})
	})
	w := do(app, "GET", "/created", "")
	if w.Code != 201 {
		t.Fatalf("status = %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Fatalf("content-type = %q", ct)
	}
}

func TestAllMethod(t *testing.T) {
	app := New()
	app.All("/any", func(req *Request, res *Response, next Next) { res.Send(req.Method()) })
	if b := do(app, "GET", "/any", "").Body.String(); b != "GET" {
		t.Fatalf("GET body=%q", b)
	}
	if b := do(app, "DELETE", "/any", "").Body.String(); b != "DELETE" {
		t.Fatalf("DELETE body=%q", b)
	}
}

func TestRecoverMiddleware(t *testing.T) {
	app := New()
	app.Use(Recover())
	app.Get("/panic", func(req *Request, res *Response, next Next) {
		panic("kaboom")
	})
	w := do(app, "GET", "/panic", "")
	if w.Code != 500 {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}
