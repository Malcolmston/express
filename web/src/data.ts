// Content for the Express-for-Go documentation site. The `Lib` shape mirrors the
// shared `Lib` interface used across every malcolmston/go library site; this
// repo carries only its own entry so the site never depends on go/data.ts at
// runtime.
export interface Lib {
  id: string; name: string; icon: string; accent: string; pkg: string; node: string;
  repo: string; docs: string; tagline: string; blurb: string; tags: string[];
  features: string[]; node_code: string; go_code: string; integrate: string;
}

export const NODE_ACCENT = '#8cc84b';

export const EXPRESS: Lib = {
  id: "express", name: "Express", icon: '<i class="fa-solid fa-route"></i>', accent: "#00add8",
  pkg: "github.com/malcolmston/express", node: "expressjs/express",
  repo: "https://github.com/malcolmston/express", docs: "https://malcolmston.github.io/express/",
  tagline: "Fast, unopinionated, minimalist web framework.",
  blurb: "Routing, middleware chains, views, content negotiation and streaming — the Express you know, in Go. " +
    "The same three-argument handler signature, path patterns (:param, :id?, :id(\\\\d+), *), and mountable routers.",
  tags: ["routing", "middleware", "views", "SSE / streaming", "QUERY method", "190+ packages"],
  features: [
    "Handlers with the classic <code>(req, res, next)</code> shape",
    "Path-to-regexp params: <code>:id</code>, optional <code>:id?</code>, regex <code>:id(\\d+)</code>, wildcard <code>*</code>",
    "Routers with <code>CaseSensitive</code> / <code>Strict</code> / <code>MergeParams</code> options",
    "Views via <code>html/template</code>, <code>res.Render</code> / <code>res.SendFile</code>",
    "Content negotiation, SSE &amp; chunked streaming, the new <code>QUERY</code> HTTP method",
    "Batteries: 100+ middleware + utility ports (<code>ms</code>, <code>bytes</code>, <code>cookie</code>, <code>qs</code>, <code>jsonwebtoken</code>, <code>uuid</code>, <code>lodash/*</code> …)",
    "<code>express.WrapHandler</code> mounts any <code>net/http</code> handler"
  ],
  node_code:
`const express = require('express')
const app = express()

app.get('/users/:id', (req, res) => {
  res.json({ id: req.params.id })
})

app.listen(3000)`,
  go_code:
`app := express.New()

app.Get("/users/:id", func(req *express.Request,
    res *express.Response, next express.Next) {
    res.JSON(map[string]any{"id": req.Params("id")})
})

app.Listen(":3000")`,
  integrate:
`<span class="tok-c">// Mount a router, add middleware, stream Server-Sent Events</span>
api := express.NewRouter(express.RouterOptions{MergeParams: true})
api.Use(func(req *express.Request, res *express.Response, next express.Next) {
    res.Set("X-App", "demo"); next()
})
api.Get("/events", func(req *express.Request, res *express.Response, next express.Next) {
    sse := res.SSE()                 <span class="tok-c">// sets SSE headers + flushes</span>
    sse.Send("tick", "hello")        <span class="tok-c">// event: tick\\ndata: hello</span>
})
app.Use("/api", api)`
};

// The single library this site documents, in the shape ReleaseList/VersionBadge
// expect (repo is the bare GitHub repo name, not a URL).
export const RELEASE_LIB = {
  name: EXPRESS.name,
  icon: EXPRESS.icon,
  accent: EXPRESS.accent,
  repo: 'express',
  url: 'https://github.com/malcolmston/express',
};
