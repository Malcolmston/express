package collection

import (
	"reflect"
	"testing"
)

func TestMapWithIndex(t *testing.T) {
	got := MapWithIndex([]string{"a", "b", "c"}, func(v string, i int) string {
		return v + string(rune('0'+i))
	})
	if !reflect.DeepEqual(got, []string{"a0", "b1", "c2"}) {
		t.Errorf("MapWithIndex = %v", got)
	}
}

func TestFilterWithIndex(t *testing.T) {
	got := FilterWithIndex([]int{10, 20, 30, 40}, func(_ int, i int) bool { return i%2 == 0 })
	if !reflect.DeepEqual(got, []int{10, 30}) {
		t.Errorf("FilterWithIndex = %v", got)
	}
}

func TestReduceWithIndex(t *testing.T) {
	got := ReduceWithIndex([]int{5, 5, 5}, func(acc, cur, i int) int { return acc + cur*i }, 0)
	if got != 0*5+1*5+2*5 {
		t.Errorf("ReduceWithIndex = %d", got)
	}
}

func TestFlatMapWithIndex(t *testing.T) {
	got := FlatMapWithIndex([]int{1, 2}, func(v, i int) []int { return []int{v, i} })
	if !reflect.DeepEqual(got, []int{1, 0, 2, 1}) {
		t.Errorf("FlatMapWithIndex = %v", got)
	}
}

func TestForEachRight(t *testing.T) {
	var order []int
	ForEachRight([]int{1, 2, 3}, func(v int) bool { order = append(order, v); return true })
	if !reflect.DeepEqual(order, []int{3, 2, 1}) {
		t.Errorf("ForEachRight order = %v", order)
	}
	var partial []int
	ForEachRight([]int{1, 2, 3, 4}, func(v int) bool { partial = append(partial, v); return v != 3 })
	if !reflect.DeepEqual(partial, []int{4, 3}) {
		t.Errorf("ForEachRight early stop = %v", partial)
	}
}

func TestCountWhere(t *testing.T) {
	if got := CountWhere([]int{1, 2, 3, 4, 5, 6}, func(n int) bool { return n%2 == 0 }); got != 3 {
		t.Errorf("CountWhere = %d", got)
	}
}

func TestTapThru(t *testing.T) {
	seen := 0
	got := Tap(42, func(v int) { seen = v })
	if got != 42 || seen != 42 {
		t.Error("Tap")
	}
	if r := Thru(10, func(v int) string { return "n" + string(rune('0'+v%10)) }); r != "n0" {
		t.Errorf("Thru = %q", r)
	}
}
