package accesslog

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
)

func TestLogsLine(t *testing.T) {
	var buf bytes.Buffer
	app := express.New()
	app.Use(New(Options{Writer: &buf}))
	app.Get("/hello", func(req *express.Request, res *express.Response, next express.Next) {
		res.Status(201).Send("hi there")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	req.Header.Set("User-Agent", "test-agent")
	app.ServeHTTP(rec, req)

	line := buf.String()
	if !strings.Contains(line, "GET /hello") {
		t.Errorf("log missing request line: %q", line)
	}
	if !strings.Contains(line, " 201 ") {
		t.Errorf("log missing status 201: %q", line)
	}
	if !strings.Contains(line, "8") { // "hi there" == 8 bytes
		t.Errorf("log missing byte count: %q", line)
	}
	if !strings.Contains(line, "test-agent") {
		t.Errorf("log missing user-agent: %q", line)
	}
}
