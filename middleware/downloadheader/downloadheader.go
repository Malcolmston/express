// Package downloadheader provides express middleware that marks responses as
// file downloads by setting the Content-Disposition response header to
// "attachment". It mirrors the behavior of Express's res.attachment() /
// res.download() helpers and the Content-Disposition disposition used by the
// Node "content-disposition" package, but applies it uniformly to every
// response that passes through the middleware rather than at a single call
// site.
//
// Use this middleware when an entire route tree serves content that browsers
// should save to disk instead of rendering inline: report exporters, generated
// CSV/PDF endpoints, raw file blobs, or any API whose payloads are meant to be
// downloaded. Because it is a plain header-setting handler, mounting it with
// app.Use on a subtree is enough to make every response in that subtree a
// download; there is no per-response opt-in required.
//
// The middleware runs early in the chain and only ever writes a single header.
// On each request it calls res.Set("Content-Disposition", value) and then
// invokes next() unconditionally, so it never short-circuits, never inspects
// the request, and never touches the body or status code. Downstream handlers
// remain in full control of the payload; setting the header before next() means
// it is already present when the eventual handler flushes the response. If a
// later handler overwrites Content-Disposition, that later value wins, since
// this middleware sets the header only once, up front.
//
// The header value is computed once, when New is called, from the optional
// Options. With no Filename a bare "attachment" is sent. With a Filename the
// value becomes attachment; filename="<name>", where any backslashes and double
// quotes in the name are escaped (\ -> \\ and " -> \") so the quoted-string is
// well formed. Note that only the legacy quoted filename parameter is emitted:
// unlike the Node content-disposition package, this port does not add a
// RFC 5987 filename* parameter, so non-ASCII characters in the filename are
// passed through verbatim inside the quoted string rather than percent-encoded.
//
// Regarding parity with the Node original: the download semantics
// (attachment disposition plus an optional suggested filename) match
// res.attachment(filename) in Express, and quote/backslash escaping matches the
// content-disposition module's handling of the quoted form. The differences are
// that the filename is fixed at construction time rather than chosen per
// response, and that extended (filename*) encoding for international filenames
// is intentionally omitted to keep the port dependency-free. It performs no
// content-type inference, so callers that need a matching Content-Type must set
// it themselves.
package downloadheader

import (
	"strings"

	"github.com/malcolmston/express"
)

// Options configures the middleware.
type Options struct {
	// Filename is the suggested download filename. When empty, a bare
	// "attachment" disposition is sent.
	Filename string
}

// New returns middleware that sets Content-Disposition: attachment so browsers
// offer to save the response rather than render it. If a Filename is given it
// is included (with quotes escaped) as the suggested name.
func New(opts ...Options) express.Handler {
	var filename string
	if len(opts) > 0 {
		filename = opts[0].Filename
	}
	value := "attachment"
	if filename != "" {
		value = `attachment; filename="` + escapeQuotes(filename) + `"`
	}
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Content-Disposition", value)
		next()
	}
}

func escapeQuotes(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
