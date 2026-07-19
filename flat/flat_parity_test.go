package flat

// Upstream parity tests for hughsk/flat (npm "flat", v6.0.1).
//
// All input -> expected-output vectors below are taken verbatim from the
// original library's own test suite:
//
//	https://raw.githubusercontent.com/hughsk/flat/master/test/test.js
//	https://raw.githubusercontent.com/hughsk/flat/master/index.js
//
// Only vectors expressible in this port's data model (map[string]any trees with
// scalar/nil/map leaves) are encoded here. Upstream options with no equivalent
// in the Go port -- maxDepth, transformKey, safe, object, overwrite -- and its
// array/Buffer/typed-array handling are documented as gaps in the task notes
// rather than tested, because the Go signature cannot express them.

import (
	"reflect"
	"testing"
)

// ---------------------------------------------------------------------------
// Flatten
// ---------------------------------------------------------------------------

// describe('Flatten Primitives') — String/Number/Boolean/null map cleanly onto
// Go scalar leaves; Date/undefined/Buffer/typed-array have no Go equivalent.
func TestParityFlattenPrimitives(t *testing.T) {
	cases := []struct {
		name string
		val  any
	}{
		{"String", "good morning"},
		{"Number", 1234.99},
		{"Boolean", true},
		{"null", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := map[string]any{"hello": map[string]any{"world": c.val}}
			got := Flatten(in)
			want := map[string]any{"hello.world": c.val}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %#v want %#v", got, want)
			}
		})
	}
}

