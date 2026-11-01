# Roadmap

Detailed findings from the project review. Each entry maps to a phased roadmap item referenced from the root [README.md](../README.md).

Severity: **P0** critical, **P1** high, **P2** medium, **P3** low.

---

## Phase 0 — Critical bug fixes

### R-001 — Unchecked type assertion panic in RegisterUser
- **Severity:** P0
- **Status:** Done
- **Location:** `userservice/internal/database/database_user.go:42`
- **Problem:** `result.Error.(*pgconn.PgError).Code` panics if the error is not a `*pgconn.PgError`.
- **Fix:** Use a type assertion with `ok`, or `errors.As`; check `gorm.ErrDuplicatedKey` first.

### R-002 — Missing records return 500 instead of 404
- **Severity:** P0
- **Status:** Done
- **Location:** `bookservice/internal/database/database_book.go:45`, `database_author.go:42`, `database_genre.go:41`
- **Problem:** `GetBookById`/`GetAuthorById`/`GetGenreById` return raw `gorm.ErrRecordNotFound`; handlers expect `*dberrors.NotFoundError`.
- **Fix:** Wrap in `&dberrors.NotFoundError{Entity, ID}` like `GetBookByTitle`.

### R-003 — HTTP response body leak in borrow flow
- **Severity:** P1
- **Status:** Done
- **Location:** `loanservice/internal/server/borrow.go:76-84`
- **Problem:** `resp.Body.Close()` only runs on the success path; non-200 returns leak the body.
- **Fix:** `defer resp.Body.Close()` immediately after a successful `client.Do`.

### R-004 — CreateAuthor success reported as "Failed"
- **Severity:** P2
- **Location:** `bookservice/internal/server/author.go:65`
- **Problem:** Response status string is `"Failed"` on a successful creation.
- **Fix:** Change to `"Success"`.

### R-005 — LoginUser default error returns 404
- **Severity:** P2
- **Location:** `userservice/internal/server/user.go:115-117`
- **Problem:** DB error path returns 404 instead of 500.
- **Fix:** Return 500 for unexpected errors.

### R-006 — Delete of missing record succeeds
- **Severity:** P2
- **Location:** `bookservice/internal/database/database_book.go:82`, `database_author.go:77`
- **Problem:** GORM `Delete` doesn't error on 0 rows; handler returns success.
- **Fix:** Check `result.RowsAffected == 0` → `NotFoundError`.

### R-007 — CreateBorrow returns nil borrow
- **Severity:** P2
- **Location:** `loanservice/internal/database/database_borrow.go:22-30`
- **Problem:** Returns `nil, nil` instead of the created borrow; handler can't return data.
- **Fix:** Return `borrow, nil`.

### R-008 — Debug Sentry call on every request
- **Severity:** P2
- **Location:** `apigatewayservice/internal/server/server.go:75`
- **Problem:** `sentry.CaptureMessage("GO HERE " + target)` in `reverseProxy` runs per request.
- **Fix:** Remove the line.

---

## Phase 1 — Security & configuration

### R-009 — Hardcoded secrets
- **Severity:** P0
- **Location:** `userservice/internal/server/user.go:15`, `server.go:115`, `apigatewayservice/internal/server/server.go:45`
- **Problem:** JWT secret and Sentry DSN are hardcoded.
- **Fix:** Read from env (`JWT_SECRET`, `SENTRY_DSN`); fail fast if missing.

### R-010 — DB credentials logged
- **Severity:** P1
- **Location:** `userservice/internal/database/client.go:38`, `bookservice/.../client.go:55`, `loanservice/.../client.go:36`
- **Problem:** `fmt.Println(... dbPass ...)` leaks credentials.
- **Fix:** Remove or log only host/dbname.

### R-011 — `.env` committed with secrets
- **Severity:** P0
- **Location:** root `.gitignore`, `.env`
- **Problem:** `.env` is tracked and contains passwords.
- **Fix:** Add `.env` to `.gitignore`, `git rm --cached .env`, rotate credentials.

### R-012 — Malformed notification table prefix
- **Severity:** P3
- **Location:** `.env:39`
- **Problem:** `NOTIFICATIONSERVICE_DB_TABLE_PREFIX=".notificationservice"` (leading dot).
- **Fix:** Set to `notificationservice.`.

### R-013 — Hardcoded Kafka broker
- **Severity:** P2
- **Location:** `userservice/internal/server/server.go:71`
- **Problem:** `"broker:29092"` not configurable.
- **Fix:** Read `KAFKA_BROKER` env var with default.

---

## Phase 2 — Docker & deployment

### R-014 — `RUN sleep 15` in Dockerfiles
- **Severity:** P2
- **Location:** all `Dockerfile`s
- **Problem:** Artificial startup delay in final image.
- **Fix:** Remove the line.

### R-015 — No healthchecks / depends_on conditions
- **Severity:** P1
- **Location:** `docker-compose.yml`
- **Problem:** Apps can start before DB is ready; no retry in `NewDatabaseClient`.
- **Fix:** Add `healthcheck` (pg_isready) and `depends_on: condition: service_healthy`; add retry/backoff in DB client.

