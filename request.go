package express

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Request wraps an *http.Request and exposes Express-style accessors for route
// parameters, query values, headers, and the parsed request body.
type Request struct {
	// Raw is the underlying standard-library request.
	Raw *http.Request

	app    *Application
	params map[string]string
	query  url.Values
	body   any
	// path is the portion of the URL path used for routing. It is trimmed as
	// the request descends into mounted sub-routers.
	path string

	// ctx holds per-request values set by middleware (res.Locals equivalent
	// for the request scope).
	values map[string]any
}

func newRequest(r *http.Request, app *Application) *Request {
	return &Request{
		Raw:    r,
		app:    app,
		params: make(map[string]string),
		path:   cleanPath(r.URL.Path),
		values: make(map[string]any),
	}
}

// withMountPath returns a shallow copy of the request whose routing path has
// the matched mount prefix stripped and whose params include the mount's
// captures. Body, values, and params map are shared with the parent.
func (req *Request) withMountPath(prefixLen int, params map[string]string) *Request {
	cp := *req
	residual := req.path[prefixLen:]
	if residual == "" || residual[0] != '/' {
		residual = "/" + residual
	}
	cp.path = cleanPath(residual)
	cp.mergeParams(params)
	return &cp
}

func (req *Request) mergeParams(params map[string]string) {
	if len(params) == 0 {
		return
	}
	if req.params == nil {
		req.params = make(map[string]string)
	}
	for k, v := range params {
		if decoded, err := url.PathUnescape(v); err == nil {
			req.params[k] = decoded
		} else {
			req.params[k] = v
		}
	}
}

// Method returns the request's HTTP method (GET, POST, ...).
func (req *Request) Method() string { return req.Raw.Method }

// Path returns the request URL path (without query string).
func (req *Request) Path() string { return req.Raw.URL.Path }

// OriginalURL returns the full request URI as received.
func (req *Request) OriginalURL() string { return req.Raw.URL.RequestURI() }

// Hostname returns the host portion of the request, honoring the Host header.
func (req *Request) Hostname() string {
	h := req.Raw.Host
	if i := strings.IndexByte(h, ':'); i >= 0 {
		return h[:i]
	}
	return h
}

// Protocol returns "https" when the request arrived over TLS, else "http".
func (req *Request) Protocol() string {
	if req.Raw.TLS != nil {
		return "https"
	}
	if p := req.Raw.Header.Get("X-Forwarded-Proto"); p != "" {
		return p
	}
	return "http"
}

// Secure reports whether the request was made over a secure connection.
func (req *Request) Secure() bool { return req.Protocol() == "https" }

// IP returns the remote address of the request.
func (req *Request) IP() string {
	if f := req.Raw.Header.Get("X-Forwarded-For"); f != "" {
		return strings.TrimSpace(strings.Split(f, ",")[0])
	}
	host := req.Raw.RemoteAddr
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		return host[:i]
	}
	return host
}

// Params returns a captured route parameter by name, or "" if absent.
func (req *Request) Params(name string) string { return req.params[name] }

// AllParams returns all captured route parameters.
func (req *Request) AllParams() map[string]string { return req.params }

// Query returns the first value of a query-string parameter.
func (req *Request) Query(name string) string {
	if req.query == nil {
		req.query = req.Raw.URL.Query()
	}
	return req.query.Get(name)
}

// QueryValues returns the parsed query string as url.Values.
func (req *Request) QueryValues() url.Values {
	if req.query == nil {
		req.query = req.Raw.URL.Query()
	}
	return req.query
}

// Get returns a request header value (case-insensitive), Express's req.get.
func (req *Request) Get(field string) string { return req.Raw.Header.Get(field) }

// Header is an alias for Get.
func (req *Request) Header(field string) string { return req.Get(field) }

// Is reports whether the request's Content-Type matches the given type, e.g.
// req.Is("json") or req.Is("text/html").
func (req *Request) Is(typ string) bool {
	ct := req.Get("Content-Type")
	if ct == "" {
		return false
	}
	ct = strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	typ = strings.ToLower(typ)
	switch typ {
	case "json":
		return strings.HasSuffix(ct, "/json") || strings.HasSuffix(ct, "+json")
	case "html":
		return ct == "text/html"
	case "text":
		return strings.HasPrefix(ct, "text/")
	case "urlencoded", "form":
		return ct == "application/x-www-form-urlencoded"
	default:
		return ct == typ || strings.HasSuffix(ct, "/"+typ)
	}
}

// Cookie returns the value of a named cookie, or "" if not present.
func (req *Request) Cookie(name string) string {
	c, err := req.Raw.Cookie(name)
	if err != nil {
		return ""
	}
	if v, err := url.QueryUnescape(c.Value); err == nil {
		return v
	}
	return c.Value
}

// Body returns the parsed request body. Call one of the Parse* helpers or the
// body-parser middleware first; otherwise it returns nil.
func (req *Request) Body() any { return req.body }

// SetBody stores a parsed body value (used by body-parsing middleware).
func (req *Request) SetBody(v any) { req.body = v }

// BodyJSON reads and unmarshals a JSON request body into dst.
func (req *Request) BodyJSON(dst any) error {
	defer req.Raw.Body.Close()
	data, err := io.ReadAll(req.Raw.Body)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	req.body = dst
	return nil
}

// FormValue parses form data (query + body) and returns a single value.
func (req *Request) FormValue(name string) string {
	return req.Raw.FormValue(name)
}

// Set stores an arbitrary value on the request for downstream handlers.
func (req *Request) Set(key string, value any) { req.values[key] = value }

// Value retrieves a value previously stored with Set.
func (req *Request) Value(key string) (any, bool) {
	v, ok := req.values[key]
	return v, ok
}

// cleanPath ensures a leading slash and strips duplicate trailing slashes
// beyond the root.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	return p
}
