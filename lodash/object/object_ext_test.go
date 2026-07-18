package object

import (
	"reflect"
	"sort"
	"testing"
)

func TestAt(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := At(m, "a", "c", "z")
	if !reflect.DeepEqual(got, []int{1, 3, 0}) {
		t.Errorf("At = %v", got)
	}
}

func TestAssignWith(t *testing.T) {
	dst := map[string]int{"a": 1, "b": 2}
	src := map[string]int{"b": 10, "c": 3}
	got := AssignWith(dst, src, func(d, s int, _ string) int { return d + s })
	want := map[string]int{"a": 1, "b": 12, "c": 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("AssignWith = %v, want %v", got, want)
	}
	// input not mutated
	if dst["b"] != 2 {
		t.Error("AssignWith mutated dst")
	}
}

func TestEverySomeEntry(t *testing.T) {
	m := map[string]int{"a": 2, "b": 4, "c": 6}
	if !EveryEntry(m, func(_ string, v int) bool { return v%2 == 0 }) {
		t.Error("EveryEntry")
	}
	if EveryEntry(m, func(_ string, v int) bool { return v > 3 }) {
		t.Error("EveryEntry false")
	}
	if !SomeEntry(m, func(_ string, v int) bool { return v == 6 }) {
		t.Error("SomeEntry")
	}
	if SomeEntry(m, func(_ string, v int) bool { return v > 100 }) {
		t.Error("SomeEntry false")
	}
	if !EveryEntry(map[string]int{}, func(string, int) bool { return false }) {
		t.Error("EveryEntry empty")
	}
}

func TestFindEntry(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	k, v, ok := FindEntry(m, func(_ string, v int) bool { return v == 2 })
	if !ok || k != "b" || v != 2 {
		t.Errorf("FindEntry = %q,%d,%v", k, v, ok)
	}
	if _, _, ok := FindEntry(m, func(string, int) bool { return false }); ok {
		t.Error("FindEntry no match")
	}
}

func TestReduceEntries(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	sum := ReduceEntries(m, func(acc int, _ string, v int) int { return acc + v }, 0)
	if sum != 6 {
		t.Errorf("ReduceEntries sum = %d", sum)
	}
	keys := ReduceEntries(m, func(acc []string, k string, _ int) []string { return append(acc, k) }, nil)
	sort.Strings(keys)
	if !reflect.DeepEqual(keys, []string{"a", "b", "c"}) {
		t.Errorf("ReduceEntries keys = %v", keys)
	}
}
