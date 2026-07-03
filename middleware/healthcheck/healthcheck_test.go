package healthcheck_test

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/healthcheck"
)

func TestAllHealthy(t *testing.T) {
	app := express.New()
	app.Use(healthcheck.New(healthcheck.Options{
		Checkers: map[string]func() error{
			"db":    func() error { return nil },
			"cache": func() error { return nil },
		},
	}))
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if out.Status != "ok" || out.Checks["db"] != "ok" {
		t.Fatalf("unexpected body: %+v", out)
	}
}

func TestUnhealthy(t *testing.T) {
	app := express.New()
	app.Use(healthcheck.New(healthcheck.Options{
		Checkers: map[string]func() error{
			"db": func() error { return errors.New("down") },
		},
	}))
	r := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatalf("expected 503, got %d", w.Code)
	}
	if !contains(w.Body.String(), "down") {
		t.Fatalf("expected error detail in body, got %q", w.Body.String())
	}
}

func TestFallThrough(t *testing.T) {
	app := express.New()
	app.Use(healthcheck.New(healthcheck.Options{Path: "/health"}))
	app.Get("/other", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("other")
	})
	r := httptest.NewRequest("GET", "/other", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "other" {
		t.Fatalf("expected fall-through, got %d %q", w.Code, w.Body.String())
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
