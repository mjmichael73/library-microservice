# Library Microservice

A microservices-based library management system built with Go. It provides user authentication, book catalog management (genres, authors, books), and book borrowing, fronted by an API gateway with observability (Prometheus/Grafana/Jaeger) and event streaming (Kafka).

## Services

| Service | Port | Responsibility |
|---|---|---|
| `userservice` | 8080 | Registration, login, JWT validation, admin checks, Kafka producer |
| `bookservice` | 8081 | Genres, authors, books (CRUD), availability check |
| `loanservice` | 8082 | Borrow flow (validates book availability via bookservice) |
| `apigatewayservice` | 8084 | Reverse proxy, token validation, admin gating, metrics |
| `notificationservice` | 8083 | (planned) Kafka consumer for user/borrow events |

Infrastructure: PostgreSQL (one DB per service), Kafka (+ Zookeeper, Kafka UI), Jaeger, Prometheus, Grafana.

## Quick start

```bash
# 1. Configure environment
cp .env.example .env   # edit secrets as needed

# 2. Start everything
make up                # docker compose build + up -d

# 3. Load schemas
make db

# 4. Check health
curl http://localhost:8084/liveness
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for local (non-Docker) development and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for system design.

## Project layout

```
.
├── userservice/            # auth service
├── bookservice/            # catalog service
├── loanservice/            # borrowing service
├── apigatewayservice/      # gateway
├── notificationservice/    # (empty, planned)
├── docker-compose.yml
├── .env / .env.example
├── Makefile
└── docs/
    ├── ARCHITECTURE.md
    ├── DEVELOPMENT.md
    ├── TESTING.md
    └── ROADMAP.md
```

## Phased roadmap

This roadmap summarizes the review findings (bugs, improvements, new features). Each item references the detailed entry in [docs/ROADMAP.md](docs/ROADMAP.md).

### Phase 0 — Critical bug fixes
- Fix unchecked type assertion panic in `RegisterUser` → [ROADMAP.md#R-001](docs/ROADMAP.md#r-001)
- Return `404` (not `500`) for missing book/author/genre → [ROADMAP.md#R-002](docs/ROADMAP.md#r-002)
- Close HTTP response body on all paths in borrow flow → [ROADMAP.md#R-003](docs/ROADMAP.md#r-003)
- Fix CreateAuthor success status `"Failed"` typo → [ROADMAP.md#R-004](docs/ROADMAP.md#r-004)
- Fix LoginUser default error status `404` → `500` → [ROADMAP.md#R-005](docs/ROADMAP.md#r-005)
- Make DeleteBook/DeleteAuthor return not-found → [ROADMAP.md#R-006](docs/ROADMAP.md#r-006)
- Fix `CreateBorrow` DB returning `nil` → return created borrow → [ROADMAP.md#R-007](docs/ROADMAP.md#r-007)
- Remove debug `sentry.CaptureMessage` from reverse proxy → [ROADMAP.md#R-008](docs/ROADMAP.md#r-008)

### Phase 1 — Security & configuration
- Move hardcoded secrets (JWT, Sentry DSN) to env → [ROADMAP.md#R-009](docs/ROADMAP.md#r-009)
- Stop leaking DB credentials in logs → [ROADMAP.md#R-010](docs/ROADMAP.md#r-010)
- Add `.env` to `.gitignore` and remove committed secrets → [ROADMAP.md#R-011](docs/ROADMAP.md#r-011)
- Fix malformed notification table prefix → [ROADMAP.md#R-012](docs/ROADMAP.md#r-012)
- Remove hardcoded Kafka broker / make configurable → [ROADMAP.md#R-013](docs/ROADMAP.md#r-013)

### Phase 2 — Docker & deployment
- Remove `RUN sleep 15` from Dockerfiles → [ROADMAP.md#R-014](docs/ROADMAP.md#r-014)
- Add healthchecks + `depends_on` conditions → [ROADMAP.md#R-015](docs/ROADMAP.md#r-015)
- Remove Grafana hardcoded admin password → [ROADMAP.md#R-016](docs/ROADMAP.md#r-016)
- Add notificationservice to compose → [ROADMAP.md#R-017](docs/ROADMAP.md#r-017)

### Phase 3 — Data & API correctness
- Fix UpdateBook/UpdateAuthor returning-slice panic risk → [ROADMAP.md#R-018](docs/ROADMAP.md#r-018)
- Make `IsBookAvailableToBorrow` reflect active loans → [ROADMAP.md#R-019](docs/ROADMAP.md#r-019)
- Standardize response envelope across services → [ROADMAP.md#R-020](docs/ROADMAP.md#r-020)
- Expose book↔author/genre associations → [ROADMAP.md#R-021](docs/ROADMAP.md#r-021)
- Add DB migrations / auto-migrate → [ROADMAP.md#R-022](docs/ROADMAP.md#r-022)
- Borrow authorization (verify user exists / caller identity) → [ROADMAP.md#R-023](docs/ROADMAP.md#r-023)
- Add foreign keys for borrows → [ROADMAP.md#R-024](docs/ROADMAP.md#r-024)

### Phase 4 — Observability & resilience
- Add request timeouts & trace propagation → [ROADMAP.md#R-025](docs/ROADMAP.md#r-025)
- Graceful shutdown + Jaeger closer cleanup → [ROADMAP.md#R-026](docs/ROADMAP.md#r-026)
- Migrate Jaeger client to OpenTelemetry → [ROADMAP.md#R-027](docs/ROADMAP.md#r-027)
- Make Prometheus registration test-safe → [ROADMAP.md#R-028](docs/ROADMAP.md#r-028)

### Phase 5 — Testing
- Add unit + integration tests → [ROADMAP.md#R-029](docs/ROADMAP.md#r-029)
- Add CI pipeline → [ROADMAP.md#R-030](docs/ROADMAP.md#r-030)

### Phase 6 — New features
- Return book + overdue detection → [ROADMAP.md#R-031](docs/ROADMAP.md#r-031)
- Inventory management (copies) → [ROADMAP.md#R-032](docs/ROADMAP.md#r-032)
- Notifications service (Kafka consumer + email) → [ROADMAP.md#R-033](docs/ROADMAP.md#r-033)
- Search, pagination, reviews, profiles, admin tools → [ROADMAP.md#R-034](docs/ROADMAP.md#r-034)

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Development](docs/DEVELOPMENT.md)
- [Testing](docs/TESTING.md)
- [Roadmap](docs/ROADMAP.md)

## License

MIT (add your license here).
