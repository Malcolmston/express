package dotprop

// Upstream-parity vectors ported from sindresorhus/dot-prop's own test suite.
// Source of every concrete input -> expected pair below:
//
//	https://raw.githubusercontent.com/sindresorhus/dot-prop/main/test.js
//	https://raw.githubusercontent.com/sindresorhus/dot-prop/main/index.js
//
// The upstream API is getProperty/setProperty/hasProperty/deleteProperty over
// arbitrary JS values. This Go port models JS objects as map[string]any and JS
// arrays as []any, and uses the dot-with-backslash-escape path grammar plus
// numeric segments as slice indices. Only the vectors expressible in that model
// are encoded here; upstream features the port deliberately omits (bracket
// "[0]" index grammar, a default-value argument, empty-string whole path as a
// real "" key, top-level arrays, escapePath/deepKeys/unflatten, JS function
// objects) are recorded in the task notes, not asserted as failing tests.

import "testing"

// TestParityGetProperty covers the dot-grammar getProperty vectors, mapping
// upstream's (value | default) return onto the port's (value, ok) pair.
func TestParityGetProperty(t *testing.T) {
	type tc struct {
		name   string
		obj    map[string]any
		path   string
		want   any
		wantOK bool
	}
	cases := []tc{
		{"foo=1", map[string]any{"foo": 1}, "foo", 1, true},
		{"foo=null present", map[string]any{"foo": nil}, "foo", nil, true},
		{"foo.bar true", map[string]any{"foo": map[string]any{"bar": true}}, "foo.bar", true, true},
		{"foo.bar.baz true", map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": true}}}, "foo.bar.baz", true, true},
		{"foo.bar.baz null", map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": nil}}}, "foo.bar.baz", nil, true},
		{"foo.fake missing", map[string]any{"foo": map[string]any{"bar": "a"}}, "foo.fake", nil, false},
		{"foo.fake.fake2 missing", map[string]any{"foo": map[string]any{"bar": "a"}}, "foo.fake.fake2", nil, false},
		{"scalar leaf descend", map[string]any{"foo": 1}, "foo.bar", nil, false},
		// Backslash-escape grammar (JS String.raw values reproduced as Go raw strings).
		{`single backslash key`, map[string]any{`\`: true}, `\`, true, true},
		{`\foo escapes to foo`, map[string]any{"foo": true}, `\foo`, true, true},
		{`\\foo escapes to \foo`, map[string]any{`\foo`: true}, `\\foo`, true, true},
		{`foo\\ -> foo\`, map[string]any{`foo\`: true}, `foo\\`, true, true},
		{`bar\ trailing -> bar\`, map[string]any{`bar\`: true}, `bar\`, true, true},
		{`foo\bar escapes to foobar`, map[string]any{"foobar": true}, `foo\bar`, true, true},
		{`\\.foo -> [\][foo]`, map[string]any{`\`: map[string]any{"foo": true}}, `\\.foo`, true, true},
		{`bar\\\. -> bar\.`, map[string]any{`bar\.`: true}, `bar\\\.`, true, true},
		{`foo\\.bar -> [foo\][bar]`, map[string]any{`foo\`: map[string]any{"bar": true}}, `foo\\.bar`, true, true},
		{`foo\.baz.bar`, map[string]any{"foo.baz": map[string]any{"bar": true}}, `foo\.baz.bar`, true, true},
		{`fo\.ob\.az.bar`, map[string]any{"fo.ob.az": map[string]any{"bar": true}}, `fo\.ob\.az.bar`, true, true},
		// Empty-string segments produced by dot-only paths.
		{`.. empty segments`, map[string]any{"": map[string]any{"": map[string]any{"": true}}}, "..", true, true},
		{`. empty segments`, map[string]any{"": map[string]any{"": true}}, ".", true, true},
		// Array (slice) indices via numeric dot segments.
		{"foo.0 into slice", map[string]any{"foo": []any{true}}, "foo.0", true, true},
		{"foo.0 numeric map key", map[string]any{"foo": map[string]any{"0": true}}, "foo.0", true, true},
		{"foo.1.bar slice of maps", map[string]any{"foo": []any{0, map[string]any{"bar": true}}}, "foo.1.bar", true, true},
		{"foo.2.bar out of range", map[string]any{"foo": []any{0, map[string]any{"bar": 2}}}, "foo.2.bar", nil, false},
		// Disallowed (prototype-pollution) keys collapse to "not found".
		{"constructor guarded", map[string]any{"constructor": 1}, "constructor", nil, false},
		{"__proto__ guarded", map[string]any{"a": map[string]any{"__proto__": map[string]any{"b": 1}}}, "a.__proto__.b", nil, false},
		{"prototype guarded", map[string]any{"prototype": 1}, "prototype", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := Get(c.obj, c.path)
			if ok != c.wantOK || (ok && got != c.want) {
				t.Fatalf("Get(%#v, %q) = %v, %v; want %v, %v", c.obj, c.path, got, ok, c.want, c.wantOK)
			}
		})
	}
}

// TestParityHasProperty covers hasProperty truthiness, including descent into
// scalar leaves (which must report false) and the null-present case.
func TestParityHasProperty(t *testing.T) {
	type tc struct {
		name string
		obj  map[string]any
		path string
		want bool
	}
	cases := []tc{
		{"foo=1", map[string]any{"foo": 1}, "foo", true},
		{"foo=null present", map[string]any{"foo": nil}, "foo", true},
		{"foo.bar.baz=null", map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": nil}}}, "foo.bar.baz", true},
		{"foo.fake.fake2", map[string]any{"foo": map[string]any{"bar": "a"}}, "foo.fake.fake2", false},
		{"null then bar", map[string]any{"foo": nil}, "foo.bar", false},
		{"empty string leaf", map[string]any{"foo": ""}, "foo.bar", false},
		{"zero leaf", map[string]any{"foo": 0}, "foo.bar", false},
		{"false leaf", map[string]any{"foo": false}, "foo.bar", false},
		{`foo\.baz.bar`, map[string]any{"foo.baz": map[string]any{"bar": true}}, `foo\.baz.bar`, true},
		{`fo\.ob\.az.bar`, map[string]any{"fo.ob.az": map[string]any{"bar": true}}, `fo\.ob\.az.bar`, true},
		{"nil obj", nil, `fo\.ob\.az.bar`, false},
		{"numeric map key subkey", map[string]any{"foo": []any{map[string]any{"bar": map[string]any{"1": "bar"}}}}, "foo.0.bar.1", true},
		{"slice subkey present", map[string]any{"foo": []any{map[string]any{"bar": []any{"bar", "bizz"}}}}, "foo.0.bar.1", true},
		{"slice subkey out of range", map[string]any{"foo": []any{map[string]any{"bar": []any{"bar", "bizz"}}}}, "foo.0.bar.2", false},
		{"slice index out of range", map[string]any{"foo": []any{map[string]any{"bar": []any{"bar", "bizz"}}}}, "foo.1.bar.1", false},
		{"constructor guarded", map[string]any{"constructor": 1}, "constructor", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Has(c.obj, c.path); got != c.want {
				t.Fatalf("Has(%#v, %q) = %v; want %v", c.obj, c.path, got, c.want)
			}
		})
	}
}

// TestParitySetProperty covers setProperty: nested creation, overwriting a null
// intermediate, and escaped-dot keys. Each vector reads the value back.
func TestParitySetProperty(t *testing.T) {
	t.Run("top-level", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, "foo", 2)
		if v, ok := Get(obj, "foo"); !ok || v != 2 {
			t.Fatalf("foo = %v, %v", v, ok)
		}
	})
	t.Run("nested overwrite existing", func(t *testing.T) {
		obj := map[string]any{"foo": map[string]any{"bar": 1}}
		Set(obj, "foo.bar", 2)
		if v, _ := Get(obj, "foo.bar"); v != 2 {
			t.Fatalf("foo.bar = %v", v)
		}
		Set(obj, "foo.bar.baz", 3)
		if v, _ := Get(obj, "foo.bar.baz"); v != 3 {
			t.Fatalf("foo.bar.baz = %v", v)
		}
	})
	t.Run("overwrite null intermediate", func(t *testing.T) {
		obj := map[string]any{"foo": nil}
		Set(obj, "foo.bar", 2)
		if v, ok := Get(obj, "foo.bar"); !ok || v != 2 {
			t.Fatalf("foo.bar = %v, %v", v, ok)
		}
	})
	t.Run("escaped dot key", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, `foo\.bar.baz`, true)
		if v, ok := obj["foo.bar"].(map[string]any); !ok || v["baz"] != true {
			t.Fatalf("obj = %#v", obj)
		}
	})
	t.Run("multiple escaped dots", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, `fo\.ob\.ar.baz`, true)
		if v, ok := obj["fo.ob.ar"].(map[string]any); !ok || v["baz"] != true {
			t.Fatalf("obj = %#v", obj)
		}
	})
	t.Run("returns obj", func(t *testing.T) {
		obj := map[string]any{}
		if got := Set(obj, "foo", 2); &got == nil {
			t.Fatal("expected obj back")
		}
	})
	t.Run("disallowed key is no-op", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, "__proto__.polluted", true)
		if _, ok := obj["__proto__"]; ok {
			t.Fatalf("guarded set mutated obj: %#v", obj)
		}
	})
}

// TestParityDeleteProperty covers deleteProperty: nested removal, escaped-dot
// keys, the null-intermediate no-op, and the guarded-key no-op.
func TestParityDeleteProperty(t *testing.T) {
	t.Run("nested removal keeps siblings", func(t *testing.T) {
		obj := map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": map[string]any{"a": "a", "b": "b", "c": "c"}}}}
		if !Delete(obj, "foo.bar.baz.c") {
			t.Fatal("delete c should be true")
		}
		if Has(obj, "foo.bar.baz.c") {
			t.Fatal("c should be gone")
		}
		if !Has(obj, "foo.bar.baz.a") {
			t.Fatal("a should remain")
		}
	})
	t.Run("delete top", func(t *testing.T) {
		obj := map[string]any{"top": map[string]any{"dog": "sindre"}}
		if !Delete(obj, "top") {
			t.Fatal("delete top should be true")
		}
		if Has(obj, "top") {
			t.Fatal("top should be gone")
		}
	})
	t.Run("escaped dot key delete", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, `foo\.bar.baz`, true)
		if !Delete(obj, `foo\.bar.baz`) {
			t.Fatal("delete escaped should be true")
		}
		if Has(obj, `foo\.bar.baz`) {
			t.Fatal("escaped key should be gone")
		}
	})
	t.Run("escaped dot leaf and sibling", func(t *testing.T) {
		obj := map[string]any{}
		Set(obj, `foo.bar\.baz`, true)
		if v, _ := Get(obj, `foo.bar\.baz`); v != true {
			t.Fatalf("setup failed: %#v", obj)
		}
		if !Delete(obj, `foo.bar\.baz`) {
			t.Fatal("delete should be true")
		}
		if Has(obj, `foo.bar\.baz`) {
			t.Fatal("should be gone")
		}
	})
	t.Run("nested dotted sibling retained", func(t *testing.T) {
		obj := map[string]any{"dotted": map[string]any{"sub": map[string]any{"dotted.prop": "foo", "other": "prop"}}}
		if !Delete(obj, `dotted.sub.dotted\.prop`) {
			t.Fatal("delete should be true")
		}
		if Has(obj, `dotted.sub.dotted\.prop`) {
			t.Fatal("dotted.prop should be gone")
		}
		if v, _ := Get(obj, "dotted.sub.other"); v != "prop" {
			t.Fatalf("other should remain: %v", v)
		}
	})
	t.Run("null intermediate no-op", func(t *testing.T) {
		obj := map[string]any{"foo": nil}
		if Delete(obj, "foo.bar") {
			t.Fatal("delete through null should be false")
		}
		if _, ok := obj["foo"]; !ok {
			t.Fatal("foo should be untouched")
		}
	})
	t.Run("disallowed key no-op", func(t *testing.T) {
		obj := map[string]any{"a": 1}
		if Delete(obj, "constructor.x") {
			t.Fatal("guarded delete should be false")
		}
	})
}
