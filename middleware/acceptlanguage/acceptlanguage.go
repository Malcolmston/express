// Package acceptlanguage provides middleware that parses the Accept-Language
// request header, optionally negotiates against a set of supported languages,
// and stores the chosen language tag on the request under the key "language".
package acceptlanguage

import (
	"sort"
	"strconv"
	"strings"

	"github.com/malcolmston/express"
)

// Key is the request value key under which the chosen language is stored.
const Key = "language"

// Options configures the acceptlanguage middleware.
type Options struct {
	// Supported is the set of language tags the application can serve. When
	// non-empty, the middleware negotiates and stores the best match; when
	// empty, it stores the client's highest-quality language verbatim.
	Supported []string

	// Default is stored when negotiation finds no acceptable match (or the
	// header is absent).
	Default string
}

type langQ struct {
	tag string
	q   float64
	pos int
}

// New returns middleware that stores the negotiated language via
// req.Set(Key, lang).
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	return func(req *express.Request, res *express.Response, next express.Next) {
		req.Set(Key, negotiate(req.Get("Accept-Language"), o.Supported, o.Default))
		next()
	}
}

// negotiate returns the chosen language tag given the header, supported set,
// and default.
func negotiate(header string, supported []string, def string) string {
	prefs := parse(header)

	if len(supported) == 0 {
		if len(prefs) > 0 {
			return prefs[0].tag
		}
		return def
	}

	for _, p := range prefs {
		for _, s := range supported {
			if matches(p.tag, s) {
				return s
			}
		}
	}
	return def
}

// matches reports whether an accepted tag matches a supported tag, either
// exactly or by primary subtag (case-insensitive).
func matches(accepted, supported string) bool {
	accepted = strings.ToLower(accepted)
	supported = strings.ToLower(supported)
	if accepted == supported || accepted == "*" {
		return true
	}
	ap := primary(accepted)
	return ap == primary(supported)
}

func primary(tag string) string {
	if i := strings.IndexByte(tag, '-'); i >= 0 {
		return tag[:i]
	}
	return tag
}

// parse returns the header's language tags sorted by descending quality,
// preserving original order for equal qualities.
func parse(header string) []langQ {
	if strings.TrimSpace(header) == "" {
		return nil
	}
	var out []langQ
	for i, part := range strings.Split(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := part
		q := 1.0
		if semi := strings.IndexByte(part, ';'); semi >= 0 {
			tag = strings.TrimSpace(part[:semi])
			params := part[semi+1:]
			for _, kv := range strings.Split(params, ";") {
				kv = strings.TrimSpace(kv)
				if strings.HasPrefix(kv, "q=") {
					if v, err := strconv.ParseFloat(strings.TrimSpace(kv[2:]), 64); err == nil {
						q = v
					}
				}
			}
		}
		if tag == "" || q <= 0 {
			continue
		}
		out = append(out, langQ{tag: tag, q: q, pos: i})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].q != out[j].q {
			return out[i].q > out[j].q
		}
		return out[i].pos < out[j].pos
	})
	return out
}
