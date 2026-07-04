package lang

import (
	"errors"
	"math"
	"reflect"
	"testing"
)

func TestIsNil(t *testing.T) {
	var p *int
	var s []int
	var m map[string]int
	cases := []struct {
		v    any
		want bool
	}{
		{nil, true},
		{p, true},
		{s, true},
		{m, true},
		{0, false},
		{"", false},
		{new(int), false},
	}
	for _, c := range cases {
		if got := IsNil(c.v); got != c.want {
			t.Errorf("IsNil(%#v)=%v want %v", c.v, got, c.want)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	cases := []struct {
		v    any
		want bool
	}{
		{nil, true},
		{"", true},
		{"a", false},
		{[]int{}, true},
		{[]int{1}, false},
		{map[string]int{}, true},
		{map[string]int{"a": 1}, false},
		{0, true},
		{1, true}, // non-collection primitive is empty in lodash
		{true, true},
	}
	for _, c := range cases {
		if got := IsEmpty(c.v); got != c.want {
			t.Errorf("IsEmpty(%#v)=%v want %v", c.v, got, c.want)
		}
	}
}

func TestIsEqual(t *testing.T) {
	if !IsEqual([]int{1, 2}, []int{1, 2}) {
		t.Error("slices should be equal")
	}
	if IsEqual(map[string]int{"a": 1}, map[string]int{"a": 2}) {
		t.Error("maps should differ")
	}
}

func TestTypePredicates(t *testing.T) {
	if !IsArray([]int{1}) || IsArray("x") {
		t.Error("IsArray")
	}
	if !IsArray([3]int{}) {
		t.Error("IsArray array")
	}
	if !IsMap(map[int]int{}) || IsMap([]int{}) {
		t.Error("IsMap")
	}
	if !IsString("x") || IsString(1) {
		t.Error("IsString")
	}
	if !IsNumber(1) || !IsNumber(1.5) || !IsNumber(uint8(3)) || IsNumber("1") || IsNumber(true) {
		t.Error("IsNumber")
	}
	if !IsBool(true) || IsBool(1) {
		t.Error("IsBool")
	}
	if !IsFunc(func() {}) || IsFunc(1) {
		t.Error("IsFunc")
	}
	if !IsPointer(new(int)) || IsPointer(1) {
		t.Error("IsPointer")
	}
}

func TestIsZero(t *testing.T) {
	if !IsZero(0) || !IsZero("") || !IsZero(nil) {
		t.Error("IsZero zero values")
	}
	if IsZero(1) || IsZero("x") {
		t.Error("IsZero nonzero")
	}
}

func TestIsPlainObject(t *testing.T) {
	type S struct{ A int }
	if !IsPlainObject(S{}) || !IsPlainObject(&S{}) {
		t.Error("struct should be plain object")
	}
	if !IsPlainObject(map[string]int{}) {
		t.Error("string-keyed map should be plain object")
	}
	if IsPlainObject(map[int]int{}) {
		t.Error("int-keyed map is not plain object")
	}
	if IsPlainObject([]int{}) || IsPlainObject(1) {
		t.Error("non-object")
	}
}

func TestIsError(t *testing.T) {
	if !IsError(errors.New("x")) {
		t.Error("error should be error")
	}
	if IsError("x") {
		t.Error("string is not error")
	}
}

func TestDefaultTo(t *testing.T) {
	if DefaultTo(nil, 5) != 5 {
		t.Error("nil default")
	}
	if DefaultTo(3, 5) != 3 {
		t.Error("value passthrough")
	}
	if DefaultTo(math.NaN(), 5) != 5 {
		t.Error("NaN default")
	}
}

func TestCastArray(t *testing.T) {
	if !reflect.DeepEqual(CastArray(1), []any{1}) {
		t.Error("scalar cast")
	}
	if !reflect.DeepEqual(CastArray([]int{1, 2}), []any{1, 2}) {
		t.Error("slice cast")
	}
	if !reflect.DeepEqual(CastArray(nil), []any{}) {
		t.Error("nil cast")
	}
}

func TestToArray(t *testing.T) {
	if !reflect.DeepEqual(ToArray([]int{1, 2}), []any{1, 2}) {
		t.Error("slice")
	}
	if !reflect.DeepEqual(ToArray("ab"), []any{"a", "b"}) {
		t.Error("string")
	}
	if !reflect.DeepEqual(ToArray(5), []any{}) {
		t.Error("scalar to empty")
	}
	got := ToArray(map[string]int{"a": 1})
	if len(got) != 1 || got[0] != 1 {
		t.Errorf("map values: %v", got)
	}
}

func TestToString(t *testing.T) {
	cases := []struct {
		v    any
		want string
	}{
		{nil, ""},
		{"x", "x"},
		{true, "true"},
		{3.5, "3.5"},
		{42, "42"},
		{errors.New("boom"), "boom"},
	}
	for _, c := range cases {
		if got := ToString(c.v); got != c.want {
			t.Errorf("ToString(%#v)=%q want %q", c.v, got, c.want)
		}
	}
}

func TestToNumber(t *testing.T) {
	cases := []struct {
		v    any
		want float64
	}{
		{nil, 0},
		{"3.5", 3.5},
		{true, 1},
		{false, 0},
		{10, 10},
		{"", 0},
	}
	for _, c := range cases {
		if got := ToNumber(c.v); got != c.want {
			t.Errorf("ToNumber(%#v)=%v want %v", c.v, got, c.want)
		}
	}
	if !math.IsNaN(ToNumber("abc")) {
		t.Error("ToNumber(abc) should be NaN")
	}
}

func TestToInteger(t *testing.T) {
	if ToInteger(3.9) != 3 {
		t.Error("trunc positive")
	}
	if ToInteger(-3.9) != -3 {
		t.Error("trunc negative")
	}
	if ToInteger(math.NaN()) != 0 {
		t.Error("NaN")
	}
	if ToInteger("5") != 5 {
		t.Error("string")
	}
}

func TestToFinite(t *testing.T) {
	if ToFinite(math.NaN()) != 0 {
		t.Error("NaN")
	}
	if ToFinite(math.Inf(1)) != math.MaxFloat64 {
		t.Error("+Inf")
	}
	if ToFinite(math.Inf(-1)) != -math.MaxFloat64 {
		t.Error("-Inf")
	}
	if ToFinite(2.5) != 2.5 {
		t.Error("finite")
	}
}

func TestEq(t *testing.T) {
	if !Eq(1, 1.0) {
		t.Error("1 == 1.0")
	}
	if !Eq(math.NaN(), math.NaN()) {
		t.Error("NaN eq NaN")
	}
	if !Eq("a", "a") {
		t.Error("string eq")
	}
	if Eq(1, 2) {
		t.Error("1 != 2")
	}
}

func TestComparisons(t *testing.T) {
	if !Gt(2, 1) || Gt(1, 2) {
		t.Error("Gt")
	}
	if !Gte(2, 2) || !Gte(3, 2) {
		t.Error("Gte")
	}
	if !Lt(1, 2) || Lt(2, 1) {
		t.Error("Lt")
	}
	if !Lte(2, 2) || !Lte(1, 2) {
		t.Error("Lte")
	}
	if !Gt("b", "a") {
		t.Error("Gt strings")
	}
}

func TestTimes(t *testing.T) {
	got := Times(3, func(i int) int { return i * i })
	if !reflect.DeepEqual(got, []int{0, 1, 4}) {
		t.Errorf("Times=%v", got)
	}
	if len(Times(0, func(i int) int { return i })) != 0 {
		t.Error("Times zero")
	}
}

func TestIdentityConstant(t *testing.T) {
	if Identity(7) != 7 {
		t.Error("Identity")
	}
	c := Constant("x")
	v1, v2 := c(), c()
	if v1 != "x" || v2 != "x" {
		t.Error("Constant")
	}
}

func TestNoop(t *testing.T) {
	Noop() // must not panic
}

func TestRange(t *testing.T) {
	if !reflect.DeepEqual(Range(4), []int{0, 1, 2, 3}) {
		t.Error("Range")
	}
	if len(Range(0)) != 0 || len(Range(-1)) != 0 {
		t.Error("Range non-positive")
	}
}

func TestUniqueId(t *testing.T) {
	a := UniqueId("id_")
	b := UniqueId("id_")
	if a == b {
		t.Errorf("UniqueId not unique: %s %s", a, b)
	}
	if a[:3] != "id_" {
		t.Errorf("prefix missing: %s", a)
	}
}

func TestOnce(t *testing.T) {
	calls := 0
	f := Once(func() int {
		calls++
		return calls
	})
	a, b := f(), f()
	if a != 1 || b != 1 || calls != 1 {
		t.Errorf("Once ran more than once: calls=%d", calls)
	}
}
