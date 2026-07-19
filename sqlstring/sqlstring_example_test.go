package sqlstring_test

import (
	"fmt"
	"time"

	"github.com/malcolmston/express/sqlstring"
)

// ExampleEscape turns individual Go values into safely quoted SQL literals.
// Numbers render as bare numeric text, booleans as true/false, and nil becomes
// NULL, while strings are wrapped in single quotes with any dangerous characters
// backslash-escaped. The final case is the important one for injection safety:
// a string containing a single quote cannot terminate the literal early because
// the quote is escaped, so the malicious "OR '1'='1" fragment stays inert data
// rather than becoming SQL syntax. Each printed line shows the literal that would
// be spliced into a query.
func ExampleEscape() {
	fmt.Println(sqlstring.Escape(42))
	fmt.Println(sqlstring.Escape(true))
	fmt.Println(sqlstring.Escape(nil))
	fmt.Println(sqlstring.Escape("bob"))
	fmt.Println(sqlstring.Escape("x' OR '1'='1"))
	// Output:
	// 42
	// true
	// NULL
	// 'bob'
	// 'x\' OR \'1\'=\'1'
}

// ExampleEscape_types shows the richer values Escape understands beyond simple
// scalars. A []byte is rendered as an X'..' hexadecimal blob literal, a time.Time
// becomes a quoted 'YYYY-MM-DD HH:MM:SS.sss' timestamp in its own location, and a
// slice or array is rendered as a comma-joined list of its escaped elements. The
// slice form is exactly what an IN clause needs, and because every element is
// escaped recursively, mixing numbers, strings, and NULL in one list stays safe.
// This example prints one value of each kind.
func ExampleEscape_types() {
	fmt.Println(sqlstring.Escape([]byte{0xde, 0xad, 0xbe, 0xef}))
	fmt.Println(sqlstring.Escape(time.Date(2026, 7, 4, 13, 5, 9, 0, time.UTC)))
	fmt.Println(sqlstring.Escape([]any{1, "two", nil}))
	// Output:
	// X'deadbeef'
	// '2026-07-04 13:05:09.000'
	// 1, 'two', NULL
}

// ExampleEscapeID makes a table or column name safe to embed in a query. The
// identifier is wrapped in backticks, and any backtick already inside the name is
// doubled so a crafted name cannot close the quoting early and inject SQL. A
// dotted identifier is treated as a qualified name: it is split on the dot and
// each segment is quoted independently, so "table.col" becomes "`table`.`col`".
// This example shows a plain name, a name containing a backtick, and a qualified
// name. Identifier escaping is the counterpart to value escaping and is what the
// "??" placeholder uses.
func ExampleEscapeID() {
	fmt.Println(sqlstring.EscapeID("users"))
	fmt.Println(sqlstring.EscapeID("weird`name"))
	fmt.Println(sqlstring.EscapeID("db.users"))
	// Output:
	// `users`
	// `weird``name`
	// `db`.`users`
}

// ExampleFormat fills placeholders in a query template from left to right. A "?"
// placeholder is replaced with the Escape of the next argument and a "??"
// placeholder is replaced with the EscapeID of the next argument, so the table
// name is backtick-quoted while the values are single-quoted or rendered as
// numbers. Because substitution always escapes as it goes, the untrusted name
// "bob" is embedded as data and could not inject SQL even if it contained quote
// characters. The result is a complete, injection-safe statement ready to send to
// MySQL. This example builds a simple SELECT with one identifier and two values.
func ExampleFormat() {
	q := sqlstring.Format(
		"SELECT * FROM ?? WHERE id = ? AND name = ?",
		[]any{"users", 5, "bob"},
	)
	fmt.Println(q)
	// Output:
	// SELECT * FROM `users` WHERE id = 5 AND name = 'bob'
}

// ExampleFormat_leftovers documents how Format handles a mismatch between the
// number of placeholders and the number of arguments. When there are fewer
// arguments than placeholders, the surplus placeholders are left in the output
// untouched rather than causing a panic, as the "b=?" fragment shows. When the
// argument list is empty the template is returned verbatim. This forgiving
// behavior mirrors the npm original and makes partial formatting predictable.
// The example prints both cases so the contract is visible.
func ExampleFormat_leftovers() {
	fmt.Println(sqlstring.Format("a=? b=?", []any{1}))
	fmt.Println(sqlstring.Format("no placeholders", nil))
	// Output:
	// a=1 b=?
	// no placeholders
}
