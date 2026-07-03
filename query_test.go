package express

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestQueryMethod(t *testing.T) {
	app := New()
	app.Query("/search", func(req *Request, res *Response, next Next) {
		// QUERY carries a body describing the query, like a safe POST.
		body, _ := req.Body().(string)
		res.JSON(map[string]string{"method": req.Method(), "q": body})
	})
	app.Use(Text()) // parse the text body

	// Reorder: body parser must run first.
	app2 := New()
	app2.Use(Text())
	app2.Query("/search", func(req *Request, res *Response, next Next) {
		body, _ := req.Body().(string)
		res.JSON(map[string]string{"method": req.Method(), "q": body})
	})

	r := httptest.NewRequest(MethodQuery, "/search", strings.NewReader("golang generics"))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	app2.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"method":"QUERY"`) || !strings.Contains(w.Body.String(), `"q":"golang generics"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}

	// A GET to the same path should NOT match the QUERY route -> 404.
	w2 := httptest.NewRecorder()
	app2.ServeHTTP(w2, httptest.NewRequest("GET", "/search", nil))
	if w2.Code != 404 {
		t.Fatalf("GET on QUERY route: status = %d, want 404", w2.Code)
	}
}
