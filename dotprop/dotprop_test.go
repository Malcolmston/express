package dotprop

import (
	"reflect"
	"testing"
)

func TestGetSimple(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"b": map[string]any{"c": 42}}}
	v, ok := Get(obj, "a.b.c")
	if !ok || v != 42 {
		t.Fatalf("Get a.b.c = %v, %v; want 42, true", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"b": 1}}
	if v, ok := Get(obj, "a.b.c"); ok {
		t.Fatalf("Get a.b.c = %v, %v; want _, false", v, ok)
	}
	if _, ok := Get(obj, "x"); ok {
		t.Fatalf("Get x should be false")
	}
}

func TestGetNilAndEmpty(t *testing.T) {
	if _, ok := Get(nil, "a"); ok {
		t.Fatal("Get on nil should be false")
	}
	if _, ok := Get(map[string]any{"a": 1}, ""); ok {
		t.Fatal("Get with empty path should be false")
	}
}

func TestGetTopLevel(t *testing.T) {
	obj := map[string]any{"a": 1}
	v, ok := Get(obj, "a")
	if !ok || v != 1 {
		t.Fatalf("Get a = %v, %v", v, ok)
	}
}

func TestSetCreatesIntermediate(t *testing.T) {
	obj := map[string]any{}
	Set(obj, "a.b.c", "hello")
	v, ok := Get(obj, "a.b.c")
	if !ok || v != "hello" {
		t.Fatalf("after Set, Get a.b.c = %v, %v", v, ok)
	}
}

func TestSetReturnsObj(t *testing.T) {
	obj := map[string]any{}
	got := Set(obj, "x", 1)
	if !reflect.DeepEqual(got, obj) {
		t.Fatal("Set should return the same object")
	}
}

func TestSetOverwritesNonMap(t *testing.T) {
	obj := map[string]any{"a": 5}
	Set(obj, "a.b", 10)
	v, ok := Get(obj, "a.b")
	if !ok || v != 10 {
		t.Fatalf("Set over non-map = %v, %v", v, ok)
	}
}

func TestSetNilNoop(t *testing.T) {
	if got := Set(nil, "a", 1); got != nil {
		t.Fatal("Set on nil should return nil")
	}
}

func TestHas(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"b": nil}}
	if !Has(obj, "a.b") {
		t.Fatal("Has a.b should be true even when value is nil")
	}
	if Has(obj, "a.c") {
		t.Fatal("Has a.c should be false")
	}
}

func TestDelete(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"b": 1, "c": 2}}
	if !Delete(obj, "a.b") {
		t.Fatal("Delete a.b should return true")
	}
	if Has(obj, "a.b") {
		t.Fatal("a.b should be gone")
	}
	if !Has(obj, "a.c") {
		t.Fatal("a.c should remain")
	}
	if Delete(obj, "a.b") {
		t.Fatal("Deleting again should return false")
	}
	if Delete(obj, "x.y.z") {
		t.Fatal("Delete non-existent path should be false")
	}
}

func TestEscapedDot(t *testing.T) {
	obj := map[string]any{}
	Set(obj, `a\.b`, 1)
	if _, ok := obj["a.b"]; !ok {
		t.Fatalf("escaped dot should create key 'a.b'; got %#v", obj)
	}
	v, ok := Get(obj, `a\.b`)
	if !ok || v != 1 {
		t.Fatalf("Get escaped = %v, %v", v, ok)
	}
	if Has(obj, "a.b") {
		t.Fatal("unescaped a.b should not resolve into nested map")
	}
}

func TestNumericMapKey(t *testing.T) {
	obj := map[string]any{"a": map[string]any{"0": "zero"}}
	v, ok := Get(obj, "a.0")
	if !ok || v != "zero" {
		t.Fatalf("numeric map key Get = %v, %v", v, ok)
	}
}

func TestArrayIndex(t *testing.T) {
	obj := map[string]any{"a": []any{
		map[string]any{"b": "first"},
		map[string]any{"b": "second"},
	}}
	v, ok := Get(obj, "a.1.b")
	if !ok || v != "second" {
		t.Fatalf("array index Get = %v, %v", v, ok)
	}
	if !Has(obj, "a.0.b") {
		t.Fatal("Has a.0.b should be true")
	}
	if Has(obj, "a.5.b") {
		t.Fatal("out-of-range index should be false")
	}
	if Has(obj, "a.x") {
		t.Fatal("non-numeric index into slice should be false")
	}
}

func TestParsePathTrailingBackslash(t *testing.T) {
	segs := parsePath(`a\`)
	if len(segs) != 1 || segs[0] != `a\` {
		t.Fatalf("trailing backslash = %#v", segs)
	}
}
