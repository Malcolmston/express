// Package contenttype upstream-parity tests.
//
// The input -> expected-output vectors below are transcribed verbatim from the
// real test suite of the upstream npm library jshttp/content-type (v2.0.0, the
// current master branch), fetched from:
//
//	https://raw.githubusercontent.com/jshttp/content-type/master/src/parse.spec.ts
//	https://raw.githubusercontent.com/jshttp/content-type/master/src/format.spec.ts
//
// Upstream source for reference:
//
//	https://raw.githubusercontent.com/jshttp/content-type/master/src/index.ts
//
// Upstream exposes parse(header) -> {type, parameters} and
// format({type, parameters}) -> string. This Go port exposes the equivalent
// Parse and Format over the ContentType struct.
//
// The port deliberately follows the stricter v1.x semantics (validate the media
// type on Parse, sort parameters on Format) rather than the tolerant v2.0.0
// parser. Vectors that the port satisfies are asserted directly in
// TestParityParse / TestParityFormat. Vectors where the port intentionally
// diverges from v2.0.0 are documented and pinned in TestParityDivergences so
// the delta is explicit and regression-tested; each carries the upstream
// expectation in its comment.
package contenttype

import (
	"reflect"
	"testing"
)

// TestParityParse covers upstream parse.spec.ts vectors the port matches.
func TestParityParse(t *testing.T) {
	cases := []struct {
		name   string
		in     string
		typ    string
		params map[string]string
	}{
		{"basic type", "text/html", "text/html", map[string]string{}},
		{"suffix", "image/svg+xml", "image/svg+xml", map[string]string{}},
		// Upstream lists "text/$plain" under invalidTypes but accepts it; the
		// port agrees because "$" is a valid token character (RFC 9110 tchar).
		{"dollar in subtype", "text/$plain", "text/$plain", map[string]string{}},
		{"surrounding OWS", " text/html ", "text/html", map[string]string{}},
		{"lower-case type", "IMAGE/SVG+XML", "image/svg+xml", map[string]string{}},
		{"parameters", "text/html; charset=utf-8; foo=bar", "text/html",
			map[string]string{"charset": "utf-8", "foo": "bar"}},
		{"parameters with extra LWS", "text/html ; charset=utf-8 ; foo=bar", "text/html",
			map[string]string{"charset": "utf-8", "foo": "bar"}},
		{"empty parameter value with quotes", `text/html; charset=""`, "text/html",
			map[string]string{"charset": ""}},
		{"OWS around equals", "text/html; charset = utf-8", "text/html",
			map[string]string{"charset": "utf-8"}},
		{"lower-case parameter names", "text/html; Charset=UTF-8", "text/html",
			map[string]string{"charset": "UTF-8"}},
		{"unquote parameter values", `text/html; charset="UTF-8"`, "text/html",
			map[string]string{"charset": "UTF-8"}},
		{"unquote with escapes", `text/html; charset = "UT\F-\\\"8\""`, "text/html",
			map[string]string{"charset": `UTF-\"8"`}},
		{"balanced quotes", `text/html; param="charset=\"utf-8\"; foo=bar"; bar=foo`, "text/html",
			map[string]string{"param": `charset="utf-8"; foo=bar`, "bar": "foo"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ct, err := Parse(c.in)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", c.in, err)
			}
			if ct.Type != c.typ {
				t.Errorf("Parse(%q) type = %q, want %q", c.in, ct.Type, c.typ)
			}
			if !reflect.DeepEqual(ct.Parameters, c.params) {
				t.Errorf("Parse(%q) params = %v, want %v", c.in, ct.Parameters, c.params)
			}
		})
	}
}

// TestParityFormat covers upstream format.spec.ts vectors the port matches,
// including the two parameter-value validation vectors closed by this sync.
func TestParityFormat(t *testing.T) {
	okCases := []struct {
		name   string
		typ    string
		params map[string]string
		want   string
	}{
		{"basic type", "text/html", nil, "text/html"},
		{"type with suffix", "image/svg+xml", nil, "image/svg+xml"},
		{"keep type case", "IMAGE/SVG+XML", nil, "IMAGE/SVG+XML"},
		{"keep parameter case", "text/html", map[string]string{"Charset": "utf-8"},
			"text/html; Charset=utf-8"},
		// Upstream keeps insertion order; the port sorts. Here sorted order
		// ("Charset" < "charset" in ASCII) coincides with the upstream output.
		{"keep duplicate parameters case", "text/html",
			map[string]string{"Charset": "utf-8", "charset": "iso-8859-1"},
			"text/html; Charset=utf-8; charset=iso-8859-1"},
		{"type with parameter", "text/html", map[string]string{"charset": "utf-8"},
			"text/html; charset=utf-8"},
		{"parameter that needs quotes", "text/html", map[string]string{"foo": `bar or "baz"`},
			`text/html; foo="bar or \"baz\""`},
		{"parameter with empty value", "text/html", map[string]string{"foo": ""},
			`text/html; foo=""`},
		{"parameter value containing HTAB", "text/html", map[string]string{"foo": "bar\tbaz"},
			"text/html; foo=\"bar\tbaz\""},
	}
	for _, c := range okCases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Format(ContentType{Type: c.typ, Parameters: c.params})
			if err != nil {
				t.Fatalf("Format(%q, %v) error: %v", c.typ, c.params, err)
			}
			if got != c.want {
				t.Errorf("Format(%q, %v) = %q, want %q", c.typ, c.params, got, c.want)
			}
		})
	}

	errCases := []struct {
		name   string
		typ    string
		params map[string]string
	}{
		{"reject invalid type", "text/", nil},
		{"reject invalid type with LWS", " text/html", nil},
		{"reject invalid parameter name", "image/svg", map[string]string{"foo/": "bar"}},
		// Closed by this sync: upstream qstring() rejects values outside the
		// quoted-string text class.
		{"reject invalid parameter value (NUL)", "image/svg", map[string]string{"foo": "bar\x00"}},
		{"reject parameter value containing vertical tab", "text/html", map[string]string{"foo": "bar\x0bbaz"}},
	}
	for _, c := range errCases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Format(ContentType{Type: c.typ, Parameters: c.params}); err == nil {
				t.Errorf("Format(%q, %v) expected error", c.typ, c.params)
			}
		})
	}
}

