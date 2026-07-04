package chunk

import (
	"reflect"
	"testing"
)

func TestChunkInts(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		size int
		want [][]int
	}{
		{"even split", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"uneven split", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"size one", []int{1, 2, 3}, 1, [][]int{{1}, {2}, {3}}},
		{"size equals length", []int{1, 2, 3}, 3, [][]int{{1, 2, 3}}},
		{"size larger than length", []int{1, 2, 3}, 10, [][]int{{1, 2, 3}}},
		{"empty input", []int{}, 3, [][]int{}},
		{"nil input", nil, 3, [][]int{}},
		{"zero size", []int{1, 2, 3}, 0, [][]int{}},
		{"negative size", []int{1, 2, 3}, -1, [][]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Chunk(tt.in, tt.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk(%v, %d) = %v, want %v", tt.in, tt.size, got, tt.want)
			}
		})
	}
}

func TestChunkStrings(t *testing.T) {
	got := Chunk([]string{"a", "b", "c", "d", "e"}, 3)
	want := [][]string{{"a", "b", "c"}, {"d", "e"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Chunk = %v, want %v", got, want)
	}
}

func TestChunkDoesNotAliasInput(t *testing.T) {
	in := []int{1, 2, 3, 4}
	got := Chunk(in, 2)
	got[0][0] = 99
	if in[0] != 1 {
		t.Errorf("Chunk mutated input: in[0] = %d, want 1", in[0])
	}
}