// test('Nested once')
func TestParityFlattenNestedOnce(t *testing.T) {
	in := map[string]any{"hello": map[string]any{"world": "good morning"}}
	want := map[string]any{"hello.world": "good morning"}
	if got := Flatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Nested twice')
func TestParityFlattenNestedTwice(t *testing.T) {
	in := map[string]any{"hello": map[string]any{"world": map[string]any{"again": "good morning"}}}
	want := map[string]any{"hello.world.again": "good morning"}
	if got := Flatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Multiple Keys')
func TestParityFlattenMultipleKeys(t *testing.T) {
	in := map[string]any{
		"hello": map[string]any{"lorem": map[string]any{"ipsum": "again", "dolor": "sit"}},
		"world": map[string]any{"lorem": map[string]any{"ipsum": "again", "dolor": "sit"}},
	}
	want := map[string]any{
		"hello.lorem.ipsum": "again",
		"hello.lorem.dolor": "sit",
		"world.lorem.ipsum": "again",
		"world.lorem.dolor": "sit",
	}
	if got := Flatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Custom Delimiter') — delimiter ':'
func TestParityFlattenCustomDelimiter(t *testing.T) {
	in := map[string]any{"hello": map[string]any{"world": map[string]any{"again": "good morning"}}}
	want := map[string]any{"hello:world:again": "good morning"}
	if got := Flatten(in, FlattenOpts{Delimiter: ":"}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Empty Objects') — an empty nested object is preserved as a leaf.
func TestParityFlattenEmptyObjects(t *testing.T) {
	in := map[string]any{"hello": map[string]any{"empty": map[string]any{"nested": map[string]any{}}}}
	want := map[string]any{"hello.empty.nested": map[string]any{}}
	if got := Flatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Should keep number in the left when object') — numeric-looking string
// keys stay as-is in flattened output.
func TestParityFlattenKeepNumberLeft(t *testing.T) {
	in := map[string]any{"hello": map[string]any{"0200": "world", "0500": "darkness my old friend"}}
	want := map[string]any{"hello.0200": "world", "hello.0500": "darkness my old friend"}
	if got := Flatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Unflatten
// ---------------------------------------------------------------------------

// test('Nested once')
func TestParityUnflattenNestedOnce(t *testing.T) {
	in := map[string]any{"hello.world": "good morning"}
	want := map[string]any{"hello": map[string]any{"world": "good morning"}}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Nested twice')
func TestParityUnflattenNestedTwice(t *testing.T) {
	in := map[string]any{"hello.world.again": "good morning"}
	want := map[string]any{"hello": map[string]any{"world": map[string]any{"again": "good morning"}}}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Multiple Keys') — includes an object-valued entry (world:{greet}) that
// upstream re-flattens and merges with the delimited siblings.
func TestParityUnflattenMultipleKeys(t *testing.T) {
	in := map[string]any{
		"hello.lorem.ipsum": "again",
		"hello.lorem.dolor": "sit",
		"world.lorem.ipsum": "again",
		"world.lorem.dolor": "sit",
		"world":             map[string]any{"greet": "hello"},
	}
	want := map[string]any{
		"hello": map[string]any{"lorem": map[string]any{"ipsum": "again", "dolor": "sit"}},
		"world": map[string]any{
			"greet": "hello",
			"lorem": map[string]any{"ipsum": "again", "dolor": "sit"},
		},
	}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('nested objects do not clobber each other when a.b inserted before a')
func TestParityUnflattenNoClobber(t *testing.T) {
	in := map[string]any{
		"foo.bar": map[string]any{"t": 123},
		"foo":     map[string]any{"p": 333},
	}
	want := map[string]any{
		"foo": map[string]any{
			"bar": map[string]any{"t": 123},
			"p":   333,
		},
	}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Custom Delimiter') — delimiter ' '
func TestParityUnflattenCustomDelimiter(t *testing.T) {
	in := map[string]any{"hello world again": "good morning"}
	want := map[string]any{"hello": map[string]any{"world": map[string]any{"again": "good morning"}}}
	if got := Unflatten(in, UnflattenOpts{Delimiter: " "}); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Messy') — object-valued entries whose own keys are themselves delimited
// are recursively re-flattened before rebuilding.
func TestParityUnflattenMessy(t *testing.T) {
	in := map[string]any{
		"hello.world": "again",
		"lorem.ipsum": "another",
		"good.morning": map[string]any{
			"hash.key": map[string]any{
				"nested.deep": map[string]any{
					"and.even.deeper.still": "hello",
				},
			},
		},
		"good.morning.again": map[string]any{
			"testing.this": "out",
		},
	}
	want := map[string]any{
		"hello": map[string]any{"world": "again"},
		"lorem": map[string]any{"ipsum": "another"},
		"good": map[string]any{
			"morning": map[string]any{
				"hash":  map[string]any{"key": map[string]any{"nested": map[string]any{"deep": map[string]any{"and": map[string]any{"even": map[string]any{"deeper": map[string]any{"still": "hello"}}}}}}},
				"again": map[string]any{"testing": map[string]any{"this": "out"}},
			},
		},
	}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Empty objects should not be removed') — empty slice/map leaves survive.
func TestParityUnflattenEmptyObjectsNotRemoved(t *testing.T) {
	in := map[string]any{"foo": []any{}, "bar": map[string]any{}}
	want := map[string]any{"foo": []any{}, "bar": map[string]any{}}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// test('Do not include keys with numbers inside them')
func TestParityUnflattenNumbersInsideKeys(t *testing.T) {
	in := map[string]any{"1key.2_key": "ok"}
	want := map[string]any{"1key": map[string]any{"2_key": "ok"}}
	if got := Unflatten(in); !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v want %#v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Round trip
// ---------------------------------------------------------------------------

// describe('Order of Keys') — unflatten(flatten(obj)) reproduces the structure.
// The array leaf is carried through opaquely by this port.
func TestParityRoundTrip(t *testing.T) {
	in := map[string]any{
		"b": 1,
		"abc": map[string]any{
			"c": []any{map[string]any{"d": 1, "bca": 1, "a": 1}},
		},
		"a": 1,
	}
	if got := Unflatten(Flatten(in)); !reflect.DeepEqual(got, in) {
		t.Errorf("round trip failed:\n got %#v\nwant %#v", got, in)
	}
}
