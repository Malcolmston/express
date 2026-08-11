package express

import (
	"encoding/json"
	"io"
	"net"
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

	// paramDone tracks which route parameters have had their app.Param
	// callbacks run, so each runs at most once per request.
	paramDone map[string]bool

	// errHandlers records whether the application's router tree contains any
	// error-handling middleware. When it does, a panic raised inside a handler
	// is caught and routed to that middleware (as Express does with a thrown
	// error); when it does not, the panic is left to propagate so Go
	// recover-based middleware can handle it.
	errHandlers bool
}

func newRequest(r *http.Request, app *Application) *Request {
	return &Request{
		Raw:       r,
		app:       app,
		params:    make(map[string]string),
		path:      cleanPath(r.URL.EscapedPath()),
		values:    make(map[string]any),
		paramDone: make(map[string]bool),
	}
}

// withMountPath returns a shallow copy of the request whose routing path has
// the matched mount prefix stripped. When inherit is true the copy shares the
// parent's params (and adds the mount's captures) — Express's mergeParams;
// otherwise it starts with a fresh, empty parameter set scoped to the
// sub-router. Body and values are shared with the parent.
func (req *Request) withMountPath(residual string, params map[string]string, inherit bool) *Request {
	cp := *req
	if residual == "" || residual[0] != '/' {
		residual = "/" + residual
	}
	cp.path = cleanPath(residual)

	if inherit {
		cp.mergeParams(params)
	} else {
		// Scope the sub-router to its own parameters.
		cp.params = make(map[string]string)
		cp.paramDone = make(map[string]bool)
	}
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
//
// Like Express's req.path this is the path *relative to the current mount
// point*: inside a router mounted at "/api" or middleware used at "/pfx", it is
// the residual portion after the prefix, not the full original path (which is
// available via OriginalURL).
func (req *Request) Path() string {
	if req.path != "" {
		return req.path
	}
	return req.Raw.URL.Path
}

// SetPath rewrites the request path and, crucially, the path used by the
// router to match routes. Middleware that rewrites URLs (rewrite, basepath,
// ...) must call SetPath rather than mutating req.Raw.URL.Path directly, so
// that route matching downstream sees the new path. It also updates
// req.Raw.URL.Path so handlers observe a consistent value.
func (req *Request) SetPath(p string) {
	p = cleanPath(p)
	req.path = p
	req.Raw.URL.Path = p
}

// OriginalURL returns the full request URI as received.
func (req *Request) OriginalURL() string { return req.Raw.URL.RequestURI() }

// URL returns the request URL relative to the current mount point — the routing
// path (residual inside a mount) together with the original query string. It is
// the analogue of Express's req.url, which is rewritten to the residual as the
// request descends into mounted routers while retaining the query.
func (req *Request) URL() string {
	p := req.Path()
	if q := req.Raw.URL.RawQuery; q != "" {
		return p + "?" + q
	}
	return p
}

// TrustProxy reports whether the app has enabled the "trust proxy" setting.
//
// Like Express, the setting is DISABLED by default: forwarding headers such as
// X-Forwarded-For and X-Forwarded-Proto are attacker-controlled on a directly
// exposed server, so IP, Protocol, Secure and Hostname ignore them until the
// operator declares that a trusted proxy sits in front of the app with
//
//	app.Enable("trust proxy")
func (req *Request) TrustProxy() bool {
	return req.app != nil && req.app.Enabled("trust proxy")
}

// Hostname returns the host portion of the request, honoring the Host header.
// The port is removed and an IPv6 literal is unwrapped from its brackets, so
// "[::1]:8080" yields "::1". When "trust proxy" is enabled a non-empty
// X-Forwarded-Host header takes precedence (its first entry, as in Express).
func (req *Request) Hostname() string {
	h := req.Raw.Host
	if req.TrustProxy() {
		if fwd := req.Raw.Header.Get("X-Forwarded-Host"); fwd != "" {
			h = strings.TrimSpace(strings.Split(fwd, ",")[0])
		}
	}
	return stripHostPort(h)
}

// stripHostPort removes a trailing ":port" from a host value and unwraps a
// bracketed IPv6 literal. A bare IPv6 address (which contains several colons and
// no brackets) is returned unchanged rather than truncated at its first colon.
func stripHostPort(h string) string {
	if h == "" {
		return h
	}
	if h[0] == '[' {
		if end := strings.IndexByte(h, ']'); end >= 0 {
			return h[1:end]
		}
		return h
	}
	// More than one colon and no brackets means a bare IPv6 literal: no port.
	if strings.Count(h, ":") > 1 {
		return h
	}
	if i := strings.IndexByte(h, ':'); i >= 0 {
		return h[:i]
	}
	return h
}

// Protocol returns "https" when the request arrived over TLS, else "http".
//
// The X-Forwarded-Proto header is consulted only when "trust proxy" is enabled;
// otherwise any client could claim to be on HTTPS and defeat middleware that
// keys off Secure. When the header carries a comma-separated list the first
// (client-most) entry wins, matching Express.
func (req *Request) Protocol() string {
	if req.Raw.TLS != nil {
		return "https"
	}
	if req.TrustProxy() {
		if p := req.Raw.Header.Get("X-Forwarded-Proto"); p != "" {
			return strings.ToLower(strings.TrimSpace(strings.Split(p, ",")[0]))
		}
	}
	return "http"
}

// Secure reports whether the request was made over a secure connection.
func (req *Request) Secure() bool { return req.Protocol() == "https" }

// IP returns the client address of the request.
//
// By default this is the address of the socket peer, with any port removed —
// the only value the server actually observed. The X-Forwarded-For header is
// honoured (its left-most, client-most entry) only when "trust proxy" is
// enabled, because on a directly exposed server that header is set by the
// client and would let anyone forge the address that IP-keyed rate limiting,
// allow lists and audit logs depend on. Behind a proxy, either enable "trust
// proxy" or put the realip middleware in front, which rewrites RemoteAddr after
// validating the hop chain.
func (req *Request) IP() string {
	if req.TrustProxy() {
		if f := req.Raw.Header.Get("X-Forwarded-For"); f != "" {
			if first := strings.TrimSpace(strings.Split(f, ",")[0]); first != "" {
				return first
			}
		}
	}
	return remoteHost(req.Raw.RemoteAddr)
}

// IPs returns the addresses in the X-Forwarded-For header, left to right
// (client first, nearest proxy last), mirroring Express's req.ips. It returns
// nil unless "trust proxy" is enabled, since the header is otherwise untrusted.
func (req *Request) IPs() []string {
	if !req.TrustProxy() {
		return nil
	}
	f := req.Raw.Header.Get("X-Forwarded-For")
	if f == "" {
		return nil
	}
	parts := strings.Split(f, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// remoteHost strips the port from an http.Request.RemoteAddr value, handling
// "1.2.3.4:80", "[::1]:80", a bracketless IPv6 literal, and a bare host.
func remoteHost(addr string) string {
	if addr == "" {
		return addr
	}
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return stripHostPort(addr)
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
