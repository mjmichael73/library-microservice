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

// TestCreateAuthor_Success_ReportsSuccess covers R-004: a successful author
// creation must report the "Success" status string, not "Failed".
func TestCreateAuthor_Success_ReportsSuccess(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{
		authorByName:    nil,
		authorByNameErr: &dberrors.NotFoundError{Entity: "author", ID: "some-name"},
		createdAuthor: &models.Author{
			AuthorID: "author-id",
			Name:     "Some Name",
		},
	}}

	body := strings.NewReader(`{"name":"Some Name"}`)
	req := httptest.NewRequest(http.MethodPost, "/admin/authors", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	ctx := e.NewContext(req, rec)

	if err := server.CreateAuthor(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if got := resp["status"]; got != "Success" {
		t.Errorf("expected status %q, got %q", "Success", got)
	}
	if got := resp["message"]; got != "Author created successfully." {
		t.Errorf("unexpected message %q", got)
	}
}
