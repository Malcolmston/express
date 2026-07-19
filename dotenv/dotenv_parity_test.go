package dotenv

// Upstream-parity tests for the Go port of npm "dotenv" (motdotla/dotenv),
// version 17.4.2 on the "master" branch. Every input/expected vector below is
// taken verbatim from the upstream test suite and fixtures:
//
//	https://raw.githubusercontent.com/motdotla/dotenv/master/tests/.env
//	https://raw.githubusercontent.com/motdotla/dotenv/master/tests/test-parse.js
//	https://raw.githubusercontent.com/motdotla/dotenv/master/tests/.env.multiline
//	https://raw.githubusercontent.com/motdotla/dotenv/master/tests/test-parse-multiline.js
//	https://raw.githubusercontent.com/motdotla/dotenv/master/lib/main.js
//
// The .env fixtures are reproduced here as Go string slices joined with "\n".
// Backticks are written as literal characters inside double-quoted Go strings
// (where a backtick is not special), and a literal backslash-n in a fixture is
// spelled "\\n" so the parser sees the two characters dotenv would see.

import (
	"reflect"
	"testing"
)

// envFixture reproduces upstream tests/.env exactly (all single-line entries).
var envFixture = joinLines(
	"BASIC=basic",
	"",
	"# previous line intentionally left blank",
	"AFTER_LINE=after_line",
	"EMPTY=",
	"EMPTY_SINGLE_QUOTES=''",
	"EMPTY_DOUBLE_QUOTES=\"\"",
	"EMPTY_BACKTICKS=``",
	"SINGLE_QUOTES='single_quotes'",
	"SINGLE_QUOTES_SPACED='    single quotes    '",
	"DOUBLE_QUOTES=\"double_quotes\"",
	"DOUBLE_QUOTES_SPACED=\"    double quotes    \"",
	"DOUBLE_QUOTES_INSIDE_SINGLE='double \"quotes\" work inside single quotes'",
	"DOUBLE_QUOTES_WITH_NO_SPACE_BRACKET=\"{ port: $MONGOLAB_PORT}\"",
	"SINGLE_QUOTES_INSIDE_DOUBLE=\"single 'quotes' work inside double quotes\"",
	"BACKTICKS_INSIDE_SINGLE='`backticks` work inside single quotes'",
	"BACKTICKS_INSIDE_DOUBLE=\"`backticks` work inside double quotes\"",
	"BACKTICKS=`backticks`",
	"BACKTICKS_SPACED=`    backticks    `",
	"DOUBLE_QUOTES_INSIDE_BACKTICKS=`double \"quotes\" work inside backticks`",
	"SINGLE_QUOTES_INSIDE_BACKTICKS=`single 'quotes' work inside backticks`",
	"DOUBLE_AND_SINGLE_QUOTES_INSIDE_BACKTICKS=`double \"quotes\" and single 'quotes' work inside backticks`",
	"EXPAND_NEWLINES=\"expand\\nnew\\nlines\"",
	"DONT_EXPAND_UNQUOTED=dontexpand\\nnewlines",
	"DONT_EXPAND_SQUOTED='dontexpand\\nnewlines'",
	"# COMMENTS=work",
	"INLINE_COMMENTS=inline comments # work #very #well",
	"INLINE_COMMENTS_SINGLE_QUOTES='inline comments outside of #singlequotes' # work",
	"INLINE_COMMENTS_DOUBLE_QUOTES=\"inline comments outside of #doublequotes\" # work",
	"INLINE_COMMENTS_BACKTICKS=`inline comments outside of #backticks` # work",
	"INLINE_COMMENTS_SPACE=inline comments start with a#number sign. no space required.",
	"EQUAL_SIGNS=equals==",
	"RETAIN_INNER_QUOTES={\"foo\": \"bar\"}",
	"RETAIN_INNER_QUOTES_AS_STRING='{\"foo\": \"bar\"}'",
	"RETAIN_INNER_QUOTES_AS_BACKTICKS=`{\"foo\": \"bar's\"}`",
	"TRIM_SPACE_FROM_UNQUOTED=    some spaced out string",
	"USERNAME=therealnerdybeast@example.tld",
	"    SPACED_KEY = parsed",
	"export EXPORT_IS_DECLARED=parsed",
	"export   EXPORT_IS_DECLARED_WITH_SPACING=parsed",
	"export EXPORT_IS_DECLARED_WITH_SOME_VALUE=some_value",
	"export EXPORT_IS_DECLARED_WITH_SOME_VALUE_SPACED=some_value",
	"export   EXPORT_IS_DECLARED_WITH_SOME_VALUE_AND_SPACING  =some_value",
)

