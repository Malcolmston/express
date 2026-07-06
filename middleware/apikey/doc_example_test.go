package apikey_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/apikey"
)

// ExampleNew demonstrates guarding routes behind a static list of API keys. The
// middleware is mounted with app.Use so every request must present a valid key
// before reaching the handler, accepting it either from the default "X-API-Key"
// header or, because Query is set, from an "api_key" query parameter. The first
// request supplies a valid header key and receives 200; the second omits any
// key and is short-circuited with 401 before the handler runs. Both status
// codes are deterministic, so an Output block asserts them.
func ExampleNew() {
	app := express.New()
	app.Use(apikey.New(apikey.Options{
		Query: "api_key",
		Keys:  []string{"key1", "key2"},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	authed := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-API-Key", "key1")
	app.ServeHTTP(authed, r)

	anon := httptest.NewRecorder()
	app.ServeHTTP(anon, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(authed.Code, anon.Code)
	// Output: 200 401
}
