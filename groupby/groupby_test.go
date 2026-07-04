package groupby

import (
	"reflect"
	"testing"
)

func TestGroupByParity(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6}
	got := GroupBy(in, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	want := map[string][]int{
		"odd":  {1, 3, 5},
		"even": {2, 4, 6},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

func TestGroupByStringLength(t *testing.T) {
	in := []string{"one", "two", "three", "four", "six"}
	got := GroupBy(in, func(s string) int { return len(s) })
	want := map[int][]string{
		3: {"one", "two", "six"},
		5: {"three"},
		4: {"four"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

func TestGroupByFloatFloor(t *testing.T) {
	in := []float64{6.1, 4.2, 6.3}
	got := GroupBy(in, func(f float64) int { return int(f) })
	want := map[int][]float64{
		6: {6.1, 6.3},
		4: {4.2},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}

func TestGroupByEmpty(t *testing.T) {
	got := GroupBy([]int(nil), func(n int) int { return n })
	if got == nil {
		t.Fatal("GroupBy(nil) returned nil map, want empty non-nil map")
	}
	if len(got) != 0 {
		t.Errorf("GroupBy(nil) = %v, want empty", got)
	}
}

func TestGroupByPreservesOrderWithinGroup(t *testing.T) {
	type item struct {
		Cat string
		N   int
	}
	in := []item{{"a", 1}, {"b", 2}, {"a", 3}, {"a", 4}, {"b", 5}}
	got := GroupBy(in, func(i item) string { return i.Cat })
	want := map[string][]item{
		"a": {{"a", 1}, {"a", 3}, {"a", 4}},
		"b": {{"b", 2}, {"b", 5}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GroupBy = %v, want %v", got, want)
	}
}
