// Package sqlstring provides MySQL-style value and identifier escaping and
// query formatting, mirroring the npm "sqlstring" library.
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
		return "'" + v.Format("2006-01-02 15:04:05") + "'"
	}

	// Handle slices and arrays (other than []byte) reflectively.
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		parts := make([]string, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			parts[i] = Escape(rv.Index(i).Interface())
		}
		return "(" + strings.Join(parts, ", ") + ")"
	case reflect.Ptr:
		if rv.IsNil() {
			return "NULL"
		}
		return Escape(rv.Elem().Interface())
	}

	return quoteString(fmt.Sprint(val))
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
// A "?" placeholder is replaced with the escaped value (via Escape) and a "??"
// placeholder is replaced with the escaped identifier (via EscapeID). Leftover
// args are ignored; leftover placeholders are left in place.
func Format(sql string, args []any) string {
	if len(args) == 0 {
		return sql
	}
	var b strings.Builder
	b.Grow(len(sql))
	argIdx := 0
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if c != '?' {
			b.WriteByte(c)
			continue
		}
		if argIdx >= len(args) {
			b.WriteByte(c)
			continue
		}
		if i+1 < len(sql) && sql[i+1] == '?' {
			// Identifier placeholder.
			if id, ok := args[argIdx].(string); ok {
				b.WriteString(EscapeID(id))
			} else {
				b.WriteString(EscapeID(fmt.Sprint(args[argIdx])))
			}
			argIdx++
			i++ // consume second '?'
			continue
		}
		b.WriteString(Escape(args[argIdx]))
		argIdx++
	}
	return b.String()
}
