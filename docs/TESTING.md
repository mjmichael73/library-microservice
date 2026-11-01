# Testing

## Current state

Automated tests exist for all services and run in CI. Unit tests use `httptest` and fake `DatabaseClient` implementations; database-layer tests use `sqlmock`.

## Recommended testing strategy

### Unit tests
- Handler tests using `httptest` + a fake `DatabaseClient` interface implementation.
- Database layer tests using `sqlmock` for query behavior.
- Validation tests for request structs.

### Integration tests
- Spin up the full compose stack in CI, run schema load, then execute an end-to-end scenario:
  1. Register user → login → get token.
  2. Create genre/author/book as admin.
  3. Borrow book as user.
  4. Verify availability flips and borrow record exists.

### Contract tests
- Gateway routing and auth middleware behavior (401 without token, 403 for non-admin).

## Conventions

- Test files live next to the code: `*_test.go`.
- Use table-driven tests.
- Mock the `DatabaseClient` interface; never hit a real DB in unit tests.
- Name tests `Test<Component>_<Scenario>`.

## Coverage targets

- Handlers: ≥ 80% line coverage.
- Database layer: ≥ 70%.
- Overall: ≥ 60% before merging new features.

## CI

The GitHub Actions workflow in `.github/workflows/ci.yml` runs `go build ./...`, `go vet ./...`, and `go test ./... -race` for each service on push and pull requests.

## Running tests

```bash
cd <service> && go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```