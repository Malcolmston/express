package mediatyper

import "testing"

// Upstream-parity tests for the jshttp/media-typer npm package.
//
// The concrete input -> expected-output vectors below are transcribed verbatim
// from the ORIGINAL library's own test suite (the classic 0.3.0 API that this
// Go package reimplements: parse/format with type, subtype, +suffix and
// parameters, case-normalization and quoted-string handling):
//
//	https://raw.githubusercontent.com/jshttp/media-typer/v0.3.0/test/test.js
//	https://raw.githubusercontent.com/jshttp/media-typer/v0.3.0/index.js
//
// The modern 2.0.0 rewrite (src/index.spec.ts on master) dropped parameter
// handling entirely and preserves case on format, so it is NOT the parity
// target for this port; its type/subtype/suffix vectors are still covered here
// because they agree with 0.3.0. Node throws where Go returns an error; those
// vectors assert only that Parse/Format return a non-nil error.

// TestParityParseBasic mirrors "should parse basic type" and "should parse with
// suffix" plus modern src/index.spec.ts parse vectors.
func TestParityParseBasic(t *testing.T) {
	cases := []struct {
		in     string
		typ    string
		sub    string
		suffix string
	}{
		{"text/html", "text", "html", ""},
		{"image/svg+xml", "image", "svg", "xml"},
		// "should lower-case type"
		{"IMAGE/SVG+XML", "image", "svg", "xml"},
		// "should parse with multiple + in subtype" (suffix = after LAST +)
		{"application/vnd.api+json+gzip", "application", "vnd.api+json", "gzip"},
	}
	for _, c := range cases {
		m, err := Parse(c.in)
		if err != nil {
			t.Fatalf("Parse(%q) unexpected error: %v", c.in, err)
		}
		if m.Type != c.typ || m.Subtype != c.sub || m.Suffix != c.suffix {
			t.Errorf("Parse(%q) = {type:%q subtype:%q suffix:%q}, want {%q %q %q}",
				c.in, m.Type, m.Subtype, m.Suffix, c.typ, c.sub, c.suffix)
		}
	}
}

// TestParityParseParameters mirrors "should parse parameters",
// "should parse parameters with extra LWS", "should lower-case parameter names",
// "should unquote parameter values" and the escape/balanced-quote vectors.
func TestParityParseParameters(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		params map[string]string
	}{
		{
			name:   "basic parameters",
			in:     "text/html; charset=utf-8; foo=bar",
			params: map[string]string{"charset": "utf-8", "foo": "bar"},
		},
		{
			name:   "extra LWS",
			in:     "text/html ; charset=utf-8 ; foo=bar",
			params: map[string]string{"charset": "utf-8", "foo": "bar"},
		},
		{
			name:   "lower-case parameter names, value preserved",
			in:     "text/html; Charset=UTF-8",
			params: map[string]string{"charset": "UTF-8"},
		},
		{
			name:   "unquote parameter values",
			in:     `text/html; charset="UTF-8"`,
			params: map[string]string{"charset": "UTF-8"},
		},
		{
			// "should unquote parameter values with escapes":
			// text/html; charset = "UT\F-\\\"8\""  ->  UTF-\"8"
			name:   "unquote with escapes",
			in:     `text/html; charset = "UT\F-\\\"8\""`,
			params: map[string]string{"charset": `UTF-\"8"`},
		},
		{
			// "should handle balanced quotes": exactly two parameters.
			name:   "balanced quotes",
			in:     `text/html; param="charset=\"utf-8\"; foo=bar"; bar=foo`,
			params: map[string]string{"param": `charset="utf-8"; foo=bar`, "bar": "foo"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, err := Parse(c.in)
			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", c.in, err)
			}
			if len(m.Parameters) != len(c.params) {
				t.Fatalf("Parse(%q) got %d params (%v), want %d",
					c.in, len(m.Parameters), m.Parameters, len(c.params))
			}
			for k, v := range c.params {
				if got := m.Parameters[k]; got != v {
					t.Errorf("Parse(%q) param %q = %q, want %q", c.in, k, got, v)
				}
			}
		})
	}
}

