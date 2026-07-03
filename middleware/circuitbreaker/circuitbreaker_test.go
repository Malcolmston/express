package circuitbreaker_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/middleware/circuitbreaker"
)

func TestTripsAfterThreshold(t *testing.T) {
	cur := time.Unix(0, 0)
	fail := true
	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 3,
		Cooldown:  time.Minute,
		Now:       func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if fail {
			res.Status(500).Send("boom")
			return
		}
		res.Send("ok")
	})

	code := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	// 3 consecutive 5xx from the handler.
	for i := 0; i < 3; i++ {
		if code() != 500 {
			t.Fatalf("request %d expected handler 500", i)
		}
	}
	// Circuit now open: short-circuited 503 without reaching handler.
	fail = false // even though handler would now succeed, circuit is open
	if got := code(); got != 503 {
		t.Fatalf("expected open circuit 503, got %d", got)
	}
}

func TestHalfOpenRecovery(t *testing.T) {
	cur := time.Unix(0, 0)
	fail := true
	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 2,
		Cooldown:  30 * time.Second,
		Now:       func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if fail {
			res.Status(500).Send("boom")
			return
		}
		res.Send("ok")
	})
	code := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}

	code()
	code() // trips open
	if got := code(); got != 503 {
		t.Fatalf("expected open 503, got %d", got)
	}
	// Advance past cooldown -> half-open allows a trial; handler now succeeds.
	cur = cur.Add(31 * time.Second)
	fail = false
	if got := code(); got != 200 {
		t.Fatalf("expected recovery 200, got %d", got)
	}
	// Circuit closed again.
	if got := code(); got != 200 {
		t.Fatalf("expected closed 200, got %d", got)
	}
}

func TestHalfOpenReopensOnFailure(t *testing.T) {
	cur := time.Unix(0, 0)
	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 1,
		Cooldown:  10 * time.Second,
		Now:       func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(500).Send("boom")
	})
	code := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}
	code() // trips open
	if got := code(); got != 503 {
		t.Fatalf("expected open 503, got %d", got)
	}
	cur = cur.Add(11 * time.Second) // half-open, trial runs and fails
	if got := code(); got != 500 {
		t.Fatalf("expected trial 500, got %d", got)
	}
	// Should be open again immediately.
	if got := code(); got != 503 {
		t.Fatalf("expected re-open 503, got %d", got)
	}
}

func TestSuccessResetsFailures(t *testing.T) {
	cur := time.Unix(0, 0)
	fail := false
	app := express.New()
	app.Use(circuitbreaker.New(circuitbreaker.Options{
		Threshold: 2,
		Cooldown:  time.Minute,
		Now:       func() time.Time { return cur },
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		if fail {
			res.Status(500).Send("boom")
			return
		}
		res.Send("ok")
	})
	code := func() int {
		r := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, r)
		return w.Code
	}
	fail = true
	code() // 1 failure
	fail = false
	code() // success resets
	fail = true
	code() // 1 failure again, below threshold
	if got := code(); got != 500 {
		// still passing through (handler 500), not open
		t.Fatalf("expected handler 500 (not open), got %d", got)
	}
}
