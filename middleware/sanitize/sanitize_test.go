package sanitize

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestStripTags(t *testing.T) {
	cases := map[string]string{
		"<script>alert(1)</script>hi": "alert(1)hi",
		"plain":                       "plain",
		"<b>bold</b>":                 "bold",
	}
	for in, want := range cases {
		if got := StripTags(in); got != want {
			t.Errorf("StripTags(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSanitizesQuery(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		got = req.Query("name")
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/?name=%3Cscript%3Ehi%3C%2Fscript%3E", nil))

	if got != "hi" {
		t.Fatalf("query name = %q, want hi", got)
	}
}
