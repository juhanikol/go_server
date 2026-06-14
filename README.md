# GoServer

GoServer is a small Go HTTP server skeleton and learning project. It is meant to show how a standard-library server can be split into reusable packages, example programs, runtime configuration, and embedded fallback assets.

This project has two intended use cases:

1. Developer use: reusable `pkg/` packages and small examples that show how to import and use the server from Go code.
2. End-user use: a simple executable + `server.json` + `web/templates` + `web/static` flow for serving a local HTML/CSS site.

This is not a production framework.

## Project Background

This project began as a personal learning exercise for understanding Go’s standard library, goroutines, HTTP server behavior, routing, middleware, and HTML template rendering. I intentionally avoided mature third-party routers and frameworks, such as chi, because the goal was to understand the underlying mechanics rather than hide them behind an existing abstraction.

The original idea was to create a lightweight, zero-boilerplate HTTP server that could be started easily from a built executable. One intended use case was a simple local server for beginners or non-specialist users who want to demonstrate HTML/CSS pages without first learning a full web framework or deployment stack. A secondary goal was to make the server small enough to run in a home network or on low-powered single-board computers such as a Raspberry Pi.

As the project grew, it became clear that the codebase needed better boundaries. Several features had been started at the same time, including configuration loading, embedded fallback pages, manifest-based route registration, domain-aware routing, logging, and graceful shutdown. The server had already been running successfully with test pages, but later changes introduced incomplete paths and made the startup flow harder to reason about. I paused the project for a while rather than continue adding features on top of unclear structure.

The current goal is to finish the project as a clean portfolio-quality Go server skeleton. The focus is now on making the example application reliable, keeping the public API small, documenting the actual behavior, and adding tests around the most important server paths.

This cleanup is also being done as a practical exercise in AI-assisted development workflows. I am using Codex together with Headroom CLI to analyze the repository, plan small commits, and improve the codebase without turning the process into a broad uncontrolled rewrite.

## What It Demonstrates

- route registration with the standard library
- middleware for method checks, logging, and panic recovery
- structured logging
- graceful shutdown
- runtime config loading from JSON
- embedded fallback pages and embedded static assets
- serving a real local HTML/CSS page from `web/templates` and `web/static`
- a reusable import example under `examples/minimal`
- a folder-based local site example under `examples/html_site`

## Current Status

- `go run ./cmd/example_server` works from the repository root.
- `/health` returns `OK`.
- `/` serves the example page when `web/templates/index.html` exists.
- embedded fallback pages remain in place as safety nets.
- `/__go_server/static/styles.css` serves the embedded stylesheet.
- `DomainMap` is not part of the current single-project flow.
- `AllowedHosts` is opt-in: empty or nil allows all hosts, configured hosts restrict requests.
- TLS and multi-domain routing are future work.

## Repository Structure

```text
.
├── cmd/
│   └── example_server/
│       ├── main.go
│       └── myproject/
│           └── myproject.go
├── examples/
│   ├── html_site/
│   │   ├── README.md
│   │   ├── main.go
│   │   └── site/
│   │       ├── index.html
│   │       └── styles.css
│   └── minimal/
│       └── main.go
├── pkg/
│   ├── httpserver/
│   │   ├── assets/
│   │   │   ├── README.md
│   │   │   ├── servererror.html
│   │   │   ├── serverindex.html
│   │   │   └── styles.css
│   │   ├── embed.go
│   │   ├── errors.go
│   │   ├── middleware.go
│   │   ├── routes.go
│   │   ├── server.go
│   │   └── server_test.go
│   └── serverapp/
│       ├── config.go
│       ├── serverapp.go
│       └── serverapp_test.go
├── server.json
├── web/
│   ├── static/
│   │   └── styles.css
│   └── templates/
│       └── index.html
├── go.mod
├── README.md
└── AGENTS.md
```

## Quick Start

From the repository root:

```bash
go run ./cmd/example_server
```

Then open:

```text
http://localhost:8081/
```

## Manual Checks

```bash
curl -i http://localhost:8081/
curl -i http://localhost:8081/health
curl -i http://localhost:8081/__go_server/static/styles.css
```

## Development Commands

```bash
go fmt ./...
go test ./...
go vet ./...
go run ./cmd/example_server
```

## Examples

### `examples/minimal`

Demonstrates importing `pkg/httpserver` from another Go program and registering a route directly in code.

Run:

```bash
go run ./examples/minimal
```

Open:

```text
http://localhost:8082/hello
```

### `examples/html_site`

Demonstrates the beginner-friendly local site flow: point the example at a folder of HTML/CSS files and run it locally.

Run:

```bash
go run ./examples/html_site
```

Open:

```text
http://localhost:8083/
```

## Configuration Overview

The repository includes a sample `server.json` at the root. It documents the current runtime knobs used by the example server and future executable-based deployments.

Common fields include:

- `server_address`
- `root_path`
- `template_dir`
- `static_dir`
- `allowed_hosts`
- `read_timeout_sec`
- `read_header_timeout_sec`
- `write_timeout_sec`
- `idle_timeout_sec`
- `shutdown_timeout_sec`
- `log_file_name`
- `log_level`

The current `cmd/example_server/main.go` also sets matching defaults directly so the documented `go run ./cmd/example_server` command works from the repository root.

## Why There Are Two README Files

This root `README.md` is for GitHub visitors, contributors, reviewers, and portfolio readers. It can mention examples, package boundaries, config shape, and project status.

`pkg/httpserver/assets/README.md` is different: it is embedded into the built server and should stay short, practical, and user-facing. It is meant to help someone who already has the executable running.

## Known Limitations

- `DomainMap` is future work and not part of the current single-project request flow.
- TLS support is not a completed end-user feature yet.
- `AllowedHosts` is opt-in rather than a default restriction.
- The project is still a learning skeleton, not a hardened production framework.
