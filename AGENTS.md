# AGENTS.md

## Project goal

This repository is a small modular Go HTTP server skeleton intended as a portfolio-quality learning project. The goal is to demonstrate clean Go structure, route registration, middleware composition, structured logging, graceful shutdown, embedded fallback templates, and testable server behavior. Goal is to rely on entirely on GO's standard library packages.

## Current priorities

Work in small, reviewable steps. Do not perform broad rewrites unless explicitly asked.

The main cleanup goals are:

1. Make the example server start successfully.
2. Fix request routing so registered routes are actually served.
3. Fix configuration loading so JSON parse errors are returned clearly and not silently ignored.
4. Lower the Go version in go.mod to a realistic stable version unless newer language features are required.
5. Add tests for routing, method handling, panic recovery, config parsing, and error rendering.
6. Align README with the actual repository structure and current capabilities.
7. Remove or clearly mark unfinished features such as DomainMap, AllowedHosts, UseTLS, and unused timeout fields.

## Repository layout

- `cmd/example_server/` contains the runnable example server.
- `cmd/example_server/server.json` contains example configuration.
- `cmd/example_server/myproject/` contains the example project manifest.
- `pkg/httpserver/` contains server, routing, middleware, embedded assets, and error rendering.
- `pkg/serverapp/` contains application configuration and startup orchestration.
- `pkg/logging/` contains logger setup.
- `examples/` contains examples or notes.

## Commands

Run these from the repository root:


go fmt ./...
go test ./...
go vet ./...
go run ./cmd/example_server

The example server should listen on :8081.

Useful manual checks:

curl -i http://localhost:8081/
curl -i http://localhost:8081/health

## Engineering rules

- Prefer simple Go standard library solutions.
- Keep public APIs small and understandable.
- Return errors instead of silently ignoring them.
- Do not introduce new third-party dependencies without a clear reason.
- Add or update tests when changing behavior.
- Keep README factual and modest.
- Do not call the project production-ready unless the code, tests, documentation, and examples support that claim.
- Do not add secrets, tokens, customer data, private company names, or confidential implementation details.

## Definition of done

A change is done only when:

1. go fmt ./... passes.
2. go test ./... passes.
3. go vet ./... passes.
4. The example server starts.
5. At least / and /health can be checked manually with curl.
6. README remains accurate.

## Verification policy

Use the smallest useful verification for the change.

For Go source code changes:

- Run go fmt ./...
- Run go test ./...
- Run go vet ./...

For behavior changes that affect startup or routing:

- Also run go run ./cmd/example_server
- If the server starts, manually check / and /health with curl when practical.

For documentation-only changes:

- Do not run Go commands unless Go source code, go.mod, tests, or runtime configuration changed.
- Check that documented commands match the current project layout.
- Clearly report if documented commands were not executed.

For config-only changes:

- Run tests if config parsing tests exist.
- Run go test ./... if parsing behavior changed.
- Do not start the server unless the task specifically concerns startup behavior.