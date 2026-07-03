package express

import "net/http"

// WrapHandler adapts a standard net/http handler into an express Handler so it
// can be mounted with app.Use, e.g. to attach a Socket.IO server:
//
//	app.Use("/socket.io", express.WrapHandler(io)) // io is an http.Handler
//
// The wrapped handler receives the raw request (req.Raw) and response writer
// (res.Writer) — so it can hijack the connection for WebSocket upgrades — and is
// treated as terminal: express does not run later handlers for it.
func WrapHandler(h http.Handler) Handler {
	return func(req *Request, res *Response, next Next) {
		res.written = true
		h.ServeHTTP(res.Writer, req.Raw)
	}
}

// WrapHandlerFunc is WrapHandler for an http.HandlerFunc.
func WrapHandlerFunc(fn func(http.ResponseWriter, *http.Request)) Handler {
	return WrapHandler(http.HandlerFunc(fn))
}
