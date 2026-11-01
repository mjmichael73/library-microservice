# Testing

## Current state

There are **no automated tests** in the repository today (`*_test.go` files: none). All verification is manual via curl/Postman. This is tracked as roadmap item R-029.

## Recommended testing strategy

### Unit tests
- Handler tests using `httptest` + a fake `DatabaseClient` interface implementation.
- Database layer tests against a real Postgres (testcontainers-go) or SQLite for pure logic.
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

A GitHub Actions workflow should run `go build ./...`, `go vet ./...`, `golangci-lint`, and `go test ./...` for each service on every PR (roadmap R-030).

## Running tests (once added)

```bash
cd <service> && go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```
