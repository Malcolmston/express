package express

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/malcolmston/express/encodeurl"
	"github.com/malcolmston/express/etag"
	"github.com/malcolmston/express/statuses"
)

// Response wraps an http.ResponseWriter and provides chainable Express-style
// helpers for setting status codes, headers, and writing bodies.
type Response struct {
	// Writer is the underlying standard-library response writer.
	Writer http.ResponseWriter

	req        *Request
	app        *Application
	statusCode int
	written    bool
	// invalidStatus is set by writeHeaderOnce when the requested status code
	// falls outside the valid [100, 999] range. Express raises a RangeError in
	// that case which its error handler turns into a 500 "Invalid status code"
	// response; the port reproduces that outcome at commit time and uses this
	// flag to suppress any body the handler tried to write afterwards.
	invalidStatus bool
	// Locals holds values scoped to this request/response cycle, mirroring
	// Express's res.locals.
	Locals map[string]any

	// beforeWrite holds hooks run exactly once, just before the response
	// headers are committed. Middleware such as Session uses this to flush a
	// Set-Cookie header before the status line is written.
	beforeWrite []func()

	// next is the continuation for the handler currently executing. Response
	// helpers that must delegate to the middleware chain — notably res.Format
	// when nothing is acceptable — call it with an error. It is refreshed by the
	// router before each handler runs.
	next Next
}

// OnBeforeWrite registers a callback invoked once, immediately before the
// response headers are committed. Use it to set headers or cookies that depend
// on work done during the request (e.g. persisting a session).
func (res *Response) OnBeforeWrite(fn func()) {
	res.beforeWrite = append(res.beforeWrite, fn)
}

func newResponse(w http.ResponseWriter, req *Request, app *Application) *Response {
	res := &Response{
		Writer:     w,
		req:        req,
		app:        app,
		statusCode: http.StatusOK,
		Locals:     make(map[string]any),
	}
	// Seed request-cycle locals from application locals.
	for k, v := range app.locals {
		res.Locals[k] = v
	}
	return res
}

// Status sets the HTTP status code and returns the response for chaining.
func (res *Response) Status(code int) *Response {
	res.statusCode = code
	return res
}

// StatusCode returns the currently set status code.
func (res *Response) StatusCode() int { return res.statusCode }

// Set assigns a response header, returning the response for chaining.
func (res *Response) Set(field, value string) *Response {
	res.Writer.Header().Set(field, value)
	return res
}

// Append adds a value to an existing response header without replacing it.
func (res *Response) Append(field, value string) *Response {
	res.Writer.Header().Add(field, value)
	return res
}

// GetHeader returns a header value already set on the response.
func (res *Response) GetHeader(field string) string {
	return res.Writer.Header().Get(field)
}

// Type sets the Content-Type header. Common shorthands ("json", "html",
// "text") are expanded; anything containing a slash is used verbatim.
func (res *Response) Type(t string) *Response {
	res.Writer.Header().Set("Content-Type", normalizeContentType(t))
	return res
}

// Vary adds a field to the Vary response header.
func (res *Response) Vary(field string) *Response {
	return res.Append("Vary", field)
}

// Location sets the Location response header.
//
// The URL is percent-encoded with encodeurl before being written, mirroring
// Express's res.location: characters that are unsafe or invalid in a URL are
// escaped while already-encoded "%xx" sequences are preserved, so a value such
// as "https://example.com?q=☃ §10" becomes
// "https://example.com?q=%E2%98%83%20%C2%A710". Values that already consist of
// safe characters (e.g. "http://google.com/") are returned unchanged.
func (res *Response) Location(url string) *Response {
	return res.Set("Location", encodeurl.Encode(url))
}

// writeHeaderOnce commits the status line exactly once.
func (res *Response) writeHeaderOnce() {
	if res.written {
		return
	}
	// Run before-write hooks while headers are still mutable.
	for _, fn := range res.beforeWrite {
		fn()
	}
	res.beforeWrite = nil
	res.written = true
	if res.app != nil && res.app.Enabled("x-powered-by") && res.GetHeader("X-Powered-By") == "" {
		res.Writer.Header().Set("X-Powered-By", "Express")
	}
	// Express validates the status code and raises a RangeError for anything
	// outside [100, 999]; its default error handler renders that as a 500 with
	// an "Invalid status code" message. Reproduce that here so an out-of-range
	// Status(code) never emits a malformed status line.
	if res.statusCode < 100 || res.statusCode > 999 {
		res.invalidStatus = true
		res.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		res.Writer.WriteHeader(http.StatusInternalServerError)
		_, _ = res.Writer.Write([]byte(fmt.Sprintf("Invalid status code: %d", res.statusCode)))
		return
	}
	res.Writer.WriteHeader(res.statusCode)
}

