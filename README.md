# GoServer

GoServer is a small modular Go HTTP server skeleton and learning project. It is meant to demonstrate how a simple standard-library web server can be organized into reusable packages, a runnable example app, and embedded fallback assets.

This is not a production framework. Some fields and ideas are intentionally still in progress.

## Project background

This project began as a personal learning exercise for understanding Go’s standard library, goroutines, HTTP server behavior, routing, middleware, and HTML template rendering. I intentionally avoided mature third-party routers and frameworks, such as chi, because the goal was to understand the underlying mechanics rather than hide them behind an existing abstraction.

The original idea was to create a lightweight, zero-boilerplate HTTP server that could be started easily from a built executable. One intended use case was a simple local server for beginners or non-specialist users who want to demonstrate HTML/CSS pages without first learning a full web framework or deployment stack. A secondary goal was to make the server small enough to run in a home network or on low-powered single-board computers such as a Raspberry Pi.

As the project grew, it became clear that the codebase needed better boundaries. Several features had been started at the same time, including configuration loading, embedded fallback pages, manifest-based route registration, domain-aware routing, logging, and graceful shutdown. The server had already been running successfully with test pages, but later changes introduced incomplete paths and made the startup flow harder to reason about. I paused the project for a while rather than continue adding features on top of unclear structure.

The current goal is to finish the project as a clean portfolio-quality Go server skeleton. The focus is now on making the example application reliable, keeping the public API small, documenting the actual behavior, and adding tests around the most important server paths.

This cleanup is also being done as a practical exercise in AI-assisted development workflows. I am using Codex together with Headroom CLI to analyze the repository, plan small commits, and improve the codebase without turning the process into a broad uncontrolled rewrite.

## What It Demonstrates

- Building an HTTP server using Go’s standard library
- Registering application routes through a small server abstraction
- Using middleware for logging, method checks, and panic recovery
- Rendering HTML templates
- Embedding fallback pages and static assets into the compiled binary
- Loading runtime configuration from JSON
- Handling graceful shutdown
- Keeping a runnable example application inside the repository
Using AI-assisted development in small, reviewable commits
- Graceful shutdown on interrupt or `SIGTERM`.
- A runnable example server under `cmd/example_server`.

## Repository Structure

```text
.
├── AGENTS.md
├── AGENTS_comds.md
├── README.md
├── go.mod
├── cmd/
│   └── example_server/
│       ├── main.go
│       ├── server.json
│       └── myproject/
│           └── myproject.go
├── examples/
│   └── api/
│       └── apicaller_examples.txt
└── pkg/
    ├── httpserver/
    │   ├── server.go
    │   ├── routes.go
    │   ├── middleware.go
    │   ├── errors.go
    │   ├── embed.go
    │   ├── server_test.go
    │   └── assets/
    │       ├── README.md
    │       ├── serverindex.html
    │       ├── servererror.html
    │       └── styles.css
    ├── logging/
    │   └── logger.go
    └── serverapp/
        ├── config.go
        ├── serverapp.go
        └── serverapp_test.go
```

## Quick Start

From the repository root:

```bash
go run ./cmd/example_server
```

The example server listens on `:8081`.

## Manual Checks

In another terminal:

```bash
curl -i http://localhost:8081/
curl -i http://localhost:8081/health
curl -i http://localhost:8081/__go_server/static/styles.css
```

Expected current behavior:

- `/` returns the embedded GoServer fallback landing page if no project homepage template exists.
- `/health` returns `OK`.
- `/__go_server/static/styles.css` returns the embedded stylesheet.

## Development Commands

```bash
go fmt ./...
go test ./...
go vet ./...
go run ./cmd/example_server
```

## Configuration

The example app provides required startup defaults in `cmd/example_server/main.go`. The active config merge code also looks for optional config files relative to the process working directory:

- `server.json`
- `web/server.json`
- `config/server.json`

The example config file is at `cmd/example_server/server.json`. It is useful as a sample, but the documented root command currently relies on explicit defaults in `main.go`.

## Known Limitations

- Multi-domain routing through `DomainMap` is future work. The current flow serves a single project through the registered router.
- TLS, `AllowedHosts`, and related security fields may be declared but are not fully implemented yet.
- Project template and static directories are not included in this repository, so the example may log that `web/templates` is missing and then use embedded fallback pages.
- This project is still a learning skeleton, not a hardened production framework.

## Current status

This project is not intended to be a production-ready web framework. It is a learning and portfolio project that demonstrates how a small Go HTTP server can be structured from few principles.

The current cleanup work focuses on:

- making the example server start reliably from the documented command,
- clarifying configuration behavior,
- restoring request routing through the registered router,
separating future multi-domain support from the current single-project server flow,
- updating the documentation to match the real implementation,
- and adding tests for the most important runtime behavior.