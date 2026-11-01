# Architecture

## Overview

Library Microservice is a set of Go microservices behind an API gateway. Each domain service owns its own PostgreSQL database (database-per-service). Services communicate over HTTP (synchronous) and Kafka (asynchronous events). Observability is provided by Prometheus/Grafana (metrics) and Jaeger (tracing).

## Services

### userservice (port 8080)
- Endpoints: `/auth/register`, `/auth/login`, `/user/validate-token`, `/user/is-admin`, `/user/profile`, `/readiness`, `/liveness`, `/metrics`.
- Stores users (bcrypt password hashes, `is_admin` flag).
- Issues HS256 JWTs (`user_id` claim, 24h expiry).
- Publishes `user-registered-topic` events via Kafka.

### bookservice (port 8081)
- Endpoints: admin CRUD for genres/authors/books under `/admin/*`, book-author/genre associations, copy inventory, reviews, pagination, plus `/isbookavailabletoboroow/:id`, readiness/liveness/metrics.
- PostgreSQL schema includes `genres`, `authors`, `books`, `books_authors`, `books_genres`, `book_reviews`.

### loanservice (port 8082)
- Endpoints: `POST /loan/borrow`, return/overdue endpoints, readiness/liveness/metrics.
- Validates the book and user via HTTP calls, then creates/updates borrow records.

### apigatewayservice (port 8084)
- Reverse proxies `/auth/*`, `/user/*`, `/loan/*`, `/admin/*`, `/books/*`.
- Middleware: token validation (delegates to userservice), admin gating, Prometheus metrics, Jaeger tracing, Sentry.

### notificationservice (port 8083)
- Kafka consumer for user/borrow/return events.

## Data flow

```text
Client -> API Gateway -> (userservice | bookservice | loanservice)
loanservice -> bookservice (HTTP, availability and inventory)
loanservice -> userservice (HTTP, token validation)
userservice -> Kafka (user-registered-topic)
loanservice -> Kafka (borrow/return events)
notificationservice <- Kafka
```

## Data ownership

Each service has its own database and schema prefix:
- `userservice.users`
- `bookservice.{genres,authors,books,books_authors,books_genres,book_reviews}`
- `loanservice.borrows`

There are no cross-service foreign keys at the DB level; integrity is enforced in application code.

## Observability

- **Metrics**: each service exposes `/metrics` (Prometheus format) via `promhttp`; gateway aggregates scrape targets in `prometheus/prometheus.yml`.
- **Tracing**: Jaeger all-in-one; each service runs `JaegerTracingMiddleware` and initializes a tracer in `initJaeger`.
- **Dashboards**: Grafana provisions `grafana/dashboards/library_microservices.json`.

## Deployment

`docker-compose.yml` defines all services, databases, Kafka/Zookeeper, Kafka UI, Prometheus, Grafana, and Jaeger on the `library_microservice_net` bridge network. Environment is loaded from the root `.env`.