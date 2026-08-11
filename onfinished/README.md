# onfinished — registers callbacks that run when an HTTP response finishes

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/onfinished.svg)](https://pkg.go.dev/github.com/malcolmston/express/onfinished)

Package onfinished registers callbacks that run when an HTTP response
finishes, mirroring the behavior of the npm on-finished library
(https://www.npmjs.com/package/on-finished) using only the Go standard
library. In Node, on-finished lets middleware attach a listener that fires
exactly once when a response has been fully sent or a request has been fully
read — or when either is torn down early by an error — so that cleanup
logic (closing files, releasing locks, writing access logs, flushing metrics)
runs deterministically at the end of the exchange.

The Node API centers on a single function, onFinished(msg, listener), that
hooks into the stream's "finish", "end", "close" and "error" events and
guarantees the listener is invoked one time with an optional error argument.
Go's net/http does not expose those lifecycle events directly, so this port
re-frames the same contract around an explicit Tracker value that wraps an
http.ResponseWriter. You create a Tracker with New, register any number of
completion callbacks with OnFinished, and signal completion yourself by
calling Done when the handler returns or the connection is closed.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/onfinished"
```

## Usage

This is the package's own `ExampleTracker`, so it compiles and its output is
asserted on every `go test ./onfinished/`.

```go
rec := httptest.NewRecorder()

	handler := func(w http.ResponseWriter, r *http.Request) {
		tr := onfinished.New(w)
		tr.OnFinished(func(err error) {
			fmt.Printf("finished err=%v\n", err)
		})
		tr.WriteHeader(http.StatusOK)
		fmt.Fprint(tr, "hello")
		tr.Done(nil)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler(rec, req)

	fmt.Printf("status=%d body=%q finished=%v\n", rec.Code, rec.Body.String(), true)
```

```
finished err=<nil>
status=200 body="hello" finished=true
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Tracker` | Tracker wraps an http.ResponseWriter and tracks whether the response has finished. |

<details>
<summary><code>Tracker</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(w http.ResponseWriter) *Tracker` | New wraps w in a Tracker. |
| `func (t *Tracker) Done(err error)` | Done marks the response as finished and invokes all registered callbacks with err. |
| `func (t *Tracker) Flush()` | Flush implements http.Flusher when the underlying ResponseWriter supports it; otherwise it is a no-op. |
| `func (t *Tracker) IsFinished() bool` | IsFinished reports whether the response has finished. |
| `func (t *Tracker) OnFinished(fn func(err error))` | OnFinished registers fn to be called when the response finishes. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/onfinished).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
