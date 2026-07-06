package hidepoweredby_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/hidepoweredby"
)

// Example wires the hidepoweredby middleware into an express application and
// drives it with net/http/httptest. It passes a SetTo decoy value so the
// middleware replaces X-Powered-By with a spoofed stack instead of deleting it,
// which keeps the emitted header deterministic. The route handler deliberately
// sets a real X-Powered-By value to prove the before-write hook overrides
// whatever downstream handlers or the framework produced. After serving the
// request we read the header back off the recorder to confirm the decoy won.
// Leaving Options empty instead (hidepoweredby.New()) would delete the header
// entirely rather than spoof it.
func Example() {
	app := express.New()
	app.Use(hidepoweredby.New(hidepoweredby.Options{SetTo: "PHP/8.0"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("X-Powered-By", "Express")
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	fmt.Println(rec.Header().Get("X-Powered-By"))
	// Output:
	// PHP/8.0
}
