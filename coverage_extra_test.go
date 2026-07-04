package express

import (
	"crypto/tls"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ---- Application settings helpers -------------------------------------------

func TestApplicationSettings(t *testing.T) {
	app := New()
	if app.GetSetting("env") != "development" {
		t.Fatalf("GetSetting(env) = %v", app.GetSetting("env"))
	}
	// x-powered-by default is enabled.
	if !app.Enabled("x-powered-by") {
		t.Fatal("x-powered-by should be enabled by default")
	}
	if app.Disabled("x-powered-by") {
		t.Fatal("x-powered-by should not be disabled")
	}
	app.Disable("x-powered-by")
	if app.Enabled("x-powered-by") {
		t.Fatal("Disable did not turn setting off")
	}
	if !app.Disabled("x-powered-by") {
		t.Fatal("Disabled should be true after Disable")
	}
	app.Enable("feature")
	if !app.Enabled("feature") {
		t.Fatal("Enable did not turn setting on")
	}
	// Locals map is usable.
	app.Locals()["site"] = "example"
	if app.Locals()["site"] != "example" {
		t.Fatal("Locals not stored")
	}
}

func TestXPoweredByHeaderDisabled(t *testing.T) {
	app := New()
	app.Disable("x-powered-by")
	app.Get("/", func(req *Request, res *Response, next Next) { res.Send("hi") })
	w := do(app, "GET", "/", "")
	if w.Header().Get("X-Powered-By") != "" {
		t.Fatalf("X-Powered-By should be absent, got %q", w.Header().Get("X-Powered-By"))
	}
}

func TestXPoweredByHeaderEnabled(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) { res.Send("hi") })
	w := do(app, "GET", "/", "")
	if w.Header().Get("X-Powered-By") != "Express" {
		t.Fatalf("X-Powered-By = %q", w.Header().Get("X-Powered-By"))
	}
}

// ---- Request accessors ------------------------------------------------------

