// Package cspnonce provides middleware that generates a fresh, cryptographically
// random nonce for every request and emits a Content-Security-Policy header whose
// script-src directive includes that nonce, enabling specific inline scripts to
// be allow-listed safely. It complements the sibling csp package: where csp sets
// a static policy from a directive map, cspnonce specializes in the nonce
// workflow that Helmet's contentSecurityPolicy exposes through per-request
// directive functions in the Node ecosystem, producing a "'nonce-<value>'"
// source on each response.
//
// Use this middleware when a page must include inline <script> blocks but you
// still want the protection of a strict Content-Security-Policy. A nonce lets
// the browser execute exactly the inline scripts whose nonce attribute matches
// the one advertised in the header, while blocking any injected script that
// lacks it — the recommended modern alternative to "'unsafe-inline'". Because a
// new nonce is minted per request and never reused, an attacker cannot predict
// or replay it. Mount it with app.Use ahead of the handlers that render HTML, so
// each of those handlers can read the current nonce and stamp it onto the tags
// it emits.
//
// Operationally the middleware sits near the front of the chain. On each request
// it generates the nonce, stores it in two places for downstream handlers —
// req.Set("nonce", n) (retrievable with Nonce) and res.Locals["nonce"] (for
// templates) — writes a Content-Security-Policy header of the form
// "default-src <DefaultSrc>; script-src <ScriptSrc> 'nonce-<n>'", and then always
// calls next(). The ContextKey constant ("nonce") is the shared lookup key for
// both stores. No request headers are consulted; the middleware only reads the
// caller's Options and writes response state, so it is purely additive.
//
// Options controls the policy shape and nonce strength. DefaultSrc sets the
// default-src value (default "'self'"), ScriptSrc holds the static script-src
// sources placed alongside the generated nonce (default "'self'"), and Bytes is
// the number of random bytes drawn from crypto/rand before base64 encoding
// (default 16, i.e. 128 bits of entropy). The zero-value Options is usable and
// yields the defaults. If the system random source fails — which does not happen
// in practice — generate returns an empty string, and callers should treat a
// missing nonce (Nonce returning "") as a signal that the middleware did not run
// or could not produce one.
//
// Compared with Helmet's nonce recipe, this port keeps the essential contract —
// one unpredictable nonce per request, surfaced to templates, reflected in the
// header — but is intentionally narrow. It emits only default-src and script-src
// rather than a full directive set, does not compute style nonces or hashes, and
// leaves it entirely to the caller to render the matching nonce attribute onto
// each inline <script> tag; a nonce in the header with no matching attribute in
// the HTML simply causes the browser to block that script. Callers who need a
// broader static policy can layer the csp package alongside this one.
package cspnonce

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/malcolmston/express"
)

// ContextKey is the key under which the generated nonce is stored on the
// request and in res.Locals.
const ContextKey = "nonce"

// Options configures the CSP nonce middleware.
type Options struct {
	// DefaultSrc sets the default-src directive value (default "'self'").
	DefaultSrc string
	// ScriptSrc holds additional script-src sources placed alongside the
	// generated nonce (default "'self'").
	ScriptSrc string
	// Bytes is the number of random bytes used to build the nonce (default 16).
	Bytes int
}

func (o *Options) applyDefaults() {
	if o.DefaultSrc == "" {
		o.DefaultSrc = "'self'"
	}
	if o.ScriptSrc == "" {
		o.ScriptSrc = "'self'"
	}
	if o.Bytes <= 0 {
		o.Bytes = 16
	}
}

// New returns middleware that generates a nonce, exposes it via
// req.Set("nonce", n) and res.Locals["nonce"], and sets a
// Content-Security-Policy header of the form:
//
//	default-src <DefaultSrc>; script-src <ScriptSrc> 'nonce-<n>'
//
// Retrieve the nonce with Nonce.
func New(opts ...Options) express.Handler {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}
	o.applyDefaults()

	return func(req *express.Request, res *express.Response, next express.Next) {
		n := generate(o.Bytes)
		req.Set(ContextKey, n)
		res.Locals[ContextKey] = n

		policy := fmt.Sprintf("default-src %s; script-src %s 'nonce-%s'", o.DefaultSrc, o.ScriptSrc, n)
		res.Set("Content-Security-Policy", policy)
		next()
	}
}

// Nonce returns the nonce generated for the request, or "" if the middleware
// did not run.
func Nonce(req *express.Request) string {
	if v, ok := req.Value(ContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func generate(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}
