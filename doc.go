// Package express is a fast, minimalist web framework for Go, modeled after the
// Node.js Express framework. It is a stdlib-only port: everything is built on
// net/http and the standard library, with no third-party dependencies, so it
// drops cleanly into any Go project and interoperates with the wider net/http
// ecosystem. The goal is to give Go programmers the ergonomics that made
// Express popular — declarative routing, a chainable request/response API, and a
// composable middleware stack — while remaining idiomatic Go underneath.
//
// # Core model
//
// The core model mirrors Express's four building blocks: the Application, the
// Router, the Request, and the Response. express.New returns an *Application,
// which embeds a *Router; because of that embedding every routing method lives
// directly on the app — app.Use, app.Get, app.Post, app.All, and so on all work
// on the returned value, exactly as app.get / app.use do in JavaScript. An
// *Application also satisfies http.Handler through ServeHTTP, so it can be
// handed to http.ListenAndServe, wrapped by other net/http middleware, or
// driven directly in tests with net/http/httptest. The convenience Listen and
// ListenWithServer helpers exist for the common serving cases, and WrapHandler /
// WrapHandlerFunc adapt any existing net/http handler into an express.Handler.
//
// # Handlers
//
// A handler has the type express.Handler, defined as
// func(req *Request, res *Response, next Next). Every handler receives the
// wrapped request, the wrapped response, and a Next function used to pass
// control down the stack. Calling next() continues to the next matching layer;
// calling next(err) diverts to error-handling middleware, which is any function
// with the ErrorHandler shape func(err error, req *Request, res *Response, next Next).
// Request exposes Express-style accessors — Params for captured route
// parameters, Query for query-string values, Get for headers, Body and the
// Parse helpers for payloads — while Response offers chainable writers such as
// Status, Set, Send, JSON, Redirect, and Cookie. This keeps handler code terse
// and reads much like its JavaScript counterpart.
//
// # Routing
//
// Routing follows Express's path syntax. Patterns support named parameters
// (":id"), optional parameters (":id?"), custom regular-expression constraints
// (":id(\\d+)"), and wildcards ("*"), all compiled into standard library
// regexp under the hood. Every HTTP verb has a method (Get, Post, Put, Delete,
// Patch, Head, Options, All), including Query for the emerging QUERY method.
// Routers are first-class: a Router created with NewRouter can register its own
// middleware and routes and then be mounted on the app (or another router) with
// app.Use("/prefix", subRouter), enabling modular, prefix-scoped route trees.
// RouterOptions mirrors Express's case-sensitive, strict, and merge-params
// toggles, and app.Param registers parameter-preprocessing callbacks that run
// once per request before the matched route.
//
// # Views, streaming, and settings
//
// View rendering is included as well. app.Engine registers a template engine
// for a file extension, res.Render resolves a view against the configured
// "views" directories and "view engine" setting and streams the rendered HTML,
// and a cache-aware html/template engine is wired up by default for ".html" and
// ".tmpl" files. For dynamic output, res.Stream writes chunked responses and
// res.SSE returns an SSEWriter for Server-Sent Events, while res.SendFile,
// res.Download, and res.Attachment cover static and downloadable content with
// conditional-request support. Application settings (Set, GetSetting, Enable,
// Disable, Enabled, Disabled) and per-request Locals round out the Express
// feature set, giving handlers a place to stash configuration and
// request-scoped data.
//
// # Batteries: middleware and utility ports
//
// Beyond the core framework, this module ships 191 importable packages in
// total. Under middleware/ live 102 connect-style middleware — body parsing,
// CORS, compression, sessions, rate limiting, security headers (helmet, CSP,
// HSTS), CSRF, and more — each exposed as express.Handler values. Alongside
// them sit 88 standalone utility packages (81 top-level plus the seven
// lodash/* subpackages) that port popular npm libraries such as accepts,
// cookie, content-type, content-disposition, bytes, ms, qs, jsonwebtoken,
// uuid, nanoid, and many others as ordinary, dependency-free Go packages. Each
// utility package documents the npm package it corresponds to and the parity it
// targets, so an Express codebase can be translated to Go piece by piece. The
// aim throughout is Express.js parity: familiar names, familiar semantics, and
// behavior that matches the Node originals closely enough to port real
// applications with minimal surprise.
//
// # Example
//
// A minimal application looks like this:
//
//	app := express.New()
//
//	app.Get("/", func(req *express.Request, res *express.Response, next express.Next) {
//		res.Send("Hello World")
//	})
//
//	log.Fatal(app.Listen(":3000"))
package express
