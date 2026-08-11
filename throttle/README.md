# throttle — rate-limits how often a function may run

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/throttle.svg)](https://pkg.go.dev/github.com/malcolmston/express/throttle)

Package throttle rate-limits how often a function may run, invoking it at most
once per a fixed wait duration. It is a stdlib-only Go port of lodash's
throttle (https://lodash.com/docs/#throttle), matching that function's
leading/trailing edge behavior along with its cancel and flush controls; the
same leading/trailing model is shared by the npm "throttle-debounce" package.
Throttling is used to tame high-frequency events — scroll and resize
handlers, keystrokes, progress callbacks, log flushing — so expensive work
runs at a bounded rate instead of on every event.

A Throttler is created with New(wait, fn, opts...). Calling its Call method
stands in for "the event fired"; the Throttler decides whether to run fn now,
schedule it for the end of the current window, or ignore it because fn already
ran recently. Regardless of how many times Call is invoked inside one wait
window, fn runs at most once for the leading edge and at most once for the
trailing edge of that window. This is distinct from debouncing, which resets
its timer on every call and only fires after activity stops; a throttle
guarantees steady progress during a sustained burst.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/throttle"
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Clock` | Clock abstracts time so that throttling can be tested with a fake clock. |
| `Option` | Option configures a Throttler. |
| `Throttler` | Throttler is a throttled wrapper around a function. |
| `Timer` | Timer represents a single scheduled call created by Clock.AfterFunc. |

<details>
<summary><code>Option</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func WithClock(c Clock) Option` | WithClock injects a custom Clock, primarily for testing. |
| `func WithLeading(leading bool) Option` | WithLeading enables or disables invoking on the leading edge of the timeout. |
| `func WithTrailing(trailing bool) Option` | WithTrailing enables or disables invoking on the trailing edge of the timeout. |

</details>

<details>
<summary><code>Throttler</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func New(wait time.Duration, fn func(), opts ...Option) *Throttler` | New creates a Throttler that invokes fn at most once per wait. |
| `func (t *Throttler) Call()` | Call invokes the throttled behavior. |
| `func (t *Throttler) Cancel()` | Cancel discards any pending trailing invocation. |
| `func (t *Throttler) Flush()` | Flush immediately invokes any pending trailing call. |
| `func (t *Throttler) Pending() bool` | Pending reports whether an invocation is scheduled. |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/throttle).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
