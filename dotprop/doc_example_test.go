package dotprop_test

import (
	"fmt"

	"github.com/malcolmston/express/dotprop"
)

// ExampleGet reads a value from deep inside a tree of maps by a single dotted
// path. The path "a.b.c" is split on dots and each segment is used as a map key,
// descending one level at a time. Get returns the value and true when the whole
// path resolves, or nil and false otherwise, so a caller can address nested
// config without hand-written lookups and type assertions.
func ExampleGet() {
	obj := map[string]any{
		"a": map[string]any{
			"b": map[string]any{"c": 42},
		},
	}
	value, ok := dotprop.Get(obj, "a.b.c")
	fmt.Println(value, ok)
	// Output: 42 true
}

// ExampleSet assigns a value at a dotted path, creating intermediate maps as
// needed. Starting from an empty map, setting "server.port" builds the nested
// "server" map automatically. Set returns the same top-level map so calls can be
// chained, and reading the value back confirms it was stored at the nested
// location.
func ExampleSet() {
	obj := map[string]any{}
	dotprop.Set(obj, "server.port", 8080)

	value, ok := dotprop.Get(obj, "server.port")
	fmt.Println(value, ok)
	// Output: 8080 true
}

// ExampleHas reports whether a dotted path resolves to a value, without
// returning the value itself. It is simply Get with the value discarded. Here a
// present path reports true and a missing sibling reports false, which is handy
// for feature-flag or optional-config checks.
func ExampleHas() {
	obj := map[string]any{"a": map[string]any{"b": 1}}
	fmt.Println(dotprop.Has(obj, "a.b"))
	fmt.Println(dotprop.Has(obj, "a.c"))
	// Output:
	// true
	// false
}
