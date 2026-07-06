// Package lang ports the "Lang" category of the npm lodash library, together
// with a handful of closely related utility helpers, to idiomatic Go using only
// the standard library. It supplies the runtime type predicates (IsNil, IsEmpty,
// IsArray, IsMap, IsString, IsNumber, IsBool, IsFunc, IsPointer, IsZero,
// IsPlainObject, IsError, IsEqual, Eq), the value coercions (ToString, ToNumber,
// ToInteger, ToFinite, ToArray, CastArray, DefaultTo), the relational
// comparators (Gt, Gte, Lt, Lte), and the general utility helpers (Times,
// Identity, Constant, Noop, Range, UniqueId, Once) that lodash groups under its
// Lang and Util documentation sections.
//
// Reach for this package when Go code needs to inspect or normalise values whose
// concrete type is not known until run time, for example while decoding
// arbitrary JSON into an any, validating loosely typed configuration, or
// re-implementing JavaScript-flavoured logic during a port. It is the Go analogue
// of scattering _.isEmpty, _.toNumber or _.defaultTo calls throughout a Node
// codebase, and it lets you keep the permissive, coercion-heavy semantics that
// front-end and Node developers expect rather than fighting Go's static type
// system with hand-written type switches at every call site.
//
// The predicate and coercion helpers operate on values of type any and lean on
// the reflect package to examine the dynamic kind of each value, mirroring the
// duck-typed behaviour of the original JavaScript. IsNumber, for instance,
// reports true for every built-in signed, unsigned and floating-point kind but
// false for bool and complex numbers; ToNumber parses numeric strings, maps
// booleans to 0 or 1 and returns NaN for anything unparseable; and ToFinite and
// ToInteger further fold NaN to 0 and clamp infinities, exactly as lodash does.
// The generic helpers (Times, Identity, Constant, Once) instead use Go type
// parameters, because a statically typed signature is both safer and more
// natural than reflection for those cases.
//
// Edge cases follow lodash rather than Go's stricter defaults. IsNil recognises
// both an untyped nil interface and a typed nil pointer, slice, map, channel,
// function or interface; IsEmpty treats nil, empty strings, empty collections
// and every non-collection scalar (numbers, booleans, structs) as empty; Eq
// treats two NaN values as equal even though Go's == operator does not; and the
// coercions never panic on unexpected input, instead returning empty strings,
// zero, NaN or empty slices the way the JavaScript versions return "", 0, NaN or
// []. Functions such as CastArray and ToArray always return a non-nil slice so
// callers can range over the result without a nil check.
//
// Parity with Node lodash is close but necessarily shaped by Go's type system.
// Because Go lacks JavaScript's single Object type, IsPlainObject accepts a Go
// struct, a pointer to a struct, or a map with string keys as the nearest
// equivalent of an object literal, and ToString formats floats compactly (via
// strconv with the 'g' verb) rather than reproducing V8's exact number-to-string
// algorithm. UniqueId returns process-local, monotonically increasing ids and is
// safe for concurrent use, and Once memoises the first result of its function
// under a sync.Once. Where lodash would return JavaScript's undefined, these
// helpers return the Go zero value or a documented sentinel instead.
package lang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// IsNil reports whether value is nil. It returns true both for an untyped nil
// interface and for a typed nil whose underlying kind is nilable (pointer,
// slice, map, channel, function or interface).
func IsNil(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return v.IsNil()
	default:
		return false
	}
}

// IsEmpty reports whether value is considered "empty". Nil values, empty
// strings, empty collections (array, slice, map, channel) and the zero value of
// numbers, booleans and structs are all treated as empty, matching lodash which
// treats primitives that are not collections as empty.
func IsEmpty(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			return true
		}
		return IsEmpty(v.Elem().Interface())
	default:
		// Numbers, booleans, structs and functions are not collections; lodash
		// considers such values empty.
		return true
	}
}

// IsEqual performs a deep comparison between two values and reports whether they
// are equivalent. It uses reflect.DeepEqual semantics.
func IsEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// IsArray reports whether value is a Go slice or array.
func IsArray(value any) bool {
	if value == nil {
		return false
	}
	k := reflect.ValueOf(value).Kind()
	return k == reflect.Slice || k == reflect.Array
}

// IsMap reports whether value is a Go map.
func IsMap(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Map
}

// IsString reports whether value is a string.
func IsString(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.String
}

// IsNumber reports whether value is any built-in integer, unsigned integer or
// floating point type. It returns false for complex numbers and booleans.
func IsNumber(value any) bool {
	if value == nil {
		return false
	}
	switch reflect.ValueOf(value).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

// IsBool reports whether value is a boolean.
func IsBool(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Bool
}

// IsFunc reports whether value is a function.
func IsFunc(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Func
}

// IsPointer reports whether value is a pointer.
func IsPointer(value any) bool {
	if value == nil {
		return false
	}
	return reflect.ValueOf(value).Kind() == reflect.Ptr
}

// IsZero reports whether value is the zero value for its type. An untyped nil is
// considered zero.
func IsZero(value any) bool {
	if value == nil {
		return true
	}
	return reflect.ValueOf(value).IsZero()
}

// IsPlainObject reports whether value is a "plain object": a Go map with string
// keys or a struct (or a pointer to one). This is the closest analogue to a
// JavaScript object literal.
func IsPlainObject(value any) bool {
	if value == nil {
		return false
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		return true
	case reflect.Map:
		return v.Type().Key().Kind() == reflect.String
	default:
		return false
	}
}

// IsError reports whether value implements the error interface.
func IsError(value any) bool {
	if value == nil {
		return false
	}
	_, ok := value.(error)
	return ok
}

// DefaultTo returns value unless it is nil or (for numbers) NaN, in which case
// it returns defaultValue.
func DefaultTo(value, defaultValue any) any {
	if IsNil(value) {
		return defaultValue
	}
	if f, ok := toFloat(value); ok && math.IsNaN(f) {
		return defaultValue
	}
	return value
}

// CastArray wraps value in a []any unless it is already a slice or array, in
// which case its elements are copied into a fresh []any. A nil value yields an
// empty slice.
func CastArray(value any) []any {
	if value == nil {
		return []any{}
	}
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		out := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = v.Index(i).Interface()
		}
		return out
	}
	return []any{value}
}

