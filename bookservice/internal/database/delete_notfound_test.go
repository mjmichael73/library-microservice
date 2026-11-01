package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newMockClient builds a Client backed by a go-sqlmock connection so that
// Delete* methods can be exercised without a live Postgres instance.
func newMockClient(t *testing.T) (Client, sqlmock.Sqlmock) {
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

func TestDeleteBook_MissingRecord_ReturnsNotFound(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "books" WHERE "books"."book_id" = \$1`).
		WithArgs("missing-id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := client.DeleteBook(context.Background(), "missing-id")

	var notFound *dberrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if notFound.Entity != "book" || notFound.ID != "missing-id" {
		t.Errorf("expected book/missing-id, got %q/%q", notFound.Entity, notFound.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteBook_ExistingRecord_ReturnsNil(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "books" WHERE "books"."book_id" = \$1`).
		WithArgs("book-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := client.DeleteBook(context.Background(), "book-id"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteAuthor_MissingRecord_ReturnsNotFound(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "authors" WHERE "authors"."author_id" = \$1`).
		WithArgs("missing-id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := client.DeleteAuthor(context.Background(), "missing-id")

	var notFound *dberrors.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %v", err)
	}
	if notFound.Entity != "author" || notFound.ID != "missing-id" {
		t.Errorf("expected author/missing-id, got %q/%q", notFound.Entity, notFound.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestDeleteAuthor_ExistingRecord_ReturnsNil(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "authors" WHERE "authors"."author_id" = \$1`).
		WithArgs("author-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := client.DeleteAuthor(context.Background(), "author-id"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