// multilineFixture reproduces upstream tests/.env.multiline exactly.
var multilineFixture = joinLines(
	"BASIC=basic",
	"",
	"# previous line intentionally left blank",
	"AFTER_LINE=after_line",
	"EMPTY=",
	"SINGLE_QUOTES='single_quotes'",
	"SINGLE_QUOTES_SPACED='    single quotes    '",
	"DOUBLE_QUOTES=\"double_quotes\"",
	"DOUBLE_QUOTES_SPACED=\"    double quotes    \"",
	"EXPAND_NEWLINES=\"expand\\nnew\\nlines\"",
	"DONT_EXPAND_UNQUOTED=dontexpand\\nnewlines",
	"DONT_EXPAND_SQUOTED='dontexpand\\nnewlines'",
	"# COMMENTS=work",
	"EQUAL_SIGNS=equals==",
	"RETAIN_INNER_QUOTES={\"foo\": \"bar\"}",
	"",
	"RETAIN_INNER_QUOTES_AS_STRING='{\"foo\": \"bar\"}'",
	"TRIM_SPACE_FROM_UNQUOTED=    some spaced out string",
	"USERNAME=therealnerdybeast@example.tld",
	"    SPACED_KEY = parsed",
	"",
	"MULTI_DOUBLE_QUOTED=\"THIS",
	"IS",
	"A",
	"MULTILINE",
	"STRING\"",
	"",
	"MULTI_SINGLE_QUOTED='THIS",
	"IS",
	"A",
	"MULTILINE",
	"STRING'",
	"",
	"MULTI_BACKTICKED=`THIS",
	"IS",
	"A",
	"\"MULTILINE'S\"",
	"STRING`",
	"",
	"MULTI_PEM_DOUBLE_QUOTED=\"-----BEGIN PUBLIC KEY-----",
	"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnNl1tL3QjKp3DZWM0T3u",
	"LgGJQwu9WqyzHKZ6WIA5T+7zPjO1L8l3S8k8YzBrfH4mqWOD1GBI8Yjq2L1ac3Y/",
	"bTdfHN8CmQr2iDJC0C6zY8YV93oZB3x0zC/LPbRYpF8f6OqX1lZj5vo2zJZy4fI/",
	"kKcI5jHYc8VJq+KCuRZrvn+3V+KuL9tF9v8ZgjF2PZbU+LsCy5Yqg1M8f5Jp5f6V",
	"u4QuUoobAgMBAAE=",
	"-----END PUBLIC KEY-----\"",
)

func joinLines(lines ...string) string {
	out := ""
	for _, l := range lines {
		out += l + "\n"
	}
	return out
}

