package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

// TestReturnBorrow_Success_Returns200 covers R-031: a successful return
// returns 200 with the updated borrow in the envelope.
func TestReturnBorrow_Success_Returns200(t *testing.T) {
	borrow := &models.Borrow{BorrowID: "borrow-1", Status: "returned"}
	db := &mockBorrowDB{
		returnBorrowFn: func(context.Context, string) (*models.Borrow, error) {
			return borrow, nil
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodPost, "/loan/return/borrow-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("borrow_id")
	c.SetParamValues("borrow-1")

	if err := s.ReturnBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !contains(rec.Body.String(), `"borrow_id":"borrow-1"`) {
		t.Errorf("response body missing borrow data: %s", rec.Body.String())
	}
}

// TestReturnBorrow_NotFound_Returns404 covers R-031: a missing borrow maps to 404.
func TestReturnBorrow_NotFound_Returns404(t *testing.T) {
	db := &mockBorrowDB{
		returnBorrowFn: func(context.Context, string) (*models.Borrow, error) {
			return nil, &dberrors.NotFoundError{Entity: "borrow", ID: "missing"}
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodPost, "/loan/return/missing", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("borrow_id")
	c.SetParamValues("missing")

	if err := s.ReturnBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestReturnBorrow_AlreadyReturned_Returns409 covers R-031: a borrow that was
// already returned maps to 409 Conflict.
func TestReturnBorrow_AlreadyReturned_Returns409(t *testing.T) {
	db := &mockBorrowDB{
		returnBorrowFn: func(context.Context, string) (*models.Borrow, error) {
			return nil, &dberrors.ConflictError{}
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodPost, "/loan/return/borrow-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("borrow_id")
	c.SetParamValues("borrow-1")

	if err := s.ReturnBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

// TestReturnBorrow_DBError_Returns500 covers R-031: an unexpected DB error
// surfaces as 500.
func TestReturnBorrow_DBError_Returns500(t *testing.T) {
	db := &mockBorrowDB{
		returnBorrowFn: func(context.Context, string) (*models.Borrow, error) {
			return nil, errors.New("db down")
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodPost, "/loan/return/borrow-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("borrow_id")
	c.SetParamValues("borrow-1")

	if err := s.ReturnBorrow(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

// TestGetOverdueBorrows_Success_Returns200 covers R-031: overdue detection
// returns 200 with the overdue borrows.
func TestGetOverdueBorrows_Success_Returns200(t *testing.T) {
	db := &mockBorrowDB{
		markOverdueFn: func(context.Context) ([]models.Borrow, error) {
			return []models.Borrow{{BorrowID: "borrow-1", Status: "overdue"}}, nil
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/loan/overdue", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.GetOverdueBorrows(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !contains(rec.Body.String(), `"status":"overdue"`) {
		t.Errorf("response body missing overdue borrow: %s", rec.Body.String())
	}
}

// TestGetOverdueBorrows_DBError_Returns500 covers R-031: a DB failure while
// marking overdue loans surfaces as 500.
func TestGetOverdueBorrows_DBError_Returns500(t *testing.T) {
	db := &mockBorrowDB{
		markOverdueFn: func(context.Context) ([]models.Borrow, error) {
			return nil, errors.New("db down")
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/loan/overdue", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.GetOverdueBorrows(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}
