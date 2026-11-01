package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
)

// withUser attaches a JWT token carrying the given user id to the context, as
// the echojwt middleware would.
func withUser(ctx echo.Context, userID string) echo.Context {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID})
	ctx.Set("user", token)
	return ctx
}

func TestGetProfile_Success(t *testing.T) {
	db := &mockUserDB{
		getByIdFn: func(ctx context.Context, id string) (*models.User, error) {
			return &models.User{UserID: id, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"}, nil
		},
	}
	s := newTestEchoServer(db)

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	rec := httptest.NewRecorder()
	ctx := withUser(s.echo.NewContext(req, rec), "user-1")

	if err := s.GetProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object, got %T", resp["data"])
	}
	if _, leaked := data["password"]; leaked {
		t.Fatal("profile response must not include the password field")
	}
	if data["first_name"] != "Ada" {
		t.Fatalf("expected first_name Ada, got %v", data["first_name"])
	}
}

func TestGetProfile_NotFound_Returns404(t *testing.T) {
	db := &mockUserDB{
		getByIdFn: func(ctx context.Context, id string) (*models.User, error) {
			return nil, &dberrors.NotFoundError{Entity: "user", ID: id}
		},
	}
	s := newTestEchoServer(db)

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	rec := httptest.NewRecorder()
	ctx := withUser(s.echo.NewContext(req, rec), "missing")

	if err := s.GetProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetProfile_MissingTokenClaim_Returns401(t *testing.T) {
	s := newTestEchoServer(&mockUserDB{})

	req := httptest.NewRequest(http.MethodGet, "/user/profile", nil)
	rec := httptest.NewRecorder()
	// No "user" token set on the context -> handler must not panic.
	ctx := s.echo.NewContext(req, rec)

	if err := s.GetProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	var got *models.User
	db := &mockUserDB{
		updateFn: func(ctx context.Context, user *models.User) (*models.User, error) {
			got = user
			return &models.User{UserID: user.UserID, FirstName: user.FirstName, LastName: user.LastName, Email: "ada@example.com"}, nil
		},
	}
	s := newTestEchoServer(db)

	body := `{"first_name":"Grace","last_name":"Hopper"}`
	req := httptest.NewRequest(http.MethodPut, "/user/profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := withUser(s.echo.NewContext(req, rec), "user-1")

	if err := s.UpdateProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got == nil || got.FirstName != "Grace" || got.LastName != "Hopper" {
		t.Fatalf("DB received unexpected user: %+v", got)
	}
	if got.UserID != "user-1" {
		t.Fatalf("expected user id from JWT, got %q", got.UserID)
	}
}

func TestUpdateProfile_ValidationError_Returns400(t *testing.T) {
	s := newTestEchoServer(&mockUserDB{})

	body := `{"first_name":"","last_name":"Hopper"}`
	req := httptest.NewRequest(http.MethodPut, "/user/profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := withUser(s.echo.NewContext(req, rec), "user-1")

	if err := s.UpdateProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateProfile_NotFound_Returns404(t *testing.T) {
	db := &mockUserDB{
		updateFn: func(ctx context.Context, user *models.User) (*models.User, error) {
			return nil, &dberrors.NotFoundError{Entity: "user", ID: user.UserID}
		},
	}
	s := newTestEchoServer(db)

	body := `{"first_name":"Grace","last_name":"Hopper"}`
	req := httptest.NewRequest(http.MethodPut, "/user/profile", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := withUser(s.echo.NewContext(req, rec), "missing")

	if err := s.UpdateProfile(ctx); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
