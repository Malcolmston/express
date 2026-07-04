package object

import (
	"reflect"
	"sort"
	"testing"
)

func sortedKeys[V any](m map[string]V) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func TestKeysValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	ks := Keys(m)
	sort.Strings(ks)
	if !reflect.DeepEqual(ks, []string{"a", "b", "c"}) {
		t.Fatalf("Keys = %v", ks)
	}
	vs := Values(m)
	sort.Ints(vs)
	if !reflect.DeepEqual(vs, []int{1, 2, 3}) {
		t.Fatalf("Values = %v", vs)
	}
}

func TestEntriesFromEntries(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	pairs := Entries(m)
	if len(pairs) != 2 {
		t.Fatalf("Entries len = %d", len(pairs))
	}
	back := FromEntries(pairs)
	if !reflect.DeepEqual(back, m) {
		t.Fatalf("FromEntries roundtrip = %v", back)
	}
	if len(ToPairs(m)) != 2 {
		t.Fatalf("ToPairs len wrong")
	}
}

func TestPickPickBy(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2, "c": 3}
	got := Pick(m, "a", "c", "missing")
	if !reflect.DeepEqual(got, map[string]any{"a": 1, "c": 3}) {
		t.Fatalf("Pick = %v", got)
	}
	gotBy := PickBy(m, func(v any, k string) bool { return v.(int) > 1 })
	if !reflect.DeepEqual(sortedKeys(gotBy), []string{"b", "c"}) {
		t.Fatalf("PickBy = %v", gotBy)
	}
}

func TestOmitOmitBy(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2, "c": 3}
	got := Omit(m, "b")
	if !reflect.DeepEqual(got, map[string]any{"a": 1, "c": 3}) {
		t.Fatalf("Omit = %v", got)
	}
	gotBy := OmitBy(m, func(v any, k string) bool { return v.(int) > 1 })
	if !reflect.DeepEqual(gotBy, map[string]any{"a": 1}) {
		t.Fatalf("OmitBy = %v", gotBy)
	}
}

func TestMapKeysMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	mk := MapKeys(m, func(v int, k string) string { return k + k })
	if !reflect.DeepEqual(mk, map[string]int{"aa": 1, "bb": 2}) {
		t.Fatalf("MapKeys = %v", mk)
	}
	mv := MapValues(m, func(v int, k string) int { return v * 10 })
	if !reflect.DeepEqual(mv, map[string]int{"a": 10, "b": 20}) {
		t.Fatalf("MapValues = %v", mv)
	}
}

func TestInvertInvertBy(t *testing.T) {
	m := map[string]string{"a": "1", "b": "2"}
	inv := Invert(m)
	if !reflect.DeepEqual(inv, map[string]string{"1": "a", "2": "b"}) {
		t.Fatalf("Invert = %v", inv)
	}
	m2 := map[string]int{"a": 1, "b": 2, "c": 1}
	by := InvertBy(m2, func(v int) int { return v })
	sort.Strings(by[1])
	if !reflect.DeepEqual(by[1], []string{"a", "c"}) || !reflect.DeepEqual(by[2], []string{"b"}) {
		t.Fatalf("InvertBy = %v", by)
	}
}

func TestDefaults(t *testing.T) {
	dst := map[string]any{"a": 1, "b": nil}
	Defaults(dst, map[string]any{"a": 99, "b": 2, "c": 3})
	want := map[string]any{"a": 1, "b": 2, "c": 3}
	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("Defaults = %v", dst)
	}
}

func TestDefaultsDeep(t *testing.T) {
	dst := map[string]any{"user": map[string]any{"name": "amy"}}
	src := map[string]any{"user": map[string]any{"name": "zed", "age": 30}}
	DefaultsDeep(dst, src)
	want := map[string]any{"user": map[string]any{"name": "amy", "age": 30}}
	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("DefaultsDeep = %v", dst)
	}
	// mutating src's nested map must not affect dst (deep clone of filled key)
	dst2 := map[string]any{}
	src2 := map[string]any{"x": map[string]any{"y": 1}}
	DefaultsDeep(dst2, src2)
	src2["x"].(map[string]any)["y"] = 999
	if Get(dst2, "x.y") != 1 {
		t.Fatalf("DefaultsDeep did not clone: %v", dst2)
	}
}

