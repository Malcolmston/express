package util

import (
	"reflect"
	"testing"
)

func TestCond(t *testing.T) {
	classify := Cond(
		CondPair[int, string]{When: func(n int) bool { return n < 0 }, Then: func(int) string { return "neg" }},
		CondPair[int, string]{When: func(n int) bool { return n == 0 }, Then: func(int) string { return "zero" }},
		CondPair[int, string]{When: func(n int) bool { return n > 0 }, Then: func(int) string { return "pos" }},
	)
	cases := map[int]string{-5: "neg", 0: "zero", 7: "pos"}
	for in, want := range cases {
		got, ok := classify(in)
		if !ok || got != want {
			t.Errorf("Cond(%d) = %q,%v want %q", in, got, ok, want)
		}
	}
	noMatch := Cond(CondPair[int, string]{When: func(int) bool { return false }, Then: func(int) string { return "x" }})
	if _, ok := noMatch(1); ok {
		t.Error("expected no match")
	}
}

func TestStubs(t *testing.T) {
	if a := StubArray[int](); a == nil || len(a) != 0 {
		t.Error("StubArray")
	}
	if m := StubObject[string, int](); m == nil || len(m) != 0 {
		t.Error("StubObject")
	}
	if !StubTrue() || StubFalse() || StubString() != "" {
		t.Error("stub scalars")
	}
}

func TestToPath(t *testing.T) {
	tests := []struct {
		in   string
		want []string
	}{
		{"a.b.c", []string{"a", "b", "c"}},
		{"a[0].b", []string{"a", "0", "b"}},
		{"a[0][1]", []string{"a", "0", "1"}},
		{`a["x.y"].z`, []string{"a", "x.y", "z"}},
		{"single", []string{"single"}},
	}
	for _, tt := range tests {
		if got := ToPath(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ToPath(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestRangeStep(t *testing.T) {
	tests := []struct {
		start, end, step int
		want             []int
	}{
		{0, 5, 1, []int{0, 1, 2, 3, 4}},
		{0, 20, 5, []int{0, 5, 10, 15}},
		{5, 0, -1, []int{5, 4, 3, 2, 1}},
		{0, 5, 0, []int{}},
		{0, 5, -1, []int{}},
	}
	for _, tt := range tests {
		if got := RangeStep(tt.start, tt.end, tt.step); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("RangeStep(%d,%d,%d) = %v, want %v", tt.start, tt.end, tt.step, got, tt.want)
		}
	}
}

func TestToIndex(t *testing.T) {
	if n, ok := ToIndex("3"); !ok || n != 3 {
		t.Error("ToIndex valid")
	}
	if _, ok := ToIndex("-1"); ok {
		t.Error("ToIndex negative")
	}
	if _, ok := ToIndex("x"); ok {
		t.Error("ToIndex non-numeric")
	}
}