// TestParityEnvFixture checks every t.equal assertion from test-parse.js.
func TestParityEnvFixture(t *testing.T) {
	m, err := Bytes([]byte(envFixture))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"BASIC":                                          "basic",
		"AFTER_LINE":                                     "after_line",
		"EMPTY":                                          "",
		"EMPTY_SINGLE_QUOTES":                            "",
		"EMPTY_DOUBLE_QUOTES":                            "",
		"EMPTY_BACKTICKS":                                "",
		"SINGLE_QUOTES":                                  "single_quotes",
		"SINGLE_QUOTES_SPACED":                           "    single quotes    ",
		"DOUBLE_QUOTES":                                  "double_quotes",
		"DOUBLE_QUOTES_SPACED":                           "    double quotes    ",
		"DOUBLE_QUOTES_INSIDE_SINGLE":                    "double \"quotes\" work inside single quotes",
		"DOUBLE_QUOTES_WITH_NO_SPACE_BRACKET":            "{ port: $MONGOLAB_PORT}",
		"SINGLE_QUOTES_INSIDE_DOUBLE":                    "single 'quotes' work inside double quotes",
		"BACKTICKS_INSIDE_SINGLE":                        "`backticks` work inside single quotes",
		"BACKTICKS_INSIDE_DOUBLE":                        "`backticks` work inside double quotes",
		"BACKTICKS":                                      "backticks",
		"BACKTICKS_SPACED":                               "    backticks    ",
		"DOUBLE_QUOTES_INSIDE_BACKTICKS":                 "double \"quotes\" work inside backticks",
		"SINGLE_QUOTES_INSIDE_BACKTICKS":                 "single 'quotes' work inside backticks",
		"DOUBLE_AND_SINGLE_QUOTES_INSIDE_BACKTICKS":      "double \"quotes\" and single 'quotes' work inside backticks",
		"EXPAND_NEWLINES":                                "expand\nnew\nlines",
		"DONT_EXPAND_UNQUOTED":                           "dontexpand\\nnewlines",
		"DONT_EXPAND_SQUOTED":                            "dontexpand\\nnewlines",
		"INLINE_COMMENTS":                                "inline comments",
		"INLINE_COMMENTS_SINGLE_QUOTES":                  "inline comments outside of #singlequotes",
		"INLINE_COMMENTS_DOUBLE_QUOTES":                  "inline comments outside of #doublequotes",
		"INLINE_COMMENTS_BACKTICKS":                      "inline comments outside of #backticks",
		"INLINE_COMMENTS_SPACE":                          "inline comments start with a",
		"EQUAL_SIGNS":                                    "equals==",
		"RETAIN_INNER_QUOTES":                            "{\"foo\": \"bar\"}",
		"RETAIN_INNER_QUOTES_AS_STRING":                  "{\"foo\": \"bar\"}",
		"RETAIN_INNER_QUOTES_AS_BACKTICKS":               "{\"foo\": \"bar's\"}",
		"TRIM_SPACE_FROM_UNQUOTED":                       "some spaced out string",
		"USERNAME":                                       "therealnerdybeast@example.tld",
		"SPACED_KEY":                                     "parsed",
		"EXPORT_IS_DECLARED":                             "parsed",
		"EXPORT_IS_DECLARED_WITH_SPACING":                "parsed",
		"EXPORT_IS_DECLARED_WITH_SOME_VALUE":             "some_value",
		"EXPORT_IS_DECLARED_WITH_SOME_VALUE_SPACED":      "some_value",
		"EXPORT_IS_DECLARED_WITH_SOME_VALUE_AND_SPACING": "some_value",
	}
	for k, w := range want {
		if got, ok := m[k]; !ok || got != w {
			t.Errorf("%s = %q (ok=%v), want %q", k, got, ok, w)
		}
	}
	// t.notOk(parsed.COMMENTS): the commented-out line must not appear.
	if v, ok := m["COMMENTS"]; ok {
		t.Errorf("COMMENTS should be absent, got %q", v)
	}
}

