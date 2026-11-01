package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

// errorRoundTripper fails every request, simulating an unreachable userservice.
type errorRoundTripper struct{}

func (errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}

func newR023DB() *mockBorrowDB {
	return &mockBorrowDB{
		activeBorrowFn: func(context.Context, string, string) (*models.Borrow, error) {
			return nil, errors.New("no active borrow")
		},
		createFn: func(_ context.Context, borrow *models.Borrow) (*models.Borrow, error) {
			return borrow, nil
		},
	}
}

// TestCreateBorrow_MissingAuthorization_Returns401 covers R-023: a request
// without an Authorization header is rejected before any downstream call.
func TestCreateBorrow_MissingAuthorization_Returns401(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")

	s := newBorrowTestServer(t, newR023DB())
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "", validBorrowBody(t)), rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestCreateBorrow_UserserviceRejects_Returns401 covers R-023: a rejected token
// (non-200 from userservice) yields 401.
func TestCreateBorrow_UserserviceRejects_Returns401(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")

	s := newBorrowTestServer(t, newR023DB())
	rt := &routedRoundTripper{
		userServiceStatus: http.StatusUnauthorized,
		userServiceBody:   &trackingReadCloser{},
	}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = originalTransport }()

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "bad-token", validBorrowBody(t)), rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestCreateBorrow_UserserviceUnreachable_Returns401 covers R-023: a transport
// error reaching userservice is treated as an invalid token (401).
func TestCreateBorrow_UserserviceUnreachable_Returns401(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")

	s := newBorrowTestServer(t, newR023DB())
	originalTransport := http.DefaultTransport
	http.DefaultTransport = errorRoundTripper{}
	defer func() { http.DefaultTransport = originalTransport }()

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "test-token", validBorrowBody(t)), rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

// TestCreateBorrow_UserMismatch_Returns403 covers R-023: a valid token whose
// user id differs from the request user_id is forbidden.
func TestCreateBorrow_UserMismatch_Returns403(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")

	s := newBorrowTestServer(t, newR023DB())
	rt := &routedRoundTripper{
		userServiceStatus: http.StatusOK,
		userServiceBody:   userServiceTokenBody("someone-else"),
	}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = originalTransport }()

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "test-token", validBorrowBody(t)), rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

// TestCreateBorrow_UserserviceNotConfigured_Returns500 covers R-023: missing
// userservice env configuration is an operator error surfaced as 500.
func TestCreateBorrow_UserserviceNotConfigured_Returns500(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "")
	t.Setenv("USERSERVICE_APP_PORT", "")

	s := newBorrowTestServer(t, newR023DB())
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "test-token", validBorrowBody(t)), rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
