package uniq

import (
	"math"
	"reflect"
	"testing"
)

func TestUniqInts(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"basic dedup", []int{2, 1, 2, 3, 1}, []int{2, 1, 3}},
		{"no duplicates", []int{1, 2, 3}, []int{1, 2, 3}},
		{"all same", []int{5, 5, 5, 5}, []int{5}},
		{"empty", []int{}, []int{}},
		{"nil", nil, []int{}},
		{"preserve first order", []int{3, 1, 3, 2, 1}, []int{3, 1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Uniq(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Uniq(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestUniqStrings(t *testing.T) {
	got := Uniq([]string{"a", "b", "a", "c", "b"})
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Uniq = %v, want %v", got, want)
	}
}

func TestUniqByLength(t *testing.T) {
	in := []string{"one", "two", "three", "four", "six"}
	got := UniqBy(in, func(s string) int { return len(s) })
	// lengths: 3,3,5,4,3 -> keep first of each length: "one"(3), "three"(5), "four"(4)
	want := []string{"one", "three", "four"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy = %v, want %v", got, want)
	}
}

func TestUniqByFloatFloor(t *testing.T) {
	in := []float64{2.1, 1.2, 2.3}
	got := UniqBy(in, func(f float64) float64 { return math.Floor(f) })
	want := []float64{2.1, 1.2}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy = %v, want %v", got, want)
	}
}

func TestUniqByStruct(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	in := []user{{1, "a"}, {2, "b"}, {1, "c"}}
	got := UniqBy(in, func(u user) int { return u.ID })
	want := []user{{1, "a"}, {2, "b"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UniqBy = %v, want %v", got, want)
	}
}

func TestUniqByEmpty(t *testing.T) {
	got := UniqBy([]int(nil), func(i int) int { return i })
	if len(got) != 0 {
		t.Errorf("UniqBy(nil) = %v, want empty", got)
	}
}
