# abtest — provides middleware that assigns each visitor a stable A/B test bucket and persists

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/abtest.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/abtest)

Package abtest provides middleware that assigns each visitor a stable A/B test
bucket and persists the assignment in a cookie so a returning visitor stays in
the same bucket across requests. It is the Go analogue of the cookie-based
split-testing middleware common in the Node ecosystem, such as express-ab,
which sticks a user to one experiment variant for the lifetime of the cookie.
It is packaged as a drop-in express.Handler rather than a route-splitting
router: it labels the request and lets downstream handlers decide what to
render for each bucket.

Use this middleware whenever you want to divide traffic into deterministic
cohorts for experiments, gradual feature rollouts, or multivariate testing,
and you need the same user to see a consistent experience on every visit. A
random-per-request choice would let a visitor flip between variants and
pollute your metrics; persisting the choice in a cookie keeps the cohort
stable so conversion can be attributed to a single variant. Because the number
and names of buckets are configurable, it also serves canary rollouts (for
example {"stable", "canary"}) as easily as a plain A/B split.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/abtest"
```

## Usage

This is the package's own `ExampleNew`, so it compiles and its output is
asserted on every `go test ./abtest/`.

```go
app := express.New()
	app.Use(abtest.New(abtest.Options{Buckets: []string{"A", "B"}}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		switch abtest.Bucket(req) {
		case "A":
			res.Send("variant A")
		default:
			res.Send("variant B")
		}
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if len(rec.Result().Cookies()) == 0 {
		panic("expected an abtest cookie")
	}
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func Bucket(req *express.Request) string` | Bucket returns the A/B bucket assigned to the request, or "" if the middleware did not run. |
| `func New(opts ...Options) express.Handler` | New returns middleware that assigns and persists a stable bucket for each visitor. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the A/B test middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/abtest).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
