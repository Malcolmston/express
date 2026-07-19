package contentdisposition

import "testing"

// Upstream parity vectors for jshttp/content-disposition.
//
// Every input/expected pair below is copied verbatim from the ORIGINAL
// library's own test suites (real values, nothing invented):
//
//   - Current release 2.0.1 (create/parse specs):
//     https://raw.githubusercontent.com/jshttp/content-disposition/master/src/create.spec.ts
//     https://raw.githubusercontent.com/jshttp/content-disposition/master/src/parse.spec.ts
//   - Canonical long-stable suite (v1.0.0 mocha tests):
//     https://raw.githubusercontent.com/jshttp/content-disposition/v1.0.0/test/test.js
//
// Mapping notes:
//   - Upstream contentDisposition()/create() maps to this port's Format; the
//     upstream {type, parameters} result of parse() maps to Parse, where the
//     decoded filename is exposed as ContentDisposition.Filename.
//   - This port emits a quoted filename (RFC 6266 quoted-string) even for
//     simple tokens, matching the v1.x line of upstream. The 2.0.x line emits a
//     bare token there (filename=plans.pdf); that stylistic divergence and a
//     handful of others are documented in the task notes, not encoded here.

// TestParityFormat covers the create() / contentDisposition(filename, opts)
// output vectors.
func TestParityFormat(t *testing.T) {
	cases := []struct {
		name string
		in   string
		opts []Option
		want string
		src  string
	}{
		// --- 2.0.1 src/create.spec.ts (non-ASCII fallback + RFC 8187) ---
		{"empty-is-attachment", "", nil, "attachment", "create.spec.ts"},
		{"escape-quotes", `the "plans".pdf`, nil,
			`attachment; filename="the \"plans\".pdf"`, "create.spec.ts"},
		{"iso8859-1-guillemets", "«plans».pdf", nil,
			`attachment; filename="?plans?.pdf"; filename*=UTF-8''%C2%ABplans%C2%BB.pdf`, "create.spec.ts"},
		{"iso8859-1-with-quotes", `the "plans" (1µ).pdf`, nil,
			`attachment; filename="the \"plans\" (1?).pdf"; filename*=UTF-8''the%20%22plans%22%20%281%C2%B5%29.pdf`, "create.spec.ts"},
		{"latin-diacritic", "foo-ä.html", nil,
			`attachment; filename="foo-?.html"; filename*=UTF-8''foo-%C3%A4.html`, "create.spec.ts"},
		{"unicode-cyrillic", "планы.pdf", nil,
			`attachment; filename="?????.pdf"; filename*=UTF-8''%D0%BF%D0%BB%D0%B0%D0%BD%D1%8B.pdf`, "create.spec.ts"},
		{"unicode-fallback-mixed", "£ and € rates.pdf", nil,
			`attachment; filename="? and ? rates.pdf"; filename*=UTF-8''%C2%A3%20and%20%E2%82%AC%20rates.pdf`, "create.spec.ts"},
		{"unicode-fallback-euro", "€ rates.pdf", nil,
			`attachment; filename="? rates.pdf"; filename*=UTF-8''%E2%82%AC%20rates.pdf`, "create.spec.ts"},
		{"special-chars", "€'*%().pdf", nil,
			`attachment; filename="?'*%().pdf"; filename*=UTF-8''%E2%82%AC%27%2A%25%28%29.pdf`, "create.spec.ts"},
		{"unicode-with-hex", "€%20£.pdf", nil,
			`attachment; filename="?%20?.pdf"; filename*=UTF-8''%E2%82%AC%2520%C2%A3.pdf`, "create.spec.ts"},
		{"fallback-false-mixed", "£ and € rates.pdf", []Option{WithFallback(false)},
			`attachment; filename*=UTF-8''%C2%A3%20and%20%E2%82%AC%20rates.pdf`, "create.spec.ts"},
		{"fallback-false-cyrillic", "планы.pdf", []Option{WithFallback(false)},
			`attachment; filename*=UTF-8''%D0%BF%D0%BB%D0%B0%D0%BD%D1%8B.pdf`, "create.spec.ts"},
		{"fallback-false-pound", "£ rates.pdf", []Option{WithFallback(false)},
			`attachment; filename*=UTF-8''%C2%A3%20rates.pdf`, "create.spec.ts"},
		{"type-inline", "", []Option{WithType("inline")}, "inline", "create.spec.ts"},
		{"type-preserve-casing", "", []Option{WithType("INLINE")}, "INLINE", "create.spec.ts"},

		// --- v1.0.0 test/test.js (quoted-filename form this port follows) ---
		{"ascii-plans", "plans.pdf", nil,
			`attachment; filename="plans.pdf"`, "v1.0.0 test.js"},
		{"inline-with-filename", "plans.pdf", []Option{WithType("inline")},
			`inline; filename="plans.pdf"`, "v1.0.0 test.js"},
		{"ascii-hex-escape", "the%20plans.pdf", nil,
			`attachment; filename="the%20plans.pdf"; filename*=UTF-8''the%2520plans.pdf`, "v1.0.0 test.js"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Format(tc.in, tc.opts...)
			if got != tc.want {
				t.Errorf("Format(%q) = %q\n want %q\n (upstream %s)", tc.in, got, tc.want, tc.src)
			}
		})
	}
}

