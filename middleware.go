package express

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// JSON returns middleware that parses JSON request bodies into a
// map[string]any (or []any) and stores the result on req via SetBody. It only
// acts on requests whose Content-Type is JSON.
func JSON() Handler {
	return func(req *Request, res *Response, next Next) {
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
		if len(data) == 0 {
			req.SetBody(map[string]any{})
			next()
			return
		}
		var parsed any
		if err := json.Unmarshal(data, &parsed); err != nil {
			next(err)
			return
		}
		req.SetBody(parsed)
		next()
	}
}

// URLEncoded returns middleware that parses application/x-www-form-urlencoded
// request bodies and stores them on req as url.Values.
func URLEncoded() Handler {
	return func(req *Request, res *Response, next Next) {
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
		values, err := url.ParseQuery(string(data))
		if err != nil {
			next(err)
			return
		}
		req.SetBody(values)
		next()
	}
}

// Static returns middleware that serves files from root. Requests that do not
// resolve to an existing file fall through to the next handler.
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
		if !strings.HasPrefix(full, root) {
			next()
			return
		}
		info, err := os.Stat(full)
		if err != nil || info.IsDir() {
			// Try an index.html for directory requests.
			if err == nil && info.IsDir() {
				idx := filepath.Join(full, "index.html")
				if _, e := os.Stat(idx); e == nil {
					http.ServeFile(res.Writer, req.Raw, idx)
					res.written = true
					return
				}
			}
			next()
			return
		}
		res.written = true
		http.ServeFile(res.Writer, req.Raw, full)
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
