package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/loanservice/internal/database"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

type mockBorrowDB struct {
	activeBorrowFn func(context.Context, string, string) (*models.Borrow, error)
	createFn       func(context.Context, *models.Borrow) (*models.Borrow, error)
}

func (m *mockBorrowDB) Ready() bool { return true }

func (m *mockBorrowDB) GetActiveBorrowByUserIdAndBookId(ctx context.Context, bookId string, userId string) (*models.Borrow, error) {
	if m.activeBorrowFn != nil {
		return m.activeBorrowFn(ctx, bookId, userId)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBorrowDB) CreateBorrow(ctx context.Context, borrow *models.Borrow) (*models.Borrow, error) {
	if m.createFn != nil {
		return m.createFn(ctx, borrow)
	}
	return nil, errors.New("not implemented")
}

// trackingReadCloser records whether Close was called so a leaked body can be
// detected (R-003).
type trackingReadCloser struct {
	closed bool
}

func (t *trackingReadCloser) Read(_ []byte) (int, error) { return 0, io.EOF }
func (t *trackingReadCloser) Close() error {
	t.closed = true
	return nil
}

// staticRoundTripper returns a canned response without any network access. It
// is installed as http.DefaultTransport, which is the transport used by the
// zero-value http.Client created inside CreateBorrow.
type staticRoundTripper struct {
	status int
	body   *trackingReadCloser
}

func (rt *staticRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: rt.status,
		Body:       rt.body,
		Header:     http.Header{},
	}, nil
}

func newBorrowTestServer(t *testing.T, db database.DatabaseClient) *EchoServer {
	t.Helper()
	s := &EchoServer{
		echo: echo.New(),
		DB:   db,
	}
	s.echo.Validator = &CustomValidator{validator: validator.New()}
	return s
}

func validBorrowBody(t *testing.T) string {
	t.Helper()
	from := time.Now().AddDate(0, 0, 2).Format(time.RFC3339)
	to := time.Now().AddDate(0, 0, 30).Format(time.RFC3339)
	return `{"user_id":"u1","book_id":"b1","from_date":"` + from + `","to_date":"` + to + `","remarks":"test"}`
}

// TestCreateBorrow_Non200ClosesBody verifies that the bookservice response
// body is closed when the service returns a non-200 status, i.e. that the body
// is no longer leaked on the early-return path (R-003).
func TestCreateBorrow_Non200ClosesBody(t *testing.T) {
	t.Setenv("BOOKSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("BOOKSERVICE_APP_PORT", "9")

	db := &mockBorrowDB{
		activeBorrowFn: func(context.Context, string, string) (*models.Borrow, error) {
			return nil, errors.New("no active borrow")
		},
	}
	s := newBorrowTestServer(t, db)

	body := &trackingReadCloser{}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = &staticRoundTripper{status: http.StatusServiceUnavailable, body: body}
	defer func() { http.DefaultTransport = originalTransport }()

	req := httptest.NewRequest(http.MethodPost, "/loan/borrow", bytes.NewBufferString(validBorrowBody(t)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if !body.closed {
		t.Fatal("response body was not closed on the non-200 path")
	}
}

// TestCreateBorrow_SuccessClosesBody verifies the success path still closes the
// body after the refactor (R-003).
func TestCreateBorrow_SuccessClosesBody(t *testing.T) {
	t.Setenv("BOOKSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("BOOKSERVICE_APP_PORT", "9")

	db := &mockBorrowDB{
		activeBorrowFn: func(context.Context, string, string) (*models.Borrow, error) {
			return nil, errors.New("no active borrow")
		},
		createFn: func(_ context.Context, borrow *models.Borrow) (*models.Borrow, error) {
			return borrow, nil
		},
	}
	s := newBorrowTestServer(t, db)

	body := &trackingReadCloser{}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = &staticRoundTripper{status: http.StatusOK, body: body}
	defer func() { http.DefaultTransport = originalTransport }()

	req := httptest.NewRequest(http.MethodPost, "/loan/borrow", bytes.NewBufferString(validBorrowBody(t)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !body.closed {
		t.Fatal("response body was not closed on the success path")
	}
}
