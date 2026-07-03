// Package originagentcluster provides middleware that sets the
// Origin-Agent-Cluster response header, requesting that the browser isolate the
// origin into its own agent cluster.
package originagentcluster

import "github.com/malcolmston/express"

// New returns middleware that sets Origin-Agent-Cluster: ?1.
func New() express.Handler {
	return func(req *express.Request, res *express.Response, next express.Next) {
		res.Set("Origin-Agent-Cluster", "?1")
		next()
	}
}
