package acceptlanguage_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/acceptlanguage"
)

// ExampleNew demonstrates negotiating a request's language against a fixed set
// of supported tags. The middleware is mounted with app.Use so the winning
// language is resolved before the handler runs, and the handler reads it back
// with req.Value(acceptlanguage.Key). The request advertises German first and
// French second, but German is not supported, so negotiation walks the
// preferences in quality order and settles on "fr". Because the supported set
// and header are fixed, the chosen language is deterministic and an Output
// block asserts it.
func ExampleNew() {
	app := express.New()
	app.Use(acceptlanguage.New(acceptlanguage.Options{
		Supported: []string{"en", "fr"},
		Default:   "en",
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		lang, _ := req.Value(acceptlanguage.Key)
		res.Send(fmt.Sprintf("%v", lang))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "de,fr;q=0.9")
	app.ServeHTTP(rec, req)

	fmt.Println(rec.Body.String())
	// Output: fr
}
