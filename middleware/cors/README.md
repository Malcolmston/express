# cors — provides configurable Cross-Origin Resource Sharing (CORS) middleware for the express

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/cors.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/cors)

Package cors provides configurable Cross-Origin Resource Sharing (CORS)
middleware for the express framework. It sets the appropriate Access-Control-*
response headers and answers CORS preflight (OPTIONS) requests with a 204 No
Content response. It is the Go analogue of the ubiquitous Express cors package
(expressjs/cors), packaged as a drop-in express.Handler.

Use this middleware whenever a browser-based client hosted on one origin must
call an API served from another origin. The same-origin policy blocks such
requests by default; CORS is the mechanism by which the server opts specific
origins back in and tells the browser which methods, request headers, and
response headers are permitted, and whether credentials (cookies,
Authorization headers) may be sent. Typical deployments mount it globally with
app.Use so both the simple requests and the preflight probes for every route
are handled, but it can equally guard a single router or path prefix.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/cors"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./cors/`.

```go
app := express.New()
	app.Use(cors.New(cors.Options{
		AllowedOrigins: []string{"https://app.example.com"},
		AllowedMethods: []string{"GET", "POST"},
		MaxAge:         600,
	}))
	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	get := httptest.NewRecorder()
	greq := httptest.NewRequest(http.MethodGet, "/", nil)
	greq.Header.Set("Origin", "https://app.example.com")
	app.ServeHTTP(get, greq)
	fmt.Println("GET status:", get.Code)
	fmt.Println("GET allow-origin:", get.Header().Get("Access-Control-Allow-Origin"))
	fmt.Println("GET body:", get.Body.String())

	pre := httptest.NewRecorder()
	preq := httptest.NewRequest(http.MethodOptions, "/", nil)
	preq.Header.Set("Origin", "https://app.example.com")
	app.ServeHTTP(pre, preq)
	fmt.Println("preflight status:", pre.Code)
	fmt.Println("preflight allow-methods:", pre.Header().Get("Access-Control-Allow-Methods"))
```

```
GET status: 200
GET allow-origin: https://app.example.com
GET body: ok
preflight status: 204
preflight allow-methods: GET, POST
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts ...Options) express.Handler` | New returns CORS middleware configured by the optional Options. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the CORS middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/cors).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