func TestAssign(t *testing.T) {
	dst := map[string]any{"a": 1}
	Assign(dst, map[string]any{"a": 2, "b": 3}, map[string]any{"c": 4})
	want := map[string]any{"a": 2, "b": 3, "c": 4}
	if !reflect.DeepEqual(dst, want) {
		t.Fatalf("Assign = %v", dst)
	}
}

func TestMergeDeep(t *testing.T) {
	dst := map[string]any{
		"a": map[string]any{"x": 1},
		"list": []any{
			map[string]any{"k": 1},
		},
	}
	src := map[string]any{
		"a":    map[string]any{"y": 2},
		"list": []any{map[string]any{"v": 3}, "extra"},
	}
	Merge(dst, src)
	if Get(dst, "a.x") != 1 || Get(dst, "a.y") != 2 {
		t.Fatalf("Merge nested map = %v", dst["a"])
	}
	if Get(dst, "list.0.k") != 1 || Get(dst, "list.0.v") != 3 {
		t.Fatalf("Merge nested slice = %v", dst["list"])
	}
	if Get(dst, "list.1") != "extra" {
		t.Fatalf("Merge slice extend = %v", dst["list"])
	}
}

func TestGetHas(t *testing.T) {
	root := map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{"c": 42},
			},
		},
	}
	if Get(root, "a.b.0.c") != 42 {
		t.Fatalf("Get nested path failed")
	}
	if Get(root, "a.b.5.c", "def") != "def" {
		t.Fatalf("Get default failed")
	}
	if !Has(root, "a.b.0.c") {
		t.Fatalf("Has should be true")
	}
	if Has(root, "a.b.9.c") {
		t.Fatalf("Has should be false")
	}
}

func TestSet(t *testing.T) {
	root := map[string]any{}
	Set(root, "a.b.c", 7)
	if Get(root, "a.b.c") != 7 {
		t.Fatalf("Set nested failed: %v", root)
	}
	// numeric segment builds a slice
	Set(root, "a.arr.2", "hi")
	if Get(root, "a.arr.2") != "hi" {
		t.Fatalf("Set slice failed: %v", root)
	}
	if arr, ok := Get(root, "a.arr").([]any); !ok || len(arr) != 3 {
		t.Fatalf("Set slice grow failed: %v", Get(root, "a.arr"))
	}
	// Set into nil root returns a fresh container
	fresh := Set(nil, "x.y", 1)
	if Get(fresh, "x.y") != 1 {
		t.Fatalf("Set nil root failed: %v", fresh)
	}
}

func TestUpdate(t *testing.T) {
	root := map[string]any{"n": 1}
	Update(root, "n", func(cur any) any { return cur.(int) + 5 })
	if Get(root, "n") != 6 {
		t.Fatalf("Update failed: %v", root)
	}
	Update(root, "missing.deep", func(cur any) any {
		if cur != nil {
			t.Fatalf("expected nil current")
		}
		return "new"
	})
	if Get(root, "missing.deep") != "new" {
		t.Fatalf("Update create failed: %v", root)
	}
}

func TestUnset(t *testing.T) {
	root := map[string]any{"a": map[string]any{"b": 1, "c": 2}}
	if !Unset(root, "a.b") {
		t.Fatalf("Unset should return true")
	}
	if Has(root, "a.b") {
		t.Fatalf("Unset did not remove key")
	}
	if Unset(root, "a.zzz") {
		t.Fatalf("Unset missing should return false")
	}
	rootSlice := map[string]any{"arr": []any{1, 2, 3}}
	if !Unset(rootSlice, "arr.1") {
		t.Fatalf("Unset slice should return true")
	}
	if Get(rootSlice, "arr.1") != nil {
		t.Fatalf("Unset slice should nil element")
	}
}

