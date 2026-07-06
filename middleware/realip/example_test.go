package realip_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/realip"
)

// ExampleNew shows the middleware resolving a client address from behind a
// proxy chain. It is configured with a single trusted proxy so that resolution
// walks X-Forwarded-For from the right and skips the proxy hop, returning the
// originating client instead. The middleware is registered with app.Use ahead
// of the route so that both req.Raw.RemoteAddr and the realip.ClientIP accessor
// report the corrected address. A request is driven through httptest with an
// X-Forwarded-For header carrying "client, proxy", and the handler prints the
// resolved value. Because the resolution is deterministic the example asserts
// its Output.
func ExampleNew() {
	app := express.New()
	app.Use(realip.New(realip.Options{
		TrustedProxies: []string{"5.6.7.8"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Printf("client: %s\n", realip.ClientIP(req))
		fmt.Printf("remote: %s\n", req.Raw.RemoteAddr)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")
	app.ServeHTTP(rec, req)

	// Output:
	// client: 1.2.3.4
	// remote: 1.2.3.4
}
