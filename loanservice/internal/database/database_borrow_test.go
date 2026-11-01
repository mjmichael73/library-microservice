package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockBorrowClient builds a Client backed by a go-sqlmock connection so
// CreateBorrow can be exercised without a live Postgres instance.
func newMockBorrowClient(t *testing.T) (Client, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}

	return Client{DB: db}, mock
}

// TestCreateBorrow_ReturnsCreatedBorrow covers R-007: the method must return
// the created borrow rather than a nil pointer.
func TestCreateBorrow_ReturnsCreatedBorrow(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "borrows"`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	borrow := &models.Borrow{
		BorrowID:  "borrow-1",
		UserID:    "user-1",
		BookID:    "book-1",
		FromDate:  time.Now(),
		ToDate:    time.Now().AddDate(0, 1, 0),
		Status:    "active",
		Remarks:   "test",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	got, err := client.CreateBorrow(context.Background(), borrow)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != borrow {
		t.Fatalf("expected returned borrow to be the created borrow, got %v", got)
	}
	if got.BorrowID != "borrow-1" {
		t.Errorf("expected borrow id borrow-1, got %q", got.BorrowID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCreateBorrow_Duplicate_ReturnsConflict verifies the existing duplicate-key
// mapping still returns a ConflictError and no borrow.
func TestCreateBorrow_Duplicate_ReturnsConflict(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "borrows"`).
		WillReturnError(gorm.ErrDuplicatedKey)
	mock.ExpectRollback()

	borrow := &models.Borrow{BorrowID: "borrow-1", Status: "active"}

	got, err := client.CreateBorrow(context.Background(), borrow)

	var conflict *dberrors.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
	if got != nil {
		t.Errorf("expected nil borrow on conflict, got %v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestGetActiveBorrowsByBookId_ReturnsRows covers R-019: the query filters by
// book_id and active status and returns every matching borrow.
func TestGetActiveBorrowsByBookId_ReturnsRows(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	rows := sqlmock.NewRows([]string{"borrow_id", "user_id", "book_id", "status"}).
		AddRow("borrow-1", "user-1", "book-1", "active")
	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE "borrows"\."book_id" = \$1 AND "borrows"\."status" = \$2`).
		WithArgs("book-1", "active").
		WillReturnRows(rows)

	got, err := client.GetActiveBorrowsByBookId(context.Background(), "book-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 borrow, got %d", len(got))
	}
	if got[0].BorrowID != "borrow-1" || got[0].UserID != "user-1" {
		t.Errorf("unexpected borrow: %+v", got[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestGetActiveBorrowsByBookId_NoRows verifies an empty result set yields an
// empty slice and no error, so the handler can treat the book as available.
func TestGetActiveBorrowsByBookId_NoRows(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE "borrows"\."book_id" = \$1 AND "borrows"\."status" = \$2`).
		WithArgs("book-1", "active").
		WillReturnRows(sqlmock.NewRows([]string{"borrow_id", "user_id", "book_id", "status"}))

	got, err := client.GetActiveBorrowsByBookId(context.Background(), "book-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 borrows, got %d", len(got))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
