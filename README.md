# ChainConnector

ChainConnector is a small Go service that demonstrates a clean, testable
architecture for connecting to blockchains and exposing an HTTP surface.
The project uses Uber Fx for dependency wiring, Fiber for the HTTP layer,
and a hexagonal-style structure that separates adapters, domain, and
application wiring.

This README summarizes repository purpose, key components, how to run and
test the project, and practical guidance for extending and testing.

## Quick Start

- Build and run (development):

```bash
go run ./cmd/chainconnector
```

- Run tests across the repo and show coverage report:

```bash
make cover
```

## Project Structure

- `cmd/chainconnector` — application entrypoint.
	- [cmd/chainconnector/main.go](cmd/chainconnector/main.go)
- `internal/app` — Fx modules and application wiring.
	- [internal/app/fx_modules.go](internal/app/fx_modules.go)
- `internal/adapters` — external adapters (HTTP server, persistence, RPC).
	- HTTP server: [internal/adapters/http/server.go](internal/adapters/http/server.go)
- `internal/domain` — core domain logic, entities and ports.
	- Transaction service: [internal/domain/service/transaction_service.go](internal/domain/service/transaction_service.go)
	- Entities (status, transaction): [internal/domain/entity/status.go](internal/domain/entity/status.go) and [internal/domain/entity/transaction.go](internal/domain/entity/transaction.go)
	- Ports: [internal/domain/ports/tx_repository_port.go](internal/domain/ports/tx_repository_port.go)
- `migrations/` — database migration files.

## Architecture & Design

High level principles used in this repository:

- Hexagonal / Ports & Adapters: domain code depends on interfaces (ports) defined
	in `internal/domain/ports`, concrete adapters implement those ports.
- Dependency injection with Uber Fx: wiring and lifecycle hooks live in
	`internal/app/fx_modules.go` and the `cmd` package boots the Fx app.
- Small, testable units: services are written to accept interfaces and a
	`*zap.Logger` so behavior can be validated with substitutions/mocks.

## HTTP Server (Fiber) notes

- The HTTP adapter builds a Fiber app via `CreateFiberServer` and exposes an
	Fx-friendly `FiberServer` (constructor `NewFiberServer`) that registers
	lifecycle hooks to start/stop the server. See [internal/adapters/http/server.go](internal/adapters/http/server.go).
- For route-level tests, you can use the returned `*fiber.App` and call
	`app.Test(req)` to exercise handlers without starting a network listener.

Example (test):

```go
app := CreateFiberServer()
req, _ := http.NewRequest("GET", "/health", nil)
resp, _ := app.Test(req)
// assert resp.StatusCode
```

## Testing guidance

- Use `zap.NewNop()` when a `*zap.Logger` is required by services in tests
	to avoid nil-pointer panics. Example: `svc := NewTransactionService(repo, zap.NewNop())`.
- The HTTP server `FiberServer` stores an app as an interface to allow tests to
	inject a fake implementation that implements `Listen` and `Shutdown` so
	lifecycle hooks can be exercised without opening sockets.
- Tests in this repo demonstrate these techniques (see `internal/*_test.go`).

## Running & Development

- Install dependencies (if needed):

```bash
go mod tidy
```

- Run unit tests for a single package:

```bash
go test ./internal/domain/service -v
```

- Run entire test suite and coverage summary (already provided by `make cover`):

```bash
make cover
```

## Extending the project

- To add new blockchain RPC adapters, follow the ports pattern in
	`internal/domain/ports` and provide concrete implementations under
	`internal/adapters/ethereum_rpc` or equivalent.
- For multi-chain support consider a router/adapter that maps chain IDs to
	configured RPC endpoints (see the project's `.github/prompts/plan-evmPlugNPlay.prompt.md` for a suggested plan).

## Best Practices & Conventions used

- Dependency injection via Fx keeps wiring explicit and test doubles easy.
- Keep side-effecting code (network, IO) in adapters; domain code stays pure
	and relies on interfaces.
- Prefer small, focused tests that mock external dependencies. Use
	`zap.NewNop()` for loggers and lightweight fakes for servers.
- Maintain clear package boundaries: `adapters`, `domain`, `app`.

## Contributing

1. Fork the repo and create a feature branch.
2. Add unit tests for new behavior.
3. Run `gofmt` and `go vet` and ensure `make cover` passes.
4. Open a pull request with a short description and changelog.

## License

This repository includes a `LICENSE` file at the root. Follow the terms
contained there when contributing or reusing code.

---

If you want, I can also:
- add a short CONTRIBUTING.md with a PR checklist,
- add CI workflow to run `make cover` on PRs,
- or commit the README changes for you.