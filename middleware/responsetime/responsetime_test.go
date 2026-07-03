package responsetime

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/malcolmston/express"
)

func TestResponseTimeHeader(t *testing.T) {
	var cur time.Time
	now = func() time.Time { return cur }
	defer func() { now = time.Now }()

	cur = time.Unix(0, 0)
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		cur = cur.Add(5 * time.Millisecond) // simulate 5ms of work
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))

	got := rr.Header().Get(HeaderName)
	if !strings.HasSuffix(got, "ms") {
		t.Fatalf("X-Response-Time = %q, want ms suffix", got)
	}
	if got != "5.00ms" {
		t.Fatalf("X-Response-Time = %q, want 5.00ms", got)
	}
}
