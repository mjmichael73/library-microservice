package server

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/loanservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/loanservice/internal/models"
)

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
	_, err := s.DB.GetActiveBorrowByUserIdAndBookId(ctx.Request().Context(), createBorrowRequest.BookID, createBorrowRequest.UserID)
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
	if resp.StatusCode != 200 {
		return ctx.JSON(http.StatusNotFound, echo.Map{
			"status":  "Failed",
			"message": "Book is not available to borrow",
		})
	}
	if resp.Body != nil {
		defer resp.Body.Close()
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
