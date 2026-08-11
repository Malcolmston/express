# validator — provides lightweight, fluent input validation for express handlers

[![Go Reference](https://pkg.go.dev/badge/github.com/malcolmston/express/validator.svg)](https://pkg.go.dev/github.com/malcolmston/express/validator)

Package validator provides lightweight, fluent input validation for express
handlers, in the spirit of the npm "express-validator" package. Instead of
mirroring express-validator's global check()/body()/query() functions and its
validationResult() accumulator, this port models validation as a Schema value:
an ordered list of per-field rule chains that you build once and reuse as a
plain function or as request middleware. Build a Schema of field rules and
either validate a map directly with Schema.Validate or mount it with
Schema.Body / Schema.Query so that invalid requests are rejected with a 400
JSON response before your handler runs.

You reach for this package when an express route needs to guard its body or
query string against missing, malformed, or out-of-range fields without
pulling in a large dependency. Each Field starts a chain, and the chained
methods (Required, Optional, Email, MinLen, MaxLen, Min, Max, IsInt, IsNumber,
In, Matches, Custom) append rules that are evaluated in the order they were
declared. Rule chains are fluent: every method returns the same *FieldRules
pointer, so a whole field can be described in a single readable expression
that reads much like the express-validator DSL it emulates.

## Install

```sh
go get github.com/malcolmston/express@v0.5.0
```

```go
import "github.com/malcolmston/express/validator"
```

## Usage

This is the package's own `ExampleErrors`, so it compiles and its output is
asserted on every `go test ./validator/`.

```go
schema := validator.Schema{
		validator.Field("email").Required(),
		validator.Field("name").Required(),
	}

	errs := schema.Validate(map[string]any{})

	fmt.Println(errs.Error())
```

```
validation failed: email is required; name is required
```

## Exported surface

### Types

| Type | What it is |
| --- | --- |
| `Errors` | Errors is a collection of validation failures. |
| `FieldError` | FieldError is a single validation failure. |
| `FieldRules` | FieldRules accumulates the rules applied to a single named field. |
| `Schema` | Schema is an ordered set of field rules. |

<details>
<summary><code>Errors</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func (e Errors) Error() string` | Error implements the error interface, summarizing the failures. |

</details>

<details>
<summary><code>FieldRules</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func Field(name string) *FieldRules` | Field starts a rule chain for the named field. |
| `func (f *FieldRules) Custom(fn func(value any) string) *FieldRules` | Custom applies a user-supplied validation function. |
| `func (f *FieldRules) Email() *FieldRules` | Email validates that the value is a syntactically valid email address. |
| `func (f *FieldRules) In(allowed ...string) *FieldRules` | In requires the value to be one of the allowed strings. |
| `func (f *FieldRules) IsInt() *FieldRules` | IsInt requires the value to be an integer. |
| `func (f *FieldRules) IsNumber() *FieldRules` | IsNumber requires the value to be numeric. |
| `func (f *FieldRules) Matches(pattern string) *FieldRules` | Matches requires the string value to match the given regular expression. |
| `func (f *FieldRules) Max(n float64) *FieldRules` | Max requires a numeric value <= n. |
| `func (f *FieldRules) MaxLen(n int) *FieldRules` | MaxLen requires the string value to be at most n characters. |
| `func (f *FieldRules) Min(n float64) *FieldRules` | Min requires a numeric value >= n. |
| `func (f *FieldRules) MinLen(n int) *FieldRules` | MinLen requires the string value to be at least n characters. |
| `func (f *FieldRules) Optional() *FieldRules` | Optional marks the field as optional: when absent, remaining rules are skipped instead of failing. |
| `func (f *FieldRules) Required() *FieldRules` | Required fails when the field is missing or an empty string. |

</details>

<details>
<summary><code>Schema</code> — constructors and methods</summary>

| Signature | What it does |
| --- | --- |
| `func (s Schema) Body() express.Handler` | Body returns express middleware that validates the parsed request body. |
| `func (s Schema) Query() express.Handler` | Query returns express middleware that validates the request's query string. |
| `func (s Schema) Validate(data map[string]any) Errors` | Validate runs the schema against a data map, returning all failures (nil when valid). |

</details>

Full signatures, doc comments and every runnable example are on
[pkg.go.dev](https://pkg.go.dev/github.com/malcolmston/express/validator).

## Deviations from upstream

Deliberate differences for this package, where any exist, are recorded in the
module-wide [`API-DEVIATIONS.md`](../API-DEVIATIONS.md).

## License

MIT, as part of [`github.com/malcolmston/express`](..).
An independent re-implementation, not affiliated with or endorsed by the
original project.
