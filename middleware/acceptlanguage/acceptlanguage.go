// Package acceptlanguage provides middleware that parses the Accept-Language
// request header, optionally negotiates it against a set of supported
// languages, and stores the chosen language tag on the request under the key
// "language". It is the Go analogue of Node content-negotiation helpers such
// as the accepts / negotiator packages and the accept-language module, exposed
// as a drop-in express.Handler that resolves a single best language up front
// so handlers need not reparse the header.
//
// Use this middleware whenever an application serves localized content and
// needs to know which language the client prefers. Rather than scattering
// header parsing and locale-matching logic across handlers, mount this once
// and read the resolved tag from the request. It is well suited to
// internationalized web pages, localized API messages, and any route that
// selects templates, currency, or copy by language, giving every downstream
// handler a consistent, already-negotiated answer.
//
// Operationally the middleware sits near the front of the chain, before any
// handler that reads the language. On each request it reads the
// Accept-Language request header via req.Get, computes the winning tag, stores
// it with req.Set(Key, lang) where Key is "language", and always calls next().
// It never writes a response or short-circuits, so it composes freely with
// other middleware; retrieve the value downstream with req.Value(Key). Header
// entries are parsed into tag/quality pairs and sorted by descending q-value,
// with original order preserved for ties, matching the precedence rules of RFC
// 7231 quality-value negotiation.
//
// Behavior depends on Options. When Supported is empty the middleware performs
// no negotiation and simply stores the client's highest-quality tag verbatim
// (or Default when the header is absent or empty). When Supported is non-empty
// it walks the client's preferences in quality order and stores the first
// supported tag that matches, where a match is exact, a wildcard "*", or a
// shared primary subtag (so a client asking for "en-US" matches a supported
// "en", case-insensitively). If no acceptable match is found, Options.Default
// is stored. Entries with q=0 are discarded as explicitly unacceptable, and a
// malformed q parameter falls back to the default quality of 1.0.
//
// Compared with the Node originals, this port keeps the essential q-value
// ordering and primary-subtag fallback but is intentionally compact: it does
// not implement the full BCP 47 lookup/filtering algorithm, does not weight or
// score partial region matches beyond the primary subtag, and negotiates only
// languages (not charset, encoding, or media types). It also differs from
// Express's res.acceptsLanguages by resolving eagerly and stashing one winner
// on the request instead of offering an on-demand negotiation call.
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
