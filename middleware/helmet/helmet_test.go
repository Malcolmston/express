package helmet

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func serve(t *testing.T, h express.Handler, tweak func(*express.Application)) *httptest.ResponseRecorder {
	t.Helper()
	app := express.New()
	if tweak != nil {
		tweak(app)
	}
	app.Use(h)
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	return rec
}

func TestDefaults(t *testing.T) {
	rec := serve(t, New(), func(app *express.Application) {
		app.Disable("x-powered-by")
	})
	checks := map[string]string{
		"X-Content-Type-Options":            "nosniff",
		"X-Frame-Options":                   "SAMEORIGIN",
		"Strict-Transport-Security":         "max-age=15552000",
		"Referrer-Policy":                   "no-referrer",
		"X-DNS-Prefetch-Control":            "off",
		"X-Permitted-Cross-Domain-Policies": "none",
		"Origin-Agent-Cluster":              "?1",
	}
	for k, want := range checks {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	if got := rec.Header().Get("X-Powered-By"); got != "" {
		t.Errorf("X-Powered-By = %q, want empty", got)
	}
}

func TestOptions(t *testing.T) {
	rec := serve(t, New(Options{
		HSTSMaxAge:            100,
		HSTSIncludeSubDomains: true,
		FrameguardAction:      "DENY",
		ReferrerPolicy:        "same-origin",
	}), nil)

	if got := rec.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("X-Frame-Options = %q", got)
	}
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=100; includeSubDomains" {
		t.Errorf("HSTS = %q", got)
	}
	if got := rec.Header().Get("Referrer-Policy"); got != "same-origin" {
		t.Errorf("Referrer-Policy = %q", got)
	}
}
