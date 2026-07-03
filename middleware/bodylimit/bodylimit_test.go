package bodylimit

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestBodyLimitRejectsOversized(t *testing.T) {
	app := express.New()
	app.Use(New(Options{MaxBytes: 4}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("way too long"))
	app.ServeHTTP(rr, r)
	if rr.Code != 413 {
		t.Fatalf("status = %d, want 413", rr.Code)
	}
}

func TestBodyLimitAllowsWithinLimit(t *testing.T) {
	app := express.New()
	app.Use(New(Options{MaxBytes: 64}))
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		res.Send(string(data))
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("small"))
	app.ServeHTTP(rr, r)
	if rr.Code != 200 || rr.Body.String() != "small" {
		t.Fatalf("code=%d body=%q", rr.Code, rr.Body.String())
	}
}

func TestBodyLimitMaxBytesReaderCaps(t *testing.T) {
	// Content-Length is within limit but reading a larger body fails.
	app := express.New()
	app.Use(New(Options{MaxBytes: 3}))
	var readErr error
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		_, readErr = io.ReadAll(req.Raw.Body)
		res.Send("done")
	})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("abcdef"))
	r.ContentLength = -1 // unknown length, so the Content-Length gate is bypassed
	app.ServeHTTP(rr, r)
	if readErr == nil {
		t.Fatalf("expected read error from MaxBytesReader, got nil")
	}
}
