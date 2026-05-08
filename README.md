# WalletTest

WalletTest is a REST API for wallet balance operations. It supports deposits, withdrawals, balance reads, PostgreSQL persistence, Docker Compose startup, and controlled handling of high concurrency on the same wallet.

## Features

- `POST /api/v1/wallet` applies `DEPOSIT` or `WITHDRAW` operations.
- `GET /api/v1/wallets/{WALLET_UUID}` returns current wallet balance.
- PostgreSQL stores wallets and transaction history.
- Migrations run on application startup through `golang-migrate`.
- Same-wallet write operations are serialized in the application to reduce database row-lock contention.
- Known validation, balance, queue, and context errors are mapped to non-5xx responses.
- Docker Compose starts both the API and PostgreSQL.
- Unit tests cover domain, config, service, handler, repository retry helpers, and wallet queue behavior.
- k6 load test is included for mixed `DEPOSIT`, `WITHDRAW`, and `GET` traffic.

## Requirements

- Go 1.25+
- Docker and Docker Compose
- Make
- k6, only for load testing
- GCC or another C compiler, only for `go test -race` on Windows

## Configuration

Environment variables are read from `config.env`.

```env
APP_HOST=0.0.0.0

POSTGRES_HOST=postgres
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=wallet_test
POSTGRES_SSLMODE=disable
```

Application settings are stored in `configs/config.yaml`.

```yaml
app:
  port: "8081"
  log_level: "info"
  shutdown_timeout: "10s"
  wallet_queue_size: "2048"

db:
  max_conns: "64"
  min_conns: "8"
```

## Quick Start

Start the whole system:

```bash
make up
```

Stop the system:

```bash
make down
```

Equivalent Docker Compose commands:

```bash
docker compose up -d --build
docker compose down
```

The API listens on:

```text
http://localhost:8081
```

PostgreSQL is published on:

```text
localhost:5432
```

## API

### Apply Wallet Operation

```http
POST /api/v1/wallet
Content-Type: application/json
```

Request:

```json
{
  "valletId": "11111111-1111-1111-1111-111111111111",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

`operationType` can be:

- `DEPOSIT`
- `WITHDRAW`

Successful response:

```json
{
  "walletId": "11111111-1111-1111-1111-111111111111",
  "balance": 1000
}
```

Example:

```bash
curl -X POST http://localhost:8081/api/v1/wallet \
  -H "Content-Type: application/json" \
  -d '{"valletId":"11111111-1111-1111-1111-111111111111","operationType":"DEPOSIT","amount":1000}'
```

### Get Wallet Balance

```http
GET /api/v1/wallets/{WALLET_UUID}
```

Example:

```bash
curl http://localhost:8081/api/v1/wallets/11111111-1111-1111-1111-111111111111
```

Successful response:

```json
{
  "walletId": "11111111-1111-1111-1111-111111111111",
  "balance": 1000
}
```

## Error Responses

Error responses use this shape:

```json
{
  "message": "insufficient funds",
  "success": false
}
```

Common statuses:

- `400` invalid JSON, wallet UUID, amount, or operation type.
- `404` wallet not found.
- `409` insufficient funds or balance overflow.
- `408` request canceled.
- `429` wallet queue full or request deadline exceeded.
- `500` unexpected internal error.

## Development Commands

Build local binary:

```bash
make build
```

Run tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-cover
```

Run race detector:

```bash
make test-race
```

On Windows, `make test-race` requires a C compiler in `PATH`. If `gcc` is missing, Go will fail with a `runtime/cgo` build error.

Build Docker image:

```bash
make docker-build
```

## Load Testing

The repository includes `load_test.js` for k6. It sends mixed traffic to one wallet:

- 40% `DEPOSIT`
- 35% `WITHDRAW`
- 25% `GET`

Start the service first:

```bash
make up
```

Run the default load test:

```bash
make load-test
```

Override load parameters:

```bash
make load-test K6_RATE=1000 K6_DURATION=60s K6_VUS=300 K6_MAX_VUS=1000
```

Use a specific wallet:

```bash
make load-test WALLET_ID=22222222-2222-2222-2222-222222222222
```

Or run k6 directly:

```bash
k6 run -e RATE=1000 -e DURATION=30s -e VUS=300 -e MAX_VUS=1000 load_test.js
```

The k6 thresholds check that no response has a 5xx status.

## Database

Migrations live in `migrations/`.

The application runs migrations on startup using `golang-migrate`. The migration tool maintains its schema version table, so already-applied migrations are skipped.

## Project Layout

```text
cmd/                       Application entrypoint
configs/                   YAML application config
internal/app/              Application wiring and HTTP server lifecycle
internal/config/           config.env and YAML loading
internal/domain/           Domain types and errors
internal/handler/          Gin HTTP handlers and DTOs
internal/repository/       PostgreSQL persistence
internal/repository/db/    PostgreSQL pool and migrations
internal/service/          Business logic and per-wallet processor
migrations/                SQL migrations
```

## Notes

- The request field name is `valletId` because it is part of the assignment contract.
- `walletId` is also accepted by the handler for compatibility.
- Writes for the same wallet are processed sequentially by a per-wallet queue.
- Writes for different wallets can be processed in parallel.
