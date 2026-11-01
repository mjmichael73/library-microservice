package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// r019RoundTripper returns a canned status without touching the network. It is
// installed as http.DefaultTransport, the transport used by the http.Client
// created inside IsBookAvailableToBorrow.
type r019RoundTripper struct {
	status int
	err    error
}

func (rt *r019RoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	if rt.err != nil {
		return nil, rt.err
	}
	return &http.Response{
		StatusCode: rt.status,
		Body:       http.NoBody,
		Header:     http.Header{},
	}, nil
}

func invokeIsBookAvailable(t *testing.T, s *EchoServer) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/isbookavailabletoboroow/book-1", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("book-1")
	if err := s.IsBookAvailableToBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	return rec
}

// TestIsBookAvailableToBorrow_MissingBook_Returns404 preserves the existing
// behaviour: a book row that does not exist is not available.
func TestIsBookAvailableToBorrow_MissingBook_Returns404(t *testing.T) {
	t.Setenv("LOANSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("LOANSERVICE_APP_PORT", "9")

	server := &EchoServer{DB: &mockDatabaseClient{
		err: &dberrors.NotFoundError{Entity: "book", ID: "book-1"},
	}}

	rec := invokeIsBookAvailable(t, server)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	assertMessage(t, rec, "Book is not available to borrow.")
}

// TestIsBookAvailableToBorrow_ActiveBorrow_Returns404 covers R-019: the book row
// exists but loanservice reports an active borrow, so it is not available.
func TestIsBookAvailableToBorrow_ActiveBorrow_Returns404(t *testing.T) {
	t.Setenv("LOANSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("LOANSERVICE_APP_PORT", "9")

	server := &EchoServer{DB: &mockDatabaseClient{
		book: &models.Book{BookID: "book-1", Title: "Title", AvailableCopies: 1},
	}}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = &r019RoundTripper{status: http.StatusOK}
	defer func() { http.DefaultTransport = originalTransport }()

	rec := invokeIsBookAvailable(t, server)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	assertMessage(t, rec, "Book is not available to borrow.")
}

// TestIsBookAvailableToBorrow_NoActiveBorrow_Returns200 covers the available
// path: loanservice returns 404 (no active borrow) and the book is returned.
func TestIsBookAvailableToBorrow_NoActiveBorrow_Returns200(t *testing.T) {
	t.Setenv("LOANSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("LOANSERVICE_APP_PORT", "9")

	book := &models.Book{BookID: "book-1", Title: "Title", AvailableCopies: 1}
	server := &EchoServer{DB: &mockDatabaseClient{book: book}}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = &r019RoundTripper{status: http.StatusNotFound}
	defer func() { http.DefaultTransport = originalTransport }()

	rec := invokeIsBookAvailable(t, server)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", body["data"])
	}
	if data["book_id"] != "book-1" {
		t.Errorf("expected book_id book-1, got %v", data["book_id"])
	}
}

// TestIsBookAvailableToBorrow_UpstreamError_Returns500 verifies that a failure
// to reach loanservice is reported as an internal error, not as availability.
func TestIsBookAvailableToBorrow_UpstreamError_Returns500(t *testing.T) {
	t.Setenv("LOANSERVICE_APP_HOST", "127.0.0.1")
	t.Setenv("LOANSERVICE_APP_PORT", "9")

	server := &EchoServer{DB: &mockDatabaseClient{
		book: &models.Book{BookID: "book-1", Title: "Title", AvailableCopies: 1},
	}}

	originalTransport := http.DefaultTransport
	http.DefaultTransport = &r019RoundTripper{err: http.ErrHandlerTimeout}
	defer func() { http.DefaultTransport = originalTransport }()

	rec := invokeIsBookAvailable(t, server)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

// TestIsBookAvailableToBorrow_MissingConfig_Returns500 verifies that missing
// loanservice configuration is reported as an internal error.
func TestIsBookAvailableToBorrow_MissingConfig_Returns500(t *testing.T) {
	t.Setenv("LOANSERVICE_APP_HOST", "")
	t.Setenv("LOANSERVICE_APP_PORT", "")

	server := &EchoServer{DB: &mockDatabaseClient{
		book: &models.Book{BookID: "book-1", Title: "Title", AvailableCopies: 1},
	}}

	rec := invokeIsBookAvailable(t, server)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func assertMessage(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if got := body["message"]; got != want {
		t.Errorf("expected message %q, got %q", want, got)
	}
}
