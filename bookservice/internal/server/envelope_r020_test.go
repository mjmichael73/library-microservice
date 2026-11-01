package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// decodeEnvelope decodes a response body into its status/message/data fields.
func decodeEnvelope(t *testing.T, rec *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return body
}

// TestGetAllAuthors_Success_UsesEnvelope covers R-020: the list endpoint must
// return the shared {status, message, data} envelope rather than a raw slice.
func TestGetAllAuthors_Success_UsesEnvelope(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{
		authors: []models.Author{{AuthorID: "author-id", Name: "Some Name"}},
	}}

	req := httptest.NewRequest(http.MethodGet, "/admin/authors", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := server.GetAllAuthors(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := decodeEnvelope(t, rec)
	if got := body["status"]; got != "Success" {
		t.Errorf("expected status %q, got %q", "Success", got)
	}
	if got := body["message"]; got != "Authors received successfully." {
		t.Errorf("unexpected message %q", got)
	}
	if _, ok := body["data"].([]interface{}); !ok {
		t.Errorf("expected data to be a slice, got %T", body["data"])
	}
}

// TestGetAllGenres_Success_UsesEnvelope covers R-020: the list endpoint must
// return the shared envelope rather than a raw slice.
func TestGetAllGenres_Success_UsesEnvelope(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{
		genres: []models.Genre{{GenreID: "genre-id", Title: "Some Genre"}},
	}}

	req := httptest.NewRequest(http.MethodGet, "/admin/genres", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := server.GetAllGenres(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	body := decodeEnvelope(t, rec)
	if got := body["status"]; got != "Success" {
		t.Errorf("expected status %q, got %q", "Success", got)
	}
	if got := body["message"]; got != "Genres received successfully." {
		t.Errorf("unexpected message %q", got)
	}
	if _, ok := body["data"].([]interface{}); !ok {
		t.Errorf("expected data to be a slice, got %T", body["data"])
	}
}

// TestCreateGenre_Success_ReturnsEnvelope201 covers R-020: genre creation must
// return 201 with the shared envelope, mirroring CreateAuthor.
func TestCreateGenre_Success_ReturnsEnvelope201(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{
		genreByTitle:  nil,
		genreTitleErr: &dberrors.NotFoundError{Entity: "genre", ID: "Some Genre"},
		createdGenre: &models.Genre{
			GenreID: "genre-id",
			Title:   "Some Genre",
		},
	}}

	body := strings.NewReader(`{"title":"Some Genre","description":"A genre"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/genres", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	ctx := e.NewContext(req, rec)

	if err := server.CreateGenre(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	resp := decodeEnvelope(t, rec)
	if got := resp["status"]; got != "Success" {
		t.Errorf("expected status %q, got %q", "Success", got)
	}
	if got := resp["message"]; got != "Genre created successfully." {
		t.Errorf("unexpected message %q", got)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if data["genre_id"] != "genre-id" {
		t.Errorf("expected genre_id genre-id, got %v", data["genre_id"])
	}
}
