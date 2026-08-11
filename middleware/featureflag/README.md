# featureflag — provides middleware that attaches a fixed set of boolean feature flags to every

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/featureflag.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/featureflag)

Package featureflag provides middleware that attaches a fixed set of boolean
feature flags to every request, letting downstream handlers gate behavior on
named features. It is a small, dependency-free analogue of the request-scoped
flag lookup exposed by Node feature-flag libraries such as
"express-feature-flags" or the flag helpers bundled with LaunchDarkly's and
Unleash's Express SDKs: rather than talking to a remote flag service, it
carries a caller-supplied map of names to enabled states and makes it
available for the lifetime of the request.

Use this middleware when you want to toggle code paths (a new UI, an
experimental endpoint, a gradual rollout) without redeploying or littering
handlers with global configuration reads. Because the flag set is captured
once when New is called, it is best for process-wide flags resolved at
startup; if you need per-user or dynamically refreshed flags, build the map
per request in your own middleware and store it under the same request state,
or wrap New so it recomputes the map on each call.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/featureflag"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./featureflag/`.

```go
app := express.New()
	app.Use(featureflag.New(featureflag.Options{
		Flags: map[string]bool{"new-ui": true, "beta": false},
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		fmt.Println("new-ui:", featureflag.Enabled(req, "new-ui"))
		fmt.Println("beta:", featureflag.Enabled(req, "beta"))
		fmt.Println("nope:", featureflag.Enabled(req, "nope"))
		res.Send("ok")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
```

```
new-ui: true
beta: false
nope: false
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Enabled(req *express.Request, name string) bool` | Enabled reports whether the named feature flag is enabled for the request. |
| `func New(opts ...Options) express.Handler` | New returns middleware that stores the configured flags on each request. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the feature flag middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/featureflag).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
