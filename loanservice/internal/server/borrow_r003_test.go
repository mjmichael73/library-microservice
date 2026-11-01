package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/loanservice/internal/database"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

type mockBorrowDB struct {
	activeBorrowFn  func(context.Context, string, string) (*models.Borrow, error)
	activeBorrowsFn func(context.Context, string) ([]models.Borrow, error)
	getBorrowFn     func(context.Context, string) (*models.Borrow, error)
	returnBorrowFn  func(context.Context, string) (*models.Borrow, error)
	markOverdueFn   func(context.Context) ([]models.Borrow, error)
	createFn        func(context.Context, *models.Borrow) (*models.Borrow, error)
}

func (m *mockBorrowDB) Ready() bool { return true }

func (m *mockBorrowDB) GetActiveBorrowByUserIdAndBookId(ctx context.Context, bookId string, userId string) (*models.Borrow, error) {
	if m.activeBorrowFn != nil {
		return m.activeBorrowFn(ctx, bookId, userId)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBorrowDB) GetActiveBorrowsByBookId(ctx context.Context, bookId string) ([]models.Borrow, error) {
	if m.activeBorrowsFn != nil {
		return m.activeBorrowsFn(ctx, bookId)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBorrowDB) GetBorrowById(ctx context.Context, borrowID string) (*models.Borrow, error) {
	if m.getBorrowFn != nil {
		return m.getBorrowFn(ctx, borrowID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBorrowDB) ReturnBorrow(ctx context.Context, borrowID string) (*models.Borrow, error) {
	if m.returnBorrowFn != nil {
		return m.returnBorrowFn(ctx, borrowID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBorrowDB) MarkOverdueBorrows(ctx context.Context) ([]models.Borrow, error) {
	if m.markOverdueFn != nil {
		return m.markOverdueFn(ctx)
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
// detected (R-003). It optionally serves a fixed payload before reporting EOF.
type trackingReadCloser struct {
	closed bool
	data   string
}

func (t *trackingReadCloser) Read(p []byte) (int, error) {
	if t.data == "" {
		return 0, io.EOF
	}
	n := copy(p, t.data)
	t.data = t.data[n:]
	return n, nil
}

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

// routedRoundTripper serves a distinct canned response for the userservice
// token-validation call and the bookservice availability call, and records the
// Authorization header forwarded to userservice (R-023).
type routedRoundTripper struct {
	userServiceStatus int
	userServiceBody   *trackingReadCloser
	bookServiceStatus int
	bookServiceBody   *trackingReadCloser
	authHeader        string
}

func (rt *routedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Path, "/user/validate-token") {
		rt.authHeader = req.Header.Get("Authorization")
		return &http.Response{StatusCode: rt.userServiceStatus, Body: rt.userServiceBody, Header: http.Header{}}, nil
	}
	return &http.Response{StatusCode: rt.bookServiceStatus, Body: rt.bookServiceBody, Header: http.Header{}}, nil
}

// userServiceTokenBody builds a userservice /user/validate-token success body
// that reports userID as the authenticated user.
func userServiceTokenBody(userID string) *trackingReadCloser {
	return &trackingReadCloser{data: `{"status":"Success","message":"User is valid","data":"` + userID + `"}`}
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

// newBorrowRequest builds a create-borrow request with the given Authorization
// header (set to "Bearer <token>") and body.
func newBorrowRequest(t *testing.T, token, body string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/loan/borrow", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}

// TestCreateBorrow_Non200ClosesBody verifies that the bookservice response
// body is closed when the service returns a non-200 status, i.e. that the body
// is no longer leaked on the early-return path (R-003). Token validation must
// succeed first so the request reaches the bookservice call (R-023).
func TestCreateBorrow_Non200ClosesBody(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")
	t.Setenv("BOOKSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("BOOKSERVICE_APP_PORT", "9")

	db := &mockBorrowDB{
		activeBorrowFn: func(context.Context, string, string) (*models.Borrow, error) {
			return nil, errors.New("no active borrow")
		},
	}
	s := newBorrowTestServer(t, db)

	body := &trackingReadCloser{}
	rt := &routedRoundTripper{
		userServiceStatus: http.StatusOK,
		userServiceBody:   userServiceTokenBody("u1"),
		bookServiceStatus: http.StatusServiceUnavailable,
		bookServiceBody:   body,
	}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = originalTransport }()

	c := s.echo.NewContext(newBorrowRequest(t, "test-token", validBorrowBody(t)), httptest.NewRecorder())

	if err := s.CreateBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rt.authHeader != "Bearer test-token" {
		t.Fatalf("userservice Authorization header = %q, want %q", rt.authHeader, "Bearer test-token")
	}
	if got := c.Response().Status; got != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", got, http.StatusNotFound)
	}
	if !body.closed {
		t.Fatal("response body was not closed on the non-200 path")
	}
}

// TestCreateBorrow_SuccessClosesBody verifies the success path still closes the
// body after the refactor (R-003), with token validation passing (R-023).
func TestCreateBorrow_SuccessClosesBody(t *testing.T) {
	t.Setenv("USERSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("USERSERVICE_APP_PORT", "9")
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
	rt := &routedRoundTripper{
		userServiceStatus: http.StatusOK,
		userServiceBody:   userServiceTokenBody("u1"),
		bookServiceStatus: http.StatusOK,
		bookServiceBody:   body,
	}
	originalTransport := http.DefaultTransport
	http.DefaultTransport = rt
	defer func() { http.DefaultTransport = originalTransport }()

	rec := httptest.NewRecorder()
	c := s.echo.NewContext(newBorrowRequest(t, "test-token", validBorrowBody(t)), rec)

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