// Send writes a response body. Strings and []byte are written as-is; any other
// value is serialized as JSON. It sets a default Content-Type when none is set.
func (res *Response) Send(body any) *Response {
	switch v := body.(type) {
	case string:
		if res.GetHeader("Content-Type") == "" {
			res.Type("html")
		}
		suppress := res.applyETag([]byte(v))
		res.writeHeaderOnce()
		if res.invalidStatus || suppress {
			return res
		}
		_, _ = res.Writer.Write([]byte(v))
	case []byte:
		if res.GetHeader("Content-Type") == "" {
			res.Type("application/octet-stream")
		}
		suppress := res.applyETag(v)
		res.writeHeaderOnce()
		if res.invalidStatus || suppress {
			return res
		}
		_, _ = res.Writer.Write(v)
	case nil:
		res.writeHeaderOnce()
	default:
		return res.JSON(v)
	}
	return res
}

// JSON serializes v as JSON and writes it with an application/json Content-Type.
func (res *Response) JSON(v any) *Response {
	if res.GetHeader("Content-Type") == "" {
		res.Type("json")
	}
	data, err := marshalJSON(v)
	if err != nil {
		res.finalError(err)
		return res
	}
	suppress := res.applyETag(data)
	res.writeHeaderOnce()
	if res.invalidStatus || suppress {
		return res
	}
	_, _ = res.Writer.Write(data)
	return res
}

// applyETag mirrors the ETag + freshness handling Express performs inside
// res.send/res.json: for a GET or HEAD request it derives a default weak ETag
// from the body (unless one was set explicitly), then — if the request is fresh
// against that validator (If-None-Match / If-Modified-Since) — downgrades the
// response to a bodyless 304 Not Modified. It returns true when the caller must
// suppress the body.
func (res *Response) applyETag(body []byte) bool {
	if res.req == nil {
		return false
	}
	m := res.req.Method()
	isGetHead := strings.EqualFold(m, "GET") || strings.EqualFold(m, "HEAD")
	if !isGetHead {
		return false
	}
	if res.GetHeader("ETag") == "" && len(body) > 0 {
		res.Set("ETag", etag.Generate(body, true))
	}
	if res.statusCode == http.StatusOK && res.req.Fresh(res) {
		res.statusCode = http.StatusNotModified
		h := res.Writer.Header()
		h.Del("Content-Type")
		h.Del("Content-Length")
		return true
	}
	return false
}

// SendStatus sets the status code and sends its standard text as the body.
//
// The body text comes from the `statuses` message table (Express uses the same
// `statuses` module), which differs from net/http.StatusText in casing for a
// few codes — e.g. 418 is "I'm a Teapot", not "I'm a teapot".
func (res *Response) SendStatus(code int) *Response {
	res.Status(code)
	body := statuses.Message(code)
	if body == "" {
		body = strconv.Itoa(code)
	}
	return res.Type("text").Send(body)
}

// End finalizes the response with no (further) body.
func (res *Response) End() {
	res.writeHeaderOnce()
}

