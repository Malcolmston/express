package express

import (
	"encoding/json"
	"strings"
)

// JSONP sends a JSONP response: the value v serialised as JSON and wrapped in a
// call to the callback function named by the request's query parameter (default
// "callback"). When no callback parameter is present it falls back to a normal
// JSON response. The body is prefixed with `/**/` and the content type is set to
// application/javascript, mirroring Express's res.jsonp.
func (res *Response) JSONP(v any) *Response {
	cb := res.req.Query("callback")
	if cb == "" {
		return res.JSON(v)
	}
	// Restrict the callback name to a safe identifier to avoid script injection.
	cb = expressSanitizeCallback(cb)
	if cb == "" {
		return res.JSON(v)
	}
	b, err := json.Marshal(v)
	if err != nil {
		b = []byte("null")
	}
	// Escape line separators that are valid in JSON but not in JavaScript.
	body := strings.NewReplacer("\u2028", "\\u2028", "\u2029", "\\u2029").Replace(string(b))
	if res.GetHeader("Content-Type") == "" {
		res.Type("application/javascript")
	}
	res.Set("X-Content-Type-Options", "nosniff")
	return res.Send("/**/ typeof " + cb + " === 'function' && " + cb + "(" + body + ");")
}

// expressSanitizeCallback returns name if it is a safe JavaScript callback
// identifier (letters, digits, _, $, and dotted/bracketed access), else "".
func expressSanitizeCallback(name string) string {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '_' || r == '$' || r == '.' || r == '[' || r == ']':
		default:
			return ""
		}
	}
	return name
}

// Links sets the Link response header from a map of rel -> URL, as in Express's
// res.links. For example {"next": "http://…?page=2"} becomes
// `<http://…?page=2>; rel="next"`.
func (res *Response) Links(links map[string]string) *Response {
	parts := make([]string, 0, len(links))
	// Deterministic ordering for stable output.
	for _, rel := range expressSortedKeys(links) {
		parts = append(parts, "<"+links[rel]+`>; rel="`+rel+`"`)
	}
	if existing := res.GetHeader("Link"); existing != "" {
		parts = append([]string{existing}, parts...)
	}
	return res.Set("Link", strings.Join(parts, ", "))
}

// CacheControl sets the Cache-Control response header to directive and returns
// res for chaining.
func (res *Response) CacheControl(directive string) *Response {
	return res.Set("Cache-Control", directive)
}

// RemoveHeader deletes the named response header and returns res for chaining.
func (res *Response) RemoveHeader(field string) *Response {
	res.Writer.Header().Del(field)
	return res
}

func expressSortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// simple insertion sort (small maps)
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j-1] > keys[j]; j-- {
			keys[j-1], keys[j] = keys[j], keys[j-1]
		}
	}
	return keys
}

// Subdomains returns the subdomain labels of the request Host in reverse order,
// excluding the last `offset` labels (default 2 — the domain and TLD). For
// "tobi.ferrets.example.com" it returns ["ferrets", "tobi"]. It mirrors
// Express's req.subdomains.
func (req *Request) Subdomains(offset ...int) []string {
	off := 2
	if len(offset) > 0 && offset[0] >= 0 {
		off = offset[0]
	}
	host := req.Hostname()
	if host == "" {
		return nil
	}
	labels := strings.Split(host, ".")
	if len(labels) <= off {
		return nil
	}
	subs := labels[:len(labels)-off]
	// Reverse in place.
	for i, j := 0, len(subs)-1; i < j; i, j = i+1, j-1 {
		subs[i], subs[j] = subs[j], subs[i]
	}
	return subs
}

// Xhr reports whether the request was made with an XMLHttpRequest, detected via
// the X-Requested-With header (case-insensitive), mirroring Express's req.xhr.
func (req *Request) Xhr() bool {
	return strings.EqualFold(req.Get("X-Requested-With"), "XMLHttpRequest")
}

// Referrer returns the request's referrer URL, accepting either the correctly
// spelled "Referer" header or the "Referrer" variant.
func (req *Request) Referrer() string {
	if v := req.Get("Referer"); v != "" {
		return v
	}
	return req.Get("Referrer")
}

// UserAgent returns the request's User-Agent header.
func (req *Request) UserAgent() string { return req.Get("User-Agent") }

// BaseURL returns the scheme and host portion of the request, e.g.
// "https://example.com", without a trailing slash.
func (req *Request) BaseURL() string {
	host := req.Raw.Host
	if host == "" {
		host = req.Hostname()
	}
	return req.Protocol() + "://" + host
}

// ContentType returns the request's Content-Type without any parameters (e.g.
// "application/json" for "application/json; charset=utf-8").
func (req *Request) ContentType() string {
	ct := req.Get("Content-Type")
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.TrimSpace(ct)
}
