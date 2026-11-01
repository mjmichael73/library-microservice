package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

// CreateReview adds a review for a book. A missing book maps to 404 so clients
// cannot attach reviews to non-existent books.
func (s *EchoServer) CreateReview(ctx echo.Context) error {
	bookID := ctx.Param("id")
	if bookID == "" {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}

	createReviewRequest := new(models.CreateReviewRequest)
	if err := ctx.Bind(createReviewRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}
	if err := ctx.Validate(createReviewRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}

	if _, err := s.DB.GetBookById(ctx.Request().Context(), bookID); err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Book not found",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}

	review := &models.BookReview{
		ReviewID:  uuid.NewString(),
		BookID:    bookID,
		UserID:    createReviewRequest.UserID,
		Rating:    createReviewRequest.Rating,
		Comment:   createReviewRequest.Comment,
		CreatedAt: time.Now(),
	}
	created, err := s.DB.CreateReview(ctx.Request().Context(), review)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	return ctx.JSON(http.StatusCreated, echo.Map{
		"status":  "Success",
		"message": "Review created successfully.",
		"data":    created,
	})
}

// GetReviews lists the reviews for a book, newest first.
func (s *EchoServer) GetReviews(ctx echo.Context) error {
	bookID := ctx.Param("id")
	if _, err := s.DB.GetBookById(ctx.Request().Context(), bookID); err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Book not found",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, echo.Map{
				"status":  "Failed",
				"message": "Internal server error, please try again later.",
			})
		}
	}

	reviews, err := s.DB.GetReviewsByBook(ctx.Request().Context(), bookID)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"status":  "Failed",
			"message": "Internal server error, please try again later.",
		})
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Reviews received successfully.",
		"data":    reviews,
	})
}
