package express

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Flush sends any buffered response data to the client immediately, if the
// underlying writer supports flushing. It commits the headers first. Returns
// true when a flush actually occurred.
func (res *Response) Flush() bool {
	res.writeHeaderOnce()
	if f, ok := res.Writer.(http.Flusher); ok {
		f.Flush()
		return true
	}
	return false
}

// Write implements io.Writer, committing the status line and headers on the
// first call. This lets a *Response be used directly as a streaming sink, e.g.
// with io.Copy or fmt.Fprintf. When no Content-Length is set, net/http streams
// the body using chunked transfer-encoding automatically.
func (res *Response) Write(p []byte) (int, error) {
	res.writeHeaderOnce()
	return res.Writer.Write(p)
}

// WriteChunk writes a chunk and immediately flushes it to the client — the
// building block for chunked streaming responses.
func (res *Response) WriteChunk(p []byte) (int, error) {
	n, err := res.Write(p)
	if err != nil {
		return n, err
	}
	res.Flush()
	return n, err
}

// WriteString writes a string chunk (without flushing).
func (res *Response) WriteString(s string) (int, error) {
	return res.Write([]byte(s))
}

// flushWriter is an io.Writer that flushes to the client after every write.
type flushWriter struct{ res *Response }

// Write implements io.Writer; it writes p as a response chunk via WriteChunk,
// flushing the data to the client after the write.
func (w flushWriter) Write(p []byte) (int, error) { return w.res.WriteChunk(p) }

// Flush flushes the underlying response.
func (w flushWriter) Flush() { w.res.Flush() }

// Stream streams a response body produced by fn. fn receives a writer that
// flushes each write to the client, so data reaches the client incrementally.
// The Content-Type defaults to application/octet-stream when unset.
//
//	res.Stream(func(w io.Writer) error {
//		for i := 0; i < 5; i++ {
//			fmt.Fprintf(w, "chunk %d\n", i)
//		}
//		return nil
//	})
func (res *Response) Stream(fn func(w io.Writer) error) error {
	if res.GetHeader("Content-Type") == "" {
		res.Type("application/octet-stream")
	}
	res.writeHeaderOnce()
	err := fn(flushWriter{res: res})
	res.Flush()
	return err
}

// SendStream copies a reader to the client in chunks, flushing as it goes. It is
// suited to large or open-ended payloads where buffering the whole body in
// memory is undesirable. chunkSize defaults to 32 KiB when <= 0.
func (res *Response) SendStream(r io.Reader, chunkSize ...int) error {
	size := 32 * 1024
	if len(chunkSize) > 0 && chunkSize[0] > 0 {
		size = chunkSize[0]
	}
	if res.GetHeader("Content-Type") == "" {
		res.Type("application/octet-stream")
	}
	res.writeHeaderOnce()

	buf := make([]byte, size)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, werr := res.Writer.Write(buf[:n]); werr != nil {
				return werr
			}
			res.Flush()
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

// SendChunked writes data to the client split into fixed-size chunks, flushing
// after each. Useful for pacing a large in-memory response.
func (res *Response) SendChunked(data []byte, chunkSize int) error {
	if chunkSize <= 0 {
		chunkSize = 32 * 1024
	}
	if res.GetHeader("Content-Type") == "" {
		res.Type("application/octet-stream")
	}
	res.writeHeaderOnce()
	for off := 0; off < len(data); off += chunkSize {
		end := off + chunkSize
		if end > len(data) {
			end = len(data)
		}
		if _, err := res.Writer.Write(data[off:end]); err != nil {
			return err
		}
		res.Flush()
	}
	return nil
}

// SSEWriter writes Server-Sent Events to the client. Obtain one with res.SSE().
type SSEWriter struct {
	res *Response
}

// SSE prepares the response for Server-Sent Events (sets the text/event-stream
// headers and flushes them) and returns a writer for emitting events. The
// handler should keep running to push events, typically until req context is
// done.
func (res *Response) SSE() *SSEWriter {
	res.Set("Content-Type", "text/event-stream")
	res.Set("Cache-Control", "no-cache")
	res.Set("Connection", "keep-alive")
	res.Set("X-Accel-Buffering", "no") // disable proxy buffering
	res.writeHeaderOnce()
	res.Flush()
	return &SSEWriter{res: res}
}

// Send emits a named event with a data payload. Multi-line data is split into
// multiple data: lines per the SSE spec.
func (s *SSEWriter) Send(event, data string) error {
	var b strings.Builder
	if event != "" {
		fmt.Fprintf(&b, "event: %s\n", event)
	}
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(&b, "data: %s\n", line)
	}
	b.WriteByte('\n')
	_, err := s.res.WriteChunk([]byte(b.String()))
	return err
}

// SendData emits an unnamed (message) event carrying data.
func (s *SSEWriter) SendData(data string) error { return s.Send("", data) }

// SendJSON marshals v to JSON and emits it as a named event.
func (s *SSEWriter) SendJSON(event string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.Send(event, string(data))
}

// SendID emits an event that also sets the SSE id field (for Last-Event-ID
// resumption).
func (s *SSEWriter) SendID(id, event, data string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "id: %s\n", id)
	if event != "" {
		fmt.Fprintf(&b, "event: %s\n", event)
	}
	for _, line := range strings.Split(data, "\n") {
		fmt.Fprintf(&b, "data: %s\n", line)
	}
	b.WriteByte('\n')
	_, err := s.res.WriteChunk([]byte(b.String()))
	return err
}

// Comment emits an SSE comment line (often used as a keep-alive heartbeat).
func (s *SSEWriter) Comment(text string) error {
	_, err := s.res.WriteChunk([]byte(": " + text + "\n\n"))
	return err
}

// Retry tells the client how long (in ms) to wait before reconnecting.
func (s *SSEWriter) Retry(ms int) error {
	_, err := s.res.WriteChunk([]byte(fmt.Sprintf("retry: %d\n\n", ms)))
	return err
}
