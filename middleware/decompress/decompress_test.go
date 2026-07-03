package decompress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/malcolmston/express"
)

func gzipBytes(t *testing.T, s string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDecompressGzip(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		got = string(data)
		res.Send("ok")
	})

	body := bytes.NewReader(gzipBytes(t, "hello compressed world"))
	r := httptest.NewRequest("POST", "/", body)
	r.Header.Set("Content-Encoding", "gzip")
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)

	if got != "hello compressed world" {
		t.Fatalf("decompressed body = %q", got)
	}
}

func TestDecompressPassthrough(t *testing.T) {
	app := express.New()
	app.Use(New())
	var got string
	app.Post("/", func(req *express.Request, res *express.Response, next express.Next) {
		data, _ := io.ReadAll(req.Raw.Body)
		got = string(data)
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("plain")))
	rr := httptest.NewRecorder()
	app.ServeHTTP(rr, r)
	if got != "plain" {
		t.Fatalf("body = %q", got)
	}
}
