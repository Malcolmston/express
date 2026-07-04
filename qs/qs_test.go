package qs

import (
	"reflect"
	"testing"
)

func TestParseFlat(t *testing.T) {
	got := Parse("a=1&b=2")
	want := map[string]any{"a": "1", "b": "2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseNested(t *testing.T) {
	got := Parse("a[b]=1&a[c]=2")
	want := map[string]any{"a": map[string]any{"b": "1", "c": "2"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseArray(t *testing.T) {
	got := Parse("a[]=1&a[]=2")
	want := map[string]any{"a": []any{"1", "2"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseDeepNesting(t *testing.T) {
	got := Parse("a[b][c]=1")
	want := map[string]any{"a": map[string]any{"b": map[string]any{"c": "1"}}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseURLDecoding(t *testing.T) {
	got := Parse("a%20b=c%20d")
	want := map[string]any{"a b": "c d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParsePlusAsSpace(t *testing.T) {
	got := Parse("key=hello+world")
	want := map[string]any{"key": "hello world"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseEmpty(t *testing.T) {
	got := Parse("")
	if len(got) != 0 {
		t.Errorf("got %#v, want empty", got)
	}
}

func TestParseLeadingQuestionMark(t *testing.T) {
	got := Parse("?a=1")
	want := map[string]any{"a": "1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseNoValue(t *testing.T) {
	got := Parse("a=&b")
	want := map[string]any{"a": "", "b": ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestParseNestedArrayOfObjects(t *testing.T) {
	got := Parse("a[][b]=1&a[][b]=2")
	want := map[string]any{"a": []any{
		map[string]any{"b": "1"},
		map[string]any{"b": "2"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestStringifyFlat(t *testing.T) {
	got := Stringify(map[string]any{"b": "2", "a": "1"})
	want := "a=1&b=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringifyNested(t *testing.T) {
	got := Stringify(map[string]any{"a": map[string]any{"c": "2", "b": "1"}})
	want := "a[b]=1&a[c]=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringifyArray(t *testing.T) {
	got := Stringify(map[string]any{"a": []any{"1", "2"}})
	want := "a[]=1&a[]=2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStringifyEncoding(t *testing.T) {
	got := Stringify(map[string]any{"a b": "c d"})
	want := "a+b=c+d"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	cases := []string{
		"a=1&b=2",
		"a[b]=1&a[c]=2",
		"a[]=1&a[]=2",
		"a[b][c]=1",
	}
	for _, s := range cases {
		parsed := Parse(s)
		out := Stringify(parsed)
		reparsed := Parse(out)
		if !reflect.DeepEqual(parsed, reparsed) {
			t.Errorf("round trip mismatch for %q: %#v -> %q -> %#v", s, parsed, out, reparsed)
		}
	}
}

func TestStringifyNonString(t *testing.T) {
	got := Stringify(map[string]any{"n": 42, "b": true})
	want := "b=true&n=42"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
