package querystringlite

import (
	"reflect"
	"testing"
)

func TestParseBasic(t *testing.T) {
	got := Parse("a=1&b=2&a=3")
	want := map[string][]string{
		"a": {"1", "3"},
		"b": {"2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseEmpty(t *testing.T) {
	got := Parse("")
	if len(got) != 0 {
		t.Fatalf("empty string should parse to empty map, got %v", got)
	}
}

func TestParseMissingEquals(t *testing.T) {
	got := Parse("a&b=2")
	want := map[string][]string{
		"a": {""},
		"b": {"2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParseEmptyValue(t *testing.T) {
	got := Parse("a=&b=2")
	want := map[string][]string{
		"a": {""},
		"b": {"2"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestParsePlusIsSpace(t *testing.T) {
	got := Parse("q=hello+world")
	if got["q"][0] != "hello world" {
		t.Fatalf("+ should decode to space, got %q", got["q"][0])
	}
}

func TestParsePercentDecoding(t *testing.T) {
	got := Parse("q=hello%20world&name=a%26b")
	if got["q"][0] != "hello world" {
		t.Fatalf("%%20 should decode to space, got %q", got["q"][0])
	}
	if got["name"][0] != "a&b" {
		t.Fatalf("%%26 should decode to &, got %q", got["name"][0])
	}
}

func TestParseEncodedKey(t *testing.T) {
	got := Parse("a%20b=1")
	if _, ok := got["a b"]; !ok {
		t.Fatalf("key should be decoded, got %v", got)
	}
}

func TestParseMalformedPercent(t *testing.T) {
	got := Parse("a=%2&b=%zz")
	// Malformed sequences are left as-is.
	if got["a"][0] != "%2" {
		t.Fatalf("malformed %%2 should stay literal, got %q", got["a"][0])
	}
	if got["b"][0] != "%zz" {
		t.Fatalf("malformed %%zz should stay literal, got %q", got["b"][0])
	}
}

func TestStringifyBasic(t *testing.T) {
	got := Stringify(map[string][]string{
		"a": {"1", "3"},
		"b": {"2"},
	})
	// Keys are sorted; repeated keys for multi-values.
	want := "a=1&a=3&b=2"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStringifyEncoding(t *testing.T) {
	got := Stringify(map[string][]string{
		"na me": {"a&b c"},
	})
	want := "na%20me=a%26b%20c"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestStringifyEmptyValue(t *testing.T) {
	got := Stringify(map[string][]string{"a": {""}})
	if got != "a=" {
		t.Fatalf("got %q, want %q", got, "a=")
	}
}

func TestStringifyEmptyMap(t *testing.T) {
	if got := Stringify(map[string][]string{}); got != "" {
		t.Fatalf("empty map should stringify to empty, got %q", got)
	}
}

func TestStringifySkipsKeyWithNoValues(t *testing.T) {
	got := Stringify(map[string][]string{
		"a": {},
		"b": {"1"},
	})
	if got != "b=1" {
		t.Fatalf("key with no values should be skipped, got %q", got)
	}
}

func TestRoundTrip(t *testing.T) {
	original := map[string][]string{
		"name":  {"John Doe"},
		"tags":  {"go", "web", "a&b"},
		"empty": {""},
		"sym":   {"100% sure!"},
	}
	encoded := Stringify(original)
	decoded := Parse(encoded)
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("round trip mismatch:\n encoded=%q\n decoded=%v\n want=%v", encoded, decoded, original)
	}
}

func TestEscapeUnescapeRoundTrip(t *testing.T) {
	cases := []string{
		"hello world",
		"a&b=c?d#e",
		"100%",
		"unicode: éè",
		"tab\tnewline\n",
		"",
	}
	for _, c := range cases {
		if got := Unescape(Escape(c)); got != c {
			t.Fatalf("round trip of %q gave %q", c, got)
		}
	}
}

func TestEscapeSpaceIsPercent20(t *testing.T) {
	if got := Escape("a b"); got != "a%20b" {
		t.Fatalf("space should encode to %%20, got %q", got)
	}
}

func TestParseSingle(t *testing.T) {
	got := ParseSingle("a=1&a=2&b=3")
	if got["a"] != "1" {
		t.Fatalf("ParseSingle should keep first value, got %q", got["a"])
	}
	if got["b"] != "3" {
		t.Fatalf("got %q", got["b"])
	}
}

func TestStringifySingle(t *testing.T) {
	got := StringifySingle(map[string]string{"a": "1", "b": "hello world"})
	// Sorted keys: a=1&b=hello%20world
	want := "a=1&b=hello%20world"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