// TestParityParseInvalid mirrors the upstream invalidTypes list and the
// "should throw on invalid parameter format" vectors. Upstream throws; the port
// returns a non-nil error.
func TestParityParseInvalid(t *testing.T) {
	invalid := []string{
		// invalidTypes (identical in v0.3.0 and master):
		" ",
		"null",
		"undefined",
		"/",
		"text/;plain",
		`text/"plain"`,
		"text/p£ain", // text/p£ain
		"text/(plain)",
		"text/@plain",
		"text/plain,wrong",
		// invalid parameter format:
		`text/plain; foo="bar`,
		"text/plain; profile=http://localhost; foo=bar",
		"text/plain; profile=http://localhost",
	}
	for _, s := range invalid {
		if _, err := Parse(s); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", s)
		}
	}
}

// TestParityFormat mirrors the "typer.format(obj)" success vectors.
func TestParityFormat(t *testing.T) {
	cases := []struct {
		name string
		m    MediaType
		want string
	}{
		{"basic", MediaType{Type: "text", Subtype: "html"}, "text/html"},
		{"suffix", MediaType{Type: "image", Subtype: "svg", Suffix: "xml"}, "image/svg+xml"},
		{
			"parameter",
			MediaType{Type: "text", Subtype: "html", Parameters: map[string]string{"charset": "utf-8"}},
			"text/html; charset=utf-8",
		},
		{
			// "should format type with parameter that needs quotes"
			"parameter needs quotes",
			MediaType{Type: "text", Subtype: "html", Parameters: map[string]string{"foo": `bar or "baz"`}},
			`text/html; foo="bar or \"baz\""`,
		},
		{
			// "should format type with parameter with empty value"
			"empty parameter value",
			MediaType{Type: "text", Subtype: "html", Parameters: map[string]string{"foo": ""}},
			`text/html; foo=""`,
		},
		{
			// "should format type with multiple parameters" (sorted output)
			"multiple parameters sorted",
			MediaType{Type: "text", Subtype: "html", Parameters: map[string]string{"charset": "utf-8", "foo": "bar", "bar": "baz"}},
			"text/html; bar=baz; charset=utf-8; foo=bar",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Format(c.m)
			if err != nil {
				t.Fatalf("Format(%+v) unexpected error: %v", c.m, err)
			}
			if got != c.want {
				t.Errorf("Format(%+v) = %q, want %q", c.m, got, c.want)
			}
		})
	}
}

// TestParityFormatInvalid mirrors the "typer.format(obj)" rejection vectors.
// Upstream throws; the port returns a non-nil error.
func TestParityFormatInvalid(t *testing.T) {
	cases := []struct {
		name string
		m    MediaType
	}{
		{"require type", MediaType{}},
		{"invalid type", MediaType{Type: "text/", Subtype: "html"}},
		{"require subtype", MediaType{Type: "text"}},
		{"invalid subtype", MediaType{Type: "text", Subtype: "html/"}},
		{"invalid suffix", MediaType{Type: "image", Subtype: "svg", Suffix: `xml\`}},
		{"invalid parameter name", MediaType{Type: "image", Subtype: "svg", Parameters: map[string]string{"foo/": "bar"}}},
		{"invalid parameter value", MediaType{Type: "image", Subtype: "svg", Parameters: map[string]string{"foo": "bar\x00"}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Format(c.m); err == nil {
				t.Errorf("Format(%+v) expected error, got nil", c.m)
			}
		})
	}
}

// TestParityRoundTrip confirms parse->format stability for canonical inputs
// drawn from the upstream vectors.
func TestParityRoundTrip(t *testing.T) {
	inputs := []string{
		"text/html",
		"image/svg+xml",
		"text/html; charset=utf-8",
	}
	for _, in := range inputs {
		m, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q): %v", in, err)
		}
		out, err := Format(m)
		if err != nil {
			t.Fatalf("Format after Parse(%q): %v", in, err)
		}
		if out != in {
			t.Errorf("round-trip %q -> %q", in, out)
		}
	}
}
