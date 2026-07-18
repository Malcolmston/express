package express

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/malcolmston/express/encodeurl"
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
		res.writeHeaderOnce()
		if res.invalidStatus {
			return res
		}
		_, _ = res.Writer.Write([]byte(v))
	case []byte:
		if res.GetHeader("Content-Type") == "" {
			res.Type("application/octet-stream")
		}
		res.writeHeaderOnce()
		if res.invalidStatus {
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
	data, err := json.Marshal(v)
	if err != nil {
		res.finalError(err)
		return res
	}
	res.writeHeaderOnce()
	if res.invalidStatus {
		return res
	}
	_, _ = res.Writer.Write(data)
	return res
}

// SendStatus sets the status code and sends its standard text as the body.
func (res *Response) SendStatus(code int) *Response {
	res.Status(code)
	return res.Type("text").Send(http.StatusText(code))
}

// End finalizes the response with no (further) body.
func (res *Response) End() {
	res.writeHeaderOnce()
}

// Redirect sends a redirect response. If a status code is provided it is used;
// otherwise 302 Found is assumed.
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
	res.writeHeaderOnce()
	if res.invalidStatus {
		return res
	}
	_, _ = res.Writer.Write([]byte("Redirecting to " + location))
	return res
}

// Cookie sets a Set-Cookie header. opts may be nil for a session cookie.
func (res *Response) Cookie(name, value string, opts *CookieOptions) *Response {
	c := &http.Cookie{Name: name, Value: url.QueryEscape(value), Path: "/"}
	if opts != nil {
		opts.apply(c)
	}
	http.SetCookie(res.Writer, c)
	return res
}

// ClearCookie expires a cookie on the client.
func (res *Response) ClearCookie(name string) *Response {
	c := &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1}
	http.SetCookie(res.Writer, c)
	return res
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

func (o *CookieOptions) apply(c *http.Cookie) {
	if o.Path != "" {
		c.Path = o.Path
	}
	c.Domain = o.Domain
	c.MaxAge = o.MaxAge
	c.Secure = o.Secure
	c.HttpOnly = o.HTTPOnly
	c.SameSite = o.SameSite
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
		if strings.HasPrefix(t, "text/") && !strings.Contains(t, "charset") {
			return t + "; charset=utf-8"
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

// itoa is a tiny helper to avoid importing strconv at call sites.
func itoa(i int) string { return strconv.Itoa(i) }
