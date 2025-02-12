package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/mjmichael73/library-microservice/bookservice/internal/dberrors"
	"github.com/mjmichael73/library-microservice/bookservice/internal/models"
)

func (s *EchoServer) GetAllGenres(ctx echo.Context) error {
	title := ctx.QueryParam("title")
	genres, err := s.DB.GetAllGenres(ctx.Request().Context(), title)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	}
	return ctx.JSON(http.StatusOK, genres)
}

func (s *EchoServer) CreateGenre(ctx echo.Context) error {
	createGenreRequest := new(models.CreateGenreRequest)
	if err := ctx.Bind(createGenreRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Invalid request",
		})
	}
	if err := ctx.Validate(createGenreRequest); err != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Please fix validation errors",
			"errors":  FormatValidationErrors(err),
		})
	}
	genre, err := s.DB.GetGenreByTitle(ctx.Request().Context(), createGenreRequest.Title)
	if err == nil && genre != nil {
		return ctx.JSON(http.StatusBadRequest, echo.Map{
			"status":  "Failed",
			"message": "Genre already exist",
		})
	}
	switch err.(type) {
	case *dberrors.NotFoundError:
		newGenre := &models.Genre{
			Title:       createGenreRequest.Title,
			Description: createGenreRequest.Description,
			GenreID:     uuid.NewString(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		newGenre, err := s.DB.CreateGenre(ctx.Request().Context(), newGenre)
		if err != nil {
			switch err.(type) {
			case *dberrors.ConflictError:
				return ctx.JSON(http.StatusConflict, err)
			default:
				return ctx.JSON(http.StatusInternalServerError, err)
			}
		}
		return ctx.JSON(http.StatusOK, newGenre)
	default:
		return ctx.JSON(http.StatusInternalServerError, err)

	}
}

func (s *EchoServer) GetGenreById(ctx echo.Context) error {
	ID := ctx.Param("id")
	genre, err := s.DB.GetGenreById(ctx.Request().Context(), ID)
	if err != nil {
		switch err.(type) {
		case *dberrors.NotFoundError:
			return ctx.JSON(http.StatusNotFound, echo.Map{
				"status":  "Failed",
				"message": "Genre not found",
			})
		default:
			return ctx.JSON(http.StatusInternalServerError, err)
		}
	}
	return ctx.JSON(http.StatusOK, echo.Map{
		"status":  "Success",
		"message": "Genre received successfully.",
		"data":    genre,
	})
}
