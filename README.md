# Library Microservice

A Go microservices platform for managing a library: users, books, authors, genres, loans, notifications, and an API gateway. Each domain service owns its own PostgreSQL database and communicates over HTTP and Kafka.

## Services

| Service | Port | Responsibility |
|---|---:|---|
| `userservice` | 8080 | Registration, login, JWT issuance, token/admin validation, profiles |
| `bookservice` | 8081 | Genres, authors, books, book-author/genre associations, reviews, copy inventory |
| `loanservice` | 8082 | Borrowing, returns, overdue detection, availability integration |
| `notificationservice` | 8083 | Kafka consumer for user/borrow/return events |
| `apigatewayservice` | 8084 | Public entrypoint, routing, auth/admin middleware, metrics |

## Architecture

- Database-per-service (PostgreSQL), schema applied with GORM auto-migration.
- Synchronous service-to-service calls use HTTP clients with timeouts and trace propagation.
- Asynchronous events use Kafka (`user-registered-topic` and related topics).
- Observability: Prometheus metrics (`/metrics`), Jaeger tracing, Grafana dashboards.

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Quick start

```bash
cp .env.example .env
make up
make db
make ps
```

The gateway is available at `http://localhost:8084`.

## Configuration

Shared configuration lives in the root `.env` (gitignored). Placeholders are in `.env.example`. Required values include database credentials, `JWT_SECRET`, `SENTRY_DSN`, `KAFKA_BROKER`, and `GRAFANA_ADMIN_PASSWORD`.

## Development

```bash
cd <service>
go mod download
go build ./...
go vet ./...
go test ./... -race
```

Format with `gofmt -w .` (and `goimports` if available). See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md).

## Testing

Each service has unit tests under `internal/.../*_test.go`. Run:

```bash
cd <service> && go test ./... -race
```

CI runs build, vet, and tests for all services. See [docs/TESTING.md](docs/TESTING.md).

## Project layout

```text
userservice/            User domain service
bookservice/            Book/catalog domain service
loanservice/            Loan domain service
notificationservice/    Kafka consumer service
apigatewayservice/      API gateway
docs/                   Architecture, development, and testing docs
```

## License

See repository license.