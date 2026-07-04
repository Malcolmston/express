package flatten

import (
	"reflect"
	"testing"
)

func TestFlattenInts(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want []int
	}{
		{"basic", [][]int{{1, 2}, {3, 4}}, []int{1, 2, 3, 4}},
		{"jagged", [][]int{{1}, {2, 3, 4}, {}, {5}}, []int{1, 2, 3, 4, 5}},
		{"empty outer", [][]int{}, []int{}},
		{"nil outer", nil, []int{}},
		{"all empty inner", [][]int{{}, {}, {}}, []int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Flatten(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFlattenStrings(t *testing.T) {
	got := Flatten([][]string{{"a"}, {"b", "c"}})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Flatten = %v, want %v", got, want)
	}
}

func TestFlattenOnlyOneLevel(t *testing.T) {
	// Flatten of [][][]int flattens exactly one level, leaving [][]int elements.
	in := [][][]int{{{1, 2}}, {{3}, {4, 5}}}
	got := Flatten(in)
	want := [][]int{{1, 2}, {3}, {4, 5}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Flatten = %v, want %v", got, want)
	}
}

func TestFlattenDeep(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want []any
	}{
		{
			name: "nested any",
			in:   []any{1, []any{2, []any{3, []any{4}}}, 5},
			want: []any{1, 2, 3, 4, 5},
		},
		{
			name: "lodash example",
			in:   []any{1, []any{2, []any{3, []any{4}}, 5}},
			want: []any{1, 2, 3, 4, 5},
		},
		{
			name: "mixed types",
			in:   []any{"a", []any{"b", []any{"c"}}, 1, []any{2}},
			want: []any{"a", "b", "c", 1, 2},
		},
		{
			name: "typed inner slices via reflection",
			in:   []any{[]int{1, 2}, []int{3}, []any{[]string{"x"}}},
			want: []any{1, 2, 3, "x"},
		},
		{
			name: "already flat",
			in:   []any{1, 2, 3},
			want: []any{1, 2, 3},
		},
		{
			name: "empty",
			in:   []any{},
			want: []any{},
		},
		{
			name: "non-slice input",
			in:   42,
			want: []any{},
		},
		{
			name: "strings not split",
			in:   []any{"hello", []any{"world"}},
			want: []any{"hello", "world"},
		},
		{
			name: "deeply nested empties",
			in:   []any{[]any{[]any{}}, 1, []any{[]any{2, []any{}}}},
			want: []any{1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlattenDeep(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlattenDeep(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestFlattenDepth(t *testing.T) {
	nested := []any{1, []any{2, []any{3, []any{4}}}, 5}
	tests := []struct {
		name  string
		depth int
		want  []any
	}{
		{"depth 0", 0, []any{1, []any{2, []any{3, []any{4}}}, 5}},
		{"depth 1", 1, []any{1, 2, []any{3, []any{4}}, 5}},
		{"depth 2", 2, []any{1, 2, 3, []any{4}, 5}},
		{"depth 3", 3, []any{1, 2, 3, 4, 5}},
		{"depth beyond", 10, []any{1, 2, 3, 4, 5}},
		{"negative depth", -1, []any{1, []any{2, []any{3, []any{4}}}, 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FlattenDepth(nested, tt.depth)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlattenDepth(depth=%d) = %v, want %v", tt.depth, got, tt.want)
			}
		})
	}
}

func TestFlattenDepthEmpty(t *testing.T) {
	got := FlattenDepth([]any{}, 3)
	if len(got) != 0 {
		t.Errorf("FlattenDepth(empty) = %v, want empty", got)
	}
}
