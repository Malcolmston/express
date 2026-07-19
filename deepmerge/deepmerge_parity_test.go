package deepmerge

// Parity tests transcribed from the upstream npm library "TehShrike/deepmerge"
// (v4.3.1). Concrete input -> expected-output vectors are taken verbatim from
// the upstream test suite:
//
//   https://raw.githubusercontent.com/TehShrike/deepmerge/master/test/merge.js
//   https://raw.githubusercontent.com/TehShrike/deepmerge/master/test/merge-all.js
//
// The Go port operates on map[string]any leaves rather than arbitrary
// JavaScript objects, so upstream vectors whose top-level value is an Array,
// or that exercise JS-only options (customMerge, isMergeableObject, clone
// flag, Symbol keys), are documented as gaps in the returned notes rather than
// forced into this file. Every vector below is a faithful map-domain
// transcription of an upstream case. JS primitives that have no Go equivalent
// are represented by the nearest stdlib leaf: Date -> time.Time,
// RegExp -> *regexp.Regexp, null -> nil, undefined -> nil.

import (
	"reflect"
	"regexp"
	"testing"
	"time"
)

// upstream: "add keys in target that do not exist at the root"
func TestParityAddKeysAtRoot(t *testing.T) {
	src := map[string]any{"key1": "value1", "key2": "value2"}
	target := map[string]any{}

	res := Merge(target, src)

	if !reflect.DeepEqual(target, map[string]any{}) {
		t.Errorf("merge should be immutable, target = %#v", target)
	}
	if !reflect.DeepEqual(res, map[string]any{"key1": "value1", "key2": "value2"}) {
		t.Errorf("got %#v", res)
	}
}

