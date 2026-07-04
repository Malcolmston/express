package deepmerge

import (
	"reflect"
	"testing"
)

func TestMergeNested(t *testing.T) {
	target := map[string]any{
		"a": 1,
		"b": map[string]any{"x": 1, "y": 2},
	}
	source := map[string]any{
		"b": map[string]any{"y": 20, "z": 30},
		"c": 3,
	}
	got := Merge(target, source)
	want := map[string]any{
		"a": 1,
		"b": map[string]any{"x": 1, "y": 20, "z": 30},
		"c": 3,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMergeConcatArrays(t *testing.T) {
	target := map[string]any{"list": []any{1, 2}}
	source := map[string]any{"list": []any{3, 4}}
	got := Merge(target, source)
	want := map[string]any{"list": []any{1, 2, 3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMergeDoesNotMutateInputs(t *testing.T) {
	target := map[string]any{
		"b":    map[string]any{"x": 1},
		"list": []any{1, 2},
	}
	source := map[string]any{
		"b":    map[string]any{"y": 2},
		"list": []any{3},
	}
	targetCopy := map[string]any{
		"b":    map[string]any{"x": 1},
		"list": []any{1, 2},
	}
	sourceCopy := map[string]any{
		"b":    map[string]any{"y": 2},
		"list": []any{3},
	}
	got := Merge(target, source)

	if !reflect.DeepEqual(target, targetCopy) {
		t.Errorf("target mutated: %#v", target)
	}
	if !reflect.DeepEqual(source, sourceCopy) {
		t.Errorf("source mutated: %#v", source)
	}

	// Mutating the result must not touch the inputs.
	got["b"].(map[string]any)["x"] = 999
	if target["b"].(map[string]any)["x"] != 1 {
		t.Errorf("result shares nested map with target")
	}
	got["list"] = append(got["list"].([]any), 4)
	if len(target["list"].([]any)) != 2 {
		t.Errorf("result shares slice with target")
	}
}

func TestMergeSourceReplacesScalar(t *testing.T) {
	target := map[string]any{"a": 1}
	source := map[string]any{"a": "two"}
	got := Merge(target, source)
	if got["a"] != "two" {
		t.Errorf("got %#v, want a=two", got)
	}
}

func TestMergeTypeMismatchReplaces(t *testing.T) {
	// Slice in target, map in source -> source replaces.
	target := map[string]any{"a": []any{1, 2}}
	source := map[string]any{"a": map[string]any{"x": 1}}
	got := Merge(target, source)
	want := map[string]any{"a": map[string]any{"x": 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMergeWithReplaceArrays(t *testing.T) {
	replace := func(target, source []any) []any { return source }
	target := map[string]any{"list": []any{1, 2}}
	source := map[string]any{"list": []any{3, 4}}
	got := MergeWith(target, source, Options{ArrayMerge: replace})
	want := map[string]any{"list": []any{3, 4}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMergeAll(t *testing.T) {
	a := map[string]any{"a": 1, "shared": map[string]any{"x": 1}}
	b := map[string]any{"b": 2, "shared": map[string]any{"y": 2}}
	c := map[string]any{"c": 3, "shared": map[string]any{"z": 3}}
	got := MergeAll(a, b, c)
	want := map[string]any{
		"a":      1,
		"b":      2,
		"c":      3,
		"shared": map[string]any{"x": 1, "y": 2, "z": 3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestMergeAllEmpty(t *testing.T) {
	got := MergeAll()
	if len(got) != 0 {
		t.Errorf("expected empty map, got %#v", got)
	}
}

func TestMergeNilInputs(t *testing.T) {
	got := Merge(nil, map[string]any{"a": 1})
	if !reflect.DeepEqual(got, map[string]any{"a": 1}) {
		t.Errorf("got %#v", got)
	}
	got2 := Merge(map[string]any{"a": 1}, nil)
	if !reflect.DeepEqual(got2, map[string]any{"a": 1}) {
		t.Errorf("got %#v", got2)
	}
}

func TestDeepNestedClone(t *testing.T) {
	target := map[string]any{
		"level1": map[string]any{
			"level2": []any{
				map[string]any{"deep": 1},
			},
		},
	}
	got := Merge(target, map[string]any{})
	inner := got["level1"].(map[string]any)["level2"].([]any)[0].(map[string]any)
	inner["deep"] = 999
	orig := target["level1"].(map[string]any)["level2"].([]any)[0].(map[string]any)
	if orig["deep"] != 1 {
		t.Errorf("deep clone failed, original mutated")
	}
}
