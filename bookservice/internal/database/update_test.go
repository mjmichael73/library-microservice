package database

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// TestUpdateBook_ReturnsRefetchedRecord covers R-018: UpdateBook no longer
// indexes into a returned slice (which could panic) and instead re-fetches the
// updated row by ID, returning the persisted record.
func TestUpdateBook_ReturnsRefetchedRecord(t *testing.T) {
	client, mock := newMockClient(t)

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "books" SET "title"=\$1,"summary"=\$2,"updated_at"=\$3 WHERE "books"."book_id" = \$4`).
		WithArgs("new title", "new summary", sqlmock.AnyArg(), "book-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "books" WHERE "books"."book_id" = \$1 ORDER BY "books"."book_id" LIMIT \$2`).
		WithArgs("book-id", 1).
		WillReturnRows(sqlmock.NewRows([]string{"book_id", "title", "summary", "created_at", "updated_at"}).
			AddRow("book-id", "new title", "new summary", now, now))

	book, err := client.UpdateBook(context.Background(), &models.Book{
		BookID:  "book-id",
		Title:   "new title",
		Summary: "new summary",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if book == nil {
		t.Fatal("expected a book, got nil")
	}
	if book.BookID != "book-id" || book.Title != "new title" || book.Summary != "new summary" {
		t.Errorf("unexpected book: %+v", book)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpdateBook_NoRows_ReturnsNotFound ensures an update that matches no rows
// still reports a NotFoundError and does not attempt to index into an empty
// result.
func TestUpdateBook_NoRows_ReturnsNotFound(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "books" SET "title"=\$1,"summary"=\$2,"updated_at"=\$3 WHERE "books"."book_id" = \$4`).
		WithArgs("new title", "new summary", sqlmock.AnyArg(), "missing-id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	book, err := client.UpdateBook(context.Background(), &models.Book{
		BookID:  "missing-id",
		Title:   "new title",
		Summary: "new summary",
	})
	if book != nil {
		t.Errorf("expected nil book, got %+v", book)
	}
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

// TestUpdateAuthor_ReturnsRefetchedRecord covers R-018 for authors.
func TestUpdateAuthor_ReturnsRefetchedRecord(t *testing.T) {
	client, mock := newMockClient(t)

	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authors" SET "name"=\$1,"updated_at"=\$2 WHERE "authors"."author_id" = \$3`).
		WithArgs("new name", sqlmock.AnyArg(), "author-id").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "authors" WHERE "authors"."author_id" = \$1 ORDER BY "authors"."author_id" LIMIT \$2`).
		WithArgs("author-id", 1).
		WillReturnRows(sqlmock.NewRows([]string{"author_id", "name", "created_at", "updated_at"}).
			AddRow("author-id", "new name", now, now))

	author, err := client.UpdateAuthor(context.Background(), &models.Author{
		AuthorID: "author-id",
		Name:     "new name",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if author == nil {
		t.Fatal("expected an author, got nil")
	}
	if author.AuthorID != "author-id" || author.Name != "new name" {
		t.Errorf("unexpected author: %+v", author)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestUpdateAuthor_NoRows_ReturnsNotFound covers the missing-record path for
// authors without indexing into an empty slice.
func TestUpdateAuthor_NoRows_ReturnsNotFound(t *testing.T) {
	client, mock := newMockClient(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "authors" SET "name"=\$1,"updated_at"=\$2 WHERE "authors"."author_id" = \$3`).
		WithArgs("new name", sqlmock.AnyArg(), "missing-id").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	author, err := client.UpdateAuthor(context.Background(), &models.Author{
		AuthorID: "missing-id",
		Name:     "new name",
	})
	if author != nil {
		t.Errorf("expected nil author, got %+v", author)
	}
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
