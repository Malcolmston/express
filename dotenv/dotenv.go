// Package dotenv parses .env style configuration content and loads it into the
// process environment, mirroring the behavior of the npm "dotenv" library.
package dotenv

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

// Parse reads .env formatted content from r and returns the parsed key/value
// pairs. It understands blank lines, full-line and trailing "#" comments,
// an optional "export " prefix, single-quoted values (kept literally),
// double-quoted values (with \n and \t escape expansion) and unquoted values
// (which are trimmed and may carry a trailing comment).
func Parse(r io.Reader) (map[string]string, error) {
	out := make(map[string]string)
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		key, val, ok := parseLine(line)
		if !ok {
			continue
		}
		out[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// Bytes parses .env formatted content from a byte slice.
func Bytes(b []byte) (map[string]string, error) {
	return Parse(bytes.NewReader(b))
}

// parseLine parses a single .env line. It returns the key, the resolved value
// and whether the line held an assignment.
func parseLine(line string) (string, string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", "", false
	}
	// Strip an optional "export " prefix.
	if strings.HasPrefix(trimmed, "export ") || strings.HasPrefix(trimmed, "export\t") {
		trimmed = strings.TrimSpace(trimmed[len("export"):])
	}
	eq := strings.IndexByte(trimmed, '=')
	if eq < 0 {
		return "", "", false
	}
	key := strings.TrimSpace(trimmed[:eq])
	if key == "" {
		return "", "", false
	}
	raw := strings.TrimSpace(trimmed[eq+1:])
	return key, parseValue(raw), true
}

// parseValue interprets the right-hand side of an assignment.
func parseValue(raw string) string {
	if raw == "" {
		return ""
	}
	switch raw[0] {
	case '\'':
		// Single-quoted: literal, find the closing quote.
		if end := strings.IndexByte(raw[1:], '\''); end >= 0 {
			return raw[1 : 1+end]
		}
		return raw[1:]
	case '"':
		// Double-quoted: expand escapes, find the closing quote.
		if end := strings.IndexByte(raw[1:], '"'); end >= 0 {
			return expandEscapes(raw[1 : 1+end])
		}
		return expandEscapes(raw[1:])
	default:
		// Unquoted: strip trailing comment, then trim.
		if hash := strings.IndexByte(raw, '#'); hash >= 0 {
			raw = raw[:hash]
		}
		return strings.TrimSpace(raw)
	}
}

// expandEscapes expands the backslash escapes that dotenv honors inside
// double-quoted values.
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