// upstream: "merge existing simple keys in target at the roots"
func TestParityMergeSimpleKeysAtRoot(t *testing.T) {
	src := map[string]any{"key1": "changed", "key2": "value2"}
	target := map[string]any{"key1": "value1", "key3": "value3"}

	expected := map[string]any{
		"key1": "changed",
		"key2": "value2",
		"key3": "value3",
	}

	if !reflect.DeepEqual(target, map[string]any{"key1": "value1", "key3": "value3"}) {
		t.Errorf("target mutated: %#v", target)
	}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "merge nested objects into target"
func TestParityMergeNestedObjects(t *testing.T) {
	src := map[string]any{
		"key1": map[string]any{"subkey1": "changed", "subkey3": "added"},
	}
	target := map[string]any{
		"key1": map[string]any{"subkey1": "value1", "subkey2": "value2"},
	}
	expected := map[string]any{
		"key1": map[string]any{"subkey1": "changed", "subkey2": "value2", "subkey3": "added"},
	}

	if !reflect.DeepEqual(target, map[string]any{
		"key1": map[string]any{"subkey1": "value1", "subkey2": "value2"},
	}) {
		t.Errorf("target mutated: %#v", target)
	}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "replace simple key with nested object in target"
func TestParityReplaceSimpleKeyWithNested(t *testing.T) {
	src := map[string]any{
		"key1": map[string]any{"subkey1": "subvalue1", "subkey2": "subvalue2"},
	}
	target := map[string]any{"key1": "value1", "key2": "value2"}
	expected := map[string]any{
		"key1": map[string]any{"subkey1": "subvalue1", "subkey2": "subvalue2"},
		"key2": "value2",
	}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should add nested object in target"
func TestParityAddNestedObject(t *testing.T) {
	src := map[string]any{"b": map[string]any{"c": map[string]any{}}}
	target := map[string]any{"a": map[string]any{}}
	expected := map[string]any{
		"a": map[string]any{},
		"b": map[string]any{"c": map[string]any{}},
	}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should clone source and target" (structure + reference independence)
func TestParityCloneSourceAndTarget(t *testing.T) {
	src := map[string]any{"b": map[string]any{"c": "foo"}}
	target := map[string]any{"a": map[string]any{"d": "bar"}}
	expected := map[string]any{
		"a": map[string]any{"d": "bar"},
		"b": map[string]any{"c": "foo"},
	}

	merged := Merge(target, src)
	if !reflect.DeepEqual(merged, expected) {
		t.Errorf("got %#v, want %#v", merged, expected)
	}

	// merged.a must not be the same map instance as target.a; mutation must not leak.
	merged["a"].(map[string]any)["d"] = "changed"
	if target["a"].(map[string]any)["d"] != "bar" {
		t.Errorf("merged.a shares reference with target.a")
	}
	merged["b"].(map[string]any)["c"] = "changed"
	if src["b"].(map[string]any)["c"] != "foo" {
		t.Errorf("merged.b shares reference with src.b")
	}
}

// upstream: "should replace object with simple key in target"
func TestParityReplaceObjectWithSimpleKey(t *testing.T) {
	src := map[string]any{"key1": "value1"}
	target := map[string]any{
		"key1": map[string]any{"subkey1": "subvalue1", "subkey2": "subvalue2"},
		"key2": "value2",
	}
	expected := map[string]any{"key1": "value1", "key2": "value2"}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should replace objects with arrays"
func TestParityReplaceObjectsWithArrays(t *testing.T) {
	target := map[string]any{"key1": map[string]any{"subkey": "one"}}
	src := map[string]any{"key1": []any{"subkey"}}
	expected := map[string]any{"key1": []any{"subkey"}}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should replace arrays with objects"
func TestParityReplaceArraysWithObjects(t *testing.T) {
	target := map[string]any{"key1": []any{"subkey"}}
	src := map[string]any{"key1": map[string]any{"subkey": "one"}}
	expected := map[string]any{"key1": map[string]any{"subkey": "one"}}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should replace dates with arrays" (Date -> time.Time leaf)
func TestParityReplaceDatesWithArrays(t *testing.T) {
	target := map[string]any{"key1": time.Now()}
	src := map[string]any{"key1": []any{"subkey"}}
	expected := map[string]any{"key1": []any{"subkey"}}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should replace null with arrays" (null -> nil)
func TestParityReplaceNullWithArrays(t *testing.T) {
	target := map[string]any{"key1": nil}
	src := map[string]any{"key1": []any{"subkey"}}
	expected := map[string]any{"key1": []any{"subkey"}}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should work on array properties" (array values concatenate)
func TestParityArrayProperties(t *testing.T) {
	src := map[string]any{
		"key1": []any{"one", "three"},
		"key2": []any{"four"},
	}
	target := map[string]any{"key1": []any{"one", "two"}}
	expected := map[string]any{
		"key1": []any{"one", "two", "one", "three"},
		"key2": []any{"four"},
	}
	if got := Merge(target, src); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should work on array properties with clone option" (reference independence)
func TestParityArrayPropertiesClone(t *testing.T) {
	src := map[string]any{
		"key1": []any{"one", "three"},
		"key2": []any{"four"},
	}
	target := map[string]any{"key1": []any{"one", "two"}}

	merged := Merge(target, src)
	// Mutating merged.key2 must not reach into src.key2 (which the port clones).
	merged["key2"].([]any)[0] = "mutated"
	if src["key2"].([]any)[0] != "four" {
		t.Errorf("merged.key2 shares reference with src.key2")
	}
}

// upstream: "should treat regular expressions like primitive values"
// (RegExp -> *regexp.Regexp leaf; source value replaces target)
func TestParityRegexpAsPrimitive(t *testing.T) {
	abc := regexp.MustCompile("abc")
	efg := regexp.MustCompile("efg")
	target := map[string]any{"key1": abc}
	src := map[string]any{"key1": efg}

	got := Merge(target, src)
	if got["key1"].(*regexp.Regexp).String() != "efg" {
		t.Errorf("regex leaf not replaced, got %v", got["key1"])
	}
	// A non-mergeable leaf is carried through by identity, not deep-cloned.
	if got["key1"] != efg {
		t.Errorf("regex leaf should be carried through by identity")
	}
	if !got["key1"].(*regexp.Regexp).MatchString("efg") {
		t.Errorf("resulting regex should match 'efg'")
	}
}

// upstream: "should treat dates like primitives" (Date -> time.Time; source wins)
func TestParityDatesAsPrimitive(t *testing.T) {
	monday, _ := time.Parse(time.RFC3339, "2016-09-27T01:08:12.761Z")
	tuesday, _ := time.Parse(time.RFC3339, "2016-09-28T01:18:12.761Z")

	target := map[string]any{"key": monday}
	source := map[string]any{"key": tuesday}
	expected := map[string]any{"key": tuesday}

	got := Merge(target, source)
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
	if !got["key"].(time.Time).Equal(tuesday) {
		t.Errorf("date leaf value should equal tuesday")
	}
}

// upstream: "should clone an array property when there is no target array"
func TestParityCloneArrayPropertyNoTarget(t *testing.T) {
	someObject := map[string]any{}
	target := map[string]any{}
	source := map[string]any{"ary": []any{someObject}}

	output := Merge(target, source)
	if !reflect.DeepEqual(output, map[string]any{"ary": []any{map[string]any{}}}) {
		t.Errorf("got %#v", output)
	}
	// output.ary[0] must be a clone, not the original someObject.
	output["ary"].([]any)[0].(map[string]any)["mutated"] = true
	if len(someObject) != 0 {
		t.Errorf("output.ary[0] shares reference with source object")
	}
}

// upstream: "should overwrite values when property is initialised but undefined"
// (undefined -> nil; the key survives with a nil value)
func TestParityOverwriteWithUndefined(t *testing.T) {
	src := map[string]any{"value": nil}

	for _, target := range []map[string]any{
		{"value": []any{}},
		{"value": nil},
		{"value": 2},
	} {
		got := Merge(target, src)
		v, ok := got["value"]
		if !ok {
			t.Errorf("expected 'value' key to be present, got %#v", got)
		}
		if v != nil {
			t.Errorf("expected 'value' to be nil, got %#v", v)
		}
	}
}

// upstream: "dates should copy correctly in an array" (as an array property;
// Date -> time.Time; arrays concatenate, elements carried through)
func TestParityDatesCopyInArray(t *testing.T) {
	monday, _ := time.Parse(time.RFC3339, "2016-09-27T01:08:12.761Z")
	tuesday, _ := time.Parse(time.RFC3339, "2016-09-28T01:18:12.761Z")

	target := map[string]any{"k": []any{monday, "dude"}}
	source := map[string]any{"k": []any{tuesday, "lol"}}
	expected := map[string]any{"k": []any{monday, "dude", tuesday, "lol"}}

	if got := Merge(target, source); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream: "should merge correctly if custom merge is not a valid function"
// The Go port has no customMerge hook, so this reduces to the default merge,
// which is exactly the upstream-expected result when customMerge is ignored.
func TestParityDefaultMergeMatchesInvalidCustomMerge(t *testing.T) {
	target := map[string]any{
		"letters": []any{"a", "b"},
		"people":  map[string]any{"first": "Alex", "second": "Bert"},
	}
	source := map[string]any{
		"letters": []any{"c"},
		"people":  map[string]any{"first": "Smith", "second": "Bertson", "third": "Car"},
	}
	expected := map[string]any{
		"letters": []any{"a", "b", "c"},
		"people":  map[string]any{"first": "Smith", "second": "Bertson", "third": "Car"},
	}
	if got := Merge(target, source); !reflect.DeepEqual(got, expected) {
		t.Errorf("got %#v, want %#v", got, expected)
	}
}

// upstream merge-all: "return an empty object if first argument is an array with no elements"
func TestParityMergeAllEmptyArray(t *testing.T) {
	if got := MergeAll(); !reflect.DeepEqual(got, map[string]any{}) {
		t.Errorf("got %#v, want empty map", got)
	}
}

// upstream merge-all: "Work just fine if first argument is an array with least than two elements"
func TestParityMergeAllSingleElement(t *testing.T) {
	got := MergeAll(map[string]any{"example": true})
	if !reflect.DeepEqual(got, map[string]any{"example": true}) {
		t.Errorf("got %#v", got)
	}
}

// upstream merge-all: "invoke merge on every item in array should result with all props"
func TestParityMergeAllAllProps(t *testing.T) {
	got := MergeAll(
		map[string]any{"first": true},
		map[string]any{"second": false},
		map[string]any{"third": 123},
		map[string]any{"fourth": "some string"},
	)
	if got["first"] != true || got["second"] != false || got["third"] != 123 || got["fourth"] != "some string" {
		t.Errorf("got %#v", got)
	}
}

// upstream merge-all: "invoke merge on every item in array with clone should clone all elements"
func TestParityMergeAllClonesElements(t *testing.T) {
	firstObject := map[string]any{"a": map[string]any{"d": 123}}
	secondObject := map[string]any{"b": map[string]any{"e": true}}
	thirdObject := map[string]any{"c": map[string]any{"f": "string"}}

	merged := MergeAll(firstObject, secondObject, thirdObject)

	// Each nested object in the result must be an independent clone.
	merged["a"].(map[string]any)["d"] = 999
	merged["b"].(map[string]any)["e"] = false
	merged["c"].(map[string]any)["f"] = "changed"
	if firstObject["a"].(map[string]any)["d"] != 123 {
		t.Errorf("merged.a shares reference with firstObject.a")
	}
	if secondObject["b"].(map[string]any)["e"] != true {
		t.Errorf("merged.b shares reference with secondObject.b")
	}
	if thirdObject["c"].(map[string]any)["f"] != "string" {
		t.Errorf("merged.c shares reference with thirdObject.c")
	}
}
