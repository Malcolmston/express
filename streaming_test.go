package express

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestStream(t *testing.T) {
	app := New()
	app.Get("/stream", func(req *Request, res *Response, next Next) {
		res.Type("text").Stream(func(w io.Writer) error {
			for i := 0; i < 3; i++ {
				fmt.Fprintf(w, "chunk-%d;", i)
			}
			return nil
		})
	})
	w := do(app, "GET", "/stream", "")
	if w.Body.String() != "chunk-0;chunk-1;chunk-2;" {
		t.Fatalf("stream body = %q", w.Body.String())
	}
	if !w.Flushed {
		t.Fatal("expected the response to have been flushed")
	}
}

func TestSendStream(t *testing.T) {
	app := New()
	payload := strings.Repeat("abcdefgh", 100) // 800 bytes
	app.Get("/download", func(req *Request, res *Response, next Next) {
		res.SendStream(strings.NewReader(payload), 64)
	})
	w := do(app, "GET", "/download", "")
	if w.Body.String() != payload {
		t.Fatalf("sendstream length = %d, want %d", w.Body.Len(), len(payload))
	}
}

func TestSendChunked(t *testing.T) {
	app := New()
	data := []byte("0123456789")
	app.Get("/c", func(req *Request, res *Response, next Next) {
		res.SendChunked(data, 3)
	})
	w := do(app, "GET", "/c", "")
	if w.Body.String() != "0123456789" {
		t.Fatalf("chunked body = %q", w.Body.String())
	}
}

func TestWriteChunkFlushes(t *testing.T) {
	app := New()
	app.Get("/w", func(req *Request, res *Response, next Next) {
		res.Status(202)
		res.WriteChunk([]byte("hello "))
		res.WriteChunk([]byte("world"))
	})
	w := do(app, "GET", "/w", "")
	if w.Code != 202 {
		t.Fatalf("status = %d", w.Code)
	}
	if w.Body.String() != "hello world" {
		t.Fatalf("body = %q", w.Body.String())
	}
}

func TestSSE(t *testing.T) {
	app := New()
	app.Get("/events", func(req *Request, res *Response, next Next) {
		sse := res.SSE()
		sse.Send("tick", "1")
		sse.SendJSON("update", map[string]int{"n": 2})
		sse.SendData("plain")
		sse.Comment("keepalive")
	})
	w := do(app, "GET", "/events", "")

	ct := w.Header().Get("Content-Type")
	if ct != "text/event-stream" {
		t.Fatalf("content-type = %q", ct)
	}
	body := w.Body.String()
	for _, want := range []string{
		"event: tick\ndata: 1\n\n",
		"event: update\ndata: {\"n\":2}\n\n",
		"data: plain\n\n",
		": keepalive\n\n",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("SSE body missing %q\ngot:\n%s", want, body)
		}
	}
}

func TestSSEMultilineData(t *testing.T) {
	app := New()
	app.Get("/m", func(req *Request, res *Response, next Next) {
		res.SSE().Send("log", "line1\nline2")
	})
	w := do(app, "GET", "/m", "")
	if !strings.Contains(w.Body.String(), "data: line1\ndata: line2\n\n") {
		t.Fatalf("multiline SSE = %q", w.Body.String())
	}
}
