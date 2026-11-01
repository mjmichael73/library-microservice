# Development

## Prerequisites

- Go 1.23+
- Docker + Docker Compose
- PostgreSQL client (`psql`) for manual schema loading
- (Optional) `golangci-lint`, `goimports`

## Environment

Copy `.env.example` to `.env` and adjust values. The root `.env` is shared by all services via `env_file` in compose. Each service also has its own `.env.example` (userservice, bookservice) for standalone runs.

> Never commit real secrets. `.env` is gitignored; keep `.env.example` placeholders only.

## Run with Docker (recommended)

```bash
make up      # build + start all containers
make db      # load schemas into each database
make ps      # status
make down    # stop and remove volumes
```

## Run a single service locally

```bash
cd userservice
go mod download
USERSERVICE_APP_PORT=8080 \
USERSERVICE_DB_HOST=localhost \
USERSERVICE_DB_INTERNAL_PORT=54322 \
USERSERVICE_DB_NAME=userservice_db \
USERSERVICE_DB_USER=userservice_db_user \
USERSERVICE_DB_PASS=userservice_db_pass \
USERSERVICE_DB_TABLE_PREFIX="userservice." \
go run .
```

Start Postgres first (e.g. via compose) and load `userservice/data/schema.sql`.

## Build

```bash
cd <service> && go build ./... && go vet ./...
```

All four services must build with zero errors before committing.

## Code style

- Run `gofmt -w .` and `goimports -w .` before committing.
- Keep handlers thin; push logic into the database/client layer.
- Wrap DB errors into `dberrors.NotFoundError` / `dberrors.ConflictError` so handlers can map to HTTP statuses.
- Return a consistent `{status, message, data}` envelope.
- Never log secrets or full DSNs.

## Adding a new service

1. Scaffold `internal/{database,models,server,dberrors}` mirroring an existing service.
2. Add Dockerfile (multi-stage, no `sleep`).
3. Add compose entry with healthcheck and `depends_on: condition: service_healthy`.
4. Add env vars to `.env.example` and README service table.
5. Add Prometheus scrape target and Grafana dashboard panel.
6. Register routes in the gateway if externally exposed.

## Troubleshooting

- **App exits immediately on startup**: DB not ready. Check `depends_on` and `pg_isready`; re-run `make db` after containers are healthy.
- **404 on borrow availability check**: ensure bookservice route and loanservice URL typo match (tracked as R-019/R-008 in roadmap).
- **Kafka connection refused**: broker hostname is `broker` inside the compose network, `localhost:9092` from host.
