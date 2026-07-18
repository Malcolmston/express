# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-07-18
### Added
- **`app.Docs()` — automatic API documentation.** A single call introspects the
  application's registered routes (including mounted sub-routers) and serves a
  generated OpenAPI 3.1 specification together with interactive Swagger UI and
  ReDoc pages, a YAML rendering, an AsyncAPI 2.6 document for socket/event
  channels, and an importable Postman v2.1 collection.
- `app.Routes()` / `router.Routes()` return the discovered routes (`RouteInfo`)
  with method, path and path parameters, de-duplicated and sorted.
- `app.Describe(method, path, RouteDoc{...})` enriches a route's generated
  operation with a summary, description, tags, parameters, request body and
  responses that introspection alone cannot infer.
- `app.Channel(name, ChannelDoc{...})` documents an event/socket channel
  (Socket.IO event, WebSocket topic, queue subject) for the AsyncAPI document.
- `app.OpenAPI()`, `app.OpenAPIYAML()`, `app.AsyncAPI()` and
  `app.PostmanCollection()` return the specifications directly for custom
  serving. `DocsOptions` configures titles, servers, endpoint paths (set any to
  `"-"` to disable), the UI asset base URL, and an `Enrich` hook for
  programmatic customisation of every operation.

## [0.1.0] - 2026-07-04
### Added
- Initial public release — a faithful, dependency-light Go port of Express.
- Routing with the classic `(req, res, next)` handler shape; path params
  (`:id`, optional `:id?`, regex `:id(\d+)`, wildcard `*`); mountable routers
  with `CaseSensitive` / `Strict` / `MergeParams`.
- Views via `html/template` (`res.Render`, `res.SendFile`), content
  negotiation, Server-Sent Events and chunked streaming, and the `QUERY` method.
- 100+ middleware and utility ports: `ms`, `bytes`, `cookie`, `qs`,
  `jsonwebtoken`, `uuid`, `lodash/*`, date helpers, and more.
- `express.WrapHandler` to mount any `net/http` handler.
- `RouteRegistrar` interface documenting the shared router/application surface.
- `TestEveryPackageShipsTests` guard — every shipped package must carry tests.
- Automated releases (VERSION-driven tags + GitHub Releases, moving `stable` tag).
- CI: build/test matrix (Go 1.23 & 1.24), `-race` + coverage, golangci-lint v2,
  govulncheck, CodeQL, benchmarks, dependency review, and a stale bot.

[Unreleased]: https://github.com/malcolmston/express/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/malcolmston/express/releases/tag/v0.1.0
