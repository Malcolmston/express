package downloadheader

import (
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func TestDownloadHeaderFilename(t *testing.T) {
	app := express.New()
	app.Use(New(Options{Filename: "report.pdf"}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Content-Disposition"); got != `attachment; filename="report.pdf"` {
		t.Fatalf("Content-Disposition = %q", got)
	}
}

func TestDownloadHeaderNoFilename(t *testing.T) {
	app := express.New()
	app.Use(New())
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, httptest.NewRequest("GET", "/", nil))
	if got := rr.Header().Get("Content-Disposition"); got != "attachment" {
		t.Fatalf("Content-Disposition = %q", got)
	}
}