func TestCloneShallow(t *testing.T) {
	orig := map[string]any{"a": 1, "nested": map[string]any{"x": 1}}
	cl := Clone(orig).(map[string]any)
	cl["a"] = 99
	if orig["a"] != 1 {
		t.Fatalf("Clone top-level not independent")
	}
	// nested is shared
	cl["nested"].(map[string]any)["x"] = 2
	if Get(orig, "nested.x") != 2 {
		t.Fatalf("Clone should share nested references")
	}
}

func TestCloneDeepIndependence(t *testing.T) {
	orig := map[string]any{
		"a": 1,
		"nested": map[string]any{
			"list": []any{1, 2, map[string]any{"deep": "val"}},
		},
	}
	cl := CloneDeep(orig).(map[string]any)
	// mutate clone extensively
	cl["a"] = 999
	cl["nested"].(map[string]any)["list"].([]any)[0] = 111
	cl["nested"].(map[string]any)["list"].([]any)[2].(map[string]any)["deep"] = "changed"

	// original must be untouched
	if orig["a"] != 1 {
		t.Fatalf("CloneDeep: top-level mutated original")
	}
	if Get(orig, "nested.list.0") != 1 {
		t.Fatalf("CloneDeep: slice mutated original")
	}
	if Get(orig, "nested.list.2.deep") != "val" {
		t.Fatalf("CloneDeep: deep map mutated original")
	}
	// and clone reflects changes
	if Get(cl, "nested.list.2.deep") != "changed" {
		t.Fatalf("CloneDeep: clone did not retain change")
	}
}

func TestIsEqual(t *testing.T) {
	a := map[string]any{
		"x": 1,
		"y": []any{1, 2, map[string]any{"z": "q"}},
	}
	b := map[string]any{
		"x": 1,
		"y": []any{1, 2, map[string]any{"z": "q"}},
	}
	if !IsEqual(a, b) {
		t.Fatalf("IsEqual should be true for deeply equal maps")
	}
	c := map[string]any{
		"x": 1,
		"y": []any{1, 2, map[string]any{"z": "DIFFERENT"}},
	}
	if IsEqual(a, c) {
		t.Fatalf("IsEqual should be false when nested scalar differs")
	}
	// different lengths
	if IsEqual([]any{1, 2}, []any{1, 2, 3}) {
		t.Fatalf("IsEqual slice length mismatch should be false")
	}
	// missing key
	if IsEqual(map[string]any{"a": 1}, map[string]any{"b": 1}) {
		t.Fatalf("IsEqual missing key should be false")
	}
	// nil handling
	if !IsEqual(nil, nil) {
		t.Fatalf("IsEqual(nil,nil) should be true")
	}
	if IsEqual(nil, 1) {
		t.Fatalf("IsEqual(nil,1) should be false")
	}
	// scalars
	if !IsEqual("s", "s") || IsEqual("s", "t") {
		t.Fatalf("IsEqual scalar comparison wrong")
	}
}

func TestTransform(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2, "c": 3}
	out := Transform(m, func(acc map[string]any, v any, k string) bool {
		acc[k] = v.(int) * 2
		return true
	}, nil)
	want := map[string]any{"a": 2, "b": 4, "c": 6}
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("Transform = %v", out)
	}
}

func TestForInForOwn(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2, "c": 3}
	sum := 0
	ForIn(m, func(v any, k string) bool { sum += v.(int); return true })
	if sum != 6 {
		t.Fatalf("ForIn sum = %d", sum)
	}
	count := 0
	ForOwn(m, func(v any, k string) bool { count++; return count < 2 })
	if count != 2 {
		t.Fatalf("ForOwn early stop failed: %d", count)
	}
}

func TestFindKey(t *testing.T) {
	m := map[string]any{"a": 1, "b": 2, "c": 3}
	k, ok := FindKey(m, func(v any, key string) bool { return v.(int) == 2 })
	if !ok || k != "b" {
		t.Fatalf("FindKey = %q, %v", k, ok)
	}
	_, ok = FindKey(m, func(v any, key string) bool { return v.(int) == 99 })
	if ok {
		t.Fatalf("FindKey should not find")
	}
}
