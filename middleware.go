package express

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/malcolmston/express/httperrors"
)

// JSON returns middleware that parses JSON request bodies into a
// map[string]any (or []any) and stores the result on req via SetBody. It only
// acts on requests whose Content-Type is JSON.
//
// It mirrors express.json(): every request has req.Body() defaulted to an empty
// object (so a handler never observes a nil body), strict mode rejects a
// top-level scalar such as `42` with a 400, and a syntactically invalid body is
// reported to the error chain as a 400 (not a 500). Those HTTP statuses are
// carried on the error via the httperrors package.
func JSON() Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Body() == nil {
			req.SetBody(map[string]any{})
		}
		if !req.Is("json") {
			next()
			return
		}
		data, err := io.ReadAll(req.Raw.Body)
		if err != nil {
			next(err)
			return
		}
		req.Raw.Body.Close()
		trimmed := bytes.TrimLeft(data, " \t\r\n")
		if len(trimmed) == 0 {
			next()
			return
		}
		// strict mode: only objects and arrays are valid at the top level.
		if trimmed[0] != '{' && trimmed[0] != '[' {
			next(httperrors.New(http.StatusBadRequest, "invalid json, only supports object and array"))
			return
		}
		var parsed any
		if err := json.Unmarshal(data, &parsed); err != nil {
			next(httperrors.New(http.StatusBadRequest, "invalid json"))
			return
		}
		req.SetBody(parsed)
		next()
	}
}

// URLEncoded returns middleware that parses application/x-www-form-urlencoded
// request bodies. Like express.urlencoded() (with the default extended:false
// querystring parser) a key seen once maps to a string and a repeated key maps
// to an array of strings; an invalid percent escape is kept verbatim rather than
// aborting the request. The result is stored on req as a map[string]any.
func URLEncoded() Handler {
	return func(req *Request, res *Response, next Next) {
		if req.Body() == nil {
			req.SetBody(map[string]any{})
		}
		if !req.Is("urlencoded") {
			next()
			return
		}
		data, err := io.ReadAll(req.Raw.Body)
		if err != nil {
			next(err)
			return
		}
		req.Raw.Body.Close()
		req.SetBody(parseURLEncoded(string(data)))
		next()
	}
}

// parseURLEncoded is a lenient application/x-www-form-urlencoded parser matching
// Node's querystring: single-valued keys become strings, repeated keys become
// []any, and a value that fails to percent-decode is preserved verbatim.
func parseURLEncoded(raw string) map[string]any {
	grouped := map[string][]string{}
	var order []string
	for _, pair := range strings.Split(raw, "&") {
		if pair == "" {
			continue
		}
		key, val := pair, ""
		if i := strings.IndexByte(pair, '='); i >= 0 {
			key, val = pair[:i], pair[i+1:]
		}
		key = ueUnescape(key)
		val = ueUnescape(val)
		if _, seen := grouped[key]; !seen {
			order = append(order, key)
		}
		grouped[key] = append(grouped[key], val)
	}
	out := make(map[string]any, len(order))
	for _, k := range order {
		vs := grouped[k]
		if len(vs) == 1 {
			out[k] = vs[0]
			continue
		}
		arr := make([]any, len(vs))
		for i, v := range vs {
			arr[i] = v
		}
		out[k] = arr
	}
	return out
}

// ueUnescape decodes one application/x-www-form-urlencoded token, treating '+'
// as a space and leaving a malformed percent escape untouched.
func ueUnescape(s string) string {
	s = strings.ReplaceAll(s, "+", " ")
	if u, err := url.QueryUnescape(s); err == nil {
		return u
	}
	return s
}

// Static returns middleware that serves files from root. Requests that do not
// resolve to an existing file fall through to the next handler.
//
// It mirrors serve-static's observable behaviour: a "Cache-Control:
// public, max-age=0" header on every hit, a directory request served from its
// index.html (without the 301 that http.ServeFile issues for /index.html),
// Content-Type resolved from the extension (with a charset for the text
// formats), Accept-Ranges plus Range and conditional-GET handling, and
// fall-through to the next handler for a missing file or a non-GET/HEAD method.
func Static(root string) Handler {
	root = filepath.Clean(root)
	return func(req *Request, res *Response, next Next) {
		if req.Method() != "GET" && req.Method() != "HEAD" {
			next()
			return
		}
		// Guard against path traversal by cleaning and confining to root.
		rel := filepath.Clean("/" + req.path)
		full := filepath.Join(root, rel)
		if full != root && !strings.HasPrefix(full, root+string(filepath.Separator)) {
			next()
			return
		}
		info, err := os.Stat(full)
		if err != nil {
			next()
			return
		}
		if info.IsDir() {
			// A directory request is served from its index.html, if present.
			idx := filepath.Join(full, "index.html")
			fi, e := os.Stat(idx)
			if e != nil || fi.IsDir() {
				next()
				return
			}
			full, info = idx, fi
		}
		f, err := os.Open(full)
		if err != nil {
			next()
			return
		}
		defer f.Close()

		res.Set("Cache-Control", "public, max-age=0")
		res.Set("Accept-Ranges", "bytes")
		if ct := mimeByExt(full); ct != "" {
			res.Set("Content-Type", ct)
		}
		res.written = true
		http.ServeContent(res.Writer, req.Raw, filepath.Base(full), info.ModTime(), f)
	}
}

// Logger returns middleware that logs each request's method, path, status, and
// duration to the standard logger, in a concise Express/morgan-like format.
func Logger() Handler {
	return func(req *Request, res *Response, next Next) {
		start := time.Now()
		method := req.Method()
		path := req.Path()
		next()
		log.Printf("%s %s %d - %s", method, path, res.statusCode, time.Since(start).Round(time.Microsecond))
	}
}

// Recover returns middleware that recovers from panics in downstream handlers
// and converts them into a 500 response, keeping the server alive.
func Recover() Handler {
	return func(req *Request, res *Response, next Next) {
		defer func() {
			if r := recover(); r != nil {
				if !res.written {
					res.Status(http.StatusInternalServerError).Send("Internal Server Error")
				}
				log.Printf("express: recovered from panic: %v", r)
			}
		}()
		next()
	}
}
