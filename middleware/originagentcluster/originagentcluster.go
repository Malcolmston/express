// Package originagentcluster provides middleware that sets the
// Origin-Agent-Cluster response header on every response, requesting that the
// browser place the origin into its own, origin-keyed agent cluster. It is a
// direct port of Helmet's originAgentCluster middleware, which emits the same
// single, fixed header value and takes no configuration.
//
// The Origin-Agent-Cluster header is a hint to the user agent that documents
// from this origin should be isolated from other origins that happen to share
// the same site (scheme plus registrable domain). When honored, the browser
// keys the agent cluster by origin rather than by site, which prevents
// same-site pages from reaching into this origin's synchronous scripting
// context via mechanisms such as document.domain and can give the origin its
// own dedicated process. Use it as a defense-in-depth hardening measure on
// applications that never rely on cross-origin, same-site DOM access.
//
// Mechanically the middleware is trivial and stateless: for each request it
// sets Origin-Agent-Cluster to the structured-header boolean "?1" (meaning
// true) and then calls next to continue the chain. It never inspects the
// request, never short-circuits, and never writes a body, so it composes
// cleanly with any other handlers. Register it early — for example via app.Use
// before your routes — so the header is present on every response, including
// error responses produced downstream, as long as it runs before the headers
// are flushed.
//
// There are no options and no defaults to reason about. The value "?1" is fixed
// because the header is a boolean serialized in HTTP structured-field syntax;
// the only alternative, "?0" (false), is already the browser default and is not
// worth emitting. Note that browsers may cache the isolation decision for the
// lifetime of a site's process, so toggling the header on and off across
// responses is discouraged. The header has no effect on non-supporting browsers
// and is safe to send unconditionally.
//
// Parity with the Node original is exact for the observable behavior: Helmet's
// originAgentCluster also sets Origin-Agent-Cluster: ?1 and accepts no
// configuration. This port does not replicate Helmet's surrounding
// option-validation plumbing because there are no options to validate; the
// resulting header and value on every response are identical.
package originagentcluster

import "github.com/malcolmston/express"

// New returns middleware that sets Origin-Agent-Cluster: ?1.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Origin-Agent-Cluster", "?1")
		next()
	}
}
