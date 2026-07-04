package flat

import (
	"reflect"
	"testing"
)

func TestFlattenSimple(t *testing.T) {
	in := map[string]any{"a": map[string]any{"b": 1}}
	got := Flatten(in)
	want := map[string]any{"a.b": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

func TestFlattenDeep(t *testing.T) {
	in := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "deep",
			},
			"d": 2,
		},
		"e": "top",
	}
	got := Flatten(in)
	want := map[string]any{"a.b.c": "deep", "a.d": 2, "e": "top"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

func TestFlattenEmptyMapLeaf(t *testing.T) {
	in := map[string]any{"a": map[string]any{}}
	got := Flatten(in)
	if len(got) != 1 {
		t.Fatalf("got %#v", got)
	}
	if m, ok := got["a"].(map[string]any); !ok || len(m) != 0 {
		t.Errorf("empty map leaf not preserved: %#v", got)
	}
}

func TestFlattenCustomDelimiter(t *testing.T) {
	in := map[string]any{"a": map[string]any{"b": 1}}
	got := Flatten(in, FlattenOpts{Delimiter: "/"})
	if got["a/b"] != 1 {
		t.Errorf("got %#v", got)
	}
}

func TestUnflatten(t *testing.T) {
	in := map[string]any{"a.b": 1}
	got := Unflatten(in)
	want := map[string]any{"a": map[string]any{"b": 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

func TestUnflattenCustomDelimiter(t *testing.T) {
	in := map[string]any{"a/b/c": 3}
	got := Unflatten(in, UnflattenOpts{Delimiter: "/"})
	want := map[string]any{"a": map[string]any{"b": map[string]any{"c": 3}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	in := map[string]any{
		"name": "root",
		"nested": map[string]any{
			"x": 1,
			"y": map[string]any{"z": "leaf", "list": "kept"},
		},
		"flag": true,
	}
	got := Unflatten(Flatten(in))
	if !reflect.DeepEqual(got, in) {
		t.Errorf("round trip failed:\n got %#v\nwant %#v", got, in)
	}
}

func TestFlattenEmptyTop(t *testing.T) {
	got := Flatten(map[string]any{})
	if len(got) != 0 {
		t.Errorf("expected empty, got %#v", got)
	}
}