// Redirect sends a redirect response. If a status code is provided it is used;
// otherwise 302 Found is assumed.
//
// Like Express's res.redirect the body is negotiated from the request's Accept
// header: clients that prefer HTML get an escaped anchor, everything else gets
// plain text. The Content-Type is always set explicitly, which matters because
// the destination URL is echoed into the body — leaving it unset would let
// net/http sniff a redirect to "<script>…" as text/html and turn an open
// redirect into a script-injection sink.
func (res *Response) Redirect(args ...any) *Response {
	code := http.StatusFound
	var location string
	switch len(args) {
	case 1:
		location, _ = args[0].(string)
	case 2:
		if c, ok := args[0].(int); ok {
			code = c
		}
		location, _ = args[1].(string)
	default:
		res.finalError(fmt.Errorf("express: Redirect requires (url) or (status, url)"))
		return res
	}
	res.Location(location)
	res.Status(code)

	// res.redirect negotiates the courtesy body through res.format, which always
	// adds "Vary: Accept". The body text is prefixed with the status phrase, e.g.
	// "Found. Redirecting to /target".
	res.Vary("Accept")
	phrase := statuses.Message(code)
	encoded := res.GetHeader("Location")
	body := phrase + ". Redirecting to " + encoded
	if res.GetHeader("Content-Type") == "" {
		// Offer order matches Express's res.redirect (text first, then html), so
		// a client that expresses no preference gets the plain-text body.
		if res.req != nil && res.req.Accepts("text", "html") == "html" {
			esc := html.EscapeString(encoded)
			body = "<p>" + phrase + ` Redirecting to <a href="` + esc + `">` + esc + `</a></p>`
			res.Type("html")
		} else {
			res.Type("text")
		}
	}

	res.writeHeaderOnce()
	if res.invalidStatus {
		return res
	}
	// A HEAD response must not carry a body.
	if res.req != nil && strings.EqualFold(res.req.Method(), "HEAD") {
		return res
	}
	_, _ = res.Writer.Write([]byte(body))
	return res
}

// Cookie sets a Set-Cookie header. opts may be nil for a session cookie.
//
// The serialization mirrors Express (the `cookie` module): attributes are
// emitted in the order Max-Age, Domain, Path, Expires, HttpOnly, Secure,
// SameSite, and — as Express does — a non-zero MaxAge additionally sets an
// Expires derived from it.
func (res *Response) Cookie(name, value string, opts *CookieOptions) *Response {
	sc := setCookie{name: name, value: url.QueryEscape(value), path: "/"}
	if opts != nil {
		if opts.Path != "" {
			sc.path = opts.Path
		}
		sc.domain = opts.Domain
		if opts.MaxAge != 0 {
			sc.maxAge, sc.hasMaxAge = opts.MaxAge, true
			sc.expires, sc.hasExpires = time.Now().Add(time.Duration(opts.MaxAge)*time.Second), true
		}
		sc.httpOnly = opts.HTTPOnly
		sc.secure = opts.Secure
		sc.sameSite = opts.SameSite
	}
	res.Append("Set-Cookie", sc.serialize())
	return res
}

// ClearCookie expires a cookie on the client. Like Express it sends an Expires
// in the past (the Unix epoch) rather than Max-Age=0.
func (res *Response) ClearCookie(name string) *Response {
	sc := setCookie{name: name, value: "", path: "/", expires: time.Unix(0, 0), hasExpires: true}
	res.Append("Set-Cookie", sc.serialize())
	return res
}

// setCookie assembles a Set-Cookie header value in Express's attribute order.
type setCookie struct {
	name, value   string
	path, domain  string
	maxAge        int
	hasMaxAge     bool
	expires       time.Time
	hasExpires    bool
	httpOnly      bool
	secure        bool
	sameSite      http.SameSite
}

func (sc setCookie) serialize() string {
	var b strings.Builder
	b.WriteString(sc.name)
	b.WriteByte('=')
	b.WriteString(sc.value)
	if sc.hasMaxAge {
		b.WriteString("; Max-Age=")
		b.WriteString(strconv.Itoa(sc.maxAge))
	}
	if sc.domain != "" {
		b.WriteString("; Domain=")
		b.WriteString(sc.domain)
	}
	if sc.path != "" {
		b.WriteString("; Path=")
		b.WriteString(sc.path)
	}
	if sc.hasExpires {
		b.WriteString("; Expires=")
		b.WriteString(sc.expires.UTC().Format(http.TimeFormat))
	}
	if sc.httpOnly {
		b.WriteString("; HttpOnly")
	}
	if sc.secure {
		b.WriteString("; Secure")
	}
	switch sc.sameSite {
	case http.SameSiteLaxMode:
		b.WriteString("; SameSite=Lax")
	case http.SameSiteStrictMode:
		b.WriteString("; SameSite=Strict")
	case http.SameSiteNoneMode:
		b.WriteString("; SameSite=None")
	}
	return b.String()
}

// CookieOptions configures a cookie set via res.Cookie.
type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HTTPOnly bool
	SameSite http.SameSite
}

// Written reports whether the response headers have been committed.
func (res *Response) Written() bool { return res.written }

