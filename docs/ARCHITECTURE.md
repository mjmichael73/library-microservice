# Architecture

## Overview

Library Microservice is a set of Go microservices behind an API gateway. Each domain service owns its own PostgreSQL database (database-per-service). Services communicate over HTTP (synchronous) and Kafka (asynchronous events). Observability is provided by Prometheus/Grafana (metrics) and Jaeger (tracing).

## Services

### userservice (port 8080)
- Endpoints: `/auth/register`, `/auth/login`, `/user/validate-token`, `/user/is-admin`, `/readiness`, `/liveness`, `/metrics`.
- Stores users (bcrypt password hashes, `is_admin` flag).
- Issues HS256 JWTs (`user_id` claim, 24h expiry).
- Publishes `user-registered-topic` events via Kafka.

### bookservice (port 8081)
- Endpoints: admin CRUD for genres/authors/books under `/admin/*`, plus `/isbookavailabletoboroow/:id`, readiness/liveness/metrics.
- PostgreSQL schema includes `genres`, `authors`, `books`, `books_authors`, `books_genres` (many-to-many).

### loanservice (port 8082)
- Endpoint: `POST /loan/borrow`.
- Validates the book exists via bookservice HTTP call, then creates a borrow record.

### apigatewayservice (port 8084)
- Reverse proxies `/auth/*`, `/user/*`, `/loan/*`, `/admin/*`.
- Middleware: token validation (delegates to userservice), admin gating, Prometheus metrics, Jaeger tracing, Sentry.

### notificationservice (port 8083, planned)
- Currently an empty directory. Intended Kafka consumer for user/borrow events.

## Data flow

```
Client -> API Gateway -> (userservice | bookservice | loanservice)
loanservice -> bookservice (HTTP, availability check)
userservice -> Kafka (user-registered-topic)
```

## Data ownership

Each service has its own database and schema prefix:
- `userservice.users`
- `bookservice.{genres,authors,books,books_authors,books_genres}`
- `loanservice.borrows`

There are no cross-service foreign keys at the DB level; integrity is enforced in application code (currently partially).

## Observability

- **Metrics**: each service exposes `/metrics` (Prometheus format) via `promhttp`; gateway aggregates scrape targets in `prometheus/prometheus.yml`.
- **Tracing**: Jaeger all-in-one; each service runs `JaegerTracingMiddleware` and initializes a tracer in `initJaeger`.
- **Dashboards**: Grafana provisions `grafana/dashboards/library_microservices.json`.

## Deployment

`docker-compose.yml` defines all services, databases, Kafka/Zookeeper, Kafka UI, Prometheus, Grafana, and Jaeger on the `library_microservice_net` bridge network. Environment is loaded from the root `.env`.

## Known architectural gaps

- No service discovery / config server; hostnames are env-driven.
- No API versioning.
- No shared module; dberrors/models/server boilerplate is duplicated.
- No graceful shutdown or health-based readiness aggregation.
- Synchronous HTTP coupling between loanservice and bookservice (no circuit breaker).
