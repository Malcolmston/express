# rolecheck — provides middleware that authorizes a request only when the caller holds at least one

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/middleware/rolecheck.svg)](https://pkg.go.dev/github.com/malcolmston/express/middleware/rolecheck)

Package rolecheck provides middleware that authorizes a request only when the
caller holds at least one of a configured set of roles. It is the express
framework's Go analogue of the small role-guard helpers that Node developers
hand-roll on top of Passport or connect-roles, and of libraries such as
express-authorization or the role checks baked into access-control kits:
authentication has already established who the caller is, and this middleware
answers the coarser question of whether that caller is allowed in at all.

Reach for this middleware to gate a route or a whole router behind a role,
after some earlier handler has attached the caller's roles to the request.
Typical uses are an admin console that only "admin" may reach, an editorial
area open to "editor" or "admin", or a moderation endpoint reserved for
"moderator". Because the roles are plain strings compared for equality, the
middleware is agnostic about where they come from: a session, a decoded JWT, a
database lookup, or a header can all feed the same guard through the Getter
callback.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/middleware/rolecheck"
```

## Usage

This is the package's own `Example`, so it compiles and its output is
asserted on every `go test ./rolecheck/`.

```go
app := express.New()

	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.Set("roles", []string{"editor"})
		next()
	})

	app.Use(rolecheck.New(rolecheck.Options{
		Roles: []string{"admin", "editor"},
		Getter: func(req *express.Request) []string {
			v, _ := req.Value("roles")
			roles, _ := v.([]string)
			return roles
		},
	}))

	app.Get("/admin", func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("welcome")
	})

	r := httptest.NewRequest("GET", "/admin", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)

	fmt.Println(w.Code)
	fmt.Println(w.Body.String())
```

```
200
welcome
```

## Exported surface

### Functions

| Function | What it does |
| --- | --- |
| `func New(opts Options) express.Handler` | New returns middleware that responds with 403 unless the request holds at least one of the required roles. |

### Types

| Type | What it is |
| --- | --- |
| `Options` | Options configures the role-check middleware. |

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/middleware/rolecheck).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](../..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
