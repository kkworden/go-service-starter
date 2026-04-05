# go-service-starter

A production-ready Go backend starter template. Clone it, replace the example `Item` domain with your own, and start building.

## Architecture

Three-layer architecture with dependency injection and consumer-defined interfaces:

```
handler (HTTP)  ->  service (business logic)  ->  store (Postgres)
```

Each layer defines the interface it **consumes** from the layer below. Compile-time checks in `main.go` catch signature drift at build time rather than at runtime.

```
.
├── main.go                 # Dependency wiring, migrations, server startup
├── config/                 # Environment-based configuration
│   └── config.go
├── domain/                 # Shared data types and sentinel errors
│   ├── item.go
│   └── errors.go
├── handler/                # HTTP request/response translation
│   ├── item.go             # Item endpoints
│   ├── health.go           # Liveness and readiness probes
│   └── response.go         # writeJSON / writeError helpers
├── service/                # Business logic
│   └── item.go
├── store/                  # Postgres persistence (Writer/Reader split)
│   ├── postgres.go         # DB struct, connection helper
│   └── item.go
├── middleware/              # HTTP middleware (decompression)
│   └── decompress.go
├── util/                   # Shared helpers (pagination)
│   └── pagination.go
├── migrations/             # SQL migrations (embedded in binary)
│   ├── 000001_create_items.up.sql
│   └── 000001_create_items.down.sql
├── Makefile                # Build, test, lint, deploy automation
├── Dockerfile              # Multi-stage build to Alpine 3.21
├── docker-compose.yml      # Local Postgres 17
└── helm/                   # Kubernetes Helm chart
```

## Prerequisites

- Go 1.26+
- Docker and Docker Compose
- Make

Optional (installed via `make setup`):
- [golangci-lint](https://golangci-lint.run/) for deeper static analysis
- [air](https://github.com/air-verse/air) for live-reload during development

## Quick Start

```bash
git clone <repo-url> && cd go-service-starter
cp .env.example .env        # defaults work with the local Postgres
make setup                   # install golangci-lint + air (optional)
make db/up                   # start Postgres in the background
make run                     # build, run migrations, start server
```

The server starts on `http://localhost:8080`. Migrations run automatically on startup.

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | -- | Postgres writer connection string |
| `DATABASE_READ_URL` | No | `DATABASE_URL` | Postgres reader replica connection string |
| `PORT` | No | `8080` | HTTP listen port |

Setting an optional variable to an unparseable value is a fatal error, not a silent fallback to the default.

The `config` package also provides `envDurationOr` and `envIntOr` helpers for adding typed config fields as your service grows.

## API Endpoints

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness probe -- always returns 200 |
| `GET` | `/readyz` | Readiness probe -- pings the database |

### Items (example domain)

| Method | Path | Description |
|---|---|---|
| `POST` | `/items` | Create an item |
| `GET` | `/items/{id}` | Retrieve an item by ID |

### Error Format

All error responses use a consistent JSON structure:

```json
{"error": "item not found", "code": "NOT_FOUND"}
```

## Example Usage

```bash
# Create an item
curl -s -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{"data": "hello world"}'

# Response: {"id":"<uuid>","data":"hello world"}

# Retrieve the item
curl -s http://localhost:8080/items/<uuid>
```

## Database

**Postgres 17** via Docker Compose for local development.

The store layer uses a **Writer/Reader split** -- all `SELECT` queries go through `db.Reader` (which can point to a replica in production), while mutations use `db.Writer`. In local dev, both point to the same instance.

**Migrations** are embedded in the binary via `go:embed` and applied automatically on startup using [golang-migrate](https://github.com/golang-migrate/migrate). Add new migrations as numbered SQL files in `migrations/`.

```bash
make db/up        # start Postgres
make db/psql      # open a psql session
make db/down      # stop (data preserved)
make db/destroy   # stop and delete all data
```

## Building and Testing

```bash
make build        # compile to out/main
make test         # run tests with race detector
make coverage     # tests + HTML coverage report (out/coverage.html)
make fmt          # auto-format Go source files
make lint         # go vet + golangci-lint (if installed)
make watch        # live-reload on source changes (requires air)
```

Store integration tests require a running database. They are skipped by default:

```bash
TEST_DATABASE_URL="postgres://starter:starter@localhost:5432/starter?sslmode=disable" make test
```

## Middleware Stack

All requests pass through (in order):

1. **Logger** -- logs every request (chi built-in)
2. **Recoverer** -- catches panics, returns 500 (chi built-in)
3. **Compress** -- gzip response compression at level 5 (chi built-in)
4. **Decompress** -- decompresses gzip request bodies, capped at 10 MB

Add authentication or other middleware as separate files in the `middleware/` package and wire them in `main.go`. Use `r.Group(func(r chi.Router) { ... })` for route-scoped middleware (e.g., JWT auth on protected endpoints).

## Deploying

### Docker

```bash
make image        # build Docker image (loads into kind if available)
```

The Dockerfile uses a multi-stage build with `CGO_ENABLED=0` for a minimal Alpine image. A `.dockerignore` prevents secrets and build artifacts from entering the image.

### Kubernetes (Helm)

```bash
make chart        # build image + update helm/values.yaml
make up           # deploy via Helm
make down         # remove Helm release
```

The Helm chart includes:
- Deployment with `/healthz` liveness and `/readyz` readiness probes
- ConfigMap for non-sensitive env vars (`PORT`)
- Secret for sensitive env vars (`DATABASE_URL`, `DATABASE_READ_URL`)
- Service, ServiceAccount, optional Ingress and HPA

Override secrets per environment -- never commit real values to `helm/values.yaml`.

## Make Targets Reference

| Target | Description |
|---|---|
| `make build` | Build the binary to `out/main` |
| `make run` | Build and run the server (reads `.env`) |
| `make watch` | Live-reload on source changes (requires `make setup`) |
| `make test` | Run all tests with race detector |
| `make coverage` | Tests + coverage summary + HTML report |
| `make fmt` | Auto-format all Go source files |
| `make lint` | `go vet` + `golangci-lint` (if installed) |
| `make clean` | Remove build artifacts and Docker image |
| `make setup` | Install `golangci-lint` and `air` |
| `make db/up` | Start Postgres via Docker Compose |
| `make db/down` | Stop Postgres (data preserved) |
| `make db/destroy` | Stop Postgres and delete all data |
| `make db/psql` | Open psql session against local DB |
| `make image` | Build Docker image |
| `make chart` | Update `helm/values.yaml` and build image |
| `make up` | Deploy to Kubernetes via Helm |
| `make down` | Remove Helm release |
| `make help` | Print all available targets |

## Key Design Decisions

- **No ORM** -- raw SQL via `database/sql` + pgx for full control and transparency.
- **Consumer-defined interfaces** -- each layer defines what it needs from the layer below, enabling clean testing with inline mocks.
- **Embedded migrations** -- SQL files compiled into the binary; no external migration tool required at runtime.
- **Writer/Reader split** -- store queries route through `db.Reader` or `db.Writer` to support read replicas in production.
- **Graceful shutdown** -- the server drains in-flight requests (up to 10s) on SIGINT/SIGTERM for safe Kubernetes rolling deploys.
- **Decompression safety** -- gzip request bodies are capped at 10 MB to prevent decompression bombs.
- **Fail-loud config** -- unparseable env vars cause a fatal error rather than silently falling back to defaults.
