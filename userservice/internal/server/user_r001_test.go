package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/userservice/internal/database"
	"github.com/mjmichael73/library-microservice/userservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/userservice/internal/models"
)

type mockUserDB struct {
	registerFn   func(context.Context, *models.User) (*models.User, error)
	getByEmailFn func(context.Context, string) (*models.User, error)
}

func (m *mockUserDB) Ready() bool { return true }

func (m *mockUserDB) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, user)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserDB) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserDB) GetUserById(ctx context.Context, id string) (*models.User, error) {
	return nil, errors.New("not implemented")
}

func newTestEchoServer(db database.DatabaseClient) *EchoServer {
	s := &EchoServer{
		echo:          echo.New(),
		DB:            db,
		KafkaProducer: NewKafkaProducer("127.0.0.1:1"),
		JWTSecret:     []byte("test-secret"),
	}
	s.echo.Validator = &CustomValidator{validator: validator.New()}
	return s
}

func TestRegisterUser_ConflictReturns409(t *testing.T) {
	db := &mockUserDB{
		registerFn: func(ctx context.Context, user *models.User) (*models.User, error) {
			return nil, &dberrors.ConflictError{}
		},
	}
	s := newTestEchoServer(db)

	body := `{"first_name":"A","last_name":"B","email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.RegisterUser(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}

func TestRegisterUser_InternalErrorReturns500(t *testing.T) {
	db := &mockUserDB{
		registerFn: func(ctx context.Context, user *models.User) (*models.User, error) {
			return nil, errors.New("db down")
		},
	}
	s := newTestEchoServer(db)

	body := `{"first_name":"A","last_name":"B","email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.RegisterUser(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestRegisterUser_SuccessReturns201(t *testing.T) {
	db := &mockUserDB{
		registerFn: func(ctx context.Context, user *models.User) (*models.User, error) {
			return &models.User{
				UserID:    user.UserID,
				FirstName: "A",
				LastName:  "B",
				Email:     "a@b.com",
			}, nil
		},
	}
	s := newTestEchoServer(db)

	body := `{"first_name":"A","last_name":"B","email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.RegisterUser(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp["status"] != "Success" {
		t.Fatalf("status field = %v, want Success", resp["status"])
	}
}

func TestLoginUser_UnexpectedDBErrorReturns500(t *testing.T) {
	db := &mockUserDB{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("db down")
		},
	}
	s := newTestEchoServer(db)

	body := `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.LoginUser(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestLoginUser_NotFoundReturns401(t *testing.T) {
	db := &mockUserDB{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return nil, &dberrors.NotFoundError{}
		},
	}
	s := newTestEchoServer(db)

	body := `{"email":"a@b.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := s.echo.NewContext(req, rec)

	if err := s.LoginUser(c); err != nil {
		t.Fatalf("handler returned error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
