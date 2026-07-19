package sqlstring

import (
	"testing"
	"time"
)

// Upstream-parity vectors ported from the real mysqljs/sqlstring test suite:
//
//	https://raw.githubusercontent.com/mysqljs/sqlstring/master/test/unit/test-SqlString.js
//
// Every input -> expected pair below is taken from an assert.equal/assert.strictEqual
// in that file, adapted to the Go port's API (Escape/EscapeID/Format). Vectors that
// exercise JS-only features the port does not implement (SqlString.raw, toSqlString,
// plain-object key/value expansion, the escapeId forbidQualified flag and array
// input, NaN/Infinity, JS timezone args) are intentionally omitted; see the task
// notes for those gaps.

func TestParityEscapeID(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"id", "`id`"},                   // 'value is quoted'
		{"i`d", "`i``d`"},                // 'value containing escapes is quoted'
		{"id1.id2", "`id1`.`id2`"},       // 'value containing separator is quoted'
		{"id`1.i`d2", "`id``1`.`i``d2`"}, // 'value containing separator and escapes'
	}
	for _, c := range cases {
		if got := EscapeID(c.in); got != c.want {
			t.Errorf("EscapeID(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestParityEscapeScalars(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{nil, "NULL"},    // 'null -> NULL' / 'undefined -> NULL'
		{false, "false"}, // 'booleans convert to strings'
		{true, "true"},
		{5, "5"},             // 'numbers convert to strings'
		{"Super", "'Super'"}, // 'strings are quoted'
	}
	for _, c := range cases {
		if got := Escape(c.in); got != c.want {
			t.Errorf("Escape(%v) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestParityEscapeStringSpecials(t *testing.T) {
	// Each pair is an exact assertion from the "gets escaped" tests.
	cases := []struct {
		in, want string
	}{
		{"Sup\x00er", `'Sup\0er'`}, // \0 gets escaped
		{"Super\x00", `'Super\0'`},
		{"Sup\ber", `'Sup\ber'`}, // \b gets escaped
		{"Super\b", `'Super\b'`},
		{"Sup\ner", `'Sup\ner'`}, // \n gets escaped
		{"Super\n", `'Super\n'`},
		{"Sup\rer", `'Sup\rer'`}, // \r gets escaped
		{"Super\r", `'Super\r'`},
		{"Sup\ter", `'Sup\ter'`}, // \t gets escaped
		{"Super\t", `'Super\t'`},
		{"Sup\\er", `'Sup\\er'`}, // \\ gets escaped
		{"Super\\", `'Super\\'`},
		{"Sup\x1aer", `'Sup\Zer'`}, // ascii 26 -> \Z
		{"Super\x1a", `'Super\Z'`},
		{"Sup'er", `'Sup\'er'`}, // single quotes get escaped
		{"Super'", `'Super\''`},
		{"Sup\"er", `'Sup\"er'`}, // double quotes get escaped
		{"Super\"", `'Super\"'`},
	}
	for _, c := range cases {
		if got := Escape(c.in); got != c.want {
			t.Errorf("Escape(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestParityEscapeBuffer(t *testing.T) {
	// 'buffers are converted to hex': new Buffer([0, 1, 254, 255]) -> X'0001feff'
	if got := Escape([]byte{0, 1, 254, 255}); got != "X'0001feff'" {
		t.Errorf("got %q want %q", got, "X'0001feff'")
	}
}

func TestParityEscapeArrays(t *testing.T) {
	// 'arrays are turned into lists': escape([1, 2, 'c']) -> "1, 2, 'c'"
	if got := Escape([]any{1, 2, "c"}); got != "1, 2, 'c'" {
		t.Errorf("flat list got %q want %q", got, "1, 2, 'c'")
	}
	// 'nested arrays are turned into grouped lists': the numeric groups from
	// escape([[1,2,3],[4,5,6],...]) render as "(1, 2, 3), (4, 5, 6)".
	if got := Escape([][]int{{1, 2, 3}, {4, 5, 6}}); got != "(1, 2, 3), (4, 5, 6)" {
		t.Errorf("nested groups got %q want %q", got, "(1, 2, 3), (4, 5, 6)")
	}
}

func TestParityEscapeDate(t *testing.T) {
	// 'dates are converted to YYYY-MM-DD HH:II:SS.sss':
	// new Date(2012, 4, 7, 11, 42, 3, 2) -> '2012-05-07 11:42:03.002'.
	// Constructed in UTC so the Go port formats without a zone shift.
	d := time.Date(2012, 5, 7, 11, 42, 3, 2_000_000, time.UTC)
	if got := Escape(d); got != "'2012-05-07 11:42:03.002'" {
		t.Errorf("got %q want %q", got, "'2012-05-07 11:42:03.002'")
	}
}

func TestParityFormat(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		args []any
		want string
	}{
		{
			"question marks replaced with escaped values",
			"? and ?", []any{"a", "b"}, "'a' and 'b'",
		},
		{
			"double quest marks replaced with escaped id",
			"SELECT * FROM ?? WHERE id = ?", []any{"table", 42},
			"SELECT * FROM `table` WHERE id = 42",
		},
		{
			"triple question marks are ignored",
			"? or ??? and ?", []any{"foo", "bar", "fizz", "buzz"},
			"'foo' or ??? and 'bar'",
		},
		{
			"extra question marks are left untouched",
			"? and ?", []any{"a"}, "'a' and ?",
		},
		{
			"extra arguments are not used",
			"? and ?", []any{"a", "b", "c"}, "'a' and 'b'",
		},
		{
			"question marks within values do not cause issues",
			"? and ?", []any{"hello?", "b"}, "'hello?' and 'b'",
		},
		{
			"sql untouched if values provided but no placeholders",
			"SELECT COUNT(*) FROM table", []any{"a", "b"},
			"SELECT COUNT(*) FROM table",
		},
	}
	for _, c := range cases {
		if got := Format(c.sql, c.args); got != c.want {
			t.Errorf("%s: Format(%q, %v) = %q want %q", c.name, c.sql, c.args, got, c.want)
		}
	}
}

func TestParityFormatNoValues(t *testing.T) {
	// 'sql is untouched if no values are provided': format('SELECT ??') -> 'SELECT ??'
	if got := Format("SELECT ??", nil); got != "SELECT ??" {
		t.Errorf("got %q want %q", got, "SELECT ??")
	}
}

func TestParityFormatIDNonString(t *testing.T) {
	// escapeId(42) -> `42`, exercised through the ?? placeholder.
	if got := Format("x ??", []any{42}); got != "x `42`" {
		t.Errorf("got %q want %q", got, "x `42`")
	}
}
