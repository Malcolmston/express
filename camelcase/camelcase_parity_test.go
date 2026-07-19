package camelcase

// Upstream-parity tests for the Go port of sindresorhus/camelcase.
//
// Every input -> expected-output vector below is taken verbatim from the
// original library's own test suite (npm "camelcase" v9.0.0). Source:
//
//	https://raw.githubusercontent.com/sindresorhus/camelcase/main/test.js
//	(algorithm cross-checked against
//	 https://raw.githubusercontent.com/sindresorhus/camelcase/main/index.js)
//
// Only the vectors expressible through this port's API are encoded: plain
// string input with the default (camelCase) behavior and with the pascalCase
// option. Upstream vectors that exercise features this port does not implement
// (array input, the preserveConsecutiveUppercase / locale /
// capitalizeAfterNumber:false options, and the throw-on-non-string check) are
// intentionally omitted; see the sync notes for that gap list.

import "testing"

// TestParityCamelCase covers upstream's main `camelCase` block (string inputs).
func TestParityCamelCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"b2b_registration_request", "b2bRegistrationRequest"},
		{"b2b-registration-request", "b2bRegistrationRequest"},
		{"b2b_registration_b2b_request", "b2bRegistrationB2bRequest"},
		{"foo", "foo"},
		{"IDs", "ids"},
		{"FooIDs", "fooIds"},
		{"foo-bar", "fooBar"},
		{"foo-bar-baz", "fooBarBaz"},
		{"foo--bar", "fooBar"},
		{"--foo-bar", "fooBar"},
		{"--foo--bar", "fooBar"},
		{"FOO-BAR", "fooBar"},
		{"FOÈ-BAR", "foèBar"},
		{"-foo-bar-", "fooBar"},
		{"--foo--bar--", "fooBar"},
		{"foo-1", "foo1"},
		{"foo.bar", "fooBar"},
		{"foo..bar", "fooBar"},
		{"..foo..bar..", "fooBar"},
		{"foo_bar", "fooBar"},
		{"__foo__bar__", "__fooBar"},
		{"foo bar", "fooBar"},
		{"  foo  bar  ", "fooBar"},
		{"-", ""},
		{" - ", ""},
		{"fooBar", "fooBar"},
		{"fooBar-baz", "fooBarBaz"},
		{"foìBar-baz", "foìBarBaz"},
		{"fooBarBaz-bazzy", "fooBarBazBazzy"},
		{"FBBazzy", "fbBazzy"},
		{"F", "f"},
		{"FooBar", "fooBar"},
		{"Foo", "foo"},
		{"FOO", "foo"},
		{"--", ""},
		{"", ""},
		{"_", "_"},
		{" ", ""},
		{".", ""},
		{"..", ""},
		{"  ", ""},
		{"__", "__"},
		{"--__--_--_", ""},
		{"foo bar?", "fooBar?"},
		{"foo bar!", "fooBar!"},
		{"foo bar$", "fooBar$"},
		{"foo-bar#", "fooBar#"},
		{"XMLHttpRequest", "xmlHttpRequest"},
		{"AjaxXMLHttpRequest", "ajaxXmlHttpRequest"},
		{"Ajax-XMLHttpRequest", "ajaxXmlHttpRequest"},
		{"mGridCol6@md", "mGridCol6@md"},
		{"A::a", "a::a"},
		{"Hello1World", "hello1World"},
		{"Hello11World", "hello11World"},
		{"hello1world", "hello1World"},
		{"Hello1World11foo", "hello1World11Foo"},
		{"Hello1", "hello1"},
		{"hello1", "hello1"},
		{"1Hello", "1Hello"},
		{"1hello", "1Hello"},
		{"h2w", "h2W"},
		{"розовый_пушистый-единороги", "розовыйПушистыйЕдинороги"},
		{"РОЗОВЫЙ_ПУШИСТЫЙ-ЕДИНОРОГИ", "розовыйПушистыйЕдинороги"},
		{"桑德在这里。", "桑德在这里。"},
		{"桑德_在这里。", "桑德在这里。"},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityPascalCase covers upstream's `pascalCase` option block (string inputs).
func TestParityPascalCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"b2b_registration_request", "B2bRegistrationRequest"},
		{"foo", "Foo"},
		{"foo-bar", "FooBar"},
		{"foo-bar-baz", "FooBarBaz"},
		{"foo--bar", "FooBar"},
		{"--foo-bar", "FooBar"},
		{"--foo--bar", "FooBar"},
		{"FOO-BAR", "FooBar"},
		{"FOÈ-BAR", "FoèBar"},
		{"-foo-bar-", "FooBar"},
		{"--foo--bar--", "FooBar"},
		{"foo-1", "Foo1"},
		{"foo.bar", "FooBar"},
		{"foo..bar", "FooBar"},
		{"..foo..bar..", "FooBar"},
		{"foo_bar", "FooBar"},
		{"__foo__bar__", "__FooBar"},
		{"foo bar", "FooBar"},
		{"  foo  bar  ", "FooBar"},
		{"-", ""},
		{" - ", ""},
		{"fooBar", "FooBar"},
		{"fooBar-baz", "FooBarBaz"},
		{"foìBar-baz", "FoìBarBaz"},
		{"fooBarBaz-bazzy", "FooBarBazBazzy"},
		{"FBBazzy", "FbBazzy"},
		{"F", "F"},
		{"FooBar", "FooBar"},
		{"Foo", "Foo"},
		{"FOO", "Foo"},
		{"--", ""},
		{"", ""},
		{"--__--_--_", ""},
		{"foo bar?", "FooBar?"},
		{"foo bar!", "FooBar!"},
		{"foo bar$", "FooBar$"},
		{"foo-bar#", "FooBar#"},
		{"XMLHttpRequest", "XmlHttpRequest"},
		{"AjaxXMLHttpRequest", "AjaxXmlHttpRequest"},
		{"Ajax-XMLHttpRequest", "AjaxXmlHttpRequest"},
		{"mGridCol6@md", "MGridCol6@md"},
		{"A::a", "A::a"},
		{"Hello1World", "Hello1World"},
		{"Hello11World", "Hello11World"},
		{"hello1world", "Hello1World"},
		{"hello1World", "Hello1World"},
		{"hello1", "Hello1"},
		{"Hello1", "Hello1"},
		{"1hello", "1Hello"},
		{"1Hello", "1Hello"},
		{"h1W", "H1W"},
		{"РозовыйПушистыйЕдинороги", "РозовыйПушистыйЕдинороги"},
		{"розовый_пушистый-единороги", "РозовыйПушистыйЕдинороги"},
		{"РОЗОВЫЙ_ПУШИСТЫЙ-ЕДИНОРОГИ", "РозовыйПушистыйЕдинороги"},
		{"桑德在这里。", "桑德在这里。"},
		{"桑德_在这里。", "桑德在这里。"},
		{"a1b", "A1B"},
	}
	for _, c := range cases {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityNumbers covers upstream's default number handling
// (capitalizeAfterNumber defaults to true): a digit run makes a word boundary
// for the following letter unless a separator immediately follows.
func TestParityNumbers(t *testing.T) {
	camel := []struct{ in, want string }{
		{"Hello1World", "hello1World"},
		{"foo2bar", "foo2Bar"},
		{"hello1world", "hello1World"},
		{"turn_on_2sv", "turnOn2Sv"},
		{"foo-2bar", "foo2Bar"},
		{"2foo-bar", "2FooBar"},
		{"XML2HTTP", "xml2Http"},
		{"2a", "2A"},
		{"version_3.14.15", "version31415"},
		{"temp_-5_degrees", "temp5Degrees"},
		{"123", "123"},
		{"123_456_789", "123456789"},
	}
	for _, c := range camel {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	pascal := []struct{ in, want string }{
		{"Hello1World", "Hello1World"},
		{"turn_on_2sv", "TurnOn2Sv"},
		{"foo-2bar", "Foo2Bar"},
		{"2foo-bar", "2FooBar"},
		{"XML2HTTP", "Xml2Http"},
		{"2a", "2A"},
	}
	for _, c := range pascal {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityLeadingPrefix covers upstream's "preserve leading underscores and
// dollar signs" block (string inputs, no preserveConsecutiveUppercase).
func TestParityLeadingPrefix(t *testing.T) {
	camel := []struct{ in, want string }{
		{"_foo_bar", "_fooBar"},
		{"$foo_bar", "$fooBar"},
		{"__foo_bar", "__fooBar"},
		{"$$foo_bar", "$$fooBar"},
		{"$_foo_bar", "$_fooBar"},
		{"_$foo_bar", "_$fooBar"},
		{"_", "_"},
		{"__", "__"},
		{"$", "$"},
		{"$$$", "$$$"},
		{"_$", "_$"},
		{"_foo-bar_baz", "_fooBarBaz"},
		{"$http_service", "$httpService"},
	}
	for _, c := range camel {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}

	pascal := []struct{ in, want string }{
		{"_foo_bar", "_FooBar"},
		{"$foo_bar", "$FooBar"},
		{"__foo_bar", "__FooBar"},
	}
	for _, c := range pascal {
		if got := PascalCase(c.in); got != c.want {
			t.Errorf("PascalCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityUnicodeAndSpecial covers upstream's emoji/unicode, special-character,
// and uppercase-transition edge-case blocks (string inputs).
func TestParityUnicodeAndSpecial(t *testing.T) {
	cases := []struct{ in, want string }{
		// emoji and unicode
		{"foo-🦄-bar", "foo-🦄Bar"},
		{"foo🦄bar", "foo🦄bar"},
		{"foo\u200dbar", "foo\u200dbar"},
		{"foo_مرحبا_bar", "fooمرحباBar"},
		{"foo_שלום_bar", "fooשלוםBar"},
		// special characters
		{"foo##bar", "foo##bar"},
		{"foo@#$bar", "foo@#$bar"},
		{"foo_@#_bar", "foo_@#Bar"},
		// uppercase transitions
		{"aAbBcC", "aAbBcC"},
		{"a1A2B3C", "a1A2B3C"},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestParityExtremeInputs covers upstream's "extreme inputs" edge-case block.
func TestParityExtremeInputs(t *testing.T) {
	longPrefix := ""
	for i := 0; i < 50; i++ {
		longPrefix += "_"
	}
	cases := []struct{ in, want string }{
		{longPrefix + "foo" + longPrefix + "bar", longPrefix + "fooBar"},
		{"_-. _-. _-.foo", "_foo"},
		{"-_.  -_. -_.", ""},
	}
	for _, c := range cases {
		if got := CamelCase(c.in); got != c.want {
			t.Errorf("CamelCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
