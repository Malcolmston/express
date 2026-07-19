package cookie

// Parity tests derived from the upstream jshttp/cookie test suite (v0.7.2),
// whose public API (cookie.parse(str) -> object, cookie.serialize(name, val,
// opts) -> string) is the one this Go port mirrors. Every input/expected value
// below is taken verbatim from those upstream test files:
//
//	https://raw.githubusercontent.com/jshttp/cookie/v0.7.2/test/parse.js
//	https://raw.githubusercontent.com/jshttp/cookie/v0.7.2/test/serialize.js
//
// The exact validation regexes (cookieNameRegExp, cookieValueRegExp,
// domainValueRegExp, pathValueRegExp) were cross-checked against:
//
//	https://raw.githubusercontent.com/jshttp/cookie/v0.7.2/index.js
//
// Deliberate, documented Go differences that are NOT asserted as parity here:
//   - Max-Age: upstream emits "Max-Age=0" for maxAge:0 and "Max-Age=-1" for
//     negative values; this port uses net/http sign conventions (0 omits the
//     attribute, negative emits "Max-Age=0"), because Options.MaxAge is a plain
//     int and cannot distinguish "unset" from 0 without an API change.
//   - Custom encode/decode callbacks and the surrounding-quote form of
//     cookie-value are not applicable: the value codec is fixed to
//     encode/decode, so those upstream vectors have no Go analogue.

import (
	"reflect"
	"testing"
	"time"
)

// TestParityParse covers cookie.parse(str) vectors from test/parse.js.
func TestParityParse(t *testing.T) {
	cases := []struct {
		in   string
		want map[string]string
	}{
		// should parse cookie string to object
		{"foo=bar", map[string]string{"foo": "bar"}},
		{"foo=123", map[string]string{"foo": "123"}},
		// should ignore OWS
		{"FOO    = bar;   baz  =   raz", map[string]string{"FOO": "bar", "baz": "raz"}},
		// should parse cookie with empty value
		{"foo=; bar=", map[string]string{"foo": "", "bar": ""}},
		// should parse cookie with minimum length
		{"f=", map[string]string{"f": ""}},
		{"f=;b=", map[string]string{"f": "", "b": ""}},
		// should URL-decode values
		{`foo="bar=123456789&name=Magic+Mouse"`, map[string]string{"foo": "bar=123456789&name=Magic+Mouse"}},
		{"email=%20%22%2c%3b%2f", map[string]string{"email": ` ",;/`}},
		// should parse quoted values
		{`foo="bar"`, map[string]string{"foo": "bar"}},
		{`foo=" a b c "`, map[string]string{"foo": " a b c "}},
		// should trim whitespace around key and value
		{`  foo  =  "bar"  `, map[string]string{"foo": "bar"}},
		{`  foo  =  bar  ;  fizz  =  buzz  `, map[string]string{"foo": "bar", "fizz": "buzz"}},
		{` foo = " a b c " `, map[string]string{"foo": " a b c "}},
		{` = bar `, map[string]string{"": "bar"}},
		{` foo = `, map[string]string{"foo": ""}},
		{`   =   `, map[string]string{"": ""}},
		{"\tfoo\t=\tbar\t", map[string]string{"foo": "bar"}},
		// should return original value on escape error
		{"foo=%1;bar=bar", map[string]string{"foo": "%1", "bar": "bar"}},
		// should ignore cookies without value
		{"foo=bar;fizz  ;  buzz", map[string]string{"foo": "bar"}},
		{"  fizz; foo=  bar", map[string]string{"foo": "bar"}},
		// should ignore duplicate cookies (first wins)
		{"foo=%1;bar=bar;foo=boo", map[string]string{"foo": "%1", "bar": "bar"}},
		{"foo=false;bar=bar;foo=true", map[string]string{"foo": "false", "bar": "bar"}},
		{"foo=;bar=bar;foo=boo", map[string]string{"foo": "", "bar": "bar"}},
		// should parse native properties
		{"toString=foo;valueOf=bar", map[string]string{"toString": "foo", "valueOf": "bar"}},
	}
	for _, tc := range cases {
		got := Parse(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Parse(%q) = %#v, want %#v", tc.in, got, tc.want)
		}
	}
}

