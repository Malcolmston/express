package vary

// Upstream parity tests for the npm "vary" package (jshttp/vary v1.1.2).
//
// Every input -> expected-output vector below is taken verbatim from the
// original library's test suite:
//
//	https://raw.githubusercontent.com/jshttp/vary/master/test/test.js
//
// (Reference implementation: https://raw.githubusercontent.com/jshttp/vary/master/index.js)
//
// Mapping notes: upstream exposes a single overloaded function whose `field`
// argument may be a string OR an array. This Go port splits it into a
// variadic Append/Vary plus Field. The variadic form corresponds to upstream's
// ARRAY form, in which each element must be a bare RFC 7230 token (upstream
// validates array elements directly against FIELD_NAME_REGEXP, without comma/
// space parsing). Therefore only upstream vectors expressible as an array of
// bare tokens are encoded here. Upstream's STRING form, which parses a single
// comma/space-separated string into multiple fields (e.g. append('', 'a, b')),
// has no Go equivalent and is intentionally not ported -- see the package
// divergence note at the bottom of this file. Those string-parsing vectors are
// excluded rather than asserted as failures.

import (
	"net/http"
	"testing"
)

// TestParityAppend covers vary.append(header, field) from test/test.js where
// field is an array of bare tokens (the shape the variadic port supports).
func TestParityAppend(t *testing.T) {
	cases := []struct {
		name   string
		header string
		fields []string
		want   string
	}{
		// describe('when header empty')
		{"empty/set", "", []string{"Origin"}, "Origin"},
		{"empty/array", "", []string{"Origin", "User-Agent"}, "Origin, User-Agent"},
		{"empty/preserve-case", "", []string{"ORIGIN", "user-agent", "AccepT"}, "ORIGIN, user-agent, AccepT"},

		// describe('when header has values')
		{"has/set", "Accept", []string{"Origin"}, "Accept, Origin"},
		{"has/array", "Accept", []string{"Origin", "User-Agent"}, "Accept, Origin, User-Agent"},
		{"has/no-duplicate", "Accept", []string{"Accept"}, "Accept"},
		{"has/case-insensitive", "Accept", []string{"accEPT"}, "Accept"},
		{"has/preserve-case", "Accept", []string{"AccepT"}, "Accept"},

		// describe('when *')
		{"star/set", "", []string{"*"}, "*"},
		{"star/all-already-set", "*", []string{"Origin"}, "*"},
		{"star/erradicate-existing", "Accept, Accept-Encoding", []string{"*"}, "*"},
		{"star/bad-existing-header", "Accept, Accept-Encoding, *", []string{"Origin"}, "*"},

		// describe('when field is string') -> single-token subset only
		{"str/set", "", []string{"Accept"}, "Accept"},

		// describe('when field is array')
		{"arr/set", "", []string{"Accept", "Accept-Language"}, "Accept, Accept-Language"},
		{"arr/ignore-double", "", []string{"Accept", "Accept"}, "Accept"},
		{"arr/case-insensitive", "", []string{"Accept", "ACCEPT"}, "Accept"},
		{"arr/contained-star", "", []string{"Origin", "User-Agent", "*", "Accept"}, "*"},
		{"arr/existing-values", "Accept, Accept-Encoding", []string{"origin", "accept", "accept-charset"}, "Accept, Accept-Encoding, origin, accept-charset"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Append(tc.header, tc.fields...)
			if err != nil {
				t.Fatalf("Append(%q, %q) unexpected error: %v", tc.header, tc.fields, err)
			}
			if got != tc.want {
				t.Fatalf("Append(%q, %q) = %q; want %q", tc.header, tc.fields, got, tc.want)
			}
		})
	}
}

// TestParityAppendInvalid covers the "should not allow ..." field-name vectors,
// which upstream asserts throw a TypeError and this port returns an error for.
func TestParityAppendInvalid(t *testing.T) {
	cases := []struct {
		name  string
		field string
	}{
		{"separator-colon", "invalid:header"},
		{"separator-space", "invalid header"},
		{"newline", "invalid\nheader"},
		{"high-byte", "invalidheader"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Append("", tc.field); err == nil {
				t.Fatalf("Append(%q, %q) = nil error; want error", "", tc.field)
			}
		})
	}
}

