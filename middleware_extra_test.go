package express

import (
	"encoding/base64"
	"net/http/httptest"
	"strings"
	"testing"
)

// doHeader drives one request through the app with a single extra header set.
func doHeader(app *Application, method, target, key, val string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	r.Header.Set(key, val)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	return w
}

func httptestNewAuth(app *Application, cred string) *httptest.ResponseRecorder {
	return doHeader(app, "GET", "/x", "Authorization", "Basic "+cred)
}

func httptestNewGzip(app *Application) *httptest.ResponseRecorder {
	return doHeader(app, "GET", "/x", "Accept-Encoding", "gzip")
}

func TestCORSMiddleware(t *testing.T) {
	app := New()
	app.Use(CORS(CORSOptions{AllowOrigins: []string{"*"}, MaxAge: 600}))
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send("ok") })

	res := do(app, "GET", "/x", "")
	if res.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("CORS origin header = %q", res.Header().Get("Access-Control-Allow-Origin"))
	}
	// Preflight short-circuits with 204 and advertises methods.
	pre := do(app, "OPTIONS", "/x", "")
	if pre.Code != 204 {
		t.Errorf("preflight status = %d, want 204", pre.Code)
	}
	if !strings.Contains(pre.Header().Get("Access-Control-Allow-Methods"), "GET") {
		t.Errorf("preflight methods = %q", pre.Header().Get("Access-Control-Allow-Methods"))
	}
	if pre.Header().Get("Access-Control-Max-Age") != "600" {
		t.Errorf("preflight max-age = %q", pre.Header().Get("Access-Control-Max-Age"))
	}
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	app := New()
	app.Use(SecurityHeaders(SecurityOptions{HSTSMaxAge: 31536000, ContentSecurityPolicy: "default-src 'self'"}))
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send("ok") })
	res := do(app, "GET", "/x", "")
	if res.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("nosniff missing")
	}
	if res.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("frame options = %q", res.Header().Get("X-Frame-Options"))
	}
	if !strings.HasPrefix(res.Header().Get("Strict-Transport-Security"), "max-age=31536000") {
		t.Errorf("HSTS = %q", res.Header().Get("Strict-Transport-Security"))
	}
	if res.Header().Get("Content-Security-Policy") != "default-src 'self'" {
		t.Errorf("CSP = %q", res.Header().Get("Content-Security-Policy"))
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	app := New()
	app.Use(RequestID())
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send(req.Get("X-Request-Id")) })
	res := do(app, "GET", "/x", "")
	id := res.Header().Get("X-Request-Id")
	if len(id) != 32 { // 16 random bytes hex-encoded
		t.Errorf("request id = %q (len %d)", id, len(id))
	}
	if strings.TrimSpace(res.Body.String()) != id {
		t.Errorf("request id not visible to handler: body=%q header=%q", res.Body.String(), id)
	}
}

func TestNoCacheAndMethodOverride(t *testing.T) {
	app := New()
	app.Use(NoCache(), MethodOverride())
	app.Delete("/item", func(req *Request, res *Response, next Next) { res.Send("deleted") })
	// POST with override should reach the DELETE handler.
	res := do(app, "POST", "/item?_method=DELETE", "")
	if strings.TrimSpace(res.Body.String()) != "deleted" {
		t.Errorf("method override failed: %q (code %d)", res.Body.String(), res.Code)
	}
	if !strings.Contains(res.Header().Get("Cache-Control"), "no-store") {
		t.Errorf("no-cache header = %q", res.Header().Get("Cache-Control"))
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	app := New()
	app.Use(BasicAuth(BasicAuthOptions{Users: map[string]string{"admin": "secret"}}))
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send("ok") })

	// No credentials -> 401 challenge.
	res := do(app, "GET", "/x", "")
	if res.Code != 401 || !strings.Contains(res.Header().Get("WWW-Authenticate"), "Basic") {
		t.Errorf("unauth = %d, challenge=%q", res.Code, res.Header().Get("WWW-Authenticate"))
	}
	// Correct credentials -> 200.
	cred := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
	r := httptestNewAuth(app, cred)
	if r.Code != 200 || strings.TrimSpace(r.Body.String()) != "ok" {
		t.Errorf("auth failed: %d %q", r.Code, r.Body.String())
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	app := New()
	app.Use(RateLimit(RateLimitOptions{Max: 2}))
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send("ok") })
	do(app, "GET", "/x", "")
	do(app, "GET", "/x", "")
	third := do(app, "GET", "/x", "")
	if third.Code != 429 {
		t.Errorf("3rd request status = %d, want 429", third.Code)
	}
	if third.Header().Get("X-RateLimit-Limit") != "2" {
		t.Errorf("limit header = %q", third.Header().Get("X-RateLimit-Limit"))
	}
	if third.Header().Get("Retry-After") == "" {
		t.Errorf("missing Retry-After")
	}
}

func TestHealthCheckAndWhen(t *testing.T) {
	app := New()
	app.Use(HealthCheck("/healthz", "up"))
	app.Use(When(func(r *Request) bool { return r.Get("X-Skip") == "" }, func(req *Request, res *Response, next Next) {
		res.Set("X-Ran", "1")
		next()
	}))
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send("x") })

	h := do(app, "GET", "/healthz", "")
	if h.Code != 200 || strings.TrimSpace(h.Body.String()) != "up" {
		t.Errorf("health = %d %q", h.Code, h.Body.String())
	}
	res := do(app, "GET", "/x", "")
	if res.Header().Get("X-Ran") != "1" {
		t.Errorf("When predicate should have run middleware")
	}
}

func TestCompressMiddleware(t *testing.T) {
	app := New()
	app.Use(Compress())
	app.Get("/x", func(req *Request, res *Response, next Next) { res.Send(strings.Repeat("hello ", 100)) })
	r := httptestNewGzip(app)
	if r.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("content-encoding = %q", r.Header().Get("Content-Encoding"))
	}
	if !strings.Contains(r.Header().Get("Vary"), "Accept-Encoding") {
		t.Errorf("missing Vary")
	}
}

func TestHelperMethods(t *testing.T) {
	// Subdomains / ContentType via a handler.
	app := New()
	app.Get("/j", func(req *Request, res *Response, next Next) {
		res.Links(map[string]string{"next": "/p/2", "prev": "/p/1"})
		res.JSONP(map[string]int{"n": 1})
	})
	res := do(app, "GET", "/j?callback=cb", "")
	if !strings.Contains(res.Body.String(), "cb(") {
		t.Errorf("JSONP body = %q", res.Body.String())
	}
	if lk := res.Header().Get("Link"); !strings.Contains(lk, `rel="next"`) || !strings.Contains(lk, `rel="prev"`) {
		t.Errorf("Link header = %q", lk)
	}
}
