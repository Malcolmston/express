// Package validator provides lightweight, fluent input validation for express
// handlers, in the spirit of Node's express-validator. Build a Schema of field
// rules and either validate a map directly or mount it as middleware that
// rejects invalid requests with a 400 JSON response.
//
//	schema := validator.Schema{
//		validator.Field("email").Required().Email(),
//		validator.Field("age").Optional().IsInt().Min(0).Max(120),
//		validator.Field("name").Required().MinLen(2).MaxLen(50),
//	}
//
//	app.Post("/users", schema.Body(), createUser)
package validator

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// rule validates a single field value. It receives the raw value and whether it
// was present, returning an error message when invalid.
type rule func(value any, present bool) string

// FieldRules accumulates the rules applied to a single named field.
type FieldRules struct {
	name     string
	optional bool
	rules    []rule
}

// Field starts a rule chain for the named field.
func Field(name string) *FieldRules {
	return &FieldRules{name: name}
}

// Optional marks the field as optional: when absent, remaining rules are
// skipped instead of failing.
func (f *FieldRules) Optional() *FieldRules {
	f.optional = true
	return f
}

// Required fails when the field is missing or an empty string.
func (f *FieldRules) Required() *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present || isEmpty(v) {
			return "is required"
		}
		return ""
	})
	return f
}

// Email validates that the value is a syntactically valid email address.
func (f *FieldRules) Email() *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		if !emailRe.MatchString(toString(v)) {
			return "must be a valid email address"
		}
		return ""
	})
	return f
}

// MinLen requires the string value to be at least n characters.
func (f *FieldRules) MinLen(n int) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if present && len(toString(v)) < n {
			return fmt.Sprintf("must be at least %d characters", n)
		}
		return ""
	})
	return f
}

// MaxLen requires the string value to be at most n characters.
func (f *FieldRules) MaxLen(n int) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if present && len(toString(v)) > n {
			return fmt.Sprintf("must be at most %d characters", n)
		}
		return ""
	})
	return f
}

// Min requires a numeric value >= n.
func (f *FieldRules) Min(n float64) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		num, ok := toNumber(v)
		if !ok {
			return "must be a number"
		}
		if num < n {
			return fmt.Sprintf("must be >= %v", n)
		}
		return ""
	})
	return f
}

// Max requires a numeric value <= n.
func (f *FieldRules) Max(n float64) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		num, ok := toNumber(v)
		if !ok {
			return "must be a number"
		}
		if num > n {
			return fmt.Sprintf("must be <= %v", n)
		}
		return ""
	})
	return f
}

// IsInt requires the value to be an integer.
func (f *FieldRules) IsInt() *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		if _, err := strconv.Atoi(strings.TrimSpace(toString(v))); err != nil {
			return "must be an integer"
		}
		return ""
	})
	return f
}

// IsNumber requires the value to be numeric.
func (f *FieldRules) IsNumber() *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		if _, ok := toNumber(v); !ok {
			return "must be a number"
		}
		return ""
	})
	return f
}

// In requires the value to be one of the allowed strings.
func (f *FieldRules) In(allowed ...string) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		s := toString(v)
		for _, a := range allowed {
			if s == a {
				return ""
			}
		}
		return "must be one of: " + strings.Join(allowed, ", ")
	})
	return f
}

// Matches requires the string value to match the given regular expression.
func (f *FieldRules) Matches(pattern string) *FieldRules {
	re := regexp.MustCompile(pattern)
	f.rules = append(f.rules, func(v any, present bool) string {
		if present && !re.MatchString(toString(v)) {
			return "has an invalid format"
		}
		return ""
	})
	return f
}

// Custom applies a user-supplied validation function. Returning a non-empty
// string records that message as an error.
func (f *FieldRules) Custom(fn func(value any) string) *FieldRules {
	f.rules = append(f.rules, func(v any, present bool) string {
		if !present {
			return ""
		}
		return fn(v)
	})
	return f
}

// FieldError is a single validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Errors is a collection of validation failures.
type Errors []FieldError

// Error implements the error interface, summarizing the failures.
func (e Errors) Error() string {
	parts := make([]string, len(e))
	for i, fe := range e {
		parts[i] = fe.Field + " " + fe.Message
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// Schema is an ordered set of field rules.
type Schema []*FieldRules

// Validate runs the schema against a data map, returning all failures (nil when
// valid).
func (s Schema) Validate(data map[string]any) Errors {
	var errs Errors
	for _, f := range s {
		val, present := data[f.name]
		if f.optional && (!present || isEmpty(val)) {
			continue
		}
		for _, r := range f.rules {
			if msg := r(val, present); msg != "" {
				errs = append(errs, FieldError{Field: f.name, Message: msg})
				break // one error per field is plenty
			}
		}
	}
	return errs
}

// Body returns express middleware that validates the parsed request body. It
// expects a preceding body-parser (express.JSON or express.URLEncoded). On
// failure it responds 400 with {"errors": [...]}; on success it calls next.
func (s Schema) Body() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		data := normalize(req.Body())
		if errs := s.Validate(data); len(errs) > 0 {
			res.Status(400).JSON(map[string]any{"errors": errs})
			return
		}
		next()
	}
}

// Query returns express middleware that validates the request's query string.
func (s Schema) Query() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		data := normalize(req.QueryValues())
		if errs := s.Validate(data); len(errs) > 0 {
			res.Status(400).JSON(map[string]any{"errors": errs})
			return
		}
		next()
	}
}

// normalize coerces the supported body/query shapes into a map[string]any.
func normalize(body any) map[string]any {
	switch v := body.(type) {
	case map[string]any:
		return v
	case url.Values:
		m := make(map[string]any, len(v))
		for k, vals := range v {
			if len(vals) > 0 {
				m[k] = vals[0]
			}
		}
		return m
	default:
		return map[string]any{}
	}
}

func isEmpty(v any) bool {
	if v == nil {
		return true
	}
	s, ok := v.(string)
	return ok && strings.TrimSpace(s) == ""
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case fmt.Stringer:
		return x.String()
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", x)
	}
}

func toNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case float32:
		return float64(x), true
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return n, err == nil
	default:
		return 0, false
	}
}