// TestParityVary covers vary(res, field) from test/test.js by mutating an
// http.Header the same way the upstream server handler mutates the Node
// response. An existing Vary that upstream sets as an array is joined with
// ", " here, matching how Node's res.getHeader returns and vary() joins it.
func TestParityVary(t *testing.T) {
	cases := []struct {
		name     string
		existing string   // pre-set Vary header; "" means unset
		setExist bool     // whether to set the existing header at all
		fields   []string // fields passed to vary()
		want     string   // expected resulting Vary header ("" means unset)
	}{
		// describe('when no Vary')
		{name: "no-vary/set", fields: []string{"Origin"}, want: "Origin"},
		{name: "no-vary/multiple", fields: []string{"Origin", "User-Agent"}, want: "Origin, User-Agent"},
		{name: "no-vary/preserve-case", fields: []string{"ORIGIN", "user-agent", "AccepT"}, want: "ORIGIN, user-agent, AccepT"},
		{name: "no-vary/empty-array", fields: []string{}, want: ""},

		// describe('when existing Vary')
		{name: "existing/set", existing: "Accept", setExist: true, fields: []string{"Origin"}, want: "Accept, Origin"},
		{name: "existing/no-duplicate", existing: "Accept", setExist: true, fields: []string{"Accept"}, want: "Accept"},
		{name: "existing/case-insensitive", existing: "Accept", setExist: true, fields: []string{"accEPT"}, want: "Accept"},
		{name: "existing/preserve-case", existing: "AccepT", setExist: true, fields: []string{"accEPT", "ORIGIN"}, want: "AccepT, ORIGIN"},

		// describe('when existing Vary as array') -> joined form
		{name: "existing-array/set", existing: "Accept, Accept-Encoding", setExist: true, fields: []string{"Origin"}, want: "Accept, Accept-Encoding, Origin"},
		{name: "existing-array/no-duplicate", existing: "Accept, Accept-Encoding", setExist: true, fields: []string{"accept", "origin"}, want: "Accept, Accept-Encoding, origin"},

		// describe('when Vary: *')
		{name: "star/set", fields: []string{"*"}, want: "*"},
		{name: "star/all-already-set", existing: "*", setExist: true, fields: []string{"Origin", "User-Agent"}, want: "*"},
		{name: "star/erradicate", existing: "Accept, Accept-Encoding", setExist: true, fields: []string{"*"}, want: "*"},
		{name: "star/bad-existing", existing: "Accept, Accept-Encoding, *", setExist: true, fields: []string{"Origin"}, want: "*"},

		// describe('when field is string') -> single-token subset only
		{name: "str/set", fields: []string{"Accept"}, want: "Accept"},

		// describe('when field is array')
		{name: "arr/set", fields: []string{"Accept", "Accept-Language"}, want: "Accept, Accept-Language"},
		{name: "arr/ignore-double", fields: []string{"Accept", "Accept"}, want: "Accept"},
		{name: "arr/case-insensitive", fields: []string{"Accept", "ACCEPT"}, want: "Accept"},
		{name: "arr/contained-star", fields: []string{"Origin", "User-Agent", "*", "Accept"}, want: "*"},
		{name: "arr/existing-values", existing: "Accept, Accept-Encoding", setExist: true, fields: []string{"origin", "accept", "accept-charset"}, want: "Accept, Accept-Encoding, origin, accept-charset"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			if tc.setExist {
				h.Set("Vary", tc.existing)
			}
			Vary(h, tc.fields...)
			if got := h.Get("Vary"); got != tc.want {
				t.Fatalf("Vary(existing=%q, %q) => %q; want %q", tc.existing, tc.fields, got, tc.want)
			}
		})
	}
}

// TestParityVaryMultipleCalls mirrors upstream's "should set value with
// multiple calls" where vary() is invoked twice on the same response.
func TestParityVaryMultipleCalls(t *testing.T) {
	h := http.Header{}
	h.Set("Vary", "Accept")
	Vary(h, "Origin")
	Vary(h, "User-Agent")
	if got := h.Get("Vary"); got != "Accept, Origin, User-Agent" {
		t.Fatalf("got %q; want %q", got, "Accept, Origin, User-Agent")
	}
}