// TestParityMultilineFixture checks every t.equal from test-parse-multiline.js.
func TestParityMultilineFixture(t *testing.T) {
	m, err := Bytes([]byte(multilineFixture))
	if err != nil {
		t.Fatal(err)
	}
	multiPem := "-----BEGIN PUBLIC KEY-----\n" +
		"MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAnNl1tL3QjKp3DZWM0T3u\n" +
		"LgGJQwu9WqyzHKZ6WIA5T+7zPjO1L8l3S8k8YzBrfH4mqWOD1GBI8Yjq2L1ac3Y/\n" +
		"bTdfHN8CmQr2iDJC0C6zY8YV93oZB3x0zC/LPbRYpF8f6OqX1lZj5vo2zJZy4fI/\n" +
		"kKcI5jHYc8VJq+KCuRZrvn+3V+KuL9tF9v8ZgjF2PZbU+LsCy5Yqg1M8f5Jp5f6V\n" +
		"u4QuUoobAgMBAAE=\n" +
		"-----END PUBLIC KEY-----"
	want := map[string]string{
		"BASIC":                         "basic",
		"AFTER_LINE":                    "after_line",
		"EMPTY":                         "",
		"SINGLE_QUOTES":                 "single_quotes",
		"SINGLE_QUOTES_SPACED":          "    single quotes    ",
		"DOUBLE_QUOTES":                 "double_quotes",
		"DOUBLE_QUOTES_SPACED":          "    double quotes    ",
		"EXPAND_NEWLINES":               "expand\nnew\nlines",
		"DONT_EXPAND_UNQUOTED":          "dontexpand\\nnewlines",
		"DONT_EXPAND_SQUOTED":           "dontexpand\\nnewlines",
		"EQUAL_SIGNS":                   "equals==",
		"RETAIN_INNER_QUOTES":           "{\"foo\": \"bar\"}",
		"RETAIN_INNER_QUOTES_AS_STRING": "{\"foo\": \"bar\"}",
		"TRIM_SPACE_FROM_UNQUOTED":      "some spaced out string",
		"USERNAME":                      "therealnerdybeast@example.tld",
		"SPACED_KEY":                    "parsed",
		"MULTI_DOUBLE_QUOTED":           "THIS\nIS\nA\nMULTILINE\nSTRING",
		"MULTI_SINGLE_QUOTED":           "THIS\nIS\nA\nMULTILINE\nSTRING",
		"MULTI_BACKTICKED":              "THIS\nIS\nA\n\"MULTILINE'S\"\nSTRING",
		"MULTI_PEM_DOUBLE_QUOTED":       multiPem,
	}
	for k, w := range want {
		if got, ok := m[k]; !ok || got != w {
			t.Errorf("%s = %q (ok=%v), want %q", k, got, ok, w)
		}
	}
	if v, ok := m["COMMENTS"]; ok {
		t.Errorf("COMMENTS should be absent, got %q", v)
	}
}

// TestParityBuffer mirrors the Buffer.from parse assertions from test-parse.js.
func TestParityBuffer(t *testing.T) {
	m, _ := Bytes([]byte("BUFFER=true"))
	if m["BUFFER"] != "true" {
		t.Errorf("BUFFER = %q, want %q", m["BUFFER"], "true")
	}

	// Last duplicate key wins.
	dup, _ := Bytes([]byte("DUP=one\nDUP=two"))
	if dup["DUP"] != "two" {
		t.Errorf("DUP = %q, want %q", dup["DUP"], "two")
	}
}

// TestParityLineEndings mirrors the \r, \n, and \r\n line-ending assertions.
func TestParityLineEndings(t *testing.T) {
	want := map[string]string{"SERVER": "localhost", "PASSWORD": "password", "DB": "tests"}
	for _, tc := range []struct {
		name string
		in   string
	}{
		{"CR", "SERVER=localhost\rPASSWORD=password\rDB=tests\r"},
		{"LF", "SERVER=localhost\nPASSWORD=password\nDB=tests\n"},
		{"CRLF", "SERVER=localhost\r\nPASSWORD=password\r\nDB=tests\r\n"},
	} {
		got, _ := Bytes([]byte(tc.in))
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %#v, want %#v", tc.name, got, want)
		}
	}
}