// TestParitySerialize covers the valid cookie.serialize outputs from
// test/serialize.js (single-attribute vectors, exact strings).
func TestParitySerialize(t *testing.T) {
	cases := []struct {
		name string
		val  string
		opts *Options
		want string
	}{
		{"foo", "bar", nil, "foo=bar"},
		{"foo", "bar +baz", nil, "foo=bar%20%2Bbaz"}, // should URL-encode value
		{"foo", "", nil, "foo="},                     // should serialize empty value
		// httpOnly
		{"foo", "bar", &Options{HttpOnly: true}, "foo=bar; HttpOnly"},
		{"foo", "bar", &Options{HttpOnly: false}, "foo=bar"},
		// maxAge (positive)
		{"foo", "bar", &Options{MaxAge: 1000}, "foo=bar; Max-Age=1000"},
		// secure
		{"foo", "bar", &Options{Secure: true}, "foo=bar; Secure"},
		{"foo", "bar", &Options{Secure: false}, "foo=bar"},
		// partitioned
		{"foo", "bar", &Options{Partitioned: true}, "foo=bar; Partitioned"},
		{"foo", "bar", &Options{Partitioned: false}, "foo=bar"},
		{"foo", "bar", &Options{}, "foo=bar"},
		// expires
		{"foo", "bar", &Options{Expires: time.Date(2000, time.December, 24, 10, 30, 59, 900e6, time.UTC)}, "foo=bar; Expires=Sun, 24 Dec 2000 10:30:59 GMT"},
		// priority
		{"foo", "bar", &Options{Priority: "Low"}, "foo=bar; Priority=Low"},
		{"foo", "bar", &Options{Priority: "loW"}, "foo=bar; Priority=Low"},
		{"foo", "bar", &Options{Priority: "Medium"}, "foo=bar; Priority=Medium"},
		{"foo", "bar", &Options{Priority: "medium"}, "foo=bar; Priority=Medium"},
		{"foo", "bar", &Options{Priority: "High"}, "foo=bar; Priority=High"},
		{"foo", "bar", &Options{Priority: "HIGH"}, "foo=bar; Priority=High"},
		// sameSite
		{"foo", "bar", &Options{SameSite: "Strict"}, "foo=bar; SameSite=Strict"},
		{"foo", "bar", &Options{SameSite: "strict"}, "foo=bar; SameSite=Strict"},
		{"foo", "bar", &Options{SameSite: "Lax"}, "foo=bar; SameSite=Lax"},
		{"foo", "bar", &Options{SameSite: "lax"}, "foo=bar; SameSite=Lax"},
		{"foo", "bar", &Options{SameSite: "None"}, "foo=bar; SameSite=None"},
		{"foo", "bar", &Options{SameSite: "none"}, "foo=bar; SameSite=None"},
		{"foo", "bar", &Options{SameSite: ""}, "foo=bar"}, // false -> omit
	}
	for _, tc := range cases {
		got, err := Serialize(tc.name, tc.val, tc.opts)
		if err != nil {
			t.Errorf("Serialize(%q, %q, %#v) unexpected error: %v", tc.name, tc.val, tc.opts, err)
			continue
		}
		if got != tc.want {
			t.Errorf("Serialize(%q, %q, %#v) = %q, want %q", tc.name, tc.val, tc.opts, got, tc.want)
		}
	}
}

// TestParitySerializeValidNames: names accepted by cookieNameRegExp.
func TestParitySerializeValidNames(t *testing.T) {
	validNames := []string{
		"foo", "foo!bar", "foo#bar", "foo$bar", "foo'bar", "foo*bar",
		"foo+bar", "foo-bar", "foo.bar", "foo^bar", "foo_bar", "foo`bar",
		"foo|bar", "foo~bar", "foo7bar",
	}
	for _, name := range validNames {
		got, err := Serialize(name, "baz", nil)
		if err != nil {
			t.Errorf("Serialize(%q, baz) unexpected error: %v", name, err)
			continue
		}
		if want := name + "=baz"; got != want {
			t.Errorf("Serialize(%q, baz) = %q, want %q", name, got, want)
		}
	}
}

