# AGENTS.md

Instructions for coding agents working in this repository.

## Project

Library Microservice — Go microservices (userservice, bookservice, loanservice, apigatewayservice, notificationservice) with shared Postgres/Kafka/Jaeger/Prometheus/Grafana infra. Module path prefix: `github.com/mjmichael73/library-microservice`.

## Ground rules

- Work in the repo root `/home/ubuntu/projects/personal/library-microservice`.
- Do not commit secrets. `.env` is gitignored; keep `.env.example` placeholders only. Never log DSNs or passwords.
- All four Go services must build (`go build ./...`) and pass `go vet ./...` before committing.
- Follow existing package layout: `internal/{database,models,server,dberrors}`.
- Wrap DB errors into `dberrors.NotFoundError` / `dberrors.ConflictError` so handlers map to HTTP statuses.
- Use the standard response envelope `{status, message, data}`.
- Run `gofmt -w .` and `goimports -w .` on changed files.
- No tests exist yet; add `_test.go` alongside changed code when feasible (see docs/TESTING.md).

## Common commands

```bash
# Build all
for s in userservice bookservice loanservice apigatewayservice; do (cd $s && go build ./... && go vet ./...); done

# Docker
make up && make db && make ps && make down

# Format
gofmt -w . && goimports -w .
```

## Architecture pointers

- Read docs/ARCHITECTURE.md before changing service boundaries or data flow.
- Database-per-service: never join across services; use HTTP calls (with timeouts + trace propagation) or Kafka events.
- Gateway is the only public entrypoint; new externally exposed routes must be proxied there with appropriate middleware.
- Inter-service HTTP calls must use a client with a timeout and inject the tracing span.

## Known pitfalls (do not reintroduce)

- Unchecked type assertions on `*pgconn.PgError` panic — always use `ok`.
- GORM `Delete` does not return `ErrRecordNotFound` on 0 rows — check `RowsAffected`.
- GORM struct conditions skip zero values — be careful with empty-string filters.
- `Model(&slice)` + `clause.Returning{}` then indexing `slice[0]` can panic — re-fetch by ID instead.
- `resp.Body` must be closed on every path (use defer right after `client.Do`).
- Do not add `RUN sleep` to Dockerfiles.
- Do not hardcode secrets, broker URLs, or ports — use env vars.

## Roadmap

Outstanding work is tracked in docs/ROADMAP.md (R-001 … R-034) and summarized in README.md. When fixing an item, reference its ID in the commit message (e.g. `fix: handle missing records (R-002)`).

## Commit style

Use conventional commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`. Keep commits focused and atomic.
