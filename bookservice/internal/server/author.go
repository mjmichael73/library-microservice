package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

func (s *EchoServer) GetAllAuthors(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	authors, err := s.DB.GetAllAuthors(ctx.Request().Context(), title)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusOK, authors)
}

func (s *EchoServer) CreateAuthor(ctx echo.Context) error {
	createAuthorRequest := new(models.CreateAuthorRequest)
	if err := ctx.Bind(createAuthorRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Bad request",
		})
	}
	if err := ctx.Validate(createAuthorRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}
	author, err := s.DB.GetAuthorByName(ctx.Request().Context(), createAuthorRequest.Name)
	if err == nil && author != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Author with this name already exists.",
		})
	}
	switch err.(type) {
	case *dberrors.NotFoundError:
		// Create Author
		newAuthor := &models.Author{
			AuthorID:  uuid.NewString(),
			Name:      createAuthorRequest.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		newAuthor, err := s.DB.CreateAuthor(ctx.Request().Context(), newAuthor)
		if err != nil {
			switch err.(type) {
			case *dberrors.ConflictError:
				return ctx.JSON(http.StatusBadRequest, echo.Map{
					"status":  "Failed",
					"message": "Author with this name already exists.",
				})
			default:
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		}
		return ctx.JSON(http.StatusCreated, echo.Map{
			"status":  "Failed",
			"message": "Author created successfully.",
			"data":    newAuthor,
		})
	default:
		return ctx.JSON(http.StatusInternalServerError, err)
	}
}
