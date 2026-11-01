package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// mockDatabaseClient implements database.DatabaseClient. Only the Get*ById
// methods used by the handlers under test return a configured error; the
// remaining methods exist to satisfy the interface.
type mockDatabaseClient struct {
	err error

	// book is returned by GetBookById when err is nil (R-019).
	book *models.Book

	// Author creation support (R-004). Zero values preserve the previous
	// behaviour for the Get*ById tests.
	authorByName    *models.Author
	authorByNameErr error
	createdAuthor   *models.Author
	createAuthorErr error
}

func (m *mockDatabaseClient) Ready() bool { return true }

func (m *mockDatabaseClient) GetAllGenres(ctx context.Context, title string) ([]models.Genre, error) {
	return nil, nil
}
func (m *mockDatabaseClient) GetGenreByTitle(ctx context.Context, title string) (*models.Genre, error) {
	return nil, nil
}
func (m *mockDatabaseClient) CreateGenre(ctx context.Context, genre *models.Genre) (*models.Genre, error) {
	return nil, nil
}
func (m *mockDatabaseClient) GetGenreById(ctx context.Context, ID string) (*models.Genre, error) {
	return nil, m.err
}

func (m *mockDatabaseClient) GetAllAuthors(ctx context.Context, name string) ([]models.Author, error) {
	return nil, nil
}
func (m *mockDatabaseClient) GetAuthorByName(ctx context.Context, name string) (*models.Author, error) {
	return m.authorByName, m.authorByNameErr
}
func (m *mockDatabaseClient) CreateAuthor(ctx context.Context, author *models.Author) (*models.Author, error) {
	return m.createdAuthor, m.createAuthorErr
}
func (m *mockDatabaseClient) GetAuthorById(ctx context.Context, ID string) (*models.Author, error) {
	return nil, m.err
}
func (m *mockDatabaseClient) UpdateAuthor(ctx context.Context, author *models.Author) (*models.Author, error) {
	return nil, nil
}
func (m *mockDatabaseClient) DeleteAuthor(ctx context.Context, ID string) error { return nil }

func (m *mockDatabaseClient) GetAllBooks(ctx context.Context, title string) ([]models.Book, error) {
	return nil, nil
}
func (m *mockDatabaseClient) GetBookByTitle(ctx context.Context, title string) (*models.Book, error) {
	return nil, nil
}
func (m *mockDatabaseClient) CreateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	return nil, nil
}
func (m *mockDatabaseClient) GetBookById(ctx context.Context, ID string) (*models.Book, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.book != nil {
		return m.book, nil
	}
	return nil, nil
}
func (m *mockDatabaseClient) UpdateBook(ctx context.Context, book *models.Book) (*models.Book, error) {
	return nil, nil
}
func (m *mockDatabaseClient) DeleteBook(ctx context.Context, ID string) error { return nil }

func TestGetByIdHandlers_MissingRecord_Returns404(t *testing.T) {
	tests := []struct {
		name         string
		handler      func(*EchoServer) echo.HandlerFunc
		messageMatch string
	}{
		{
			name:         "book",
			handler:      func(s *EchoServer) echo.HandlerFunc { return s.GetBookById },
			messageMatch: "Book not found",
		},
		{
			name:         "author",
			handler:      func(s *EchoServer) echo.HandlerFunc { return s.GetAuthorById },
			messageMatch: "Author not found",
		},
		{
			name:         "genre",
			handler:      func(s *EchoServer) echo.HandlerFunc { return s.GetGenreById },
			messageMatch: "Genre not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			notFound := &dberrors.NotFoundError{Entity: tc.name, ID: "missing-id"}
			server := &EchoServer{DB: &mockDatabaseClient{err: notFound}}

			req := httptest.NewRequest(http.MethodGet, "/admin/"+tc.name+"s/missing-id", nil)
			rec := httptest.NewRecorder()
			ctx := echo.New().NewContext(req, rec)
			ctx.SetParamNames("id")
			ctx.SetParamValues("missing-id")

			if err := tc.handler(server)(ctx); err != nil {
				t.Fatalf("handler returned error: %v", err)
			}

			if rec.Code != http.StatusNotFound {
				t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
			}

			var body map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}
			if got := body["message"]; got != tc.messageMatch {
				t.Errorf("expected message %q, got %q", tc.messageMatch, got)
			}
			if got := body["status"]; got != "Failed" {
				t.Errorf("expected status %q, got %q", "Failed", got)
			}
		})
	}
}
