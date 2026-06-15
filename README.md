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

- Start a local stack with Docker Compose:

```bash
docker compose up -d --build
```

> If you are still using legacy Docker Compose, you can also run:
>
> ```bash
docker-compose up -d --build
> ```

This repository includes a local Sepolia-compatible node powered by Anvil inside `docker-compose.yml`. The backend is configured to use `SEPOLIA_RPC_URL` and `ETH_RPC_URL` against the local Anvil node for development.

The Compose stack also includes LocalStack with SQS:

- `localstack` emulates AWS SQS locally on `http://localhost:4566`.
- `localstack-init` creates the `chainconnector-network-registrations` and `chainconnector-block-events` queues.
- `chainconnector` runs the HTTP API, produces network-registration messages, and subscribes to new blocks over WebSocket.
- `chainconnector-consumer` consumes network-registration messages and block events. For each block event, it calls `eth_getBlockByNumber` and stores transactions whose `from` or `to` address matches a wallet registered in PostgreSQL.

Active block monitoring uses these local defaults:

- HTTP RPC: `http://anvil:8545`
- WebSocket RPC: `ws://anvil:8545`
- Block queue: `chainconnector-block-events`
- Producer enabled in the API container with `BLOCK_PRODUCER_ENABLED=true`
- Consumer enabled in the worker container with `BLOCK_CONSUMER_ENABLED=true`

Register a network through the API:

```bash
curl -X POST http://localhost:3001/networks \
  -H "Content-Type: application/json" \
  -d '{
    "name": "sepolia",
    "chain_id": 11155111,
    "rpc_url": "http://anvil:8545",
    "currency_symbol": "ETH",
    "explorer_url": "https://sepolia.etherscan.io"
  }'
```

List persisted networks after the consumer processes the queue:

```bash
curl http://localhost:3001/networks
```

Run the block-monitor integration test against the local Docker stack:

```bash
docker compose up -d --build
$env:CHAINCONNECTOR_INTEGRATION="1"; go test -tags=integration ./test/integration -v
```

On Unix-like shells:

```bash
CHAINCONNECTOR_INTEGRATION=1 go test -tags=integration ./test/integration -v
```

This test generates a wallet, imports it through the API, sends an Anvil transaction to that wallet, waits for the receipt, and verifies that the block consumer captured the transaction in PostgreSQL.

Alternatively, to connect against a real Sepolia client, use:

```bash
docker compose -f docker-compose.sepolia.yml up -d
```

## Frontend

The project includes a sophisticated React frontend for managing the ChainConnector system:

### Development Setup

**Requirements**: Node.js 18+ and npm

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173` with API proxy to `http://localhost:3001`.

### Production with Docker

```bash
docker-compose up --build
```

Frontend available at `http://localhost:8080`, API proxy configured automatically.

### Features

- **Transaction Management**: Create, list, and monitor transaction status
- **Balance Control**: Query and update offchain balances
- **Network Management**: Register networks through the LocalStack/SQS producer-consumer flow
- **Wallet Monitoring**: Registered wallets are automatically checked by the block consumer
- **Address Monitoring**: Register interests and view blockchain logs
- **Block Events**: WebSocket producer publishes new block events; worker consumes blocks by RPC

### Compatibility Notes

- **Backend**: Go 1.19+ (tested with Go 1.21)
- **Frontend**: Node.js 18+ required for development (React 18 + Vite 5)
- **Database**: PostgreSQL 13+
- **Blockchain**: Ethereum-compatible RPC (Sepolia, Mainnet, etc.)

For development environments with older Node.js versions, use Docker for frontend builds:

```bash
docker build -t chainconnector-frontend ./frontend
```

## Project Structure

- `cmd/chainconnector` — application entrypoint.
	- [cmd/chainconnector/main.go](cmd/chainconnector/main.go)
- `internal/app` — Fx modules and application wiring.
	- [internal/app/fx_modules.go](internal/app/fx_modules.go)
- `internal/adapters` — external adapters (HTTP server, persistence, RPC).
	- HTTP server: [internal/adapters/http/server.go](internal/adapters/http/server.go)
	- LocalStack/SQS queue: [internal/adapters/sqsqueue/network_queue.go](internal/adapters/sqsqueue/network_queue.go)
- `internal/domain` — core domain logic, entities and ports.
	- Transaction service: [internal/domain/service/transaction_service.go](internal/domain/service/transaction_service.go)
	- Network service: [internal/domain/service/network_service.go](internal/domain/service/network_service.go)
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
