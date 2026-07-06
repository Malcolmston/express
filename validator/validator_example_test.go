package validator_test

import (
	"fmt"
	"net/http/httptest"
	"strings"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/validator"
)

// ExampleSchema_Validate builds a Schema with a required e-mail field and an
// optional, bounded integer field, then runs it against a map that violates
// both constraints. Schema.Validate walks the fields in declaration order and
// records at most one message per field, so the resulting Errors slice is fully
// deterministic. The e-mail value fails the Email rule and the age value, once
// coerced from its string form, fails the Max rule. Each FieldError pairs the
// offending field name with a human-readable message. The Output block below
// confirms both failures are reported in field order.
func ExampleSchema_Validate() {
	schema := validator.Schema{
		validator.Field("email").Required().Email(),
		validator.Field("age").Optional().IsInt().Min(0).Max(120),
	}

	errs := schema.Validate(map[string]any{
		"email": "not-an-email",
		"age":   "200",
	})

	for _, e := range errs {
		fmt.Printf("%s: %s\n", e.Field, e.Message)
	}
	// Output:
	// email: must be a valid email address
	// age: must be <= 120
}

// ExampleField demonstrates that a valid input map produces no errors. The
// chain marks name as required with a minimum length and treats age as an
// optional integer, so omitting age entirely is acceptable because Optional
// skips the rest of that field's chain when the value is absent. Every supplied
// value satisfies its rule, so Schema.Validate returns a nil Errors slice whose
// length is zero. This mirrors how you would gate a handler: an empty result
// means the data is safe to use. The example prints the error count to show the
// happy path.
func ExampleField() {
	schema := validator.Schema{
		validator.Field("name").Required().MinLen(2).MaxLen(50),
		validator.Field("age").Optional().IsInt().Min(0),
	}

	errs := schema.Validate(map[string]any{
		"name": "Ada",
	})

	fmt.Println("errors:", len(errs))
	// Output:
	// errors: 0
}

// ExampleErrors shows how the Errors slice satisfies the error interface. When
// a schema produces one or more FieldError values, calling Error on the slice
// joins them into a single summary string prefixed with "validation failed:".
// Here both required fields are missing, so two failures are collected. The
// combined message lists each field and its reason separated by semicolons.
// This is convenient when you want to log or return one flat string rather than
// iterating the individual failures.
func ExampleErrors() {
	schema := validator.Schema{
		validator.Field("email").Required(),
		validator.Field("name").Required(),
	}

	errs := schema.Validate(map[string]any{})

	fmt.Println(errs.Error())
	// Output:
	// validation failed: email is required; name is required
}

// ExampleSchema_Body mounts a Schema as express request middleware and drives
// it with httptest to show the rejection path. Schema.Body validates the parsed
// request body and, on failure, responds with HTTP 400 and a JSON object of the
// form {"errors": [...]} before the downstream handler ever runs. Here the body
// is registered directly as a map so the validator sees a missing required
// field. The recorder captures the 400 status and the JSON payload naming the
// field. Because the response is deterministic, the Output block asserts both.
func ExampleSchema_Body() {
	schema := validator.Schema{
		validator.Field("email").Required().Email(),
	}

	app := express.New()
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.SetBody(map[string]any{"email": ""})
		next()
	})
	app.Post("/users", schema.Body(), func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("created")
	})

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, httptest.NewRequest("POST", "/users", strings.NewReader("")))

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", strings.TrimSpace(rec.Body.String()))
	// Output:
	// status: 400
	// body: {"errors":[{"field":"email","message":"is required"}]}
}
