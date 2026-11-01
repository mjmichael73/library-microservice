package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

// TestGetActiveBorrowsByBookId_ActiveBorrow_Returns200 covers R-019: when the
// book has an active borrow the endpoint returns 200 with the borrows.
func TestGetActiveBorrowsByBookId_ActiveBorrow_Returns200(t *testing.T) {
	db := &mockBorrowDB{
		activeBorrowsFn: func(context.Context, string) ([]models.Borrow, error) {
			return []models.Borrow{{BorrowID: "borrow-1", BookID: "book-1", Status: "active"}}, nil
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/loan/active/book-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("book_id")
	c.SetParamValues("book-1")

	if err := s.GetActiveBorrowsByBookId(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

// TestGetActiveBorrowsByBookId_NoActiveBorrow_Returns404 covers the available
// case: no active borrow yields 404 so bookservice can treat it as borrowable.
func TestGetActiveBorrowsByBookId_NoActiveBorrow_Returns404(t *testing.T) {
	db := &mockBorrowDB{
		activeBorrowsFn: func(context.Context, string) ([]models.Borrow, error) {
			return []models.Borrow{}, nil
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/loan/active/book-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("book_id")
	c.SetParamValues("book-1")

	if err := s.GetActiveBorrowsByBookId(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestGetActiveBorrowsByBookId_DBError_Returns500 verifies an unexpected DB
// error surfaces as 500 rather than a misleading available/unavailable result.
func TestGetActiveBorrowsByBookId_DBError_Returns500(t *testing.T) {
	db := &mockBorrowDB{
		activeBorrowsFn: func(context.Context, string) ([]models.Borrow, error) {
			return nil, errors.New("db down")
		},
	}
	s := newBorrowTestServer(t, db)

	req := httptest.NewRequest(http.MethodGet, "/loan/active/book-1", nil)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)
	c.SetParamNames("book_id")
	c.SetParamValues("book-1")

	if err := s.GetActiveBorrowsByBookId(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
