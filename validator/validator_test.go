package validator_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/validator"
)

func TestValidateMap(t *testing.T) {
	schema := validator.Schema{
		validator.Field("email").Required().Email(),
		validator.Field("age").Optional().IsInt().Min(0).Max(120),
		validator.Field("name").Required().MinLen(2).MaxLen(50),
	}

	// Valid input.
	if errs := schema.Validate(map[string]any{"email": "a@b.com", "name": "Ada", "age": "30"}); len(errs) != 0 {
		t.Fatalf("expected no errors, got %v", errs)
	}

	// Invalid input: bad email, missing name, out-of-range age.
	errs := schema.Validate(map[string]any{"email": "nope", "age": "200"})
	if len(errs) != 3 {
		t.Fatalf("expected 3 errors, got %d: %v", len(errs), errs)
	}
}

func TestOptionalSkips(t *testing.T) {
	schema := validator.Schema{
		validator.Field("nickname").Optional().MinLen(3),
	}
	if errs := schema.Validate(map[string]any{}); len(errs) != 0 {
		t.Fatalf("optional absent field should pass, got %v", errs)
	}
	if errs := schema.Validate(map[string]any{"nickname": "ab"}); len(errs) != 1 {
		t.Fatalf("present-but-invalid optional should fail, got %v", errs)
	}
}

func TestInAndMatches(t *testing.T) {
	schema := validator.Schema{
		validator.Field("role").Required().In("admin", "user"),
		validator.Field("slug").Required().Matches(`^[a-z-]+$`),
	}
	if errs := schema.Validate(map[string]any{"role": "guest", "slug": "Bad Slug"}); len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %v", errs)
	}
	if errs := schema.Validate(map[string]any{"role": "admin", "slug": "good-slug"}); len(errs) != 0 {
		t.Fatalf("expected valid, got %v", errs)
	}
}

func TestBodyMiddlewareRejects(t *testing.T) {
	schema := validator.Schema{
		validator.Field("email").Required().Email(),
	}
	app := express.New()
	app.Use(express.JSON())
	app.Post("/signup", schema.Body(), func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("created")
	})

	// Invalid -> 400 with errors.
	r := httptest.NewRequest("POST", "/signup", strings.NewReader(`{"email":"bad"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 400 || !strings.Contains(w.Body.String(), "valid email") {
		t.Fatalf("expected 400 with error, got code=%d body=%q", w.Code, w.Body.String())
	}

	// Valid -> handler runs.
	r2 := httptest.NewRequest("POST", "/signup", strings.NewReader(`{"email":"a@b.com"}`))
	r2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	if w2.Code != 200 || w2.Body.String() != "created" {
		t.Fatalf("expected 200 created, got code=%d body=%q", w2.Code, w2.Body.String())
	}
}