func TestRequestAccessors(t *testing.T) {
	app := New()
	var snapshot map[string]string
	app.Get("/a/b", func(req *Request, res *Response, next Next) {
		snapshot = map[string]string{
			"original": req.OriginalURL(),
			"hostname": req.Hostname(),
			"protocol": req.Protocol(),
			"ip":       req.IP(),
			"header":   req.Header("X-Custom"),
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/a/b?x=1", nil)
	r.Host = "example.com:8080"
	r.RemoteAddr = "203.0.113.5:54321"
	r.Header.Set("X-Custom", "custom-val")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	if snapshot["hostname"] != "example.com" {
		t.Errorf("Hostname = %q", snapshot["hostname"])
	}
	if snapshot["protocol"] != "http" {
		t.Errorf("Protocol = %q", snapshot["protocol"])
	}
	if snapshot["ip"] != "203.0.113.5" {
		t.Errorf("IP = %q", snapshot["ip"])
	}
	if snapshot["header"] != "custom-val" {
		t.Errorf("Header = %q", snapshot["header"])
	}
	if !strings.Contains(snapshot["original"], "/a/b?x=1") {
		t.Errorf("OriginalURL = %q", snapshot["original"])
	}
}

func TestRequestProtocolAndSecure(t *testing.T) {
	app := New()
	var proto string
	var secure bool
	app.Get("/", func(req *Request, res *Response, next Next) {
		proto = req.Protocol()
		secure = req.Secure()
		res.Send("ok")
	})
	// X-Forwarded-Proto branch.
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if proto != "https" || !secure {
		t.Fatalf("forwarded proto = %q secure=%v", proto, secure)
	}
}

func TestRequestIPForwardedFor(t *testing.T) {
	app := New()
	var ip string
	app.Get("/", func(req *Request, res *Response, next Next) { ip = req.IP(); res.Send("ok") })
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "198.51.100.7, 10.0.0.1")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if ip != "198.51.100.7" {
		t.Fatalf("IP = %q", ip)
	}
}

func TestRequestSecureViaTLS(t *testing.T) {
	req := newRequest(httptest.NewRequest("GET", "/", nil), New())
	req.Raw.TLS = &tlsConnState
	if !req.Secure() {
		t.Fatal("Secure should be true with TLS set")
	}
	if req.Protocol() != "https" {
		t.Fatalf("Protocol = %q", req.Protocol())
	}
}

func TestRequestQueryValuesAndAllParams(t *testing.T) {
	app := New()
	var q, all string
	app.Get("/u/:id", func(req *Request, res *Response, next Next) {
		q = req.QueryValues().Get("sort")
		all = req.AllParams()["id"]
		res.Send("ok")
	})
	do(app, "GET", "/u/99?sort=asc", "")
	if q != "asc" {
		t.Fatalf("QueryValues sort = %q", q)
	}
	if all != "99" {
		t.Fatalf("AllParams id = %q", all)
	}
}

func TestRequestCookie(t *testing.T) {
	app := New()
	var got, missing string
	app.Get("/", func(req *Request, res *Response, next Next) {
		got = req.Cookie("sid")
		missing = req.Cookie("nope")
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.AddCookie(&http.Cookie{Name: "sid", Value: "abc%20123"})
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if got != "abc 123" {
		t.Fatalf("Cookie decoded = %q", got)
	}
	if missing != "" {
		t.Fatalf("missing cookie = %q", missing)
	}
}

func TestRequestBodyJSON(t *testing.T) {
	app := New()
	type payload struct {
		Name string `json:"name"`
	}
	var got payload
	app.Post("/", func(req *Request, res *Response, next Next) {
		if err := req.BodyJSON(&got); err != nil {
			res.Status(400).Send("bad")
			return
		}
		res.Send("ok")
	})
	do(app, "POST", "/", `{"name":"alice"}`)
	if got.Name != "alice" {
		t.Fatalf("BodyJSON name = %q", got.Name)
	}
}

func TestRequestBodyJSONEmpty(t *testing.T) {
	app := New()
	var callErr error
	app.Post("/", func(req *Request, res *Response, next Next) {
		var dst map[string]any
		callErr = req.BodyJSON(&dst)
		res.Send("ok")
	})
	do(app, "POST", "/", "")
	if callErr != nil {
		t.Fatalf("empty body should not error: %v", callErr)
	}
}

func TestRequestBodyJSONInvalid(t *testing.T) {
	app := New()
	var callErr error
	app.Post("/", func(req *Request, res *Response, next Next) {
		var dst map[string]any
		callErr = req.BodyJSON(&dst)
		res.Send("ok")
	})
	do(app, "POST", "/", `{not json`)
	if callErr == nil {
		t.Fatal("invalid JSON should error")
	}
}

func TestRequestIsVariants(t *testing.T) {
	cases := []struct {
		ct   string
		typ  string
		want bool
	}{
		{"application/json", "json", true},
		{"application/vnd.api+json", "json", true},
		{"text/html; charset=utf-8", "html", true},
		{"text/plain", "text", true},
		{"application/x-www-form-urlencoded", "urlencoded", true},
		{"application/x-www-form-urlencoded", "form", true},
		{"image/png", "png", true},
		{"image/png", "image/png", true},
		{"", "json", false},
		{"application/json", "html", false},
	}
	for _, c := range cases {
		req := newRequest(httptest.NewRequest("GET", "/", nil), New())
		if c.ct != "" {
			req.Raw.Header.Set("Content-Type", c.ct)
		}
		if got := req.Is(c.typ); got != c.want {
			t.Errorf("Is(%q) with CT %q = %v, want %v", c.typ, c.ct, got, c.want)
		}
	}
}

// ---- Response methods -------------------------------------------------------

func TestResponseStatusCodeAndLocation(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Status(201)
		if res.StatusCode() != 201 {
			t.Errorf("StatusCode = %d", res.StatusCode())
		}
		res.Location("/elsewhere")
		res.Send("ok")
	})
	w := do(app, "GET", "/", "")
	if w.Code != 201 {
		t.Fatalf("code = %d", w.Code)
	}
	if w.Header().Get("Location") != "/elsewhere" {
		t.Fatalf("Location = %q", w.Header().Get("Location"))
	}
}

func TestResponseSendBytesAndNil(t *testing.T) {
	app := New()
	app.Get("/bytes", func(req *Request, res *Response, next Next) {
		res.Send([]byte{1, 2, 3})
	})
	app.Get("/nil", func(req *Request, res *Response, next Next) {
		res.Status(204).Send(nil)
	})
	w := do(app, "GET", "/bytes", "")
	if w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("bytes CT = %q", w.Header().Get("Content-Type"))
	}
	if w.Body.Len() != 3 {
		t.Fatalf("bytes body len = %d", w.Body.Len())
	}
	w2 := do(app, "GET", "/nil", "")
	if w2.Code != 204 {
		t.Fatalf("nil code = %d", w2.Code)
	}
	if w2.Body.Len() != 0 {
		t.Fatalf("nil body should be empty")
	}
}

func TestResponseSendStatus(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) { res.SendStatus(404) })
	w := do(app, "GET", "/", "")
	if w.Code != 404 {
		t.Fatalf("code = %d", w.Code)
	}
	if w.Body.String() != "Not Found" {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestResponseRedirect(t *testing.T) {
	app := New()
	app.Get("/one", func(req *Request, res *Response, next Next) { res.Redirect("/dest") })
	app.Get("/two", func(req *Request, res *Response, next Next) { res.Redirect(301, "/perm") })
	app.Get("/bad", func(req *Request, res *Response, next Next) { res.Redirect() })

	w := do(app, "GET", "/one", "")
	if w.Code != http.StatusFound || w.Header().Get("Location") != "/dest" {
		t.Fatalf("one: code=%d loc=%q", w.Code, w.Header().Get("Location"))
	}
	w = do(app, "GET", "/two", "")
	if w.Code != 301 || w.Header().Get("Location") != "/perm" {
		t.Fatalf("two: code=%d loc=%q", w.Code, w.Header().Get("Location"))
	}
	w = do(app, "GET", "/bad", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("bad redirect code = %d", w.Code)
	}
}

func TestResponseCookieAndClearCookie(t *testing.T) {
	app := New()
	app.Get("/set", func(req *Request, res *Response, next Next) {
		res.Cookie("token", "value 1", &CookieOptions{
			Path:     "/x",
			Domain:   "example.com",
			MaxAge:   3600,
			Secure:   true,
			HTTPOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
		res.Send("ok")
	})
	app.Get("/nilopts", func(req *Request, res *Response, next Next) {
		res.Cookie("plain", "v", nil).Send("ok")
	})
	app.Get("/clear", func(req *Request, res *Response, next Next) {
		res.ClearCookie("token").Send("ok")
	})

	w := do(app, "GET", "/set", "")
	sc := w.Header().Get("Set-Cookie")
	if !strings.Contains(sc, "token=value+1") || !strings.Contains(sc, "Path=/x") ||
		!strings.Contains(sc, "Domain=example.com") || !strings.Contains(sc, "HttpOnly") ||
		!strings.Contains(sc, "Secure") {
		t.Fatalf("Set-Cookie = %q", sc)
	}
	w2 := do(app, "GET", "/nilopts", "")
	if !strings.Contains(w2.Header().Get("Set-Cookie"), "plain=v") {
		t.Fatalf("nil opts cookie = %q", w2.Header().Get("Set-Cookie"))
	}
	w3 := do(app, "GET", "/clear", "")
	if !strings.Contains(w3.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Fatalf("clear cookie = %q", w3.Header().Get("Set-Cookie"))
	}
}

func TestResponseWrittenAndFinalError(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		if res.Written() {
			t.Error("should not be written yet")
		}
		res.Send("done")
		if !res.Written() {
			t.Error("should be written after Send")
		}
	})
	do(app, "GET", "/", "")
}

func TestNormalizeContentType(t *testing.T) {
	cases := map[string]string{
		"json":            "application/json; charset=utf-8",
		"html":            "text/html; charset=utf-8",
		"text":            "text/plain; charset=utf-8",
		"txt":             "text/plain; charset=utf-8",
		"xml":             "application/xml; charset=utf-8",
		"js":              "application/javascript; charset=utf-8",
		"javascript":      "application/javascript; charset=utf-8",
		"css":             "text/css; charset=utf-8",
		"text/markdown":   "text/markdown; charset=utf-8",
		"application/pdf": "application/pdf",
		"unknownshort":    "unknownshort",
	}
	for in, want := range cases {
		if got := normalizeContentType(in); got != want {
			t.Errorf("normalizeContentType(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestResponseTypeVerbatimCharset(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Type("text/plain; charset=iso-8859-1").Send("hi")
	})
	w := do(app, "GET", "/", "")
	if w.Header().Get("Content-Type") != "text/plain; charset=iso-8859-1" {
		t.Fatalf("CT = %q", w.Header().Get("Content-Type"))
	}
}

// ---- Router method helpers --------------------------------------------------

func TestAllHTTPMethods(t *testing.T) {
	app := New()
	h := func(name string) Handler {
		return func(req *Request, res *Response, next Next) { res.Send(name) }
	}
	app.Get("/r", h("get"))
	app.Post("/r", h("post"))
	app.Put("/r", h("put"))
	app.Delete("/r", h("delete"))
	app.Patch("/r", h("patch"))
	app.Head("/r", h("head"))
	app.Options("/r", h("options"))

	for method, want := range map[string]string{
		"GET": "get", "POST": "post", "PUT": "put", "DELETE": "delete",
		"PATCH": "patch", "OPTIONS": "options",
	} {
		w := do(app, method, "/r", "")
		if w.Body.String() != want {
			t.Errorf("%s => %q, want %q", method, w.Body.String(), want)
		}
	}
	// HEAD has no body but should be 200.
	w := do(app, "HEAD", "/r", "")
	if w.Code != 200 {
		t.Errorf("HEAD code = %d", w.Code)
	}
}

func TestRouteChainingAllMethods(t *testing.T) {
	app := New()
	app.Route("/thing").
		Get(func(req *Request, res *Response, next Next) { res.Send("g") }).
		Post(func(req *Request, res *Response, next Next) { res.Send("p") }).
		Put(func(req *Request, res *Response, next Next) { res.Send("u") }).
		Delete(func(req *Request, res *Response, next Next) { res.Send("d") }).
		Patch(func(req *Request, res *Response, next Next) { res.Send("a") })

	for method, want := range map[string]string{
		"GET": "g", "POST": "p", "PUT": "u", "DELETE": "d", "PATCH": "a",
	} {
		if b := do(app, method, "/thing", "").Body.String(); b != want {
			t.Errorf("%s => %q, want %q", method, b, want)
		}
	}
}

func TestRouteAllMethod(t *testing.T) {
	app := New()
	app.Route("/any").All(func(req *Request, res *Response, next Next) { res.Send("all") })
	for _, m := range []string{"GET", "POST", "DELETE"} {
		if b := do(app, m, "/any", "").Body.String(); b != "all" {
			t.Errorf("%s => %q", m, b)
		}
	}
}

func TestWildcardRoute(t *testing.T) {
	app := New()
	app.Get("/files/*", func(req *Request, res *Response, next Next) {
		res.Send(req.Params("*"))
	})
	if b := do(app, "GET", "/files/a/b/c.txt", "").Body.String(); b != "a/b/c.txt" {
		t.Fatalf("wildcard = %q", b)
	}
}

// ---- Middleware -------------------------------------------------------------

func TestJSONMiddleware(t *testing.T) {
	app := New()
	app.Use(JSON())
	app.Post("/", func(req *Request, res *Response, next Next) {
		m, _ := req.Body().(map[string]any)
		res.JSON(m)
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"k":"v"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	var got map[string]any
	json.Unmarshal(w.Body.Bytes(), &got)
	if got["k"] != "v" {
		t.Fatalf("json mw body = %v", got)
	}
}

func TestJSONMiddlewareEmptyBody(t *testing.T) {
	app := New()
	app.Use(JSON())
	app.Post("/", func(req *Request, res *Response, next Next) {
		m, ok := req.Body().(map[string]any)
		if !ok || len(m) != 0 {
			t.Errorf("expected empty map, got %v", req.Body())
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
}

func TestJSONMiddlewareInvalid(t *testing.T) {
	app := New()
	app.Use(JSON())
	app.Post("/", func(req *Request, res *Response, next Next) { res.Send("unreached") })
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{bad`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("invalid json code = %d", w.Code)
	}
}

func TestJSONMiddlewareNonJSONPassthrough(t *testing.T) {
	app := New()
	app.Use(JSON())
	app.Post("/", func(req *Request, res *Response, next Next) {
		if req.Body() != nil {
			t.Errorf("body should be nil for non-json")
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader("plain"))
	r.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
}

func TestURLEncodedMiddleware(t *testing.T) {
	app := New()
	app.Use(URLEncoded())
	app.Post("/", func(req *Request, res *Response, next Next) {
		vals, _ := req.Body().(interface{ Get(string) string })
		res.Send(vals.Get("name"))
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader("name=bob&age=5"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Body.String() != "bob" {
		t.Fatalf("urlencoded name = %q", w.Body.String())
	}
}

func TestURLEncodedPassthrough(t *testing.T) {
	app := New()
	app.Use(URLEncoded())
	app.Post("/", func(req *Request, res *Response, next Next) {
		if req.Body() != nil {
			t.Error("body should be nil")
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader("x"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
}

func TestStaticMiddleware(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("static-content"), 0o644); err != nil {
		t.Fatal(err)
	}
	// index.html for directory request.
	sub := filepath.Join(dir, "docs")
	os.Mkdir(sub, 0o755)
	os.WriteFile(filepath.Join(sub, "index.html"), []byte("<h1>index</h1>"), 0o644)

	app := New()
	app.Use(Static(dir))
	app.Get("/missing", func(req *Request, res *Response, next Next) { res.Send("fallthrough") })

	w := do(app, "GET", "/hello.txt", "")
	if !strings.Contains(w.Body.String(), "static-content") {
		t.Fatalf("static file body = %q", w.Body.String())
	}
	// directory serves index.html
	w2 := do(app, "GET", "/docs", "")
	if !strings.Contains(w2.Body.String(), "index") {
		t.Fatalf("index body = %q", w2.Body.String())
	}
	// non-GET falls through
	w3 := do(app, "POST", "/hello.txt", "")
	if w3.Code != 404 {
		t.Fatalf("POST to static = %d", w3.Code)
	}
	// missing file falls through to handler
	w4 := do(app, "GET", "/missing", "")
	if w4.Body.String() != "fallthrough" {
		t.Fatalf("missing fallthrough = %q", w4.Body.String())
	}
}

func TestLoggerMiddleware(t *testing.T) {
	app := New()
	app.Use(Logger())
	app.Get("/", func(req *Request, res *Response, next Next) { res.Send("logged") })
	w := do(app, "GET", "/", "")
	if w.Body.String() != "logged" {
		t.Fatalf("logger body = %q", w.Body.String())
	}
}

// ---- Negotiation ------------------------------------------------------------

func TestAcceptsLanguagesCharsetsEncodings(t *testing.T) {
	app := New()
	var lang, charset, enc string
	app.Get("/", func(req *Request, res *Response, next Next) {
		lang = req.AcceptsLanguages("en", "fr")
		charset = req.AcceptsCharsets("utf-8", "iso-8859-1")
		enc = req.AcceptsEncodings("gzip", "br")
		res.Send("ok")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Accept-Language", "fr, en;q=0.5")
	r.Header.Set("Accept-Charset", "iso-8859-1")
	r.Header.Set("Accept-Encoding", "br;q=0.9, gzip;q=0.1")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if lang != "fr" {
		t.Errorf("lang = %q", lang)
	}
	if charset != "iso-8859-1" {
		t.Errorf("charset = %q", charset)
	}
	if enc != "br" {
		t.Errorf("enc = %q", enc)
	}
}

func TestAcceptsNoHeaderReturnsFirstOffer(t *testing.T) {
	req := newRequest(httptest.NewRequest("GET", "/", nil), New())
	if got := req.Accepts("json", "html"); got != "json" {
		t.Fatalf("no Accept header should return first offer, got %q", got)
	}
}

func TestAcceptsNoOffersReturnsPreferred(t *testing.T) {
	req := newRequest(httptest.NewRequest("GET", "/", nil), New())
	req.Raw.Header.Set("Accept", "text/html, application/json;q=0.9")
	if got := req.Accepts(); got != "text/html" {
		t.Fatalf("no offers preferred = %q", got)
	}
}

func TestMimeOfFullTypePassthrough(t *testing.T) {
	req := newRequest(httptest.NewRequest("GET", "/", nil), New())
	req.Raw.Header.Set("Accept", "application/vnd.custom+json")
	if got := req.Accepts("application/vnd.custom+json"); got != "application/vnd.custom+json" {
		t.Fatalf("full type = %q", got)
	}
}

// ---- WrapHandler ------------------------------------------------------------

func TestWrapHandler(t *testing.T) {
	app := New()
	std := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Wrapped", "1")
		w.WriteHeader(202)
		w.Write([]byte("wrapped"))
	})
	app.Use("/wrap", WrapHandler(std))
	w := do(app, "GET", "/wrap", "")
	if w.Code != 202 || w.Body.String() != "wrapped" || w.Header().Get("X-Wrapped") != "1" {
		t.Fatalf("wrap: code=%d body=%q hdr=%q", w.Code, w.Body.String(), w.Header().Get("X-Wrapped"))
	}
}

func TestWrapHandlerFunc(t *testing.T) {
	app := New()
	app.Use("/wf", WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hf"))
	}))
	w := do(app, "GET", "/wf", "")
	if w.Body.String() != "hf" {
		t.Fatalf("wrapfunc body = %q", w.Body.String())
	}
}

// ---- Conditional / Fresh / Stale --------------------------------------------

func TestFreshIfNoneMatchStar(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		if req.Fresh(res) {
			res.NotModified()
			return
		}
		res.Send("body")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("If-None-Match", "*")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != http.StatusNotModified {
		t.Fatalf("If-None-Match * code = %d", w.Code)
	}
}

func TestFreshIfModifiedSince(t *testing.T) {
	app := New()
	mod := time.Now().Add(-2 * time.Hour)
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.LastModified(mod)
		if res.Fresh() {
			res.NotModified()
			return
		}
		res.Send("body")
	})
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("If-Modified-Since", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != http.StatusNotModified {
		t.Fatalf("If-Modified-Since fresh code = %d", w.Code)
	}
}

func TestStale(t *testing.T) {
	req := newRequest(httptest.NewRequest("GET", "/", nil), New())
	res := newResponse(httptest.NewRecorder(), req, req.app)
	// No conditional headers => not fresh => stale.
	if !req.Stale(res) {
		t.Fatal("expected stale with no conditional headers")
	}
}

func TestFreshNonGetIsFalse(t *testing.T) {
	req := newRequest(httptest.NewRequest("POST", "/", nil), New())
	res := newResponse(httptest.NewRecorder(), req, req.app)
	req.Raw.Header.Set("If-None-Match", "*")
	if req.Fresh(res) {
		t.Fatal("POST should never be fresh")
	}
}

// ---- SendFile / Download / Attachment ---------------------------------------

func TestSendFileMimeTypes(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a.css":  "text/css; charset=utf-8",
		"a.js":   "application/javascript; charset=utf-8",
		"a.json": "application/json; charset=utf-8",
		"a.png":  "image/png",
		"a.svg":  "image/svg+xml",
		"a.pdf":  "application/pdf",
		"a.xyz":  "application/octet-stream",
	}
	for name := range files {
		os.WriteFile(filepath.Join(dir, name), []byte("data"), 0o644)
	}
	for name, wantCT := range files {
		app := New()
		p := filepath.Join(dir, name)
		app.Get("/f", func(req *Request, res *Response, next Next) {
			if err := res.SendFile(p); err != nil {
				t.Errorf("SendFile(%s): %v", name, err)
			}
		})
		w := do(app, "GET", "/f", "")
		if w.Header().Get("Content-Type") != wantCT {
			t.Errorf("%s CT = %q, want %q", name, w.Header().Get("Content-Type"), wantCT)
		}
	}
}

func TestSendFileNotFound(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		if err := res.SendFile("/no/such/file.txt"); err == nil {
			t.Error("expected error for missing file")
		}
		res.Send("done")
	})
	do(app, "GET", "/", "")
}

func TestSendFileDirectory(t *testing.T) {
	dir := t.TempDir()
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		if err := res.SendFile(dir); err == nil {
			t.Error("expected error for directory")
		}
		res.Send("done")
	})
	do(app, "GET", "/", "")
}

func TestAttachmentAndDownload(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "report.pdf")
	os.WriteFile(p, []byte("%PDF"), 0o644)

	app := New()
	app.Get("/dl", func(req *Request, res *Response, next Next) {
		res.Download(p)
	})
	app.Get("/dlname", func(req *Request, res *Response, next Next) {
		res.Download(p, "custom.pdf")
	})
	app.Get("/att", func(req *Request, res *Response, next Next) {
		res.Attachment().Send("x")
	})

	w := do(app, "GET", "/dl", "")
	if !strings.Contains(w.Header().Get("Content-Disposition"), `filename="report.pdf"`) {
		t.Fatalf("download disposition = %q", w.Header().Get("Content-Disposition"))
	}
	w2 := do(app, "GET", "/dlname", "")
	if !strings.Contains(w2.Header().Get("Content-Disposition"), `filename="custom.pdf"`) {
		t.Fatalf("named disposition = %q", w2.Header().Get("Content-Disposition"))
	}
	w3 := do(app, "GET", "/att", "")
	if w3.Header().Get("Content-Disposition") != "attachment" {
		t.Fatalf("bare attachment = %q", w3.Header().Get("Content-Disposition"))
	}
}

// ---- Streaming / SSE extras -------------------------------------------------

func TestWriteStringAndFlush(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Type("text").WriteString("part1")
		res.WriteString("part2")
		if !res.Flush() {
			// httptest recorder supports Flush
			t.Error("Flush should return true on recorder")
		}
	})
	w := do(app, "GET", "/", "")
	if w.Body.String() != "part1part2" {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestSSEExtras(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		s := res.SSE()
		s.SendData("hello")
		s.SendJSON("update", map[string]int{"n": 1})
		s.SendID("42", "named", "line1\nline2")
		s.Comment("keepalive")
		s.Retry(3000)
	})
	w := do(app, "GET", "/", "")
	body := w.Body.String()
	for _, want := range []string{
		"data: hello",
		"event: update",
		`data: {"n":1}`,
		"id: 42",
		"event: named",
		"data: line1",
		"data: line2",
		": keepalive",
		"retry: 3000",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("SSE body missing %q\nbody=%q", want, body)
		}
	}
}

func TestSendStreamReader(t *testing.T) {
	app := New()
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.SendStream(strings.NewReader("streamed-data"), 4)
	})
	w := do(app, "GET", "/", "")
	if w.Body.String() != "streamed-data" {
		t.Fatalf("stream body = %q", w.Body.String())
	}
}

// ---- Views ------------------------------------------------------------------

func TestEngineRegistration(t *testing.T) {
	app := New()
	app.Engine("md", func(path string, data any) (string, error) {
		return "rendered-md", nil
	})
	dir := t.TempDir()
	app.Set("views", dir)
	os.WriteFile(filepath.Join(dir, "page.md"), []byte("x"), 0o644)
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Render("page.md")
	})
	w := do(app, "GET", "/", "")
	if w.Body.String() != "rendered-md" {
		t.Fatalf("render md = %q", w.Body.String())
	}
}

func TestRenderNoEngineError(t *testing.T) {
	app := New()
	dir := t.TempDir()
	app.Set("views", dir)
	app.Get("/", func(req *Request, res *Response, next Next) {
		if err := res.Render("page.unknownext"); err == nil {
			t.Error("expected error for unregistered engine")
		}
	})
	w := do(app, "GET", "/", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("no engine code = %d", w.Code)
	}
}

func TestRenderUsesDefaultViewEngine(t *testing.T) {
	app := New()
	dir := t.TempDir()
	app.Set("views", dir)
	os.WriteFile(filepath.Join(dir, "hello.html"), []byte("<b>{{.Name}}</b>"), 0o644)
	app.Get("/", func(req *Request, res *Response, next Next) {
		res.Render("hello", map[string]string{"Name": "World"})
	})
	w := do(app, "GET", "/", "")
	if !strings.Contains(w.Body.String(), "<b>World</b>") {
		t.Fatalf("default engine render = %q", w.Body.String())
	}
}

// ---- Session extras ---------------------------------------------------------

func TestSessionGetStringDeleteRegenerate(t *testing.T) {
	app := New()
	app.Use(Session())
	app.Get("/set", func(req *Request, res *Response, next Next) {
		s := req.Session()
		s.Set("user", "alice")
		res.Send(s.GetString("user"))
	})
	app.Get("/del", func(req *Request, res *Response, next Next) {
		s := req.Session()
		s.Set("user", "bob")
		s.Delete("user")
		res.Send("[" + s.GetString("user") + "]")
	})
	app.Get("/regen", func(req *Request, res *Response, next Next) {
		s := req.Session()
		s.Set("x", "1")
		if err := s.Regenerate(); err != nil {
			t.Errorf("regen: %v", err)
		}
		res.Send("regenerated")
	})

	if b := do(app, "GET", "/set", "").Body.String(); b != "alice" {
		t.Fatalf("GetString = %q", b)
	}
	if b := do(app, "GET", "/del", "").Body.String(); b != "[]" {
		t.Fatalf("Delete then GetString = %q", b)
	}
	if b := do(app, "GET", "/regen", "").Body.String(); b != "regenerated" {
		t.Fatalf("Regenerate body = %q", b)
	}
}

// ---- Forms extras -----------------------------------------------------------

func TestFormValues(t *testing.T) {
	app := New()
	var name string
	var all int
	app.Post("/", func(req *Request, res *Response, next Next) {
		form := req.Form()
		name = form.Get("name")
		all = len(form)
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/?extra=1", strings.NewReader("name=carol&city=nyc"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if name != "carol" {
		t.Fatalf("Form name = %q", name)
	}
	if all < 2 {
		t.Fatalf("Form count = %d", all)
	}
}

func TestFormFilesAccessor(t *testing.T) {
	var body strings.Builder
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("docs", "one.txt")
	fw.Write([]byte("first"))
	fw2, _ := mw.CreateFormFile("docs", "two.txt")
	fw2.Write([]byte("second"))
	mw.WriteField("title", "hi")
	mw.Close()

	app := New()
	var count int
	var first string
	var single string
	app.Post("/", func(req *Request, res *Response, next Next) {
		files := req.Files("docs")
		count = len(files)
		if len(files) > 0 {
			first = files[0].Filename
		}
		f, hdr, err := req.FormFile("docs")
		if err == nil {
			single = hdr.Filename
			f.Close()
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(body.String()))
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if count != 2 {
		t.Fatalf("Files count = %d, want 2", count)
	}
	if first != "one.txt" {
		t.Fatalf("first file = %q", first)
	}
	if single != "one.txt" {
		t.Fatalf("FormFile = %q", single)
	}
}

func TestFilesMissingField(t *testing.T) {
	var body strings.Builder
	mw := multipart.NewWriter(&body)
	mw.WriteField("x", "y")
	mw.Close()
	app := New()
	app.Post("/", func(req *Request, res *Response, next Next) {
		if f := req.Files("nope"); f != nil {
			t.Errorf("expected nil for missing field, got %v", f)
		}
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(body.String()))
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
}

// tlsConnState is a minimal non-nil TLS state for Secure() tests.
var tlsConnState = tls.ConnectionState{Version: tls.VersionTLS13}
