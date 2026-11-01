package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

const userServiceClientTimeout = 5 * time.Second

var (
	// errUserServiceUnavailable signals that the userservice location is not
	// configured. This is an operator error and maps to HTTP 500.
	errUserServiceUnavailable = errors.New("userservice is not configured")
	// errInvalidToken signals that the caller's token could not be validated.
	// It maps to HTTP 401.
	errInvalidToken = errors.New("invalid token")
)

// validateTokenWithUserService validates the caller's bearer token against
// userservice and returns the authenticated user id. It returns
// errUserServiceUnavailable when the service location env vars are missing, and
// errInvalidToken when the request cannot be completed or the token is rejected.
func validateTokenWithUserService(ctx context.Context, authHeader string) (string, error) {
	userServiceAppHost := os.Getenv("USERSERVICE_APP_HOST")
	if userServiceAppHost == "" {
		return "", errUserServiceUnavailable
	}
	userServiceAppPort := os.Getenv("USERSERVICE_APP_PORT")
	if userServiceAppPort == "" {
		return "", errUserServiceUnavailable
	}
	userServiceURL := "http://" + userServiceAppHost + ":" + userServiceAppPort + "/user/validate-token"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userServiceURL, nil)
	if err != nil {
		return "", errInvalidToken
	}
	req.Header.Set("Authorization", authHeader)

	client := &http.Client{Timeout: userServiceClientTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return "", errInvalidToken
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errInvalidToken
	}

	var payload struct {
		Data string `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", errInvalidToken
	}
	if payload.Data == "" {
		return "", errInvalidToken
	}
	return payload.Data, nil
}

// GetActiveBorrowsByBookId reports whether a book currently has any active
// borrow. It returns 200 with the borrows when at least one is active, and 404
// when none are, so callers can treat a missing/available book uniformly.
func (s *EchoServer) GetActiveBorrowsByBookId(ctx echo.Context) error {
	bookId := ctx.Param("book_id")
	borrows, err := s.DB.GetActiveBorrowsByBookId(ctx.Request().Context(), bookId)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	if len(borrows) == 0 {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"status":  "Failed",
			"message": "Book is not borrowed.",
		})
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Book is currently borrowed.",
		"data":    borrows,
	})
}

func (s *EchoServer) CreateBorrow(ctx echo.Context) error {
	createBorrowRequest := new(models.CreateBorrowRequest)
	if err := ctx.Bind(createBorrowRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}
	if err := ctx.Validate(createBorrowRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}
	if createBorrowRequest.ToDate.Before(time.Now().AddDate(0, 0, 1)) {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "to_date must be greather than today",
		})
	}
	authHeader := ctx.Request().Header.Get("Authorization")
	if authHeader == "" {
		return ctx.JSON(http.StatusUnauthorized, echo.Map{
			"status":  "Failed",
			"message": "Unauthorized access",
		})
	}
	authenticatedUserID, err := validateTokenWithUserService(ctx.Request().Context(), authHeader)
	if err != nil {
		if errors.Is(err, errUserServiceUnavailable) {
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
		return ctx.JSON(http.StatusUnauthorized, echo.Map{
			"status":  "Failed",
			"message": "Unauthorized access",
		})
	}
	if authenticatedUserID != createBorrowRequest.UserID {
		return ctx.JSON(http.StatusForbidden, echo.Map{
			"status":  "Failed",
			"message": "You can only borrow books for your own account.",
		})
	}
	_, err = s.DB.GetActiveBorrowByUserIdAndBookId(ctx.Request().Context(), createBorrowRequest.BookID, createBorrowRequest.UserID)
	if err == nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "You have an active borrow of this book.",
		})
	}
	bookId := createBorrowRequest.BookID
	bookServiceAppHost := os.Getenv("BOOKSERVICE_APP_HOST")
	if bookServiceAppHost == "" {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	bookServiceAppPort := os.Getenv("BOOKSERVICE_APP_PORT")
	if bookServiceAppPort == "" {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	bookServiceUrl := "http://" + bookServiceAppHost + ":" + bookServiceAppPort
	req, err := http.NewRequest("GET", bookServiceUrl+"/isbookavailabletoboroow/"+bookId, nil)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Failed to borrow the book",
			"data":    err.Error(),
		})
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Failed to borrow the book",
			"data":    err.Error(),
		})
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"status":  "Failed",
			"message": "Book is not available to borrow",
		})
	}
	newBorrow := &models.Borrow{
		BorrowID:  uuid.NewString(),
		UserID:    createBorrowRequest.UserID,
		BookID:    createBorrowRequest.BookID,
		FromDate:  createBorrowRequest.FromDate,
		ToDate:    createBorrowRequest.ToDate,
		Status:    "active",
		Remarks:   createBorrowRequest.Remarks,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, newErr := s.DB.CreateBorrow(ctx.Request().Context(), newBorrow)
	if newErr != nil {
		fmt.Println("G 7")
		switch newErr.(type) {
		case *dberrors.ConflictError:
			return ctx.JSON(http.StatusBadRequest, echo.Map{
				"status":  "Failed",
				"message": "Bad Reqouest",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}
	// TODO: Publish a notification to Notification Service
	return ctx.JSON(http.StatusCreated, echo.Map{
		"status":  "Success",
		"message": "Borrow was successful",
	})
}
