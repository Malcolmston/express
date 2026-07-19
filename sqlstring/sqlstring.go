// Package sqlstring provides MySQL-style value and identifier escaping and
// query formatting. It is a Go port of the npm "sqlstring" library (the escaping
// engine used by the popular mysqljs/mysql driver), reimplemented with only the
// Go standard library. Its purpose is to let callers build MySQL statements from
// dynamic data without opening the door to SQL injection.
//
// The package is used whenever an application composes SQL text that embeds
// user-supplied or otherwise untrusted values and cannot (or does not wish to)
// rely solely on the database driver's own parameter binding. Rather than
// concatenating raw values into a query, callers pass those values through
// Escape (for data) or EscapeID (for identifiers), or let Format do both at once
// while filling placeholders. The escaped output is safe to splice into a query
// string because every value is turned into a self-contained, correctly quoted
// SQL literal.
//
// Escape converts a Go value into a SQL literal. NULL is produced for nil (and
// nil pointers); booleans render as true/false; every integer and floating-point
// type renders as its numeric text; strings are single-quoted with dangerous
// characters escaped; []byte becomes an X'..' hex blob literal; time.Time becomes
// a quoted 'YYYY-MM-DD HH:MM:SS.sss' timestamp (with three-digit milliseconds);
// and slices or arrays are rendered as a comma-joined list of their escaped
// elements (ideal for IN clauses). Following the upstream library, a top-level
// list is not parenthesized, but any nested slice/array element is wrapped in
// parentheses so multi-row lists render as "(a, b), (c, d)". Pointers are
// dereferenced and any other type is formatted with fmt
// and single-quoted. The critical step is quoteString, which wraps a string in
// single quotes and backslash-escapes the characters MySQL treats specially
// inside a string literal: NUL, backspace, tab, newline, carriage return,
// Ctrl-Z, the double quote, the single quote, and the backslash itself. Because
// an attacker's quote characters are neutralized this way, injected text can
// never break out of the literal and be interpreted as SQL syntax.
//
// EscapeID makes an identifier (a table or column name) safe by wrapping it in
// backticks and doubling any embedded backtick, so a malicious name cannot close
// the quoting early. A dotted identifier such as "table.col" is split on the dot
// and each segment is quoted independently, yielding "`table`.`col`" so that
// qualified names keep working. This is the identifier counterpart to Escape's
// value quoting and is what the "??" placeholder uses.
//
// Format performs placeholder substitution left to right over runs of "?". A
// single "?" is replaced with the Escape of the next argument, and a double "??"
// is replaced with the EscapeID of the next argument (coercing non-strings via
// fmt first). A run of three or more "?" is left literal and consumes no
// argument. Arguments are consumed in order; if there are more placeholders than
// arguments the surplus placeholders are left untouched, and if there are more
// arguments than placeholders the surplus arguments are ignored. Because values
// and identifiers are always escaped as they are substituted, a query built with
// Format is injection-safe by construction: untrusted input becomes data or a
// quoted identifier, never executable SQL. The behavior tracks the npm original;
// the API is adapted to Go by taking a []any of arguments and returning a string
// instead of accepting a variadic JavaScript array.
package sqlstring

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Escape returns a safely-quoted SQL literal for val suitable for MySQL. It
// handles nil (NULL), booleans, numeric types, strings, []byte, time.Time and
// slices/arrays (rendered as a parenthesized, comma-joined list). Unknown types
// are formatted with fmt and single-quoted.
func Escape(val any) string {
	if val == nil {
		return "NULL"
	}
	switch v := val.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case string:
		return quoteString(v)
	case []byte:
		return "X'" + hex.EncodeToString(v) + "'"
	case time.Time:
		return "'" + v.Format("2006-01-02 15:04:05.000") + "'"
	}

	// Handle slices and arrays (other than []byte) reflectively.
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		return arrayToList(rv)
	case reflect.Ptr:
		if rv.IsNil() {
			return "NULL"
		}
		return Escape(rv.Elem().Interface())
	}

	return quoteString(fmt.Sprint(val))
}

// arrayToList renders a slice or array as a comma-separated list of escaped
// elements. Mirroring the upstream library's arrayToList, the top-level list is
// not wrapped in parentheses, but any element that is itself a slice or array
// (other than a []byte blob, which is a hex literal) is wrapped, so nested lists
// render as grouped "(a, b), (c, d)" output suitable for multi-row VALUES.
func arrayToList(rv reflect.Value) string {
	parts := make([]string, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i).Interface()
		if ev := reflect.ValueOf(elem); ev.IsValid() {
			switch ev.Kind() {
			case reflect.Slice, reflect.Array:
				if _, isBytes := elem.([]byte); !isBytes {
					parts[i] = "(" + arrayToList(ev) + ")"
					continue
				}
			}
		}
		parts[i] = Escape(elem)
	}
	return strings.Join(parts, ", ")
}

// quoteString single-quotes s and escapes the characters MySQL requires to be
// escaped inside string literals.
func quoteString(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('\'')
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case 0:
			b.WriteString(`\0`)
		case '\b':
			b.WriteString(`\b`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case 0x1a:
			b.WriteString(`\Z`)
		case '"':
			b.WriteString(`\"`)
		case '\'':
			b.WriteString(`\'`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteByte(c)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

// EscapeID wraps an identifier in backticks, doubling any embedded backticks so
// the result is a single safe MySQL identifier. A dotted identifier such as
// "table.col" is quoted per-segment: "`table`.`col`".
func EscapeID(id string) string {
	parts := strings.Split(id, ".")
	for i, p := range parts {
		parts[i] = "`" + strings.ReplaceAll(p, "`", "``") + "`"
	}
	return strings.Join(parts, ".")
}

// Format substitutes placeholders in sql left-to-right with values from args.
// It matches runs of consecutive "?" characters: a single "?" is replaced with
// the escaped value (via Escape) and a double "??" is replaced with the escaped
// identifier (via escapeIDArg). A run of three or more "?" is left untouched and
// consumes no argument, mirroring the upstream library. Substitution stops once
// the arguments are exhausted, so surplus placeholders are left in place; surplus
// arguments are ignored.
func Format(sql string, args []any) string {
	if len(args) == 0 {
		return sql
	}
	var b strings.Builder
	b.Grow(len(sql))
	chunkStart := 0
	valuesIndex := 0
	i := 0
	for i < len(sql) && valuesIndex < len(args) {
		if sql[i] != '?' {
			i++
			continue
		}
		// Measure the maximal run of '?' starting at i.
		j := i
		for j < len(sql) && sql[j] == '?' {
			j++
		}
		runLen := j - i
		if runLen > 2 {
			// Runs longer than "??" are left literal and consume no argument.
			i = j
			continue
		}
		var value string
		if runLen == 2 {
			value = escapeIDArg(args[valuesIndex])
		} else {
			value = Escape(args[valuesIndex])
		}
		b.WriteString(sql[chunkStart:i])
		b.WriteString(value)
		chunkStart = j
		valuesIndex++
		i = j
	}
	if chunkStart == 0 {
		return sql
	}
	if chunkStart < len(sql) {
		b.WriteString(sql[chunkStart:])
	}
	return b.String()
}

// escapeIDArg escapes a "??" placeholder argument as an identifier. A string is
// passed straight to EscapeID; any other value is stringified via fmt first,
// matching the upstream behavior of coercing the identifier operand to text.
func escapeIDArg(v any) string {
	if s, ok := v.(string); ok {
		return EscapeID(s)
	}
	return EscapeID(fmt.Sprint(v))
}
