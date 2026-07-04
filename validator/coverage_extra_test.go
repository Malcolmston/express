package validator_test

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/malcolmston/express"
	"github.com/malcolmston/express/validator"
)

// TestNumericRules exercises Min, Max, IsInt, IsNumber including error branches
// and non-numeric / non-string value coercion via toNumber/toString.
func TestNumericRules(t *testing.T) {
	cases := []struct {
		name    string
		schema  validator.Schema
		data    map[string]any
		wantErr bool
	}{
		{"min ok float64", validator.Schema{validator.Field("n").Min(5)}, map[string]any{"n": float64(10)}, false},
		{"min fail", validator.Schema{validator.Field("n").Min(5)}, map[string]any{"n": 3}, true},
		{"min non-number", validator.Schema{validator.Field("n").Min(5)}, map[string]any{"n": "abc"}, true},
		{"max ok int", validator.Schema{validator.Field("n").Max(100)}, map[string]any{"n": 50}, false},
		{"max fail", validator.Schema{validator.Field("n").Max(100)}, map[string]any{"n": 150}, true},
		{"max non-number", validator.Schema{validator.Field("n").Max(100)}, map[string]any{"n": []int{1}}, true},
		{"isint ok", validator.Schema{validator.Field("n").IsInt()}, map[string]any{"n": " 42 "}, false},
		{"isint fail", validator.Schema{validator.Field("n").IsInt()}, map[string]any{"n": "3.14"}, true},
		{"isnumber ok string", validator.Schema{validator.Field("n").IsNumber()}, map[string]any{"n": "3.14"}, false},
		{"isnumber ok float32", validator.Schema{validator.Field("n").IsNumber()}, map[string]any{"n": float32(1.5)}, false},
		{"isnumber ok int64", validator.Schema{validator.Field("n").IsNumber()}, map[string]any{"n": int64(7)}, false},
		{"isnumber fail", validator.Schema{validator.Field("n").IsNumber()}, map[string]any{"n": "notnum"}, true},
		{"isnumber fail type", validator.Schema{validator.Field("n").IsNumber()}, map[string]any{"n": true}, true},
	}
	for _, c := range cases {
		errs := c.schema.Validate(c.data)
		if (len(errs) > 0) != c.wantErr {
			t.Errorf("%s: errs=%v wantErr=%v", c.name, errs, c.wantErr)
		}
	}
}

// TestMinMaxAbsentSkips ensures numeric rules skip when the field is absent.
func TestMinMaxAbsentSkips(t *testing.T) {
	schema := validator.Schema{
		validator.Field("n").Min(5).Max(10).IsInt().IsNumber(),
	}
	if errs := schema.Validate(map[string]any{}); len(errs) != 0 {
		t.Fatalf("absent field should skip numeric rules, got %v", errs)
	}
}

// TestCustomRule covers the Custom validator including the absent-skip branch.
func TestCustomRule(t *testing.T) {
	schema := validator.Schema{
		validator.Field("code").Custom(func(v any) string {
			if v == "secret" {
				return ""
			}
			return "wrong code"
		}),
	}
	if errs := schema.Validate(map[string]any{"code": "secret"}); len(errs) != 0 {
		t.Fatalf("valid custom should pass, got %v", errs)
	}
	errs := schema.Validate(map[string]any{"code": "nope"})
	if len(errs) != 1 || errs[0].Message != "wrong code" {
		t.Fatalf("custom errs = %v", errs)
	}
	// Absent field: Custom skips.
	if errs := schema.Validate(map[string]any{}); len(errs) != 0 {
		t.Fatalf("absent custom should skip, got %v", errs)
	}
}

// TestErrorsErrorString covers Errors.Error().
func TestErrorsErrorString(t *testing.T) {
	schema := validator.Schema{
		validator.Field("email").Required(),
		validator.Field("age").Required(),
	}
	errs := schema.Validate(map[string]any{})
	msg := errs.Error()
	if !strings.HasPrefix(msg, "validation failed: ") {
		t.Fatalf("Error() = %q", msg)
	}
	if !strings.Contains(msg, "email is required") || !strings.Contains(msg, "age is required") {
		t.Fatalf("Error() missing fields = %q", msg)
	}
}

// TestQueryMiddleware covers Schema.Query() and normalize(url.Values).
func TestQueryMiddleware(t *testing.T) {
	schema := validator.Schema{
		validator.Field("page").Required().IsInt(),
	}
	app := express.New()
	app.Get("/list", schema.Query(), func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})

	// Valid query.
	r := httptest.NewRequest("GET", "/list?page=2", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "ok" {
		t.Fatalf("valid query: code=%d body=%q", w.Code, w.Body.String())
	}

	// Invalid query => 400 JSON.
	r2 := httptest.NewRequest("GET", "/list?page=abc", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, r2)
	if w2.Code != 400 {
		t.Fatalf("invalid query code = %d", w2.Code)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(w2.Body.Bytes(), &payload); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if _, ok := payload["errors"]; !ok {
		t.Fatalf("expected errors key, got %s", w2.Body.String())
	}

	// Missing required query => 400.
	r3 := httptest.NewRequest("GET", "/list", nil)
	w3 := httptest.NewRecorder()
	app.ServeHTTP(w3, r3)
	if w3.Code != 400 {
		t.Fatalf("missing query code = %d", w3.Code)
	}
}

// TestBodyMiddlewarePasses covers the success path of Schema.Body().
func TestBodyMiddlewarePasses(t *testing.T) {
	schema := validator.Schema{
		validator.Field("name").Required().MinLen(2),
	}
	app := express.New()
	app.Use(express.JSON())
	app.Post("/", schema.Body(), func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("created")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"Ada"}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 200 || w.Body.String() != "created" {
		t.Fatalf("valid body: code=%d body=%q", w.Code, w.Body.String())
	}
}

// TestBodyNormalizeUnsupportedType ensures a non-map/non-Values body yields an
// empty map (so Required rules fire).
func TestBodyNormalizeUnsupportedType(t *testing.T) {
	schema := validator.Schema{validator.Field("x").Required()}
	app := express.New()
	// Store a plain string body (unsupported shape for normalize).
	app.Use(func(req *express.Request, res *express.Response, next express.Next) {
		req.SetBody("just a string")
		next()
	})
	app.Post("/", schema.Body(), func(req *express.Request, res *express.Response, next express.Next) {
		res.Send("ok")
	})
	r := httptest.NewRequest("POST", "/", strings.NewReader(""))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatalf("unsupported body type should fail Required, code = %d", w.Code)
	}
}