// TestParityParse covers the parse() vectors, comparing the disposition type
// and the decoded filename (the port's primary user-facing contract).
func TestParityParse(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		wantType string
		wantName string
		src      string
	}{
		// --- basic types ---
		{"attachment", "attachment", "attachment", "", "parse.spec.ts"},
		{"inline", "inline", "inline", "", "parse.spec.ts"},
		{"form-data", "form-data", "form-data", "", "parse.spec.ts"},
		{"uppercase-type", "ATTACHMENT", "attachment", "", "parse.spec.ts"},
		{"extension-type", "foobar", "foobar", "", "parse.spec.ts"},

		// --- quoted / token filenames ---
		{"quoted-filename", `attachment; filename="foo.html"`, "attachment", "foo.html", "parse.spec.ts"},
		{"token-filename", "attachment; filename=foo.html", "attachment", "foo.html", "parse.spec.ts"},
		{"unescape-quoted", `attachment; filename="the \"plans\".pdf"`, "attachment", `the "plans".pdf`, "parse.spec.ts"},
		{"quoting-tested", `attachment; filename="\"quoting\" tested.html"`, "attachment", `"quoting" tested.html`, "parse.spec.ts"},
		{"backslash-o", `attachment; filename="f\oo.html"`, "attachment", "foo.html", "parse.spec.ts"},
		{"leading-slash", `attachment; filename="/foo.html"`, "attachment", "/foo.html", "parse.spec.ts"},
		{"double-backslash", `attachment; filename="\\foo.html"`, "attachment", `\foo.html`, "parse.spec.ts"},
		{"iso8859-1-quoted", `attachment; filename="£ rates.pdf"`, "attachment", "£ rates.pdf", "parse.spec.ts"},
		{"uppercase-param", `attachment; FILENAME="foo.html"`, "attachment", "foo.html", "parse.spec.ts"},
		{"literal-percent41", `attachment; filename="foo-%41.html"`, "attachment", "foo-%41.html", "parse.spec.ts"},
		{"trailing-percent", `attachment; filename="50%.html"`, "attachment", "50%.html", "parse.spec.ts"},
		{"single-quotes", `attachment; filename='foo.bar'`, "attachment", "'foo.bar'", "parse.spec.ts"},
		{"inline-filename", `inline; filename="foo.html"`, "inline", "foo.html", "parse.spec.ts"},
		{"not-an-attachment", `inline; filename="Not an attachment!"`, "inline", "Not an attachment!", "parse.spec.ts"},

		// --- extended (RFC 8187) parameter values ---
		{"ext-utf8", "attachment; filename*=UTF-8''%E2%82%AC%20rates.pdf", "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-utf8-alias", "attachment; filename*=utf8''%E2%82%AC%20rates.pdf", "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-utf8-lowercase", "attachment; filename*=utf-8''%E2%82%AC%20rates.pdf", "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-iso8859-1", "attachment; filename*=ISO-8859-1''%A3%20rates.pdf", "attachment", "£ rates.pdf", "parse.spec.ts"},
		{"ext-iso8859-1-e4", "attachment; filename*=iso-8859-1''foo-%E4.html", "attachment", "foo-ä.html", "parse.spec.ts"},
		{"ext-utf8-multibyte", "attachment; filename*=UTF-8''foo-%c3%a4-%e2%82%ac.html", "attachment", "foo-ä-€.html", "parse.spec.ts"},
		{"ext-embedded-language", "attachment; filename*=UTF-8'en'%E2%82%AC%20rates.pdf", "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-prefer-over-plain", `attachment; filename="EURO rates.pdf"; filename*=UTF-8''%E2%82%AC%20rates.pdf`, "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-prefer-reversed", `attachment; filename*=UTF-8''%E2%82%AC%20rates.pdf; filename="EURO rates.pdf"`, "attachment", "€ rates.pdf", "parse.spec.ts"},
		{"ext-double-percent", "attachment; filename*=UTF-8''A-%2541.html", "attachment", "A-%41.html", "parse.spec.ts"},
		{"ext-backslash", "attachment; filename*=UTF-8''%5cfoo.html", "attachment", `\foo.html`, "parse.spec.ts"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cd, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v (upstream %s)", tc.in, err, tc.src)
			}
			if cd.Type != tc.wantType {
				t.Errorf("Parse(%q).Type = %q, want %q (upstream %s)", tc.in, cd.Type, tc.wantType, tc.src)
			}
			if cd.Filename != tc.wantName {
				t.Errorf("Parse(%q).Filename = %q, want %q (upstream %s)", tc.in, cd.Filename, tc.wantName, tc.src)
			}
		})
	}
}

// TestParityParseParameters checks that non-filename parameters survive parsing
// with the exact keys/values upstream reports. Vectors from parse.spec.ts.
func TestParityParseParameters(t *testing.T) {
	cases := []struct {
		name string
		in   string
		key  string
		val  string
	}{
		{"extra-foo", `attachment; foo="bar"; filename="foo.html"`, "foo", "bar"},
		{"name-param", `attachment; name="foo-%41.html"`, "name", "foo-%41.html"},
		{"creation-date", `attachment; creation-date="Wed, 12 Feb 1997 16:29:51 -0500"`, "creation-date", "Wed, 12 Feb 1997 16:29:51 -0500"},
		{"embedded-equals", `attachment; example="filename=example.txt"`, "example", "filename=example.txt"},
		{"xfilename", "attachment; xfilename=foo.html", "xfilename", "foo.html"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cd, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.in, err)
			}
			if got := cd.Parameters[tc.key]; got != tc.val {
				t.Errorf("Parse(%q).Parameters[%q] = %q, want %q", tc.in, tc.key, got, tc.val)
			}
		})
	}
}
