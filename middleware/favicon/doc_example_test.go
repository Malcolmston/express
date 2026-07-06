package favicon_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/favicon"
)

// Example demonstrates how the favicon middleware short-circuits /favicon.ico
// requests while letting everything else fall through. It builds an express app,
// mounts favicon.New with a tiny embedded icon and a cache lifetime, and adds a
// catch-all handler that would answer any other path. Two requests are then
// driven with net/http/httptest: one for /favicon.ico, which the middleware
// serves directly with a 200 status and the configured image content type, and
// one for /index.html, which passes through to the catch-all handler. The
// deterministic status code, Content-Type, and pass-through body are asserted in
// the Output comment below.
func Example() {
	app := express.New()
	app.Use(favicon.New(favicon.Options{
		Data:   []byte{0x00, 0x00, 0x01, 0x00},
		MaxAge: 3600,
	}))
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("app")
	})

	// Favicon request: handled and terminated by the middleware.
	icon := httptest.NewRecorder()
	app.ServeHTTP(icon, httptest.NewRequest(http.MethodGet, favicon.Path, nil))
	fmt.Println(icon.Code, icon.Header().Get("Content-Type"))

	// Any other path: passed through to the downstream handler.
	page := httptest.NewRecorder()
	app.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/index.html", nil))
	fmt.Println(page.Body.String())

	// Output:
	// 200 image/x-icon
	// app
}