// finalError writes a 500 response for an otherwise unhandled error.
func (res *Response) finalError(err error) {
	if res.written {
		return
	}
	res.Status(http.StatusInternalServerError)
	res.Type("text")
	res.writeHeaderOnce()
	_, _ = res.Writer.Write([]byte(err.Error()))
}

func normalizeContentType(t string) string {
	if strings.Contains(t, "/") {
		if !strings.Contains(strings.ToLower(t), "charset") {
			if cs := mimeCharset(t); cs != "" {
				return t + "; charset=" + cs
			}
		}
		return t
	}
	// A token carrying an explicit extension — a filename such as "foo.js",
	// "file.tar.gz" or ".json" — is resolved by its final extension, mirroring
	// Express's res.type, which routes such values through mime.contentType.
	// Bare tokens fall through to the shorthand table below and are otherwise
	// returned verbatim, matching the port's established behaviour.
	if strings.Contains(t, ".") {
		ext := t[strings.LastIndexByte(t, '.')+1:]
		return extContentType(ext)
	}
	switch t {
	case "json":
		return "application/json; charset=utf-8"
	case "html":
		return "text/html; charset=utf-8"
	case "text", "txt":
		return "text/plain; charset=utf-8"
	case "xml":
		return "application/xml; charset=utf-8"
	case "js", "javascript":
		return "application/javascript; charset=utf-8"
	case "css":
		return "text/css; charset=utf-8"
	default:
		return t
	}
}

// extContentType resolves a filename extension (without the leading dot) to a
// full Content-Type value, appending "; charset=utf-8" for the text-based
// formats that carry a charset. It mirrors the subset of the mime database that
// Express's res.type exercises; an unrecognised extension yields
// "application/octet-stream", exactly as mime.contentType returns for an
// unknown lookup.
func extContentType(ext string) string {
	switch strings.ToLower(ext) {
	case "json":
		return "application/json; charset=utf-8"
	case "js", "mjs", "cjs", "javascript":
		return "application/javascript; charset=utf-8"
	case "html", "htm":
		return "text/html; charset=utf-8"
	case "css":
		return "text/css; charset=utf-8"
	case "xml":
		return "application/xml; charset=utf-8"
	case "txt", "text":
		return "text/plain; charset=utf-8"
	case "csv":
		return "text/csv; charset=utf-8"
	case "md", "markdown":
		return "text/markdown; charset=utf-8"
	case "gz", "gzip", "tgz":
		return "application/gzip"
	case "tar":
		return "application/x-tar"
	case "zip":
		return "application/zip"
	case "pdf":
		return "application/pdf"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

// marshalJSON serialises v the way Express's res.json does — via the equivalent
// of JSON.stringify, which does NOT escape the HTML-sensitive bytes <, > and &.
// Go's encoding/json escapes them to </>/& by default, which would
// make an otherwise identical body differ from upstream, so HTML escaping is
// disabled here. Any escaping needed for safety is the caller's responsibility
// (e.g. res.jsonp pins X-Content-Type-Options: nosniff).
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	// Encoder.Encode appends a trailing newline; drop it to match json.Marshal.
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// mimeCharset mirrors the `mime-types` charset() lookup that Express's res.set
// applies to a Content-Type: text/* types and the handful of application/*
// types the mime database marks as UTF-8 carry an explicit charset. It returns
// "utf-8" for those and "" otherwise. The input may include parameters
// ("application/json; foo=bar"); only the media type is examined.
func mimeCharset(ct string) string {
	base := ct
	if i := strings.IndexByte(base, ';'); i >= 0 {
		base = base[:i]
	}
	base = strings.ToLower(strings.TrimSpace(base))
	if strings.HasPrefix(base, "text/") {
		return "utf-8"
	}
	switch base {
	case "application/json",
		"application/javascript",
		"application/xml",
		"application/rss+xml",
		"application/manifest+json",
		"application/ld+json",
		"application/x-www-form-urlencoded":
		return "utf-8"
	}
	if strings.HasPrefix(base, "application/") &&
		(strings.HasSuffix(base, "+json") || strings.HasSuffix(base, "+xml")) {
		return "utf-8"
	}
	return ""
}

// itoa is a tiny helper to avoid importing strconv at call sites.
func itoa(i int) string { return strconv.Itoa(i) }
