package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
)

var borrowColumns = []string{
	"borrow_id", "user_id", "book_id", "from_date", "to_date",
	"returned_date", "status", "remarks", "created_at", "updated_at",
}

// TestGetBorrowById_NotFound covers R-031: a missing borrow yields a populated
// NotFoundError so the handler can map it to 404.
func TestGetBorrowById_NotFound(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE borrow_id = \$1 ORDER BY "borrows"\."borrow_id" LIMIT \$2`).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows(borrowColumns))

	_, err := client.GetBorrowById(context.Background(), "missing")

	var notFound *dberrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if notFound.ID != "missing" {
		t.Errorf("expected id missing, got %q", notFound.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReturnBorrow_OnTime_ReturnsReturned covers R-031: returning before the
// due date sets status "returned" and populates returned_date/updated_at.
func TestReturnBorrow_OnTime_ReturnsReturned(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	due := time.Now().AddDate(0, 0, 5)
	rows := sqlmock.NewRows(borrowColumns).AddRow(
		"borrow-1", "user-1", "book-1", time.Now().AddDate(0, 0, -5), due,
		time.Time{}, "active", "test", time.Now(), time.Now(),
	)
	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE borrow_id = \$1 ORDER BY "borrows"\."borrow_id" LIMIT \$2`).
		WithArgs("borrow-1", 1).
		WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "borrows" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := client.ReturnBorrow(context.Background(), "borrow-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Status != "returned" {
		t.Errorf("status = %q, want returned", got.Status)
	}
	if got.ReturnedDate.IsZero() {
		t.Error("returned_date was not set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReturnBorrow_Late_ReturnsOverdue covers R-031: returning after the due
// date sets status "overdue" while still recording the return.
func TestReturnBorrow_Late_ReturnsOverdue(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	rows := sqlmock.NewRows(borrowColumns).AddRow(
		"borrow-1", "user-1", "book-1", time.Now().AddDate(0, 0, -40), time.Now().AddDate(0, 0, -10),
		time.Time{}, "active", "test", time.Now(), time.Now(),
	)
	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE borrow_id = \$1 ORDER BY "borrows"\."borrow_id" LIMIT \$2`).
		WithArgs("borrow-1", 1).
		WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "borrows" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := client.ReturnBorrow(context.Background(), "borrow-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Status != "overdue" {
		t.Errorf("status = %q, want overdue", got.Status)
	}
	if got.ReturnedDate.IsZero() {
		t.Error("returned_date was not set")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReturnBorrow_AlreadyReturned_ReturnsConflict covers R-031: a borrow whose
// status is already "returned" is rejected with a ConflictError.
func TestReturnBorrow_AlreadyReturned_ReturnsConflict(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	rows := sqlmock.NewRows(borrowColumns).AddRow(
		"borrow-1", "user-1", "book-1", time.Now().AddDate(0, 0, -40), time.Now().AddDate(0, 0, -10),
		time.Now(), "returned", "test", time.Now(), time.Now(),
	)
	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE borrow_id = \$1 ORDER BY "borrows"\."borrow_id" LIMIT \$2`).
		WithArgs("borrow-1", 1).
		WillReturnRows(rows)

	_, err := client.ReturnBorrow(context.Background(), "borrow-1")

	var conflict *dberrors.ConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected ConflictError, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestReturnBorrow_NotFound_ReturnsNotFound covers R-031: returning a missing
// borrow propagates the NotFoundError from the lookup.
func TestReturnBorrow_NotFound_ReturnsNotFound(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE borrow_id = \$1 ORDER BY "borrows"\."borrow_id" LIMIT \$2`).
		WithArgs("missing", 1).
		WillReturnRows(sqlmock.NewRows(borrowColumns))

	_, err := client.ReturnBorrow(context.Background(), "missing")

	var notFound *dberrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestMarkOverdueBorrows_MarksLateActiveLoans covers R-031: active borrows past
// their due date are flipped to "overdue" and returned.
func TestMarkOverdueBorrows_MarksLateActiveLoans(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	rows := sqlmock.NewRows(borrowColumns).AddRow(
		"borrow-1", "user-1", "book-1", time.Now().AddDate(0, 0, -40), time.Now().AddDate(0, 0, -10),
		time.Time{}, "active", "test", time.Now(), time.Now(),
	)
	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE "borrows"\."status" = \$1 AND to_date < \$2`).
		WithArgs("active", sqlmock.AnyArg()).
		WillReturnRows(rows)
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "borrows" SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := client.MarkOverdueBorrows(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 overdue borrow, got %d", len(got))
	}
	if got[0].Status != "overdue" {
		t.Errorf("status = %q, want overdue", got[0].Status)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestMarkOverdueBorrows_NoneLate verifies that no active loan past its due
// date performs no update and returns an empty slice.
func TestMarkOverdueBorrows_NoneLate(t *testing.T) {
	client, mock := newMockBorrowClient(t)

	mock.ExpectQuery(`SELECT \* FROM "borrows" WHERE "borrows"\."status" = \$1 AND to_date < \$2`).
		WithArgs("active", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows(borrowColumns))

	got, err := client.MarkOverdueBorrows(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 overdue borrows, got %d", len(got))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