// ToArray converts value into a []any. Slices and arrays are copied element by
// element, maps yield their values, strings yield their runes (as strings) and
// any other value yields an empty slice, matching lodash's handling of
// non-collection values.
func ToArray(value any) []any {
	if value == nil {
		return []any{}
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		out := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = v.Index(i).Interface()
		}
		return out
	case reflect.Map:
		out := make([]any, 0, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			out = append(out, iter.Value().Interface())
		}
		return out
	case reflect.String:
		runes := []rune(v.String())
		out := make([]any, len(runes))
		for i, r := range runes {
			out[i] = string(r)
		}
		return out
	default:
		return []any{}
	}
}

// ToString converts value to its string representation. Nil becomes an empty
// string, floating point numbers are formatted without a trailing exponent when
// possible and everything else uses fmt's default formatting.
func ToString(value any) string {
	if value == nil {
		return ""
	}
	switch x := value.(type) {
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case error:
		return x.Error()
	default:
		return fmt.Sprintf("%v", value)
	}
}

// ToNumber converts value to a float64. Booleans become 0 or 1, numeric strings
// are parsed and unparseable values yield NaN, mirroring lodash's coercion.
func ToNumber(value any) float64 {
	if value == nil {
		return 0
	}
	if f, ok := toFloat(value); ok {
		return f
	}
	switch x := value.(type) {
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
		return math.NaN()
	default:
		return math.NaN()
	}
}

// ToInteger converts value to an integer by truncating toward zero. NaN becomes
// 0 and infinities are clamped to the extremes of int.
func ToInteger(value any) int {
	f := ToFinite(value)
	return int(math.Trunc(f))
}

// ToFinite converts value to a finite float64. NaN becomes 0 and infinities are
// clamped to the largest finite float64 magnitude, matching lodash.
func ToFinite(value any) float64 {
	f := ToNumber(value)
	if math.IsNaN(f) {
		return 0
	}
	if math.IsInf(f, 1) {
		return math.MaxFloat64
	}
	if math.IsInf(f, -1) {
		return -math.MaxFloat64
	}
	return f
}

// Eq reports whether a and b are equivalent using deep equality, treating two
// NaN values as equal (unlike Go's == operator).
func Eq(a, b any) bool {
	fa, aok := toFloat(a)
	fb, bok := toFloat(b)
	if aok && bok {
		if math.IsNaN(fa) && math.IsNaN(fb) {
			return true
		}
		return fa == fb
	}
	return reflect.DeepEqual(a, b)
}

// Gt reports whether a is greater than b. Numbers are compared numerically and
// strings lexicographically; mismatched types return false.
func Gt(a, b any) bool { return compare(a, b) > 0 }

// Gte reports whether a is greater than or equal to b.
func Gte(a, b any) bool { return compare(a, b) >= 0 }

// Lt reports whether a is less than b.
func Lt(a, b any) bool { return compare(a, b) < 0 }

// Lte reports whether a is less than or equal to b.
func Lte(a, b any) bool { return compare(a, b) <= 0 }

// compare returns -1, 0 or 1 for a<b, a==b, a>b. Incomparable pairs return -2.
func compare(a, b any) int {
	if fa, aok := toFloat(a); aok {
		if fb, bok := toFloat(b); bok {
			switch {
			case fa < fb:
				return -1
			case fa > fb:
				return 1
			default:
				return 0
			}
		}
		return -2
	}
	sa, aok := a.(string)
	sb, bok := b.(string)
	if aok && bok {
		return strings.Compare(sa, sb)
	}
	return -2
}

// Times invokes fn n times, collecting the results into a slice. If n is not
// positive, an empty slice is returned.
func Times[T any](n int, fn func(i int) T) []T {
	if n <= 0 {
		return []T{}
	}
	out := make([]T, n)
	for i := 0; i < n; i++ {
		out[i] = fn(i)
	}
	return out
}

// Identity returns its argument unchanged.
func Identity[T any](value T) T { return value }

// Constant returns a function that always returns value.
func Constant[T any](value T) func() T {
	return func() T { return value }
}

// Noop does nothing and returns nothing.
func Noop() {}

// Range returns a slice of ints from 0 up to, but not including, n. A
// non-positive n yields an empty slice.
func Range(n int) []int {
	if n <= 0 {
		return []int{}
	}
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = i
	}
	return out
}

var uniqueIDCounter uint64

// UniqueId returns a unique string identifier, optionally prefixed. Successive
// calls return monotonically increasing suffixes.
func UniqueId(prefix string) string {
	id := atomic.AddUint64(&uniqueIDCounter, 1)
	return prefix + strconv.FormatUint(id, 10)
}

// Once returns a function that invokes fn at most once. Every call after the
// first returns the value computed by the first invocation.
func Once[T any](fn func() T) func() T {
	var (
		once   sync.Once
		result T
	)
	return func() T {
		once.Do(func() { result = fn() })
		return result
	}
}

// toFloat converts a numeric value to float64. The second return reports whether
// value was a recognised numeric (non-boolean) kind.
func toFloat(value any) (float64, bool) {
	if value == nil {
		return 0, false
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(v.Uint()), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	default:
		return 0, false
	}
}