// TestParitySerializeInvalidName: names rejected by cookieNameRegExp.
func TestParitySerializeInvalidName(t *testing.T) {
	invalidNames := []string{
		"foo\n", "foo⠊", "foo/foo", "foo,foo", "foo;foo", "foo@foo",
		"foo[foo]", "foo?foo", "foo:foo", "foo{foo}", "foo foo", "foo\tfoo",
		"foo\"foo", "foo<script>foo",
	}
	for _, name := range invalidNames {
		if _, err := Serialize(name, "bar", nil); err == nil {
			t.Errorf("Serialize(%q, bar) expected error, got nil", name)
		}
	}
}

// TestParitySerializeValidDomain: domains accepted by domainValueRegExp.
func TestParitySerializeValidDomain(t *testing.T) {
	validDomains := []string{
		"example.com", "sub.example.com", ".example.com",
		"localhost", ".localhost", "my-site.org",
	}
	for _, d := range validDomains {
		got, err := Serialize("foo", "bar", &Options{Domain: d})
		if err != nil {
			t.Errorf("Serialize domain %q unexpected error: %v", d, err)
			continue
		}
		if want := "foo=bar; Domain=" + d; got != want {
			t.Errorf("Serialize domain %q = %q, want %q", d, got, want)
		}
	}
}

// TestParitySerializeInvalidDomain: domains rejected by domainValueRegExp.
func TestParitySerializeInvalidDomain(t *testing.T) {
	invalidDomains := []string{
		"example.com\n", "sub.example.com\x00", "my site.org",
		"domain..com", "example.com; Path=/", "example.com /* inject a comment */",
	}
	for _, d := range invalidDomains {
		if _, err := Serialize("foo", "bar", &Options{Domain: d}); err == nil {
			t.Errorf("Serialize domain %q expected error, got nil", d)
		}
	}
}

// TestParitySerializeValidPath: paths accepted by pathValueRegExp.
func TestParitySerializeValidPath(t *testing.T) {
	validPaths := []string{
		"/", "/login", "/foo.bar/baz", "/foo-bar", "/foo=bar?baz",
		`/foo"bar"`, "/../foo/bar", "../foo/", "./",
	}
	for _, p := range validPaths {
		got, err := Serialize("foo", "bar", &Options{Path: p})
		if err != nil {
			t.Errorf("Serialize path %q unexpected error: %v", p, err)
			continue
		}
		if want := "foo=bar; Path=" + p; got != want {
			t.Errorf("Serialize path %q = %q, want %q", p, got, want)
		}
	}
}

// TestParitySerializeInvalidPath: paths rejected by pathValueRegExp.
func TestParitySerializeInvalidPath(t *testing.T) {
	invalidPaths := []string{
		"/\n", "/foo\x00", "/path/with\rnewline",
		"/; Path=/sensitive-data", `/login"><script>alert(1)</script>`,
	}
	for _, p := range invalidPaths {
		if _, err := Serialize("foo", "bar", &Options{Path: p}); err == nil {
			t.Errorf("Serialize path %q expected error, got nil", p)
		}
	}
}

// TestParitySerializeInvalidPriorityAndSameSite: invalid enum values error.
func TestParitySerializeInvalidPriorityAndSameSite(t *testing.T) {
	if _, err := Serialize("foo", "bar", &Options{Priority: "foo"}); err == nil {
		t.Errorf("Serialize priority %q expected error, got nil", "foo")
	}
	if _, err := Serialize("foo", "bar", &Options{SameSite: "foo"}); err == nil {
		t.Errorf("Serialize sameSite %q expected error, got nil", "foo")
	}
}