// TestParityDivergences pins vectors where this port intentionally diverges
// from upstream v2.0.0. Each subtest asserts the PORT's behavior and states the
// upstream v2.0.0 expectation in a comment. These are recorded as open gaps
// (large design differences), not defects to fix.
func TestParityDivergences(t *testing.T) {
	// --- Parse: v2.0.0 is a tolerant parser; the port follows v1.x strict
	// validation and returns an error where v2.0.0 returns a value. ---
	//
	// Upstream v2.0.0 returns, respectively:
	//   ""                                          -> {type:"",              parameters:{}}
	//   " "                                         -> {type:"",              parameters:{}}
	//   "null"                                      -> {type:"null",          parameters:{}}
	//   "undefined"                                 -> {type:"undefined",     parameters:{}}
	//   "/"                                         -> {type:"/",             parameters:{}}
	//   "text / plain"                              -> {type:"text / plain",  parameters:{}}
	//   `text/"plain"`                              -> {type:`text/"plain"`,  parameters:{}}
	//   "text/p£ain"                                -> {type:"text/p£ain",    parameters:{}}
	//   "text/(plain)"                              -> {type:"text/(plain)",  parameters:{}}
	//   "text/@plain"                               -> {type:"text/@plain",   parameters:{}}
	//   "text/plain,wrong"                          -> {type:"text/plain,wrong", parameters:{}}
	//   "text/html; charset="                       -> {charset:""}
	//   "text/html; charset= "                      -> {charset:""}
	//   "text/html;;;; charset=utf-8;; foo=bar;"    -> {charset:"utf-8", foo:"bar"}
	//   `text/plain; foo="bar`                      -> {} (unterminated quote ignored)
	//   `text/plain; foo="bar\`                     -> {} (unterminated quote ignored)
	//   `text/plain; foo="bar"baz`                  -> {foo:"bar"} (non-OWS after quote ignored)
	//   `text/plain; foo="bar"baz; charset=utf-8`   -> {foo:"bar", charset:"utf-8"}
	//   `text/plain; foo=bar"baz`                   -> {foo:`bar"baz`}
	//   "text/plain; foo=bar=baz"                   -> {foo:"bar=baz"}
	parseErrors := []string{
		"",
		" ",
		"null",
		"undefined",
		"/",
		"text / plain",
		`text/"plain"`,
		"text/p£ain",
		"text/(plain)",
		"text/@plain",
		"text/plain,wrong",
		"text/html; charset=",
		"text/html; charset= ",
		"text/html;;;; charset=utf-8;; foo=bar;",
		`text/plain; foo="bar`,
		`text/plain; foo="bar\`,
		`text/plain; foo="bar"baz`,
		`text/plain; foo="bar"baz; charset=utf-8`,
		`text/plain; foo=bar"baz`,
		"text/plain; foo=bar=baz",
	}
	for _, in := range parseErrors {
		t.Run("parse-strict/"+in, func(t *testing.T) {
			if _, err := Parse(in); err == nil {
				t.Errorf("Parse(%q): port returns no error, but upstream v2.0.0 tolerates this input (see comment)", in)
			}
		})
	}

	// Duplicate parameters: upstream v2.0.0 keeps the FIRST occurrence; this
	// port (a Go map) keeps the LAST. Upstream returns {charset:"utf-8"} for
	// each of these; the port returns {charset:"iso-8859-1"}.
	dupCases := []string{
		"text/html; charset=utf-8; charset=iso-8859-1",
		"text/html; Charset=utf-8; charset=iso-8859-1",
		`text/html; Charset="utf-8"; charset="iso-8859-1"`,
	}
	for _, in := range dupCases {
		t.Run("parse-dup-last-wins/"+in, func(t *testing.T) {
			ct, err := Parse(in)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", in, err)
			}
			if ct.Parameters["charset"] != "iso-8859-1" {
				t.Errorf("Parse(%q) charset = %q, want %q (port keeps last; upstream keeps first: %q)",
					in, ct.Parameters["charset"], "iso-8859-1", "utf-8")
			}
		})
	}

	// Format parameter ordering: upstream v2.0.0 preserves insertion order; the
	// port sorts parameter names. Upstream would emit
	// "text/html; charset=utf-8; foo=bar; bar=baz"; the port emits sorted order.
	t.Run("format-sorted-vs-insertion", func(t *testing.T) {
		got, err := Format(ContentType{
			Type:       "text/html",
			Parameters: map[string]string{"charset": "utf-8", "foo": "bar", "bar": "baz"},
		})
		if err != nil {
			t.Fatalf("Format error: %v", err)
		}
		const wantSorted = "text/html; bar=baz; charset=utf-8; foo=bar"
		if got != wantSorted {
			t.Errorf("Format = %q, want %q (sorted; upstream v2.0.0 insertion order = %q)",
				got, wantSorted, "text/html; charset=utf-8; foo=bar; bar=baz")
		}
	})
}
