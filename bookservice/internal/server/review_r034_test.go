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

// TestGetAllGenres_Pagination covers R-034: `limit`/`offset` slice the list and
// the standard envelope is preserved.
func TestGetAllGenres_Pagination(t *testing.T) {
	all := []models.Genre{
		{GenreID: "g1", Title: "G1"},
		{GenreID: "g2", Title: "G2"},
		{GenreID: "g3", Title: "G3"},
	}
	server := &EchoServer{DB: &mockDatabaseClient{genres: all}}

	req := httptest.NewRequest(http.MethodGet, "/admin/genres?limit=2&offset=1", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := server.GetAllGenres(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body struct {
		Data []models.Genre `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body.Data) != 2 || body.Data[0].GenreID != "g2" || body.Data[1].GenreID != "g3" {
		t.Fatalf("expected [g2 g3], got %+v", body.Data)
	}
}

// TestGetAllAuthors_OffsetPastEnd covers R-034: an offset past the end returns
// an empty, non-nil array rather than null.
func TestGetAllAuthors_OffsetPastEnd(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{authors: []models.Author{{AuthorID: "a1", Name: "A1"}}}}

	req := httptest.NewRequest(http.MethodGet, "/admin/authors?offset=50", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)

	if err := server.GetAllAuthors(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	body := decodeEnvelope(t, rec)
	data, ok := body["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be an empty slice, got %T", body["data"])
	}
	if len(data) != 0 {
		t.Fatalf("expected empty data, got %d items", len(data))
	}
}

func TestPaginationParams(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int
		wantOffset int
	}{
		{"defaults", "", defaultLimit, 0},
		{"explicit", "?limit=5&offset=3", 5, 3},
		{"invalid limit falls back", "?limit=abc", defaultLimit, 0},
		{"negative limit falls back", "?limit=-4", defaultLimit, 0},
		{"limit capped", "?limit=9999", maxLimit, 0},
		{"negative offset falls back", "?offset=-1", defaultLimit, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin/books"+tc.query, nil)
			ctx := echo.New().NewContext(req, httptest.NewRecorder())
			limit, offset := paginationParams(ctx)
			if limit != tc.wantLimit || offset != tc.wantOffset {
				t.Fatalf("got limit=%d offset=%d, want limit=%d offset=%d", limit, offset, tc.wantLimit, tc.wantOffset)
			}
		})
	}
}

// TestCreateReview_BookNotFound_Returns404 covers R-034.
func TestCreateReview_BookNotFound_Returns404(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{err: &dberrors.NotFoundError{Entity: "book", ID: "missing-id"}}}

	req := httptest.NewRequest(http.MethodPost, "/books/missing-id/reviews", strings.NewReader(`{"user_id":"u1","rating":5}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("missing-id")

	if err := server.CreateReview(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body=%s", rec.Code, rec.Body.String())
	}
}

// TestCreateReview_Success_Returns201 covers R-034.
func TestCreateReview_Success_Returns201(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{book: &models.Book{BookID: "b1"}}}

	req := httptest.NewRequest(http.MethodPost, "/books/b1/reviews", strings.NewReader(`{"user_id":"u1","rating":4,"comment":"good"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("b1")

	if err := server.CreateReview(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body=%s", rec.Code, rec.Body.String())
	}
	body := decodeEnvelope(t, rec)
	if body["status"] != "Success" {
		t.Fatalf("expected Success, got %v", body["status"])
	}
}

// TestCreateReview_InvalidRating_Returns400 covers R-034 rating validation.
func TestCreateReview_InvalidRating_Returns400(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{book: &models.Book{BookID: "b1"}}}

	req := httptest.NewRequest(http.MethodPost, "/books/b1/reviews", strings.NewReader(`{"user_id":"u1","rating":9}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("b1")

	if err := server.CreateReview(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body=%s", rec.Code, rec.Body.String())
	}
}

// TestGetReviews_Success_UsesEnvelope covers R-034.
func TestGetReviews_Success_UsesEnvelope(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{
		book:    &models.Book{BookID: "b1"},
		reviews: []models.BookReview{{ReviewID: "r1", BookID: "b1", UserID: "u1", Rating: 5}},
	}}

	req := httptest.NewRequest(http.MethodGet, "/books/b1/reviews", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("b1")

	if err := server.GetReviews(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := decodeEnvelope(t, rec)
	if body["status"] != "Success" {
		t.Fatalf("expected Success, got %v", body["status"])
	}
	if _, ok := body["data"].([]interface{}); !ok {
		t.Fatalf("expected data to be a slice, got %T", body["data"])
	}
}

// TestGetReviews_BookNotFound_Returns404 covers R-034.
func TestGetReviews_BookNotFound_Returns404(t *testing.T) {
	server := &EchoServer{DB: &mockDatabaseClient{err: &dberrors.NotFoundError{Entity: "book", ID: "missing"}}}

	req := httptest.NewRequest(http.MethodGet, "/books/missing/reviews", nil)
	rec := httptest.NewRecorder()
	ctx := echo.New().NewContext(req, rec)
	ctx.SetParamNames("id")
	ctx.SetParamValues("missing")

	if err := server.GetReviews(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
