package slowlog

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/malcolmston/express"
)

func TestLogsSlowRequest(t *testing.T) {
	var buf bytes.Buffer
	app := express.New()
	app.Use(New(Options{Threshold: time.Millisecond, Logger: log.New(&buf, "", 0)}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		time.Sleep(10 * time.Millisecond)
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !strings.Contains(buf.String(), "WARNING") {
		t.Fatalf("expected a slow warning, got %q", buf.String())
	}
}

func TestFastRequestNotLogged(t *testing.T) {
	var buf bytes.Buffer
	app := express.New()
	app.Use(New(Options{Threshold: time.Second, Logger: log.New(&buf, "", 0)}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if buf.Len() != 0 {
		t.Fatalf("expected no log for fast request, got %q", buf.String())
	}
}