### R-016 — Grafana admin password hardcoded
- **Severity:** P2
- **Location:** `docker-compose.yml:156`
- **Problem:** `GF_SECURITY_ADMIN_PASSWORD=Pass123!@`.
- **Fix:** Use env var with a default only for dev.

### R-017 — notificationservice missing from compose
- **Severity:** P2
- **Location:** `docker-compose.yml`
- **Problem:** Service referenced in env but not defined.
- **Fix:** Add compose entry (or remove references until implemented).

---

## Phase 3 — Data & API correctness

### R-018 — UpdateBook/UpdateAuthor returning-slice panic risk
- **Severity:** P1
- **Location:** `bookservice/internal/database/database_book.go:56-80`, `database_author.go:53-75`
- **Problem:** `Model(&slice)` + `clause.Returning{}` then `&slice[0]` can index out of range.
- **Fix:** Use `Model(&models.Book{})`, `Updates(...).Scan(&result)` or re-fetch by ID.

### R-019 — Availability check ignores active loans
- **Severity:** P1
- **Location:** `bookservice/internal/server/book.go:164-178`
- **Problem:** Returns "available" whenever the book row exists.
- **Fix:** Check loanservice for active borrows (or track copy counts).

### R-020 — Inconsistent response envelopes
- **Severity:** P2
- **Location:** multiple handlers
- **Problem:** Mixed shapes: raw slices, `{status,message,data}`, `Failed` on success.
- **Fix:** Adopt one envelope; document in DEVELOPMENT.md.

### R-021 — Book↔author/genre associations not exposed
- **Severity:** P2
- **Location:** models `BookAuthor`/`BookGenre` exist but unused
- **Problem:** No endpoints to attach; books returned without authors/genres.
- **Fix:** Add endpoints and preload associations in `GetBookById`/`GetAllBooks`.

### R-022 — No DB migrations / auto-migrate
- **Severity:** P1
- **Location:** all `client.go`
- **Problem:** Schema applied manually; fresh DB fails at runtime.
- **Fix:** Add `AutoMigrate` or golang-migrate with embedded SQL.

### R-023 — Borrow authorization missing
- **Severity:** P0
- **Location:** `loanservice/internal/server/borrow.go`
- **Problem:** Any caller can borrow for any `user_id`; no user existence check.
- **Fix:** Validate JWT/user via userservice; verify the authenticated user matches `user_id`.

### R-024 — No foreign keys on borrows
- **Severity:** P2
- **Location:** `loanservice/data/schema.sql`
- **Problem:** `user_id`/`book_id` are untyped-ish UUIDs without FKs.
- **Fix:** Add FK constraints (or enforce in app if cross-DB).

---

## Phase 4 — Observability & resilience

### R-025 — No request timeouts / trace propagation
- **Severity:** P2
- **Location:** `loanservice/internal/server/borrow.go:67-68`
- **Problem:** `http.Client{}` with no timeout; trace context not injected.
- **Fix:** Set `Timeout`; inject span via `opentracing.GlobalTracer().Inject`.

### R-026 — No graceful shutdown
- **Severity:** P2
- **Location:** all `main.go` / `Start()`
- **Problem:** No SIGTERM handling; Jaeger closer never closed.
- **Fix:** Use `echo` graceful shutdown + `defer closer.Close()`.

### R-027 — Deprecated Jaeger client
- **Severity:** P3
- **Location:** all `jaeger.go`
- **Problem:** `uber/jaeger-client-go` is deprecated; 100% sampling.
- **Fix:** Migrate to OpenTelemetry; make sample rate configurable.

### R-028 — Prometheus MustRegister in init()
- **Severity:** P3
- **Location:** all `metrics.go:63`
- **Problem:** Panics on duplicate registration (blocks in-process tests).
- **Fix:** Use `prometheus.Register` with duplicate handling or a sync.Once.

---

## Phase 5 — Testing

### R-029 — No automated tests
- **Severity:** P1
- **Location:** repository-wide
- **Problem:** Zero `*_test.go`.
- **Fix:** See [docs/TESTING.md](TESTING.md).

### R-030 — No CI pipeline
- **Severity:** P2
- **Location:** `.github/workflows` (missing)
- **Problem:** No automated build/vet/test on PRs.
- **Fix:** Add GitHub Actions workflow.

---

## Phase 6 — New features

### R-031 — Return book + overdue detection
- **Severity:** P3
- **Location:** loanservice
- **Description:** Add return endpoint, set `returned_date`, mark status `returned`, detect overdue.

### R-032 — Inventory management (copies)
- **Severity:** P3
- **Location:** bookservice
- **Description:** Track total/available copies; decrement on borrow, increment on return.

### R-033 — Notifications service
- **Severity:** P3
- **Location:** `notificationservice/`
- **Description:** Implement Kafka consumer for `user-registered-topic` and borrow events; send emails.

### R-034 — Search, pagination, reviews, profiles, admin tools
- **Severity:** P3
- **Location:** repository-wide
- **Description:** Full-text search, pagination, book reviews/ratings, user profile management, admin user management, export reports, soft delete, cover images, recommendations.
