// Package dotenv parses .env style configuration content and loads it into the
// process environment, mirroring the behavior of the npm "dotenv" library
// (motdotla/dotenv). It exposes Parse and Bytes for turning raw .env content
// into a map of key/value pairs, and Load and LoadOverride for reading a file
// from disk and applying it to the running process via os.Setenv.
//
// The classic use case is keeping configuration out of source code: secrets,
// connection strings, feature flags, and per-environment tunables live in a
// local .env file that is not committed, and the application reads them at
// startup. Parse and Bytes are also handy on their own when you have .env
// formatted text from some other source (an embedded asset, a network response,
// a test fixture) and want the parsed values without touching the environment.
//
// Parsing matches the upstream single regular-expression engine rather than a
// line scanner, so it handles the same shapes the Node original does. Line
// endings are first normalized: a lone "\r" and a "\r\n" pair both become "\n".
// A key is a run of word characters, dots, and dashes; it may carry a leading
// "export " prefix so shell-sourceable files work, and it is separated from its
// value by "=" (with optional surrounding spaces) or by ": ". Blank lines and
// lines whose value region is a "#" comment contribute nothing.
//
// Value handling follows dotenv's conventions. A value may be single-quoted,
// double-quoted, or backtick-quoted, and the surrounding quotes are stripped in
// every case; a quoted value may span multiple physical lines, which makes the
// captured newlines part of the value (PEM blocks, multi-line strings). Only a
// double-quoted value has escape sequences expanded: \n and \r become their
// control characters (and this port additionally honors \t, \\, and \" as a
// superset that upstream never exercises). Single-quoted and backtick-quoted
// values are literal, so backslashes and dollar signs survive verbatim. An
// unquoted value runs up to the first unquoted "#", after which the remainder is
// treated as a trailing comment, and the value is then whitespace-trimmed. An
// empty right-hand side yields an empty string.
//
// Load and LoadOverride differ only in precedence. Load never overwrites a
// variable that is already present in the environment, matching dotenv's default
// of treating the real environment as authoritative; LoadOverride writes every
// parsed variable unconditionally. Both return any error from opening or reading
// the file, or from os.Setenv, and a missing file surfaces the underlying
// os.Open error. Relative to the Node original this port implements the same
// comment, quote, escape, multi-line, and export handling, but it does not
// perform ${VAR} variable interpolation (that lives in dotenv-expand upstream
// too) and it returns an ordinary Go map and error instead of dotenv's
// { parsed, error } result object.
package dotenv

import (
	"io"
	"os"
	"regexp"
	"strings"
)

// bt is a single backtick, spelled out so the assignment regex below (which
// itself matches backtick-quoted values) can be written as a Go raw string.
const bt = "`"

// lineRE mirrors the upstream dotenv LINE regular expression. Each match binds
// group 1 to the key and group 2 to the raw (still-quoted) value region. The
// value alternation tries single-, double-, and backtick-quoted forms (each of
// which may span newlines) before falling back to an unquoted run that stops at
// the first '#'. Source: motdotla/dotenv lib/main.js.
var lineRE = regexp.MustCompile(
	`(?m)^\s*(?:export\s+)?([\w.-]+)(?:\s*=\s*?|:\s+?)` +
		`(\s*'(?:\\'|[^'])*'` +
		`|\s*"(?:\\"|[^"])*"` +
		`|\s*` + bt + `(?:\\` + bt + `|[^` + bt + `])*` + bt +
		`|[^#\r\n]+)?\s*(?:#.*)?$`)

// Parse reads .env formatted content from r and returns the parsed key/value
// pairs. It understands blank lines, full-line and trailing "#" comments, an
// optional "export " prefix, single-, double-, and backtick-quoted values
// (including multi-line quoted values), double-quoted escape expansion, and
// unquoted values (which are trimmed and may carry a trailing comment).
func Parse(r io.Reader) (map[string]string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return parseContent(data), nil
}

// Bytes parses .env formatted content from a byte slice.
func Bytes(b []byte) (map[string]string, error) {
	return parseContent(b), nil
}

// parseContent runs the upstream regex over the whole (newline-normalized)
// content and reduces each match to a key/value pair, with later assignments to
// the same key overwriting earlier ones.
func parseContent(data []byte) map[string]string {
	out := make(map[string]string)

	content := string(data)
	// Normalize line endings the way upstream does: \r\n and lone \r -> \n.
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	for _, m := range lineRE.FindAllStringSubmatch(content, -1) {
		key := m[1]
		out[key] = parseValue(m[2])
	}
	return out
}

// parseValue resolves the raw value region captured by lineRE: it trims
// surrounding whitespace, strips one pair of matching surrounding quotes, and
// (only for double-quoted values) expands backslash escapes.
func parseValue(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	quote := value[0]
	if len(value) >= 2 && (quote == '\'' || quote == '"' || quote == bt[0]) && value[len(value)-1] == quote {
		value = value[1 : len(value)-1]
	}
	if quote == '"' {
		value = expandEscapes(value)
	}
	return value
}

// expandEscapes expands the backslash escapes that dotenv honors inside
// double-quoted values. Upstream expands \n and \r; this port also honors \t,
// \\, and \" as a superset, which none of upstream's fixtures exercise.
func expandEscapes(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i+1])
			}
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// Load reads the .env file at path and sets each variable into the process
// environment. Existing environment variables are NOT overridden, matching the
// default dotenv behavior.
func Load(path string) error {
	return load(path, false)
}

// LoadOverride reads the .env file at path and sets each variable into the
// process environment, overriding any variables that are already set.
func LoadOverride(path string) error {
	return load(path, true)
}

func load(path string, override bool) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	vars, err := Parse(f)
	if err != nil {
		return err
	}
	for k, v := range vars {
		if !override {
			if _, exists := os.LookupEnv(k); exists {
				continue
			}
		}
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}
