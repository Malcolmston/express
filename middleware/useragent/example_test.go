package useragent_test

import (
	"fmt"
	"net/http/httptest"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/useragent"
)

// Example wires the useragent middleware into an express application and drives
// it with a synthetic request through httptest. The middleware parses the
// incoming User-Agent header and stores the result on the request, so a later
// handler can recover it with useragent.From. Here the downstream handler
// reports the detected browser, operating system and mobile flag back in the
// response body. Because the parser is deterministic for a fixed input string,
// the output is stable and can be checked with an Output comment. This mirrors
// how a real backend might branch on coarse client categories.
func Example() {
	app := express.New()
	app.Use(useragent.New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		ua, _ := useragent.From(req)
		res.Send(fmt.Sprintf("browser=%s os=%s mobile=%t", ua.Browser, ua.OS, ua.Mobile))
	})

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 13; Pixel) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0 Mobile Safari/537.36")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Body.String())
	// Output: browser=Chrome os=Android mobile=true
}
